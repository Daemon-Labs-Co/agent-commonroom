package main

import (
	"agentroom/internal/oidc/oidctest"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

type ssoFixture struct {
	*idFixture
	idp *oidctest.Server
}

func newSSOFixture(t *testing.T) *ssoFixture {
	t.Helper()
	idp := oidctest.New(t)
	f := newIdentityFixture(t, func(pf *platformFixture) {
		pf.ps.oidc = idp.Provider(t)
		pf.ps.oidcName = "Test SSO"
	})
	return &ssoFixture{idFixture: f, idp: idp}
}

// start begins a flow in a browser and returns the identity provider address and the
// binding cookie the browser keeps.
func (f *ssoFixture) start(ws, path string, s *idSession) (authURL string, binding *http.Cookie) {
	f.t.Helper()
	w := f.request(ws, "GET", path, nil, s, nil)
	if w.Code != http.StatusSeeOther {
		f.t.Fatal("start", path, w.Code, w.Body.String())
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == "commonroom_oidc" {
			binding = c
		}
	}
	if binding == nil || !binding.HttpOnly || binding.SameSite != http.SameSiteLaxMode || binding.Path != "/api/oidc" {
		f.t.Fatal("the flow is not bound to the browser", binding)
	}
	return w.Header().Get("Location"), binding
}

// approve plays the identity provider: it returns the callback address the browser is sent to.
func (f *ssoFixture) approve(authURL string) string {
	f.t.Helper()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Get(authURL)
	if err != nil {
		f.t.Fatal(err)
	}
	resp.Body.Close()
	return resp.Header.Get("Location")
}

func (f *ssoFixture) callback(ws, callbackURL string, binding *http.Cookie, mutate func(*http.Request)) *httptest.ResponseRecorder {
	f.t.Helper()
	u, _ := url.Parse(callbackURL)
	return f.request(ws, "GET", u.RequestURI(), nil, nil, func(r *http.Request) {
		r.Header.Del("Origin")
		r.Header.Set("Sec-Fetch-Site", "cross-site") // the browser arrives from the identity provider
		if binding != nil {
			r.AddCookie(&http.Cookie{Name: "commonroom_oidc", Value: binding.Value})
		}
		if mutate != nil {
			mutate(r)
		}
	})
}

func sessionCookie(w *httptest.ResponseRecorder) string {
	for _, c := range w.Result().Cookies() {
		if c.Name == "commonroom_session" && c.Value != "" {
			return c.Value
		}
	}
	return ""
}

// signUp invites a person and has them accept with single sign-on as the given subject.
func (f *ssoFixture) signUp(owner idSession, subject, name string) *httptest.ResponseRecorder {
	f.t.Helper()
	token, _ := f.invite(owner, "Developer")
	f.idp.Subject = subject
	authURL, binding := f.start(f.ws, "/api/invitations/oidc/start?token="+token+"&name="+url.QueryEscape(name), nil)
	return f.callback(f.ws, f.approve(authURL), binding, nil)
}

func TestSingleSignOnInvitationLoginAndLinking(t *testing.T) {
	t.Parallel()
	f := newSSOFixture(t)
	owner := f.ownerSession("acme")
	if w := f.do("GET", "/api/auth/methods", nil, nil); !strings.Contains(w.Body.String(), `"oidc":true`) || !strings.Contains(w.Body.String(), "Test SSO") {
		t.Fatal(w.Body.String())
	}
	w := f.signUp(owner, "subject-olive", "Olive")
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/" || sessionCookie(w) == "" {
		t.Fatal("accepting an invitation with single sign-on", w.Code, w.Header().Get("Location"))
	}
	var olive string
	for _, a := range f.state("acme").Actors {
		if a.Name == "Olive" {
			olive = a.ID
			if len(a.OIDCLinks) != 1 || a.OIDCLinks[0].Subject != "subject-olive" || a.OIDCLinks[0].Issuer != f.idp.URL || a.Role != "Developer" {
				t.Fatal("account not linked as expected", a)
			}
		}
	}
	if olive == "" {
		t.Fatal("no account created")
	}
	// Signing in again as the same identity reaches the same account, with a new session.
	f.idp.Subject = "subject-olive"
	authURL, binding := f.start(f.ws, "/api/oidc/login", nil)
	login := f.callback(f.ws, f.approve(authURL), binding, nil)
	if login.Code != http.StatusSeeOther || login.Header().Get("Location") != "/" || sessionCookie(login) == "" || sessionCookie(login) == sessionCookie(w) {
		t.Fatal("login", login.Code, login.Header().Get("Location"))
	}
	if w := f.do("GET", "/api/session", nil, &idSession{cookie: sessionCookie(login)}); w.Code != 200 || !strings.Contains(w.Body.String(), olive) {
		t.Fatal("the session is not for the linked account", w.Code, w.Body.String())
	}
	// An identity that was never linked gets nothing.
	f.resetThrottle("acme")
	f.idp.Subject = "stranger"
	authURL, binding = f.start(f.ws, "/api/oidc/login", nil)
	unlinked := f.callback(f.ws, f.approve(authURL), binding, nil)
	if unlinked.Header().Get("Location") != "/#signin=unlinked" || sessionCookie(unlinked) != "" {
		t.Fatal("an unknown identity was signed in", unlinked.Header().Get("Location"))
	}
	// The owner links their own identity; the session cookie is not needed on the way back.
	f.idp.Subject = "subject-owner"
	authURL, binding = f.start(f.ws, "/api/oidc/link", &owner)
	linked := f.callback(f.ws, f.approve(authURL), binding, nil)
	if linked.Header().Get("Location") != "/#signin=linked" {
		t.Fatal(linked.Header().Get("Location"))
	}
	for _, a := range f.state("acme").Actors {
		if a.ID == ownerID && (len(a.OIDCLinks) != 1 || a.OIDCLinks[0].Subject != "subject-owner") {
			t.Fatal("owner link missing", a.OIDCLinks)
		}
	}
	// One identity belongs to one account.
	f.idp.Subject = "subject-olive"
	authURL, binding = f.start(f.ws, "/api/oidc/link", &owner)
	if taken := f.callback(f.ws, f.approve(authURL), binding, nil); taken.Header().Get("Location") != "/#signin=taken" {
		t.Fatal("an identity was linked to two accounts", taken.Header().Get("Location"))
	}
}

func TestSingleSignOnCallbackRefusals(t *testing.T) {
	t.Parallel()
	f := newSSOFixture(t)
	owner := f.ownerSession("acme")
	f.signUp(owner, "subject-pat", "Pat")
	f.idp.Subject = "subject-pat"
	fresh := func() (string, *http.Cookie) {
		f.resetThrottle("acme")
		authURL, binding := f.start(f.ws, "/api/oidc/login", nil)
		return f.approve(authURL), binding
	}
	ok := func(w *httptest.ResponseRecorder) bool { return sessionCookie(w) != "" }
	// A callback replayed into another browser (no cookie, or someone else's) is refused.
	callbackURL, binding := fresh()
	if w := f.callback(f.ws, callbackURL, nil, nil); ok(w) || w.Header().Get("Location") != "/#signin=failed" {
		t.Fatal("a callback without the browser binding was accepted")
	}
	callbackURL, _ = fresh()
	_, otherBrowser := f.start(f.ws, "/api/oidc/login", nil)
	if w := f.callback(f.ws, callbackURL, otherBrowser, nil); ok(w) {
		t.Fatal("a callback with another browser's binding was accepted")
	}
	// State is single use even when the provider would accept the code again.
	f.idp.ReusableCodes = true
	callbackURL, binding = fresh()
	if w := f.callback(f.ws, callbackURL, binding, nil); !ok(w) {
		t.Fatal("the genuine callback failed")
	}
	if w := f.callback(f.ws, callbackURL, binding, nil); ok(w) {
		t.Fatal("a callback state was usable twice")
	}
	f.idp.ReusableCodes = false
	// State is single use.
	callbackURL, binding = fresh()
	if w := f.callback(f.ws, callbackURL, binding, nil); !ok(w) {
		t.Fatal("the genuine callback failed")
	}
	if w := f.callback(f.ws, callbackURL, binding, nil); ok(w) {
		t.Fatal("a callback was usable twice")
	}
	// Unknown or missing state, a missing code and a provider error.
	callbackURL, binding = fresh()
	u, _ := url.Parse(callbackURL)
	for name, q := range map[string]string{"unknown state": "code=x&state=nope", "no state": "code=x", "no code": "state=" + u.Query().Get("state"), "provider error": "error=access_denied&state=" + u.Query().Get("state")} {
		f.resetThrottle("acme")
		w := f.callback(f.ws, "https://"+f.hostName("acme")+"/api/oidc/callback?"+q, binding, nil)
		if ok(w) || w.Header().Get("Location") != "/#signin=failed" {
			t.Fatal("accepted:", name, w.Header().Get("Location"))
		}
	}
	// Expired pending flows are refused.
	f.resetThrottle("acme")
	authURL, binding := f.start(f.ws, "/api/oidc/login", nil)
	callbackURL = f.approve(authURL)
	rt := f.runtime("acme")
	rt.s.identityStore.mu.Lock()
	for id, p := range rt.s.identityStore.pending {
		p.Expires = time.Now().Add(-time.Second)
		rt.s.identityStore.pending[id] = p
	}
	rt.s.identityStore.mu.Unlock()
	if w := f.callback(f.ws, callbackURL, binding, nil); ok(w) {
		t.Fatal("an expired flow was accepted")
	}
	// A flow begun on one workspace means nothing on another.
	f.create("beta")
	f.resetThrottle("acme")
	authURL, binding = f.start(f.ws, "/api/oidc/login", nil)
	if w := f.callback("beta", f.approve(authURL), binding, nil); ok(w) {
		t.Fatal("a flow crossed into another workspace")
	}
	// Disabled accounts cannot sign in.
	for _, a := range f.state("acme").Actors {
		if a.Name == "Pat" {
			a.Enabled = false
			a.Revision++
			if err := rt.s.append(t.Context(), "actor.revised", a); err != nil {
				t.Fatal(err)
			}
		}
	}
	callbackURL, binding = fresh()
	if w := f.callback(f.ws, callbackURL, binding, nil); ok(w) || w.Header().Get("Location") != "/#signin=unlinked" {
		t.Fatal("a disabled account signed in", w.Header().Get("Location"))
	}
}

func TestOnlyTheCallbackAcceptsACrossSiteNavigation(t *testing.T) {
	t.Parallel()
	f := newSSOFixture(t)
	owner := f.ownerSession("acme")
	crossSite := func(r *http.Request) { r.Header.Del("Origin"); r.Header.Set("Sec-Fetch-Site", "cross-site") }
	if w := f.request(f.ws, "GET", "/api/state", nil, &owner, crossSite); w.Code != 403 {
		t.Fatal("a cross-site request reached the state endpoint", w.Code)
	}
	if w := f.request(f.ws, "GET", "/api/oidc/login", nil, nil, crossSite); w.Code != 403 {
		t.Fatal("a cross-site request started a sign-in", w.Code)
	}
	if w := f.request(f.ws, "POST", "/api/oidc/callback", nil, nil, crossSite); w.Code != 403 {
		t.Fatal("a cross-site POST reached the callback", w.Code)
	}
	wrongHost := func(r *http.Request) { r.Host = "beta.rooms.example.com"; crossSite(r) }
	if w := f.request(f.ws, "GET", "/api/oidc/callback?code=x&state=y", nil, nil, wrongHost); w.Code == 303 {
		t.Fatal("the callback answered for a host that is not a workspace of this installation")
	}
	if w := f.request(f.ws, "GET", "/api/oidc/callback?code=x&state=y", nil, nil, crossSite); w.Code != 303 || w.Header().Get("Location") != "/#signin=failed" {
		t.Fatal("the callback must answer a cross-site return with a fixed redirect", w.Code, w.Header().Get("Location"))
	}
}

func TestSessionIsReplacedWhenSigningInAgain(t *testing.T) {
	t.Parallel()
	f := newSSOFixture(t)
	owner := f.ownerSession("acme")
	f.signUp(owner, "subject-ray", "Ray")
	f.idp.Subject = "subject-ray"
	authURL, binding := f.start(f.ws, "/api/oidc/login", nil)
	first := f.callback(f.ws, f.approve(authURL), binding, nil)
	old := sessionCookie(first)
	f.resetThrottle("acme")
	authURL, binding = f.start(f.ws, "/api/oidc/login", nil)
	second := f.callback(f.ws, f.approve(authURL), binding, func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: old}) })
	if sessionCookie(second) == "" || sessionCookie(second) == old {
		t.Fatal("no fresh session was issued")
	}
	if w := f.do("GET", "/api/session", nil, &idSession{cookie: old}); w.Code == 200 {
		t.Fatal("the previous session survived a new sign-in")
	}
}

func TestSingleSignOnUnavailableWhenNotConfiguredOrOff(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil) // no provider configured
	owner := f.ownerSession("acme")
	token, _ := f.invite(owner, "QA")
	for _, path := range []string{"/api/oidc/login", "/api/invitations/oidc/start?token=" + token + "&name=X"} {
		if w := f.do("GET", path, nil, nil); w.Code == http.StatusSeeOther {
			t.Fatal("single sign-on started without a provider", path)
		}
	}
	g := newSSOFixture(t)
	gOwner := g.ownerSession("acme")
	if w := g.do("POST", "/api/events", map[string]any{"kind": "feature.flag.updated", "payload": map[string]any{"name": "oidcSignIn", "enabled": false}}, &gOwner); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := g.do("GET", "/api/oidc/login", nil, nil); w.Code != 403 {
		t.Fatal("single sign-on worked while turned off", w.Code)
	}
	if w := g.do("GET", "/api/auth/methods", nil, nil); !strings.Contains(w.Body.String(), `"oidc":false`) {
		t.Fatal(w.Body.String())
	}
}

func TestOIDCSettingsFromTheEnvironment(t *testing.T) {
	t.Parallel()
	env := func(kv ...string) func(string) string {
		m := map[string]string{}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return func(k string) string { return m[k] }
	}
	if p, _, err := loadOIDCSettings(env()); p != nil || err != nil {
		t.Fatal("unset must mean no provider", err)
	}
	good := []string{"COMMONROOM_OIDC_ISSUER", "https://idp.example.com", "COMMONROOM_OIDC_CLIENT_ID", "cid"}
	if p, name, err := loadOIDCSettings(env(append(good, "COMMONROOM_OIDC_CLIENT_SECRET", "s", "COMMONROOM_OIDC_NAME", "Acme SSO")...)); p == nil || name != "Acme SSO" || err != nil {
		t.Fatal(name, err)
	}
	secretFile := t.TempDir() + "/secret"
	if err := writePrivateFile(secretFile, []byte("file-secret\n")); err != nil {
		t.Fatal(err)
	}
	if p, name, err := loadOIDCSettings(env(append(good, "COMMONROOM_OIDC_CLIENT_SECRET_FILE", secretFile)...)); p == nil || name != "Single sign-on" || err != nil {
		t.Fatal(name, err)
	}
	for name, kv := range map[string][]string{
		"issuer without client id": {"COMMONROOM_OIDC_ISSUER", "https://idp.example.com"},
		"client id without issuer": {"COMMONROOM_OIDC_CLIENT_ID", "cid"},
		"insecure issuer":          {"COMMONROOM_OIDC_ISSUER", "http://idp.example.com", "COMMONROOM_OIDC_CLIENT_ID", "cid"},
		"both secrets":             append(append([]string{}, good...), "COMMONROOM_OIDC_CLIENT_SECRET", "s", "COMMONROOM_OIDC_CLIENT_SECRET_FILE", secretFile),
		"missing secret file":      append(append([]string{}, good...), "COMMONROOM_OIDC_CLIENT_SECRET_FILE", "/nonexistent/secret"),
		"bad name":                 append(append([]string{}, good...), "COMMONROOM_OIDC_NAME", strings.Repeat("n", 41)),
	} {
		if _, _, err := loadOIDCSettings(env(kv...)); err == nil {
			t.Fatal("accepted:", name)
		}
	}
	if err := os.Chmod(secretFile, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadOIDCSettings(env(append(good, "COMMONROOM_OIDC_CLIENT_SECRET_FILE", secretFile)...)); err == nil {
		t.Fatal("a world-readable secret file was accepted")
	}
}
