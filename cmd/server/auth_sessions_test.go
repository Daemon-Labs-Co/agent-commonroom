package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHumanSessionSurvivesReloadAndRestart(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "sessions.json")
	key := strings.Repeat("a", 64)
	a, _ := newOwnerAuth(key, "127.0.0.1:8787")
	if err := a.configureSessionPersistence(path, true); err != nil {
		t.Fatal(err)
	}
	s := server{auth: a}
	w := httptest.NewRecorder()
	s.login(w, httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"key":"`+key+`"}`)))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	request := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/logout", nil)
	request.AddCookie(cookie)
	initial, ok := a.session(request)
	if !ok {
		t.Fatal("reload lost session")
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), cookie.Value) || strings.Contains(string(raw), key) {
		t.Fatal("store exposed bearer token or bootstrap key")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal("session store must be private")
	}
	fresh, _ := newOwnerAuth(key, "127.0.0.1:8787")
	if err := fresh.configureSessionPersistence(path, true); err != nil {
		t.Fatal(err)
	}
	resumed, ok := fresh.session(request)
	if !ok || resumed.ActorID != initial.ActorID || resumed.CSRF != initial.CSRF || !resumed.Expires.Equal(initial.Expires) {
		t.Fatal("restart lost session identity, expiry or CSRF")
	}
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.MaxAge != 8*3600 {
		t.Fatal("cookie protections changed")
	}
	request.Header.Set("X-CSRF-Token", initial.CSRF)
	s.auth = fresh
	w = httptest.NewRecorder()
	s.logout(w, request)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	third, _ := newOwnerAuth(key, "127.0.0.1:8787")
	if err := third.configureSessionPersistence(path, true); err != nil {
		t.Fatal(err)
	}
	if _, ok := third.session(request); ok {
		t.Fatal("restart resurrected signed-out session")
	}
}
func TestPrivateSessionStoreFailuresAndDisabledState(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"invalid", "public", "symlink", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "sessions.json")
			switch kind {
			case "invalid":
				os.WriteFile(path, []byte("bad"), 0600)
			case "public":
				os.WriteFile(path, []byte("{}"), 0644)
			case "symlink":
				os.WriteFile(filepath.Join(dir, "target"), []byte("{}"), 0600)
				os.Symlink(filepath.Join(dir, "target"), path)
			case "oversized":
				os.WriteFile(path, []byte(strings.Repeat("x", 65537)), 0600)
			}
			a, _ := newOwnerAuth("key", "127.0.0.1:8787")
			if err := a.configureSessionPersistence(path, true); err == nil {
				t.Fatal("unsafe store accepted")
			}
		})
	}
	path := filepath.Join(t.TempDir(), "sessions.json")
	a, _ := newOwnerAuth("key", "127.0.0.1:8787")
	a.sessions["token"] = ownerSession{ActorID: ownerID, CSRF: strings.Repeat("c", 64), Expires: time.Now().Add(time.Hour)}
	a.configureSessionPersistence(path, true)
	if err := a.persistSessionsLocked(); err != nil {
		t.Fatal(err)
	}
	fresh, _ := newOwnerAuth("key", "127.0.0.1:8787")
	if err := fresh.configureSessionPersistence(path, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("disabled persistence retained reusable sessions")
	}
	if err := fresh.configureSessionPersistence(path, true); err != nil || len(fresh.sessions) != 0 {
		t.Fatal("reenabling resurrected session")
	}
}
func TestSessionExpiryIdentityAndWriteFailure(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "sessions.json")
	a, _ := newOwnerAuth("key", "127.0.0.1:8787")
	a.configureSessionPersistence(path, true)
	a.sessions["expired"] = ownerSession{ActorID: ownerID, CSRF: strings.Repeat("c", 64), Expires: time.Now().Add(-time.Hour)}
	a.sessions["active"] = ownerSession{ActorID: ownerID, CSRF: strings.Repeat("c", 64), Expires: time.Now().Add(time.Hour)}
	if err := a.persistSessionsLocked(); err != nil {
		t.Fatal(err)
	}
	fresh, _ := newOwnerAuth("key", "127.0.0.1:8787")
	if err := fresh.configureSessionPersistence(path, true); err != nil || len(fresh.sessions) != 1 {
		t.Fatal("expiry not preserved")
	}
	rotated, _ := newOwnerAuth("other-key", "127.0.0.1:8787")
	if err := rotated.configureSessionPersistence(path, true); err != nil || len(rotated.sessions) != 0 {
		t.Fatal("changed identity inherited sessions")
	}
	a.sessionPath = filepath.Join(path, "blocked")
	s := server{auth: a}
	r := httptest.NewRequest("POST", "/api/logout", nil)
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "active"})
	r.Header.Set("X-CSRF-Token", strings.Repeat("c", 64))
	w := httptest.NewRecorder()
	s.logout(w, r)
	if w.Code != 503 {
		t.Fatal("failed durable logout reported success")
	}
	if _, ok := a.session(r); !ok {
		t.Fatal("failed logout did not restore retryable session")
	}
}
func TestPersistentSessionFlagParsing(t *testing.T) {
	for _, value := range []string{"", "true", "false", "invalid"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("COMMONROOM_PERSISTENT_SESSIONS", value)
			enabled, err := persistentHumanSessionsEnabled()
			if value == "invalid" {
				if err == nil {
					t.Fatal("invalid flag accepted")
				}
				return
			}
			if err != nil || enabled != (value != "false") {
				t.Fatal(enabled, err)
			}
		})
	}
}
