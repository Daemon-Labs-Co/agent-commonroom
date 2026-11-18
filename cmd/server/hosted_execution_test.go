package main

import (
	"agentroom/internal/room"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestExecutionPoolCapsAreGenericAndNeverGoNegative(t *testing.T) {
	t.Parallel()
	p := newExecutionPool(3)
	if msg := p.refusal("a", 3, 0); msg != "" {
		t.Fatal("the installation cap itself must be usable", msg)
	}
	p.add("a", 2)
	p.add("b", 1)
	if msg := p.refusal("c", 1, 0); !strings.Contains(msg, "installation") {
		t.Fatal("a full installation must refuse", msg)
	}
	if msg := p.refusal("c", 1, 0); strings.Contains(msg, "a") && strings.Contains(msg, "workspace a") || strings.ContainsAny(msg, "0123456789") {
		t.Fatal("the refusal revealed other workspaces' work:", msg)
	}
	if msg := p.refusal("a", 1, 2); !strings.Contains(msg, "limit of 2 agent turns") {
		t.Fatal("a workspace at its own cap must say so", msg)
	}
	p.add("a", -2)
	p.add("a", -5) // finishing more than started must not corrupt the count
	if p.running("a") != 0 || p.total < 0 {
		t.Fatal("counts went wrong", p.running("a"), p.total)
	}
	if msg := p.refusal("c", 2, 0); msg != "" {
		t.Fatal("capacity was not released", msg)
	}
}

func TestActiveMarkingKeepsTheSharedCountExact(t *testing.T) {
	t.Parallel()
	pool := newExecutionPool(10)
	s := &server{workspaceID: "acme", active: map[string]bool{}, pool: pool}
	s.setActive("p1")
	s.setActive("p1") // marking twice must not count twice
	s.setActive("p2")
	if pool.running("acme") != 2 || !s.active["p1"] {
		t.Fatal(pool.running("acme"))
	}
	s.clearActive("p1")
	s.clearActive("p1") // clearing twice must not release twice
	s.clearActive("never-started")
	if pool.running("acme") != 1 || s.active["p1"] {
		t.Fatal(pool.running("acme"))
	}
	single := &server{workspaceID: defaultWorkspaceID, active: map[string]bool{}} // no pool: unchanged behaviour
	single.setActive("x")
	single.clearActive("x")
	if len(single.active) != 0 {
		t.Fatal("single mode marking broke")
	}
}

func TestWorkspacesShareInstallationCapacityFairly(t *testing.T) {
	t.Parallel()
	pool := newExecutionPool(3)
	mk := func(id string, perWorkspace int) *server {
		return &server{workspaceID: id, active: map[string]bool{}, pool: pool, limits: &workspaceLimits{ConcurrentTurns: perWorkspace}}
	}
	a, b := mk("a", 2), mk("b", 2)
	st := mockStateForAdmission(t)
	if err := a.agentAdmission(st, "engineering", 2); err != nil {
		t.Fatal(err)
	}
	a.setActive("a1")
	a.setActive("a2")
	if err := a.agentAdmission(st, "engineering", 1); err == nil || !strings.Contains(err.Error(), "limit of 2 agent turns") {
		t.Fatal("a workspace must be held to its own cap", err)
	}
	if err := b.agentAdmission(st, "engineering", 1); err != nil {
		t.Fatal("another workspace's turns must not block this one while capacity remains", err)
	}
	b.setActive("b1")
	if err := b.agentAdmission(st, "engineering", 1); err == nil || !strings.Contains(err.Error(), "installation") || strings.Contains(err.Error(), "a") && strings.Contains(err.Error(), "workspace a") {
		t.Fatal("the installation cap must hold across workspaces", err)
	}
	a.clearActive("a1")
	if err := b.agentAdmission(st, "engineering", 1); err != nil {
		t.Fatal("finished turns must release capacity for everyone", err)
	}
}

func TestSharedPoolIsSafeUnderConcurrency(t *testing.T) {
	t.Parallel()
	pool := newExecutionPool(1000)
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := &server{workspaceID: "w" + string(rune('a'+w)), active: map[string]bool{}, pool: pool}
			for i := 0; i < 200; i++ {
				s.setActive("p")
				s.clearActive("p")
			}
		}()
	}
	wg.Wait()
	if pool.total != 0 || len(pool.per) != 0 {
		t.Fatal("counts leaked", pool.total, pool.per)
	}
}

func TestHostedExecutionOptInAndWhatStaysRefused(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{false, true} {
		pf := newPlatformFixture(t)
		pf.ps.hostedExecution = enabled
		pf.ps.pool = newExecutionPool(4)
		f := &idFixture{platformFixture: pf, ws: "acme"}
		pf.create("acme")
		rt := f.runtime("acme")
		st, err := rt.s.load(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		rt.s.mu.Lock()
		admission := rt.s.agentAdmission(st, "engineering", 1)
		rt.s.mu.Unlock()
		if enabled && admission != nil {
			t.Fatal("opted-in hosted execution must admit turns within limits", admission)
		}
		if !enabled && (admission == nil || !strings.Contains(admission.Error(), "not available")) {
			t.Fatal("hosted execution must stay off by default", admission)
		}
		owner := f.ownerSession("acme")
		reach := func(method, path string) int { return f.do(method, path, `{}`, &owner).Code }
		for _, path := range []string{"/api/provider-auth", "/api/job-diagnostics", "/api/agent-recovery"} {
			blocked := reach("POST", path) == http.StatusServiceUnavailable
			if blocked == enabled {
				t.Fatal("worker routes must follow the opt-in:", path, enabled)
			}
		}
		// Reading the host's provider history, deleting workspace disks and the shared extension
		// registry are refused whatever the operator chose.
		for _, body := range []string{`{"action":"list-conversations"}`, `{"action":"read-context","threadId":"x"}`, `{"action":"attach","threadId":"x","mode":"fork"}`} {
			if w := f.do("POST", "/api/agents", body, &owner); w.Code != 503 || !strings.Contains(w.Body.String(), hostedExecutionMessage) {
				t.Fatal("conversation import must stay refused", enabled, body, w.Code, w.Body.String())
			}
		}
		for _, path := range []string{"/api/workspace-disposal", "/ext/v1/rooms/engineering/messages"} {
			if code := reach("POST", path); code != 503 {
				t.Fatal("must stay refused for hosted workspaces:", path, enabled, code)
			}
		}
	}
}

func TestHostedExecutionStartupValidation(t *testing.T) {
	t.Parallel()
	env := func(kv ...string) func(string) string {
		m := map[string]string{}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return func(k string) string { return m[k] }
	}
	if err := validateHostedExecution(env("COMMONROOM_REMOTE_WORKER", "true")); err != nil {
		t.Fatal(err)
	}
	if err := validateHostedExecution(env("COMMONROOM_SERVER_WORKER", "true", "COMMONROOM_DEVELOPMENT_WORKSPACES", "false")); err != nil {
		t.Fatal(err)
	}
	for name, e := range map[string]func(string) string{
		"no worker":              env(),
		"development workspaces": env("COMMONROOM_REMOTE_WORKER", "true", "COMMONROOM_DEVELOPMENT_WORKSPACES", "true"),
		"isolated commands":      env("COMMONROOM_REMOTE_WORKER", "true", "COMMONROOM_ISOLATED_COMMANDS", "true"),
		"worker development":     env("COMMONROOM_REMOTE_WORKER", "true", "COMMONROOM_WORKER_DEVELOPMENT", "true"),
	} {
		if err := validateHostedExecution(e); err == nil {
			t.Fatal("accepted:", name)
		}
	}
	for _, bad := range []string{"yes", "1", "TRUE"} {
		if _, err := loadPlatformSettings(env("COMMONROOM_HOSTED_EXECUTION", bad), 100); err == nil {
			t.Fatal("accepted COMMONROOM_HOSTED_EXECUTION", bad)
		}
	}
	for _, bad := range []string{"0", "257", "x", "-1"} {
		if _, err := loadPlatformSettings(env("COMMONROOM_HOSTED_MAX_TURNS", bad), 100); err == nil {
			t.Fatal("accepted COMMONROOM_HOSTED_MAX_TURNS", bad)
		}
	}
	ps, err := loadPlatformSettings(env("COMMONROOM_HOSTED_EXECUTION", "true", "COMMONROOM_HOSTED_MAX_TURNS", "20"), 100)
	if err != nil || !ps.hostedExecution || ps.pool.max != 20 {
		t.Fatal(ps.hostedExecution, err)
	}
	if ps, _ = loadPlatformSettings(env(), 100); ps.hostedExecution || ps.pool.max != 8 {
		t.Fatal("hosted execution must default off with a pool of 8")
	}
}

func TestConcurrentTurnLimitIsValidatedAndLive(t *testing.T) {
	t.Parallel()
	if l, err := parseWorkspaceLimits([]byte(`{"concurrentTurns":4}`)); err != nil || l.ConcurrentTurns != 4 || l.Members != defaultWorkspaceLimits.Members {
		t.Fatal(l, err)
	}
	for _, bad := range []string{`{"concurrentTurns":0}`, `{"concurrentTurns":17}`, `{"concurrentTurns":"two"}`} {
		if _, err := parseWorkspaceLimits([]byte(bad)); err == nil {
			t.Fatal("accepted", bad)
		}
	}
	if defaultWorkspaceLimits.ConcurrentTurns != 2 {
		t.Fatal("default must stay small")
	}
}

func mockStateForAdmission(t *testing.T) room.State {
	t.Helper()
	s, _, _ := mockServer(t)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return st
}
