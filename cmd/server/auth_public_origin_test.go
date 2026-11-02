package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConfiguredPublicOriginBoundary(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{false, true} {
		a, _ := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
		if enabled {
			if err := a.configurePublicOrigin("https://agentcommonroom.home.arpa"); err != nil {
				t.Fatal(err)
			}
		}
		for _, tc := range []struct {
			host, origin, site string
			want               bool
		}{
			{"agentcommonroom.home.arpa", "", "", enabled},
			{"agentcommonroom.home.arpa", "https://agentcommonroom.home.arpa", "same-origin", enabled},
			{"agentcommonroom.home.arpa", "http://agentcommonroom.home.arpa", "", false},
			{"agentcommonroom.home.arpa", "https://evil.example", "", false},
			{"agentcommonroom.home.arpa", "", "cross-site", false},
			{"evil.example", "https://agentcommonroom.home.arpa", "", false},
			{"agentcommonroom.home.arpa.evil.example", "", "", false},
			{"127.0.0.1:8787", "http://127.0.0.1:8787", "", true},
			{"127.0.0.1:8787", "https://agentcommonroom.home.arpa", "", false},
		} {
			r := httptest.NewRequest("GET", "http://127.0.0.1:8787/", nil)
			r.Host = tc.host
			r.Header.Set("Origin", tc.origin)
			r.Header.Set("Sec-Fetch-Site", tc.site)
			r.Header.Set("X-Forwarded-Host", "agentcommonroom.home.arpa")
			r.Header.Set("X-Forwarded-Proto", "https")
			if got := a.trustedRequest(r); got != tc.want {
				t.Fatalf("enabled=%v host=%s origin=%s site=%s got=%v", enabled, tc.host, tc.origin, tc.site, got)
			}
		}
	}
}

func TestPublicOriginValidation(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"http://agentcommonroom.home.arpa", "https://*.home.arpa", "https://user:pass@host", "https://host/", "https://host?q=1", "https://host?", "https://host#fragment", "https://HOST", "https://host:99999", "https://"} {
		a, _ := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
		if a.configurePublicOrigin(value) == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}

func TestPublicOriginLoginAndLogoutCookies(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{false, true} {
		a, _ := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
		origin := "http://127.0.0.1:8787"
		if enabled {
			origin = "https://agentcommonroom.home.arpa"
			if err := a.configurePublicOrigin(origin); err != nil {
				t.Fatal(err)
			}
		}
		s := &server{auth: a}
		r := httptest.NewRequest("POST", origin+"/api/login", strings.NewReader(`{"key":"`+a.key+`"}`))
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		a.protect(http.HandlerFunc(s.login)).ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatal(w.Code)
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Secure != enabled || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
			t.Fatal("login cookie protections")
		}
		var body struct {
			CSRF string `json:"csrf"`
		}
		if json.Unmarshal(w.Body.Bytes(), &body) != nil {
			t.Fatal("login response")
		}
		r = httptest.NewRequest("POST", origin+"/api/logout", nil)
		r.Header.Set("Origin", origin)
		r.Header.Set("X-CSRF-Token", body.CSRF)
		r.AddCookie(cookies[0])
		w = httptest.NewRecorder()
		a.protect(http.HandlerFunc(s.logout)).ServeHTTP(w, r)
		if w.Code != 200 || len(w.Result().Cookies()) != 1 || w.Result().Cookies()[0].Secure != enabled {
			t.Fatal("logout cookie protections", w.Code)
		}
	}
}
