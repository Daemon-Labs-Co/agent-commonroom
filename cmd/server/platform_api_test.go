package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type memoryStore struct {
	mu      sync.Mutex
	records map[string]workspaceRecord
	clock   func() time.Time
}

func (m *memoryStore) Lookup(_ context.Context, id string) (workspaceRecord, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[id]
	return rec, ok, nil
}
func (m *memoryStore) Create(_ context.Context, id, name string, limits workspaceLimits) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.records[id]; exists {
		return false, nil
	}
	m.records[id] = workspaceRecord{ID: id, Name: name, Status: workspaceActive, Limits: limits, Created: m.clock()}
	return true, nil
}
func (m *memoryStore) SetStatus(_ context.Context, id, status string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[id]
	if !ok || rec.Status == workspaceDeleted || rec.Status == status {
		return false, nil
	}
	rec.Status = status
	m.records[id] = rec
	return true, nil
}
func (m *memoryStore) SetLimits(_ context.Context, id string, limits workspaceLimits) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[id]
	if !ok || rec.Status == workspaceDeleted {
		return false, nil
	}
	rec.Limits = limits
	m.records[id] = rec
	return true, nil
}
func (m *memoryStore) List(_ context.Context, after string, limit int) ([]workspaceRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := []string{}
	for id := range m.records {
		if id > after {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	out := []workspaceRecord{}
	for _, id := range ids {
		if len(out) == limit {
			break
		}
		out = append(out, m.records[id])
	}
	return out, nil
}

type platformFixture struct {
	t     *testing.T
	host  *workspaceHost
	store *memoryStore
	priv  ed25519.PrivateKey
	keyID string
	mu    sync.Mutex
	now   time.Time
	dbs   map[string]*mockDB
	ps    platformSettings
}

func newPlatformFixture(t *testing.T) *platformFixture {
	t.Helper()
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	dir := t.TempDir()
	keys := filepath.Join(dir, "platform-keys.json")
	body, _ := json.Marshal([]map[string]any{{"id": "control-plane", "name": "Control plane", "publicKey": base64.StdEncoding.EncodeToString(pub), "rooms": []string{}}})
	if err := os.WriteFile(keys, body, 0o600); err != nil {
		t.Fatal(err)
	}
	f := &platformFixture{t: t, priv: priv, keyID: "control-plane", now: time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC), dbs: map[string]*mockDB{}}
	clock := func() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.now }
	f.store = &memoryStore{records: map[string]workspaceRecord{}, clock: clock}
	f.ps = platformSettings{dailyTurnLimit: 100, dataRoot: filepath.Join(dir, "data"), listenAddr: "127.0.0.1:8787", defaultKeyPath: filepath.Join(dir, "owner-key")}
	tenancy := tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}
	f.host = newWorkspaceHost(tenancy, f.store, func(id string) (*workspaceRuntime, error) {
		f.mu.Lock()
		db, ok := f.dbs[id]
		if !ok {
			db = newMockDB(t)
			f.dbs[id] = db
		}
		f.mu.Unlock()
		return newWorkspaceRuntime(f.ps, tenancy, db, id)
	})
	f.host.now = clock
	f.host.platform = &platformAPI{tenancy: tenancy, store: f.store, host: f.host, keyFile: keys, now: clock}
	t.Cleanup(f.host.stopAll)
	return f
}

func (f *platformFixture) advance(d time.Duration) { f.mu.Lock(); f.now = f.now.Add(d); f.mu.Unlock() }

func (f *platformFixture) request(method, target, body string, mutate func(*http.Request)) *httptest.ResponseRecorder {
	// Each call is a distinct moment, as a real control plane retry would be; identical
	// signatures within one second are (correctly) refused as replays.
	f.advance(time.Second)
	f.mu.Lock()
	stamp := strconv.FormatInt(f.now.Unix(), 10)
	f.mu.Unlock()
	r := httptest.NewRequest(method, "https://platform.rooms.example.com"+target, strings.NewReader(body))
	r.Host = "platform.rooms.example.com"
	sig := ed25519.Sign(f.priv, platformSigningInput(f.keyID, stamp, method, target, []byte(body)))
	r.Header.Set("X-Platform-Id", f.keyID)
	r.Header.Set("X-Platform-Timestamp", stamp)
	r.Header.Set("X-Platform-Signature", base64.StdEncoding.EncodeToString(sig))
	if mutate != nil {
		mutate(r)
	}
	w := httptest.NewRecorder()
	f.host.ServeHTTP(w, r)
	return w
}

func (f *platformFixture) call(method, target, body string) *httptest.ResponseRecorder {
	return f.request(method, target, body, nil)
}

func (f *platformFixture) create(id string) platformWorkspaceView {
	f.t.Helper()
	w := f.call("POST", "/platform/v1/workspaces", `{"id":"`+id+`","name":"`+strings.ToUpper(id)+` Ltd"}`)
	if w.Code != 201 {
		f.t.Fatal("create", id, w.Code, w.Body.String())
	}
	var v platformWorkspaceView
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		f.t.Fatal(err)
	}
	return v
}

func (f *platformFixture) visit(id, path string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "https://"+id+".rooms.example.com"+path, nil)
	r.Host = id + ".rooms.example.com"
	w := httptest.NewRecorder()
	f.host.ServeHTTP(w, r)
	return w
}

func TestPlatformAuthenticationRefusesEverythingUnsigned(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	other, otherKey, _ := ed25519.GenerateKey(rand.Reader)
	_ = other
	sign := func(r *http.Request, key ed25519.PrivateKey, input []byte) {
		r.Header.Set("X-Platform-Signature", base64.StdEncoding.EncodeToString(ed25519.Sign(key, input)))
	}
	stamp := strconv.FormatInt(f.now.Unix(), 10)
	list := "/platform/v1/workspaces"
	cases := map[string]func(*http.Request){
		"no headers":     func(r *http.Request) { r.Header = http.Header{} },
		"unknown key id": func(r *http.Request) { r.Header.Set("X-Platform-Id", "someone-else") },
		"wrong key":      func(r *http.Request) { sign(r, otherKey, platformSigningInput(f.keyID, stamp, "GET", list, nil)) },
		"stale": func(r *http.Request) {
			r.Header.Set("X-Platform-Timestamp", strconv.FormatInt(f.now.Add(-2*time.Minute).Unix(), 10))
		},
		"from the future": func(r *http.Request) {
			r.Header.Set("X-Platform-Timestamp", strconv.FormatInt(f.now.Add(2*time.Minute).Unix(), 10))
		},
		"junk timestamp": func(r *http.Request) { r.Header.Set("X-Platform-Timestamp", "yesterday") },
		"junk signature": func(r *http.Request) { r.Header.Set("X-Platform-Signature", "!!!") },
		"signed for another path": func(r *http.Request) {
			sign(r, f.priv, platformSigningInput(f.keyID, stamp, "GET", "/platform/v1/workspaces/acme", nil))
		},
		"signed for another method": func(r *http.Request) {
			sign(r, f.priv, platformSigningInput(f.keyID, stamp, "DELETE", list, nil))
		},
		"extension contract signature reused": func(r *http.Request) {
			sign(r, f.priv, extensionSigningInput(f.keyID, stamp, "GET", list, nil))
		},
	}
	var first string
	for name, mutate := range cases {
		w := f.request("GET", list, "", mutate)
		if w.Code != 403 {
			t.Fatal(name, w.Code, w.Body.String())
		}
		if first == "" {
			first = w.Body.String()
		} else if w.Body.String() != first {
			t.Fatal("refusals differ, so they reveal why:", name, w.Body.String())
		}
	}
	// The signed body is the body that is acted on.
	signed := `{"id":"acme","name":"Acme"}`
	w := f.request("POST", list, `{"id":"evil","name":"Evil"}`, func(r *http.Request) {
		sign(r, f.priv, platformSigningInput(f.keyID, stamp, "POST", list, []byte(signed)))
	})
	if w.Code != 403 {
		t.Fatal("tampered body accepted", w.Code)
	}
	// A genuine request works once and cannot be replayed.
	r := httptest.NewRequest("GET", "https://platform.rooms.example.com"+list, nil)
	r.Host = "platform.rooms.example.com"
	r.Header.Set("X-Platform-Id", f.keyID)
	r.Header.Set("X-Platform-Timestamp", stamp)
	sign(r, f.priv, platformSigningInput(f.keyID, stamp, "GET", list, nil))
	replay := r.Clone(r.Context())
	for i, want := range []int{200, 429} {
		req := r
		if i == 1 {
			req = replay
		}
		rec := httptest.NewRecorder()
		f.host.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatal("replay handling", i, rec.Code)
		}
	}
}

func TestPlatformAPIIsUnavailableWithoutKeysAndOnOrdinaryHosts(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	f.host.platform.(*platformAPI).keyFile = ""
	if w := f.call("GET", "/platform/v1/workspaces", ""); w.Code != 403 {
		t.Fatal("no key file configured must refuse", w.Code)
	}
	f.host.platform = nil
	if w := f.call("GET", "/platform/v1/workspaces", ""); w.Code != 404 {
		t.Fatal("unconfigured platform host must be a bare 404", w.Code)
	}
	// The platform path means nothing on a workspace host.
	f2 := newPlatformFixture(t)
	f2.create("acme")
	if w := f2.visit("acme", "/platform/v1/workspaces"); w.Code == 200 && strings.Contains(w.Body.String(), "workspaces") {
		t.Fatal("platform API reachable on a workspace host")
	}
	single := newWorkspaceHost(tenancySettings{Mode: tenancySingle}, &fakeRegistry{}, identityRuntime)
	single.platform = f2.host.platform
	if code, body := hostGet(single, "platform.rooms.example.com"); code != 200 || body != "workspace:default" {
		t.Fatal("single mode must never route to the platform API", code, body)
	}
}

func TestPlatformCreateIsIdempotentAndValidated(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	v := f.create("acme")
	if v.Origin != "https://acme.rooms.example.com" || v.Status != workspaceActive || v.SetupPending == nil || !*v.SetupPending {
		t.Fatal(v)
	}
	prefix := "https://acme.rooms.example.com/#setup="
	if !strings.HasPrefix(v.SetupLink, prefix) || len(strings.TrimPrefix(v.SetupLink, prefix)) != 64 {
		t.Fatal("setup link malformed:", v.SetupLink)
	}
	if v.SetupExpires == nil || time.Until(*v.SetupExpires) < 23*time.Hour || time.Until(*v.SetupExpires) > 25*time.Hour {
		t.Fatal("hosted setup links must expire in about a day", v.SetupExpires)
	}
	if v.Limits != defaultWorkspaceLimits {
		t.Fatal(v.Limits)
	}
	// The workspace's own history records who created it.
	found := false
	for _, e := range f.dbs["acme"].snapshot() {
		if e.kind == "platform.workspace.action" && strings.Contains(e.payload, `"created"`) {
			found = true
		}
	}
	if !found {
		t.Fatal("creation was not recorded in the workspace's own history")
	}
	if w := f.visit("acme", "/api/setup-status"); w.Code != 200 || !strings.Contains(w.Body.String(), `"setupRequired":true`) {
		t.Fatal(w.Code, w.Body.String())
	}
	// Repeating the same request returns the same workspace and setup link.
	again := f.call("POST", "/platform/v1/workspaces", `{"id":"acme","name":"ACME Ltd"}`)
	var v2 platformWorkspaceView
	json.Unmarshal(again.Body.Bytes(), &v2)
	if again.Code != 200 || v2.SetupLink != v.SetupLink {
		t.Fatal("create is not idempotent", again.Code, v2.SetupLink)
	}
	if w := f.call("POST", "/platform/v1/workspaces", `{"id":"acme","name":"Someone Else"}`); w.Code != 409 {
		t.Fatal("a different name must conflict", w.Code)
	}
	for name, body := range map[string]string{
		"reserved":       `{"id":"www","name":"X"}`,
		"default":        `{"id":"default","name":"X"}`,
		"uppercase":      `{"id":"Acme","name":"X"}`,
		"dot":            `{"id":"a.b","name":"X"}`,
		"long id":        `{"id":"` + strings.Repeat("a", 64) + `","name":"X"}`,
		"empty name":     `{"id":"ok","name":""}`,
		"padded name":    `{"id":"ok","name":" X "}`,
		"control name":   `{"id":"ok","name":"a\u0000b"}`,
		"long name":      `{"id":"ok","name":"` + strings.Repeat("n", 101) + `"}`,
		"unknown field":  `{"id":"ok","name":"X","owner":"me"}`,
		"bad limits":     `{"id":"ok","name":"X","limits":{"rooms":0}}`,
		"unknown limits": `{"id":"ok","name":"X","limits":{"gold":1}}`,
		"not json":       `nope`,
	} {
		if w := f.call("POST", "/platform/v1/workspaces", body); w.Code != 400 {
			t.Fatal(name, w.Code, w.Body.String())
		}
	}
	if len(f.store.records) != 1 {
		t.Fatal("a refused create changed the registry", len(f.store.records))
	}
}

func TestPlatformLifecycleTakesEffectImmediately(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	f.create("acme")
	f.create("beta")
	if w := f.visit("acme", "/api/setup-status"); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := f.call("POST", "/platform/v1/workspaces/acme/suspend", ""); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := f.visit("acme", "/api/setup-status"); w.Code != 503 {
		t.Fatal("suspension did not stop service at once", w.Code)
	}
	if w := f.visit("beta", "/api/setup-status"); w.Code != 200 {
		t.Fatal("suspending one workspace affected another", w.Code)
	}
	if w := f.call("POST", "/platform/v1/workspaces/acme/resume", ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := f.visit("acme", "/api/setup-status"); w.Code != 200 {
		t.Fatal("resume did not restore service", w.Code)
	}
	actions := map[string]bool{}
	for _, e := range f.dbs["acme"].snapshot() {
		if e.kind == "platform.workspace.action" {
			for _, a := range []string{"suspend", "resume", "created"} {
				if strings.Contains(e.payload, `"`+a+`"`) {
					actions[a] = true
				}
			}
		}
	}
	if len(actions) != 3 {
		t.Fatal("platform actions missing from the workspace history", actions)
	}
	if w := f.call("POST", "/platform/v1/workspaces/acme/delete", ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := f.visit("acme", "/api/setup-status"); w.Code != 404 {
		t.Fatal("a deleted workspace must be a bare 404", w.Code)
	}
	for _, path := range []string{"/platform/v1/workspaces/acme/suspend", "/platform/v1/workspaces/acme/delete", "/platform/v1/workspaces/acme/setup-link"} {
		if w := f.call("POST", path, ""); w.Code != 404 {
			t.Fatal("deleted workspaces are terminal", path, w.Code)
		}
	}
	if w := f.call("POST", "/platform/v1/workspaces", `{"id":"acme","name":"ACME Ltd"}`); w.Code != 409 {
		t.Fatal("a deleted id must never be reused", w.Code)
	}
	if w := f.call("GET", "/platform/v1/workspaces/nobody", ""); w.Code != 404 {
		t.Fatal(w.Code)
	}
}

func TestPlatformReadsReturnNoWorkspaceContent(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	for _, id := range []string{"aaa", "bbb", "ccc"} {
		f.create(id)
	}
	w := f.call("GET", "/platform/v1/workspaces?limit=2", "")
	var page struct {
		Workspaces []map[string]any `json:"workspaces"`
		Next       string           `json:"next"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || w.Code != 200 || len(page.Workspaces) != 2 || page.Next != "bbb" {
		t.Fatal(w.Code, w.Body.String())
	}
	w = f.call("GET", "/platform/v1/workspaces?limit=2&after=bbb", "")
	json.Unmarshal(w.Body.Bytes(), &page)
	if len(page.Workspaces) != 1 || page.Next != "" {
		t.Fatal(w.Body.String())
	}
	allowed := map[string]bool{"id": true, "name": true, "status": true, "origin": true, "limits": true, "created": true, "setupPending": true, "setupLink": true, "setupExpires": true}
	for _, ws := range page.Workspaces {
		for key := range ws {
			if !allowed[key] {
				t.Fatal("platform API returned a field outside the contract:", key)
			}
		}
		if _, leaked := ws["setupLink"]; leaked {
			t.Fatal("listing leaked a setup link")
		}
	}
	for _, q := range []string{"limit=0", "limit=101", "limit=x", "after=" + strings.Repeat("a", 64)} {
		if w := f.call("GET", "/platform/v1/workspaces?"+q, ""); w.Code != 400 {
			t.Fatal(q, w.Code)
		}
	}
	got := f.call("GET", "/platform/v1/workspaces/aaa", "")
	if strings.Contains(got.Body.String(), "setupLink") || !strings.Contains(got.Body.String(), `"setupPending":true`) {
		t.Fatal("get must report setup state without issuing a link", got.Body.String())
	}
}

func TestSetupLinkReissueReplacesTheOldCode(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	first := f.create("acme")
	oldCode := strings.TrimPrefix(first.SetupLink, "https://acme.rooms.example.com/#setup=")
	w := f.call("POST", "/platform/v1/workspaces/acme/setup-link", "")
	var v platformWorkspaceView
	json.Unmarshal(w.Body.Bytes(), &v)
	newCode := strings.TrimPrefix(v.SetupLink, "https://acme.rooms.example.com/#setup=")
	if w.Code != 200 || len(newCode) != 64 || newCode == oldCode {
		t.Fatal(w.Code, w.Body.String())
	}
	login := func(code string) int {
		r := httptest.NewRequest("POST", "https://acme.rooms.example.com/api/login", strings.NewReader(`{"key":"`+code+`"}`))
		r.Host = "acme.rooms.example.com"
		r.Header.Set("Origin", "https://acme.rooms.example.com")
		rec := httptest.NewRecorder()
		f.host.ServeHTTP(rec, r)
		return rec.Code
	}
	if login(oldCode) != 401 || login(newCode) != 200 {
		t.Fatal("the reissued code must replace the old one")
	}
	raw, _ := os.ReadFile(filepath.Join(f.ps.dataRoot, "workspaces", "acme", "owner-key"))
	if strings.TrimSpace(string(raw)) != newCode {
		t.Fatal("the reissued code was not persisted")
	}
	// An expired link stops working; reissuing gives a fresh window.
	f.advance(25 * time.Hour)
	f.host.entries["acme"].runtime.s.auth.keyExpires = time.Now().Add(-time.Hour)
	if login(newCode) != 401 {
		t.Fatal("expired setup code accepted")
	}
	// Once the owner has claimed the workspace nothing can be reissued.
	f.host.entries["acme"].runtime.s.auth.setupPending = false
	if w := f.call("POST", "/platform/v1/workspaces/acme/setup-link", ""); w.Code != 409 {
		t.Fatal("reissue after claim", w.Code)
	}
}

func TestWorkspaceLimitsAreEnforcedAndAppliedLive(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	f.create("acme")
	f.create("beta")
	for name, body := range map[string]string{"zero rooms": `{"rooms":0}`, "huge": `{"requestsPerMinute":100001}`, "unknown": `{"gold":1}`, "not an object": `[]`, "string": `{"rooms":"many"}`} {
		if w := f.call("PUT", "/platform/v1/workspaces/acme/limits", body); w.Code != 400 {
			t.Fatal(name, w.Code)
		}
	}
	if w := f.call("PUT", "/platform/v1/workspaces/acme/limits", `{"requestsPerMinute":10,"rooms":3,"dailyTurns":7}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	srv := f.host.entries["acme"].runtime.s
	srv.mu.Lock()
	applied := *srv.limits
	turns := srv.dailyTurnLimit
	srv.mu.Unlock()
	if applied.Rooms != 3 || turns != 7 || applied.Members != defaultWorkspaceLimits.Members {
		t.Fatal("new limits were not applied to the running workspace", applied, turns)
	}
	// Requests are metered per workspace; another workspace is unaffected.
	codes := []int{}
	for i := 0; i < 12; i++ {
		codes = append(codes, f.visit("acme", "/api/setup-status").Code)
	}
	ok, limited := 0, 0
	for _, c := range codes {
		switch c {
		case 200:
			ok++
		case 429:
			limited++
		}
	}
	if ok != 10 || limited != 2 {
		t.Fatal("request limit not enforced as configured", codes)
	}
	if w := f.visit("beta", "/api/setup-status"); w.Code != 200 {
		t.Fatal("one workspace's limit affected another", w.Code)
	}
	f.advance(61 * time.Second)
	if w := f.visit("acme", "/api/setup-status"); w.Code != 200 {
		t.Fatal("the window never reopened", w.Code)
	}
	if w := f.visit("acme", "/index.html"); w.Code == 429 {
		t.Fatal("static assets must not be metered")
	}
}

func TestCreationLimitsRefuseAtTheCap(t *testing.T) {
	t.Parallel()
	s, _, h := mockServer(t)
	s.limits = &workspaceLimits{Members: 0, Rooms: 1, Agents: 0, RequestsPerMinute: 600, DailyTurns: 10}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.created","payload":{"name":"Second"}}`); w.Code != 409 || !strings.Contains(w.Body.String(), "limit of 1 rooms") {
		t.Fatal("room cap not enforced", w.Code, w.Body.String())
	}
	membersMux := http.NewServeMux()
	membersMux.HandleFunc("/api/members", s.members)
	if w := mockRequest(s.auth.protect(s.actorGuard(membersMux)), "POST", "/api/members", `{"name":"Dev","role":"Developer","roomIds":[],"publicKey":"x"}`); w.Code != 409 || !strings.Contains(w.Body.String(), "limit of 0 members") {
		t.Fatal("member cap not enforced", w.Code, w.Body.String())
	}
	s.limits = nil // a single-workspace installation has no limits
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.created","payload":{"name":"Second"}}`); w.Code != 200 {
		t.Fatal("no limits must mean no refusal", w.Code, w.Body.String())
	}
	if msg := (&server{limits: &workspaceLimits{Agents: 2}}).limitRefusal("agents", 2); !strings.Contains(msg, "limit of 2 agents") {
		t.Fatal(msg)
	}
	if (&server{limits: &workspaceLimits{Agents: 2}}).limitRefusal("agents", 1) != "" {
		t.Fatal("refused below the cap")
	}
}

type sqlTx struct {
	pgx.Tx
	log     *[]string
	rows    int64
	failing bool
}

func (tx *sqlTx) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	*tx.log = append(*tx.log, sql+" "+strings.TrimSpace(strings.Join(argStrings(args), "|")))
	if tx.failing {
		return pgconn.CommandTag{}, os.ErrInvalid
	}
	return pgconn.NewCommandTag("UPDATE " + strconv.FormatInt(tx.rows, 10)), nil
}
func (tx *sqlTx) Commit(context.Context) error   { *tx.log = append(*tx.log, "commit"); return nil }
func (tx *sqlTx) Rollback(context.Context) error { return nil }

func argStrings(args []any) []string {
	out := []string{}
	for _, a := range args {
		out = append(out, strings.TrimSpace(strings.ReplaceAll(toString(a), "\n", " ")))
	}
	return out
}
func toString(a any) string { b, _ := json.Marshal(a); return string(b) }

type sqlDB struct {
	log     *[]string
	rows    int64
	failing bool
}

func (d sqlDB) Begin(context.Context) (pgx.Tx, error) {
	return &sqlTx{log: d.log, rows: d.rows, failing: d.failing}, nil
}
func (d sqlDB) Query(context.Context, string, ...any) (pgx.Rows, error) { panic("not used") }

func TestRegistryWritesAreParameterisedValidatedAndGeneric(t *testing.T) {
	t.Parallel()
	var log []string
	reg := databaseWorkspaceRegistry{db: sqlDB{log: &log, rows: 1}}
	created, err := reg.Create(t.Context(), "acme", "Acme' Ltd; DROP TABLE workspaces", defaultWorkspaceLimits)
	if err != nil || !created || len(log) != 2 || !strings.HasPrefix(log[0], workspaceInsertSQL) || log[1] != "commit" {
		t.Fatal(created, err, log)
	}
	if strings.Contains(workspaceInsertSQL, "DROP") || strings.Contains(workspaceStatusSQL, "DELETE") || strings.Contains(workspaceSetLimitsSQL, "DELETE") {
		t.Fatal("registry SQL must never delete rows")
	}
	if !strings.Contains(workspaceStatusSQL, "status<>'deleted'") || !strings.Contains(workspaceSetLimitsSQL, "status<>'deleted'") {
		t.Fatal("deleted must be terminal in SQL as well")
	}
	log = nil
	for _, bad := range []func() error{
		func() error { _, e := reg.Create(t.Context(), "www", "X", defaultWorkspaceLimits); return e },
		func() error { _, e := reg.Create(t.Context(), "acme", "", defaultWorkspaceLimits); return e },
		func() error { _, e := reg.SetStatus(t.Context(), "acme", "archived"); return e },
		func() error { _, e := reg.SetStatus(t.Context(), "Bad Id", workspaceActive); return e },
		func() error { _, e := reg.SetLimits(t.Context(), "default", defaultWorkspaceLimits); return e },
		func() error { _, e := reg.List(t.Context(), "", 0); return e },
		func() error { _, e := reg.List(t.Context(), "", 101); return e },
	} {
		if bad() == nil {
			t.Fatal("invalid registry input accepted")
		}
	}
	if len(log) != 0 {
		t.Fatal("invalid input reached the database", log)
	}
	failing := databaseWorkspaceRegistry{db: sqlDB{log: &log, failing: true}}
	if _, err := failing.SetStatus(t.Context(), "acme", workspaceSuspended); err == nil || err != errRegistryUnavailable {
		t.Fatal("database errors must be generic", err)
	}
	none := databaseWorkspaceRegistry{db: sqlDB{log: &log, rows: 0}}
	if changed, err := none.SetStatus(t.Context(), "acme", workspaceActive); err != nil || changed {
		t.Fatal("an update that matched nothing must report no change", changed, err)
	}
}

func TestWorkspaceNameValidation(t *testing.T) {
	t.Parallel()
	for _, ok := range []string{"A", "Acme Ltd", "Zoë & Co.", strings.Repeat("é", 100)} {
		if !validWorkspaceName(ok) {
			t.Fatal("rejected", ok)
		}
	}
	for _, bad := range []string{"", " ", " lead", "trail ", "a\nb", "a\x00b", "a b", strings.Repeat("a", 101), "bad\xffutf8"} {
		if validWorkspaceName(bad) {
			t.Fatal("accepted", bad)
		}
	}
}

func TestPlatformCommandLineClientEndToEnd(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t)
	keyFile := filepath.Join(t.TempDir(), "platform.key")
	var out, diag strings.Builder
	if code := platformKeygenMain([]string{keyFile, "ops-tool"}, &out, &diag); code != 0 {
		t.Fatal(code, diag.String())
	}
	if strings.Contains(out.String(), "PRIVATE") || strings.Contains(out.String(), strings.TrimSpace(readFile(t, keyFile))) {
		t.Fatal("keygen printed the private key")
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(out.String()), &entries); err != nil || entries[0]["id"] != "ops-tool" {
		t.Fatal(out.String(), err)
	}
	if info, _ := os.Stat(keyFile); info.Mode().Perm() != 0o600 {
		t.Fatal("key file must be private", info.Mode())
	}
	if code := platformKeygenMain([]string{keyFile, "ops-tool"}, &out, &diag); code != 1 {
		t.Fatal("keygen overwrote an existing key")
	}
	// Register the generated key beside the fixture's own and call through the real client.
	registryPath := f.host.platform.(*platformAPI).keyFile
	if err := os.WriteFile(registryPath, []byte(out.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(f.host.platform)
	defer server.Close()
	run := func(args ...string) (int, string) {
		var o, e strings.Builder
		code := platformRequestMain(args, &o, &e, func() time.Time { f.advance(time.Second); return f.now })
		return code, o.String()
	}
	code, body := run(keyFile, "ops-tool", "POST", server.URL+"/platform/v1/workspaces", `{"id":"cli","name":"CLI Co"}`)
	if code != 0 || !strings.Contains(body, `"id":"cli"`) || !strings.Contains(body, "#setup=") {
		t.Fatal(code, body)
	}
	if code, body = run(keyFile, "ops-tool", "GET", server.URL+"/platform/v1/workspaces"); code != 0 || !strings.Contains(body, `"cli"`) {
		t.Fatal(code, body)
	}
	if code, _ = run(keyFile, "wrong-id", "GET", server.URL+"/platform/v1/workspaces"); code != 1 {
		t.Fatal("an unknown key id must fail")
	}
	for _, bad := range [][]string{
		{keyFile, "ops-tool", "GET", "http://example.com/platform/v1/workspaces"},
		{keyFile, "ops-tool", "GET", "ftp://x/platform/v1/workspaces"},
		{keyFile, "ops-tool", "GET", server.URL + "/api/state"},
		{keyFile, "ops-tool", "DELETE", server.URL + "/platform/v1/workspaces"},
		{keyFile, "ops-tool"},
	} {
		if code, _ := run(bad...); code != 2 {
			t.Fatal("accepted", bad, code)
		}
	}
	if err := os.Chmod(keyFile, 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _ := run(keyFile, "ops-tool", "GET", server.URL+"/platform/v1/workspaces"); code != 1 {
		t.Fatal("a world-readable key file was used")
	}
	if code := platformKeygenMain([]string{filepath.Join(t.TempDir(), "k"), "Bad_ID"}, &out, &diag); code != 2 {
		t.Fatal("invalid key id accepted")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
