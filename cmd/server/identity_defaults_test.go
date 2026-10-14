package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMissingRequestActorCannotAcquireOwnerAuthority(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest("POST", "/api/members", strings.NewReader(`{}`))
	if requestActor(r) != "" {
		t.Fatal("missing identity became owner")
	}
	r = r.WithContext(context.WithValue(r.Context(), actorContextKey{}, 42))
	if requestActor(r) != "" {
		t.Fatal("malformed identity became owner")
	}
	// These owner-only handlers must refuse absent middleware identity before
	// CSRF/database access. A nil server dependency makes an accidental bypass
	// visible as a panic rather than accidentally exercising real storage.
	s := &server{}
	for _, c := range []struct {
		method  string
		handler http.HandlerFunc
	}{
		{"POST", s.members}, {"POST", s.providerAuth}, {"POST", s.agentSafety},
		{"GET", s.organization}, {"POST", s.agentRecovery}, {"POST", s.taskDispatch},
		{"POST", s.downloadOwnerKey},
	} {
		w := httptest.NewRecorder()
		c.handler(w, httptest.NewRequest(c.method, "/api/members", strings.NewReader(`{}`)))
		if w.Code != http.StatusForbidden {
			t.Fatal("unidentified administrative request admitted", w.Code)
		}
	}
}

func TestSessionWithoutActorIsRejectedAndRemoved(t *testing.T) {
	t.Parallel()
	for _, hashed := range []bool{false, true} {
		a, err := newOwnerAuth("synthetic", "127.0.0.1:8787")
		if err != nil {
			t.Fatal(err)
		}
		key := "synthetic-cookie"
		if hashed {
			key = sessionDigest(key)
		}
		a.sessions[key] = ownerSession{CSRF: "synthetic", Expires: time.Now().Add(time.Hour)}
		r := httptest.NewRequest("GET", "http://127.0.0.1:8787/api/session", nil)
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "synthetic-cookie"})
		if _, ok := a.session(r); ok {
			t.Fatal("identity-free session accepted")
		}
		if len(a.sessions) != 0 {
			t.Fatal("invalid session retained")
		}
		called := false
		w := httptest.NewRecorder()
		a.protect(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(w, r)
		if called || w.Code != http.StatusUnauthorized {
			t.Fatal("invalid session reached protected handler", w.Code)
		}
	}
}
