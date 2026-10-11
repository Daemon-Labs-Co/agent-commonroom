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

func TestPersistWorkerCredential(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"codex", "claude"} {
		t.Run(provider, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			source := filepath.Join(dir, "export.json")
			root := filepath.Join(dir, "connection", provider)
			raw := []byte(`{"fixture":"first"}`)
			if err := os.WriteFile(source, raw, 0600); err != nil {
				t.Fatal(err)
			}
			if err := persistWorkerCredential(root, provider, source); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(root, providerCredentialName(provider))
			info, err := os.Stat(target)
			if err != nil || info.Mode().Perm() != 0600 || !credentialPresent(target) {
				t.Fatal("private login not saved", err)
			}
			for _, invalid := range []string{"{}", "not-json", strings.Repeat("x", 65537)} {
				os.WriteFile(source, []byte(invalid), 0600)
				if err := persistWorkerCredential(root, provider, source); err == nil {
					t.Fatal("invalid export accepted")
				}
				got, err := os.ReadFile(target)
				if err != nil || string(got) != string(raw) {
					t.Fatal("previous login replaced", err)
				}
			}
			os.Remove(source)
			os.Symlink(target, source)
			if persistWorkerCredential(root, provider, source) == nil {
				t.Fatal("symlink export accepted")
			}
			os.Remove(source)
			if persistWorkerCredential(root, provider, source) == nil {
				t.Fatal("missing export accepted")
			}
			os.WriteFile(source, []byte(`{"fixture":"second"}`), 0600)
			if err := persistWorkerCredential(root, provider, source); err != nil {
				t.Fatal(err)
			}
			got, _ := os.ReadFile(target)
			if string(got) != `{"fixture":"second"}` {
				t.Fatal("replacement not saved")
			}
			files, _ := filepath.Glob(filepath.Join(root, ".login-*"))
			if len(files) != 0 {
				t.Fatal("temporary credential leaked")
			}
		})
	}
}

func TestWorkerCredentialExportPolicy(t *testing.T) {
	t.Parallel()
	for _, provider := range []string{"codex", "claude"} {
		args, err := workerCredentialExportArgs("sha256:trusted", "private-login-volume", provider)
		if err != nil {
			t.Fatal(err)
		}
		joined := strings.Join(args, " ")
		for _, required := range []string{"--network=none", "--read-only", "1000:1000", "--cap-drop=ALL", "no-new-privileges:true", "--log-driver=none", "type=volume,src=private-login-volume,dst=/home/agent,readonly", "O_NOFOLLOW", "65537", "JSON.parse"} {
			if !strings.Contains(joined, required) {
				t.Fatalf("missing export boundary %s", required)
			}
		}
		for _, forbidden := range []string{"type=bind", "provider.sock", "docker.sock", "HTTP_PROXY", "auth-login", "--privileged"} {
			if strings.Contains(joined, forbidden) {
				t.Fatalf("unexpected export capability %s", forbidden)
			}
		}
		mounts := 0
		for _, arg := range args {
			if arg == "--mount" {
				mounts++
			}
		}
		if mounts != 1 {
			t.Fatal("export must mount only the login volume")
		}
		expected := "/home/agent/.codex/auth.json"
		if provider == "claude" {
			expected = "/home/agent/.claude/.credentials.json"
		}
		if args[len(args)-1] != expected {
			t.Fatal("wrong credential file")
		}
	}
	if _, err := workerCredentialExportArgs("image", "volume", "shell"); err == nil {
		t.Fatal("unsupported provider accepted")
	}
}

func TestCredentialExportOutputLimit(t *testing.T) {
	t.Parallel()
	var output credentialExportOutput
	if n, err := output.Write([]byte(strings.Repeat("x", 65536))); err != nil || n != 65536 {
		t.Fatal(n, err)
	}
	if n, err := output.Write([]byte("x")); err == nil || n != 0 || output.Len() != 65536 {
		t.Fatal("overflow must fail without retaining extra bytes")
	}
}
