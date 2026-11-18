package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newSignupFixture(t *testing.T, settings signupSettings) (*platformFixture, *signupService) {
	t.Helper()
	f := newPlatformFixture(t)
	settings.enabled = true
	if settings.dailyLimit == 0 {
		settings.dailyLimit = 50
	}
	if settings.maxTotal == 0 {
		settings.maxTotal = 1000
	}
	svc := newSignupService(settings, f.host.tenancy, f.store, f.host)
	svc.now = func() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.now }
	f.host.signup = svc
	return f, svc
}

func (f *platformFixture) signupCall(body string, mutate func(*http.Request)) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "https://rooms.example.com/api/signup", strings.NewReader(body))
	r.Host = "rooms.example.com"
	r.RemoteAddr = "203.0.113.7:4444"
	r.Header.Set("Origin", "https://rooms.example.com")
	r.Header.Set("Content-Type", "application/json")
	if mutate != nil {
		mutate(r)
	}
	w := httptest.NewRecorder()
	f.host.ServeHTTP(w, r)
	return w
}

func TestSignupCreatesAWorkspaceAndReturnsItsSetupLink(t *testing.T) {
	t.Parallel()
	f, _ := newSignupFixture(t, signupSettings{})
	w := f.signupCall(`{"id":"acme","name":"Acme Ltd"}`, nil)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `"setupLink":"https://acme.rooms.example.com/#setup=`) || !strings.Contains(body, `"origin":"https://acme.rooms.example.com"`) {
		t.Fatal(body)
	}
	rec, ok, _ := f.store.Lookup(context.Background(), "acme")
	if !ok || rec.Name != "Acme Ltd" || rec.Limits != defaultWorkspaceLimits || rec.Status != workspaceActive {
		t.Fatalf("%+v %v", rec, ok)
	}
}

func TestSignupNeverRevealsAnExistingWorkspaceOrItsSetupLink(t *testing.T) {
	t.Parallel()
	f, _ := newSignupFixture(t, signupSettings{})
	if w := f.signupCall(`{"id":"acme","name":"Acme Ltd"}`, nil); w.Code != 201 {
		t.Fatal(w.Code)
	}
	// Same id and same name from another visitor: the platform API would return the existing
	// workspace; sign-up must not, or the second visitor could claim the first one's workspace.
	other := func(r *http.Request) { r.RemoteAddr = "198.51.100.9:1" }
	w := f.signupCall(`{"id":"acme","name":"Acme Ltd"}`, other)
	if w.Code != 409 || strings.Contains(w.Body.String(), "setup") {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestSignupFlagOffAndOtherHostsServeNothing(t *testing.T) {
	t.Parallel()
	f := newPlatformFixture(t) // no signup service
	if w := f.signupCall(`{"id":"acme","name":"Acme"}`, nil); w.Code != 404 {
		t.Fatalf("flag off: %d", w.Code)
	}
	if _, ok, _ := f.store.Lookup(context.Background(), "acme"); ok {
		t.Fatal("workspace created while sign-up is off")
	}
	f, svc := newSignupFixture(t, signupSettings{})
	for _, host := range []string{"evil.example.com", "rooms.example.com.evil.test", "platform.rooms.example.com"} {
		if f.host.signupHost(host) {
			t.Fatal("sign-up served for", host)
		}
	}
	if !f.host.signupHost("ROOMS.example.com.:443") || svc == nil {
		t.Fatal("bare domain not recognised")
	}
}

func TestSignupRefusesCrossSiteAndMalformedRequests(t *testing.T) {
	t.Parallel()
	f, _ := newSignupFixture(t, signupSettings{})
	cases := map[string]func(*http.Request){
		"no origin":       func(r *http.Request) { r.Header.Del("Origin") },
		"other origin":    func(r *http.Request) { r.Header.Set("Origin", "https://evil.test") },
		"subdomain":       func(r *http.Request) { r.Header.Set("Origin", "https://acme.rooms.example.com") },
		"http origin":     func(r *http.Request) { r.Header.Set("Origin", "http://rooms.example.com") },
		"form submission": func(r *http.Request) { r.Header.Set("Content-Type", "application/x-www-form-urlencoded") },
	}
	for name, mutate := range cases {
		if w := f.signupCall(`{"id":"acme","name":"Acme"}`, mutate); w.Code != 403 {
			t.Errorf("%s: %d", name, w.Code)
		}
	}
	for _, body := range []string{``, `{`, `{"id":"acme","name":"Acme","extra":1}`, `{"id":"admin","name":"Acme"}`, `{"id":"xn--abc","name":"Acme"}`, `{"id":"Acme","name":"Acme"}`, `{"id":"acme","name":" Acme"}`, `{"id":"acme","name":""}`, `{"id":"acme","name":"` + strings.Repeat("a", 3000) + `"}`} {
		if w := f.signupCall(body, nil); w.Code != 400 {
			t.Errorf("%.30q: %d", body, w.Code)
		}
	}
	if recs, _ := f.store.List(context.Background(), "", 10); len(recs) != 0 {
		t.Fatal("a refused request created a workspace")
	}
}

func TestSignupCodeIsRequiredWhenConfigured(t *testing.T) {
	t.Parallel()
	f, _ := newSignupFixture(t, signupSettings{code: "private-beta-code"})
	for _, body := range []string{`{"id":"acme","name":"Acme"}`, `{"id":"acme","name":"Acme","code":"wrong"}`, `{"id":"acme","name":"Acme","code":"private-beta-code "}`} {
		if w := f.signupCall(body, nil); w.Code != 403 {
			t.Fatalf("%s: %d", body, w.Code)
		}
	}
	if w := f.signupCall(`{"id":"acme","name":"Acme","code":"private-beta-code"}`, nil); w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestSignupLimitsPerAddressAndAttempts(t *testing.T) {
	t.Parallel()
	f, _ := newSignupFixture(t, signupSettings{})
	for i := 0; i < signupPerAddressLimit; i++ {
		if w := f.signupCall(`{"id":"team-`+strconv.Itoa(i)+`","name":"Team"}`, nil); w.Code != 201 {
			t.Fatal(i, w.Code, w.Body.String())
		}
	}
	if w := f.signupCall(`{"id":"team-x","name":"Team"}`, nil); w.Code != 429 {
		t.Fatal("per-address creation limit not enforced", w.Code)
	}
	// Another address is unaffected.
	if w := f.signupCall(`{"id":"team-y","name":"Team"}`, func(r *http.Request) { r.RemoteAddr = "198.51.100.1:9" }); w.Code != 201 {
		t.Fatal(w.Code)
	}
	// An hour later the first address may sign up again.
	f.advance(signupWindow + time.Second)
	if w := f.signupCall(`{"id":"team-z","name":"Team"}`, nil); w.Code != 201 {
		t.Fatal("limit did not reset", w.Code)
	}
}

func TestSignupWrongCodeGuessesAreRateLimited(t *testing.T) {
	t.Parallel()
	f, _ := newSignupFixture(t, signupSettings{code: "private-beta-code"})
	for i := 0; i < signupAttemptsPerAddress; i++ {
		if w := f.signupCall(`{"id":"acme","name":"Acme","code":"guess-`+strconv.Itoa(i)+`"}`, nil); w.Code != 403 {
			t.Fatal(i, w.Code)
		}
	}
	// Even the right code is refused once the attempt budget is spent.
	if w := f.signupCall(`{"id":"acme","name":"Acme","code":"private-beta-code"}`, nil); w.Code != 429 {
		t.Fatal(w.Code)
	}
}

func TestSignupDailyAndTotalCaps(t *testing.T) {
	t.Parallel()
	f, _ := newSignupFixture(t, signupSettings{dailyLimit: 2, maxTotal: 1000})
	for i, addr := range []string{"198.51.100.1:1", "198.51.100.2:1"} {
		if w := f.signupCall(`{"id":"day-`+strconv.Itoa(i)+`","name":"Day"}`, func(r *http.Request) { r.RemoteAddr = addr }); w.Code != 201 {
			t.Fatal(w.Code)
		}
	}
	if w := f.signupCall(`{"id":"day-9","name":"Day"}`, func(r *http.Request) { r.RemoteAddr = "198.51.100.3:1" }); w.Code != 503 {
		t.Fatal("daily cap not enforced", w.Code)
	}
	f.advance(25 * time.Hour)
	if w := f.signupCall(`{"id":"day-9","name":"Day"}`, func(r *http.Request) { r.RemoteAddr = "198.51.100.3:1" }); w.Code != 201 {
		t.Fatal("daily cap did not reset", w.Code)
	}

	f, _ = newSignupFixture(t, signupSettings{dailyLimit: 100, maxTotal: 2})
	for i := 0; i < 2; i++ {
		addr := "198.51.100." + strconv.Itoa(i+1) + ":1"
		if w := f.signupCall(`{"id":"tot-`+strconv.Itoa(i)+`","name":"Total"}`, func(r *http.Request) { r.RemoteAddr = addr }); w.Code != 201 {
			t.Fatal(w.Code)
		}
	}
	if w := f.signupCall(`{"id":"tot-9","name":"Total"}`, func(r *http.Request) { r.RemoteAddr = "198.51.100.9:1" }); w.Code != 503 {
		t.Fatal("total cap not enforced", w.Code)
	}
}

func TestSignupTrustsTheProxyAddressOnlyFromLoopback(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "127.0.0.1:5000"
	r.Header.Set("X-Real-IP", "203.0.113.50")
	if got := clientAddress(r); got != "203.0.113.50" {
		t.Fatal(got)
	}
	r.RemoteAddr = "198.51.100.4:5000" // direct connection: the header is the client's own claim
	if got := clientAddress(r); got != "198.51.100.4" {
		t.Fatal(got)
	}
	r.RemoteAddr = "127.0.0.1:5000"
	r.Header.Set("X-Real-IP", "not-an-ip")
	if got := clientAddress(r); got != "127.0.0.1" {
		t.Fatal(got)
	}
}

func TestSignupServesOnlyItsOwnPageAndConfig(t *testing.T) {
	t.Parallel()
	f, svc := newSignupFixture(t, signupSettings{code: "private-beta-code"})
	svc.web = staticHandler("../../web") // tests run in cmd/server; the service runs from the repository root
	get := func(path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "https://rooms.example.com"+path, nil)
		r.Host = "rooms.example.com"
		w := httptest.NewRecorder()
		f.host.ServeHTTP(w, r)
		return w
	}
	if w := get("/api/signup/config"); w.Code != 200 || !strings.Contains(w.Body.String(), `"codeRequired":true`) || !strings.Contains(w.Body.String(), `"domain":"rooms.example.com"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, path := range []string{"/api/state", "/api/webauthn/login/options", "/app.js", "/identity.js", "/../cmd/server/main.go", "/api/signup"} {
		if w := get(path); w.Code == 200 {
			t.Fatalf("%s served on the sign-up host", path)
		}
	}
	for _, path := range []string{"/", "/signup.js", "/signup.css", "/style.css", "/assets/commonroom-mark.png"} {
		if w := get(path); w.Code != 200 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
	if w := get("/"); w.Header().Get("Content-Security-Policy") == "" || w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("security headers missing")
	}
}

func TestLoadSignupSettings(t *testing.T) {
	t.Parallel()
	multi := tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if s, err := loadSignupSettings(env(nil), multi); err != nil || s.enabled {
		t.Fatal("default must be off", s, err)
	}
	if s, err := loadSignupSettings(env(map[string]string{"COMMONROOM_SIGNUP": "closed"}), tenancySettings{Mode: tenancySingle}); err != nil || s.enabled {
		t.Fatal(s, err)
	}
	s, err := loadSignupSettings(env(map[string]string{"COMMONROOM_SIGNUP": "open", "COMMONROOM_SIGNUP_CODE": "long-enough-code", "COMMONROOM_SIGNUP_DAILY_LIMIT": "7", "COMMONROOM_SIGNUP_MAX_WORKSPACES": "9"}), multi)
	if err != nil || !s.enabled || s.code != "long-enough-code" || s.dailyLimit != 7 || s.maxTotal != 9 {
		t.Fatal(s, err)
	}
	for name, m := range map[string]map[string]string{
		"bad flag":    {"COMMONROOM_SIGNUP": "yes"},
		"short code":  {"COMMONROOM_SIGNUP": "open", "COMMONROOM_SIGNUP_CODE": "short"},
		"spaced code": {"COMMONROOM_SIGNUP": "open", "COMMONROOM_SIGNUP_CODE": " long-enough-code"},
		"zero daily":  {"COMMONROOM_SIGNUP": "open", "COMMONROOM_SIGNUP_DAILY_LIMIT": "0"},
		"junk total":  {"COMMONROOM_SIGNUP": "open", "COMMONROOM_SIGNUP_MAX_WORKSPACES": "many"},
		"huge daily":  {"COMMONROOM_SIGNUP": "open", "COMMONROOM_SIGNUP_DAILY_LIMIT": "100001"},
	} {
		if _, err := loadSignupSettings(env(m), multi); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := loadSignupSettings(env(map[string]string{"COMMONROOM_SIGNUP": "open"}), tenancySettings{Mode: tenancySingle}); err == nil {
		t.Fatal("sign-up must require multi-workspace mode")
	}
}
