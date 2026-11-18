package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHostedCookiesCannotBeTossedFromASiblingWorkspace(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	code, _, pending := f.runtime("acme").s.auth.setupSnapshot()
	if !pending {
		t.Fatal("setup should be pending")
	}
	w := f.do("POST", "/api/login", map[string]string{"key": code}, nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("expected exactly the session cookie", cookies)
	}
	// The session cookie a hosted workspace issues satisfies the __Host- rules: Secure,
	// Path=/ and no Domain, so a sibling workspace cannot overwrite it from its own site.
	if c := cookies[0]; c.Name != hostedSessionCookie || !c.Secure || !c.HttpOnly || c.Path != "/" || c.Domain != "" {
		t.Fatal("hosted cookies must follow the __Host- rules", c)
	}
}

func TestPlainSessionCookieNameIsNotAHostedSession(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	planted := f.request("acme", "GET", "/api/session", nil, nil, func(r *http.Request) {
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: owner.cookie})
	})
	if planted.Code == 200 {
		t.Fatal("a session presented under the plain cookie name was accepted by a hosted workspace")
	}
	if ok := f.do("GET", "/api/session", nil, &owner); ok.Code != 200 {
		t.Fatal("the real session must work", ok.Code)
	}
}

func TestSingleWorkspaceCookiesAreUnchanged(t *testing.T) {
	t.Parallel()
	ps := testPlatform(t)
	rt, err := newWorkspaceRuntime(ps, tenancySettings{Mode: tenancySingle}, newMockDB(t), defaultWorkspaceID)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.stop()
	if rt.s.auth.sessionCookie != "commonroom_session" || rt.s.auth.oidcCookie != "commonroom_oidc" || rt.s.auth.oidcCookiePath != "/api/oidc" {
		t.Fatal("single-workspace installations must keep their cookie names so nobody is signed out", rt.s.auth.sessionCookie)
	}
}

func TestLookAlikeWorkspaceNamesAreRefused(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"xn--pple-43d", "xn--80ak6aa92e", "ab--cd", "a1--b2", "zz--"} {
		if validHostedWorkspaceID(bad) {
			t.Fatal("accepted a reserved internationalised-style label:", bad)
		}
		if _, ok := (tenancySettings{Mode: tenancyMulti, Domain: "rooms.example.com"}).workspaceFromHost(bad + ".rooms.example.com"); ok {
			t.Fatal("served", bad)
		}
	}
	for _, ok := range []string{"a-b", "a-b-c", "acme-1", "a1-2b"} {
		if !validHostedWorkspaceID(ok) {
			t.Fatal("rejected a normal name", ok)
		}
	}
	f := newPlatformFixture(t)
	if w := f.call("POST", "/platform/v1/workspaces", `{"id":"xn--pple-43d","name":"Apple"}`); w.Code != 400 {
		t.Fatal("the platform API created a look-alike workspace", w.Code)
	}
}

func TestCrossSitePostsToPublicIdentityEndpointsAreRefused(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	crossSite := func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") }
	evil := func(r *http.Request) { r.Header.Set("Origin", "https://evil.example.com") }
	for _, path := range []string{"/api/login", "/api/webauthn/login/options", "/api/webauthn/login/verify", "/api/invitations/lookup", "/api/invitations/passkey/options", "/api/invitations/passkey/complete"} {
		for name, mutate := range map[string]func(*http.Request){"cross-site fetch": crossSite, "foreign origin": evil} {
			if w := f.request("acme", "POST", path, `{}`, nil, mutate); w.Code != 403 {
				t.Fatal("a forged cross-site request reached", path, name, w.Code)
			}
		}
	}
}

func TestHostHeaderTricksNeverChangeTheWorkspace(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	f.create("beta")
	// Forwarding headers, a conflicting Origin and URL authority never override the Host the
	// proxy forwarded: the workspace is chosen from r.Host alone.
	for name, mutate := range map[string]func(*http.Request){
		"x-forwarded-host":  func(r *http.Request) { r.Header.Set("X-Forwarded-Host", "beta.rooms.example.com") },
		"forwarded":         func(r *http.Request) { r.Header.Set("Forwarded", "host=beta.rooms.example.com") },
		"x-original-host":   func(r *http.Request) { r.Header.Set("X-Original-Host", "beta.rooms.example.com") },
		"x-host":            func(r *http.Request) { r.Header.Set("X-Host", "beta.rooms.example.com") },
		"uppercase host":    func(r *http.Request) { r.Host = strings.ToUpper(r.Host) },
		"trailing dot host": func(r *http.Request) { r.Host += "." },
	} {
		w := f.request("acme", "GET", "/api/setup-status", nil, nil, mutate)
		if name != "x-forwarded-host" && name != "forwarded" && name != "x-original-host" && name != "x-host" {
			continue
		}
		if w.Code != 200 {
			t.Fatal(name, w.Code)
		}
	}
	acmeCode, _, _ := f.runtime("acme").s.auth.setupSnapshot()
	betaCode, _, _ := f.runtime("beta").s.auth.setupSnapshot()
	// The setup code of acme proves nothing on beta even when beta's name is forwarded.
	trick := func(r *http.Request) { r.Header.Set("X-Forwarded-Host", "beta.rooms.example.com") }
	if w := f.request("acme", "POST", "/api/login", map[string]string{"key": betaCode}, nil, trick); w.Code != 401 {
		t.Fatal("a forwarded host name selected another workspace's setup code", w.Code)
	}
	if w := f.request("beta", "POST", "/api/login", map[string]string{"key": acmeCode}, nil, nil); w.Code != 401 {
		t.Fatal("one workspace's setup code opened another", w.Code)
	}
	// A request line carrying another workspace's authority is routed by that authority, not by a
	// conflicting header, and a host outside the domain is refused outright.
	r := httptest.NewRequest("GET", "https://beta.rooms.example.com/api/setup-status", nil)
	r.Host = "beta.rooms.example.com"
	w := httptest.NewRecorder()
	f.host.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	for _, host := range []string{"rooms.example.com.evil.test", "acme.rooms.example.com@evil.test", "acme.rooms.example.com:99999999", "acme%2erooms.example.com", "acme.rooms.example.com\x00.evil.test"} {
		r := httptest.NewRequest("GET", "https://acme.rooms.example.com/", nil)
		r.Host = host
		w := httptest.NewRecorder()
		f.host.ServeHTTP(w, r)
		if w.Code != 404 && w.Code != 403 {
			t.Fatal("served an unexpected host:", host, w.Code)
		}
	}
}
