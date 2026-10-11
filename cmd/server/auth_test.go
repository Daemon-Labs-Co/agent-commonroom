package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOwnerBoundary(t *testing.T) {
	a, e := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
	if e != nil {
		t.Fatal(e)
	}
	s := &server{auth: a}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", s.login)
	mux.HandleFunc("/api/logout", s.logout)
	mux.HandleFunc("/api/session", s.session)
	for _, path := range []string{"/api/state", "/api/events", "/api/agents", "/api/integrity", "/api/session"} {
		r := httptest.NewRequest("GET", "http://127.0.0.1:8787"+path, nil)
		w := httptest.NewRecorder()
		a.protect(mux).ServeHTTP(w, r)
		if w.Code != 401 {
			t.Fatalf("anonymous %s: %d", path, w.Code)
		}
	}
	for _, v := range []struct{ host, origin, site string }{{"evil.example", "", ""}, {"127.0.0.1:8787", "http://evil.example", ""}, {"127.0.0.1:8787", "", "cross-site"}} {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/login", strings.NewReader(`{"key":"`+a.key+`"}`))
		r.Host = v.host
		r.Header.Set("Origin", v.origin)
		r.Header.Set("Sec-Fetch-Site", v.site)
		w := httptest.NewRecorder()
		a.protect(mux).ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal("untrusted browser accepted")
		}
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/login", strings.NewReader(`{"key":"wrong"}`))
	w := httptest.NewRecorder()
	a.protect(mux).ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal("wrong key accepted")
	}
	r = httptest.NewRequest("POST", "http://127.0.0.1:8787/api/login", strings.NewReader(`{"key":"`+a.key+`"}`))
	w = httptest.NewRecorder()
	a.protect(mux).ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("cookie protections missing")
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	r = httptest.NewRequest("POST", "http://127.0.0.1:8787/api/logout", nil)
	r.AddCookie(cookie)
	w = httptest.NewRecorder()
	a.protect(mux).ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("CSRF missing accepted")
	}
	r.Header.Set("X-CSRF-Token", body["csrf"])
	w = httptest.NewRecorder()
	a.protect(mux).ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("logout failed")
	}
	if _, ok := a.session(r); ok {
		t.Fatal("revoked session usable")
	}
	a.sessions[cookie.Value] = ownerSession{Expires: time.Now().Add(-time.Minute)}
	if _, ok := a.session(r); ok {
		t.Fatal("expired session usable")
	}
	if _, e = newOwnerAuth(a.key, "0.0.0.0:8787"); e == nil {
		t.Fatal("remote binding allowed")
	}
}
func TestOwnerLoginRateLimit(t *testing.T) {
	a, _ := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
	s := &server{auth: a}
	for i := 0; i < 6; i++ {
		r := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"key":"wrong"}`))
		w := httptest.NewRecorder()
		s.login(w, r)
		want := 401
		if i == 5 {
			want = 429
		}
		if w.Code != want {
			t.Fatalf("attempt %d: %d", i, w.Code)
		}
	}
}
func TestOwnerKeyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private", "key")
	first, e := loadOwnerKey(path)
	if e != nil {
		t.Fatal(e)
	}
	second, e := loadOwnerKey(path)
	if e != nil || first != second {
		t.Fatal("key not persistent")
	}
	if e = os.Chmod(path, 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = loadOwnerKey(path); e == nil {
		t.Fatal("unsafe permissions accepted")
	}
}
