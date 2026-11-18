package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeRegistry struct {
	mu      sync.Mutex
	records map[string]workspaceRecord
	calls   atomic.Int64
	err     error
}

func (f *fakeRegistry) Lookup(_ context.Context, id string) (workspaceRecord, bool, error) {
	f.calls.Add(1)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return workspaceRecord{}, false, f.err
	}
	rec, ok := f.records[id]
	return rec, ok, nil
}

func (f *fakeRegistry) set(id, status string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.records[id] = workspaceRecord{ID: id, Name: id, Status: status, Limits: defaultWorkspaceLimits}
}

func identityRuntime(id string) (*workspaceRuntime, error) {
	return &workspaceRuntime{id: id, s: &server{}, stop: func() {}, handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("workspace:" + id))
	})}, nil
}

func hostGet(h http.Handler, host string) (int, string) {
	r := httptest.NewRequest("GET", "http://"+host+"/", nil)
	r.Host = host
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code, strings.TrimSpace(w.Body.String())
}

func TestSingleModeServesDefaultWorkspaceOnceForEveryHost(t *testing.T) {
	t.Parallel()
	var builds atomic.Int64
	h := newWorkspaceHost(tenancySettings{Mode: tenancySingle}, &fakeRegistry{}, func(id string) (*workspaceRuntime, error) {
		builds.Add(1)
		return identityRuntime(id)
	})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if code, body := hostGet(h, "127.0.0.1:8787"); code != 200 || body != "workspace:default" {
				t.Error(code, body)
			}
		}()
	}
	wg.Wait()
	if code, body := hostGet(h, "agentcommonroom.home.arpa"); code != 200 || body != "workspace:default" || builds.Load() != 1 {
		t.Fatal(code, body, builds.Load())
	}
}

func TestMultiModeRoutesByHostAndRefusesWithoutRevealing(t *testing.T) {
	t.Parallel()
	reg := &fakeRegistry{records: map[string]workspaceRecord{}}
	reg.set("alpha", workspaceActive)
	reg.set("beta", workspaceActive)
	reg.set("held", workspaceSuspended)
	reg.set("gone", workspaceDeleted)
	h := newWorkspaceHost(tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}, reg, identityRuntime)
	for host, want := range map[string]string{"alpha.rooms.example.com": "workspace:alpha", "beta.rooms.example.com": "workspace:beta"} {
		if code, body := hostGet(h, host); code != 200 || body != want {
			t.Fatal(host, code, body)
		}
	}
	for host, want := range map[string]int{
		"held.rooms.example.com":    503,
		"gone.rooms.example.com":    404,
		"nobody.rooms.example.com":  404,
		"www.rooms.example.com":     404,
		"alpha.other.example.com":   404,
		"rooms.example.com":         404,
		"a.b.rooms.example.com":     404,
		"default.rooms.example.com": 404,
	} {
		code, body := hostGet(h, host)
		if code != want {
			t.Fatal(host, code, body)
		}
		if strings.Contains(body, "alpha") || strings.Contains(body, "workspace") && code == 404 {
			t.Fatal("refusal revealed workspace information:", body)
		}
	}
	// A registry outage is a generic 503, never an answer about a workspace.
	reg.err = errors.New("postgres://user:secret@host/db")
	h2 := newWorkspaceHost(tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}, reg, identityRuntime)
	if code, body := hostGet(h2, "alpha.rooms.example.com"); code != 503 || strings.Contains(body, "secret") {
		t.Fatal(code, body)
	}
}

func TestSuspensionTakesEffectAfterTheRegistryRecheck(t *testing.T) {
	t.Parallel()
	reg := &fakeRegistry{records: map[string]workspaceRecord{}}
	reg.set("alpha", workspaceActive)
	h := newWorkspaceHost(tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}, reg, identityRuntime)
	now := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	h.now = func() time.Time { return now }
	if code, _ := hostGet(h, "alpha.rooms.example.com"); code != 200 {
		t.Fatal(code)
	}
	reg.set("alpha", workspaceSuspended)
	if code, _ := hostGet(h, "alpha.rooms.example.com"); code != 200 || reg.calls.Load() != 1 {
		t.Fatal("registry answers are trusted briefly", code, reg.calls.Load())
	}
	now = now.Add(registryRecheck + time.Second)
	if code, _ := hostGet(h, "alpha.rooms.example.com"); code != 503 {
		t.Fatal("suspension did not take effect", code)
	}
	reg.set("alpha", workspaceActive)
	now = now.Add(registryRecheck + time.Second)
	if code, _ := hostGet(h, "alpha.rooms.example.com"); code != 200 {
		t.Fatal("resume did not take effect", code)
	}
	reg.set("alpha", workspaceDeleted)
	now = now.Add(registryRecheck + time.Second)
	if code, _ := hostGet(h, "alpha.rooms.example.com"); code != 404 {
		t.Fatal("deletion did not take effect", code)
	}
}

func TestUnknownWorkspacesAreRememberedAndInvalidHostsNeverReachTheRegistry(t *testing.T) {
	t.Parallel()
	reg := &fakeRegistry{records: map[string]workspaceRecord{}}
	h := newWorkspaceHost(tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}, reg, identityRuntime)
	now := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	h.now = func() time.Time { return now }
	for i := 0; i < 50; i++ {
		hostGet(h, "ghost.rooms.example.com")
	}
	if reg.calls.Load() != 1 {
		t.Fatal("probing an unknown workspace hit the registry repeatedly", reg.calls.Load())
	}
	now = now.Add(missingRemember + time.Second)
	hostGet(h, "ghost.rooms.example.com")
	if reg.calls.Load() != 2 {
		t.Fatal("unknown workspace was never rechecked", reg.calls.Load())
	}
	before := reg.calls.Load()
	for _, host := range []string{"www.rooms.example.com", "a.b.rooms.example.com", "evil.example.org", "default.rooms.example.com"} {
		hostGet(h, host)
	}
	if reg.calls.Load() != before {
		t.Fatal("an invalid host reached the registry")
	}
	for i := 0; i < maxMissingMemory+10; i++ {
		hostGet(h, "w"+strings.Repeat("x", i%5)+string(rune('a'+i%26))+strings.Repeat("y", i/26%7)+".rooms.example.com")
	}
	h.mu.Lock()
	size := len(h.missing)
	h.mu.Unlock()
	if size > maxMissingMemory {
		t.Fatal("unknown-workspace memory is unbounded", size)
	}
}

func TestWorkspaceBuiltOnceUnderConcurrencyAndFailedBuildRetries(t *testing.T) {
	t.Parallel()
	reg := &fakeRegistry{records: map[string]workspaceRecord{}}
	reg.set("alpha", workspaceActive)
	reg.set("flaky", workspaceActive)
	var builds, flakyBuilds atomic.Int64
	h := newWorkspaceHost(tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}, reg, func(id string) (*workspaceRuntime, error) {
		if id == "flaky" {
			if flakyBuilds.Add(1) == 1 {
				return nil, errors.New("first start fails: /home/claude/.config secret")
			}
			return identityRuntime(id)
		}
		builds.Add(1)
		time.Sleep(10 * time.Millisecond)
		return identityRuntime(id)
	})
	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if code, _ := hostGet(h, "alpha.rooms.example.com"); code != 200 {
				t.Error(code)
			}
		}()
	}
	wg.Wait()
	if builds.Load() != 1 {
		t.Fatal("workspace server built more than once", builds.Load())
	}
	if code, body := hostGet(h, "flaky.rooms.example.com"); code != 503 || strings.Contains(body, "secret") {
		t.Fatal("start failure must be a generic 503", code, body)
	}
	if code, _ := hostGet(h, "flaky.rooms.example.com"); code != 200 {
		t.Fatal("a failed start was cached forever", code)
	}
}

func testPlatform(t *testing.T) platformSettings {
	t.Helper()
	root := t.TempDir()
	return platformSettings{dailyTurnLimit: 100, dataRoot: root, listenAddr: "127.0.0.1:8787", defaultKeyPath: filepath.Join(root, "owner-key")}
}

func TestHostedWorkspacesAreFullyIndependentServers(t *testing.T) {
	t.Parallel()
	ps := testPlatform(t)
	tenancy := tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}
	dbs := map[string]*mockDB{"alpha": newMockDB(t), "beta": newMockDB(t)}
	reg := &fakeRegistry{records: map[string]workspaceRecord{}}
	reg.set("alpha", workspaceActive)
	reg.set("beta", workspaceActive)
	host := newWorkspaceHost(tenancy, reg, func(id string) (*workspaceRuntime, error) {
		return newWorkspaceRuntime(ps, tenancy, dbs[id], id)
	})
	defer host.stopAll()

	do := func(method, hostName, path, body, cookie string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "https://"+hostName+path, strings.NewReader(body))
		r.Host = hostName
		r.Header.Set("Origin", "https://"+hostName)
		r.Header.Set("Content-Type", "application/json")
		if cookie != "" {
			r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: cookie})
		}
		w := httptest.NewRecorder()
		host.ServeHTTP(w, r)
		return w
	}
	alpha, beta := "alpha.rooms.example.com", "beta.rooms.example.com"
	for _, h := range []string{alpha, beta} {
		if w := do("GET", h, "/api/setup-status", "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"setupRequired":true`) {
			t.Fatal(h, w.Code, w.Body.String())
		}
	}
	// Each workspace has its own setup code, key files and data folder.
	code := func(id string) string {
		raw, err := os.ReadFile(filepath.Join(ps.dataRoot, "workspaces", id, "owner-key"))
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(raw))
	}
	if code("alpha") == code("beta") || len(code("alpha")) != 64 {
		t.Fatal("workspaces share a setup code")
	}
	for _, id := range []string{"alpha", "beta"} {
		if _, err := os.Stat(filepath.Join(ps.dataRoot, "workspaces", id, "owner-key.public.pem")); err != nil {
			t.Fatal("owner identity missing for", id)
		}
	}
	if len(dbs["alpha"].snapshot()) != 1 || len(dbs["beta"].snapshot()) != 1 {
		t.Fatal("first-owner registration must write to each workspace's own history")
	}
	// A workspace's setup code opens that workspace only.
	if w := do("POST", beta, "/api/login", `{"key":"`+code("alpha")+`"}`, ""); w.Code != 401 {
		t.Fatal("alpha's setup code was accepted by beta", w.Code)
	}
	login := do("POST", alpha, "/api/login", `{"key":"`+code("alpha")+`"}`, "")
	if login.Code != 200 {
		t.Fatal(login.Code, login.Body.String())
	}
	var token string
	for _, c := range login.Result().Cookies() {
		if c.Name == "commonroom_session" {
			token = c.Value
			if c.Domain != "" || !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
				t.Fatal("session cookie must be host-only, secure, HttpOnly and strict", c)
			}
		}
	}
	if token == "" {
		t.Fatal("no session issued")
	}
	if w := do("GET", alpha, "/api/session", "", token); w.Code != 200 {
		t.Fatal("session rejected on its own workspace", w.Code)
	}
	if w := do("GET", beta, "/api/session", "", token); w.Code == 200 {
		t.Fatal("alpha's session was accepted on beta's host")
	}
	// A request naming another workspace's origin is refused even on the right host.
	r := httptest.NewRequest("POST", "https://"+alpha+"/api/login", strings.NewReader(`{"key":"`+code("alpha")+`"}`))
	r.Host = alpha
	r.Header.Set("Origin", "https://"+beta)
	w := httptest.NewRecorder()
	host.ServeHTTP(w, r)
	if w.Code == 200 {
		t.Fatal("a cross-workspace Origin was accepted")
	}
	// Hosted installations never expose worker-touching routes.
	if w := do("POST", alpha, "/api/provider-auth", `{}`, token); w.Code != 503 || !strings.Contains(w.Body.String(), hostedExecutionMessage) {
		t.Fatal("provider sign-in must be unavailable to hosted workspaces", w.Code, w.Body.String())
	}
}

func TestSingleWorkspaceRuntimeKeepsConfiguredKeyLocation(t *testing.T) {
	t.Parallel()
	ps := testPlatform(t)
	tenancy := tenancySettings{Mode: tenancySingle}
	rt, err := newWorkspaceRuntime(ps, tenancy, newMockDB(t), defaultWorkspaceID)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.stop()
	if _, err := os.Stat(ps.defaultKeyPath); err != nil {
		t.Fatal("the default workspace must keep the configured key file", err)
	}
	if _, err := os.Stat(filepath.Join(ps.dataRoot, "workspaces")); err == nil {
		t.Fatal("single mode must not create workspace folders")
	}
	if rt.s.hostedExecutionBlocked || rt.s.workspaceID != defaultWorkspaceID {
		t.Fatal("single mode must not block execution", rt.s.hostedExecutionBlocked)
	}
	if _, err := newWorkspaceRuntime(ps, tenancy, newMockDB(t), "../escape"); err == nil {
		t.Fatal("invalid workspace id accepted")
	}
}

func TestHostedExecutionIsRefusedAtEveryAdmissionPoint(t *testing.T) {
	t.Parallel()
	s, _, _ := mockServer(t)
	s.hostedExecutionBlocked = true
	st, loadErr := s.load(t.Context())
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if err := s.agentAdmission(st, "engineering", 1); err == nil || err.Error() != hostedExecutionMessage {
		t.Fatal("admission allowed hosted execution", err)
	}
	for _, action := range []string{"list-conversations", "read-context", "attach"} {
		w := mockRequest(http.HandlerFunc(s.agents), "POST", "/api/agents", `{"action":"`+action+`","threadId":"x"}`)
		if w.Code != 503 || !strings.Contains(w.Body.String(), hostedExecutionMessage) {
			t.Fatal(action, w.Code, w.Body.String())
		}
	}
	if next := (http.HandlerFunc)(func(http.ResponseWriter, *http.Request) { t.Fatal("gate let a hosted request through") }); true {
		w := httptest.NewRecorder()
		hostedGate(true, next)(w, httptest.NewRequest("POST", "/api/provider-auth", nil))
		if w.Code != 503 {
			t.Fatal(w.Code)
		}
	}
	called := false
	hostedGate(false, func(http.ResponseWriter, *http.Request) { called = true })(httptest.NewRecorder(), httptest.NewRequest("POST", "/", nil))
	if !called {
		t.Fatal("the gate must not affect a single-workspace installation")
	}
}
