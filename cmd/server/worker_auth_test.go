package main

import (
	"agentroom/internal/room"
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConnectionAuthProtocol(t *testing.T) {
	t.Parallel()
	v := workerAuthRequest{1, "connection-a", "codex"}
	encoded, err := encodeWorkerAuth(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeWorkerAuth(encoded)
	if err != nil || got != v {
		t.Fatal(got, err)
	}
	for _, raw := range []string{`{"version":1,"connection":"../escape","provider":"codex"}`, `{"version":1,"connection":"a","provider":"shell"}`, `{"version":2,"connection":"a","provider":"claude"}`, `{"version":1,"connection":"a","provider":"codex","command":"sh"}`, `{"version":1,"connection":"a","provider":"codex"}{}`} {
		if _, err = decodeWorkerAuth(base64.RawURLEncoding.EncodeToString([]byte(raw))); err == nil {
			t.Fatal("accepted invalid request", raw)
		}
	}
	if _, err = decodeWorkerAuth(strings.Repeat("a", 1025)); err == nil {
		t.Fatal("unbounded input")
	}
}
func TestConnectionCredentialScope(t *testing.T) {
	t.Parallel()
	a, err := executionConnectionID("room/member/connection-a/ssh://worker")
	if err != nil || a != "connection-a" {
		t.Fatal(a, err)
	}
	b, _ := executionConnectionID("other/member2/connection-a/ssh://worker")
	if connectionAuthKey(a) != connectionAuthKey(b) || connectionAuthKey(a) == connectionAuthKey("connection-b") {
		t.Fatal("incorrect connection scope")
	}
	for _, key := range []string{"", "key", "r/p/../endpoint", "r/p/a;sh/endpoint"} {
		if _, err = executionConnectionID(key); err == nil {
			t.Fatal("unsafe identity accepted")
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	os.WriteFile(path, []byte(`{"fixture":true}`), 0600)
	if !credentialPresent(path) {
		t.Fatal("private fixture rejected")
	}
	os.Chmod(path, 0644)
	if credentialPresent(path) {
		t.Fatal("public credentials accepted")
	}
	os.Chmod(path, 0600)
	os.WriteFile(path, []byte("invalid"), 0600)
	if credentialPresent(path) {
		t.Fatal("invalid JSON accepted")
	}
	os.Remove(path)
	os.Symlink(filepath.Join(dir, "missing"), path)
	if credentialPresent(path) {
		t.Fatal("symlink accepted")
	}
}
func TestConnectionSignInFlagAndOwner(t *testing.T) {
	t.Parallel()
	if !connectionSignInEnabled(nil) || connectionSignInEnabled(map[string]bool{"connectionSignIn": false}) || !connectionSignInEnabled(map[string]bool{"connectionSignIn": true}) {
		t.Fatal("flag policy")
	}
	s, db, _ := mockServer(t)
	h := s.auth.protect(s.actorGuard(http.HandlerFunc(s.providerAuth)))
	db.seed("connection.created", room.Connection{ID: "test-connection", Name: "Test", Protocol: "CLI", Endpoint: "ssh://commonroom-worker"})
	db.seed("feature.flag.updated", map[string]any{"name": "connectionSignIn", "enabled": false})
	w := mockRequest(h, "POST", "/api/provider-auth", `{"action":"login","connectionId":"test-connection"}`)
	if w.Code != 403 || !strings.Contains(w.Body.String(), "disabled") {
		t.Fatal(w.Code, w.Body.String())
	}
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	w = mockRequest(h, "POST", "/api/provider-auth", `{"action":"login","connectionId":"test-connection"}`)
	if w.Code != 403 {
		t.Fatal("non-owner sign-in accepted", w.Code)
	}

}
func TestInteractiveLoginCommand(t *testing.T) {
	t.Parallel()
	got := interactiveWorkerLogin("worker", "safe")
	for _, part := range []string{"ssh -tt", "BatchMode=yes", "StrictHostKeyChecking=yes", "auth-login-v1"} {
		if !strings.Contains(got, part) {
			t.Fatal("missing boundary", part)
		}
	}
	if strings.Contains(got, "auth.json") || strings.Contains(got, "--with-api-key") {
		t.Fatal("secret transport")
	}
}
