package main

import (
	"agentroom/internal/room"
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const syntheticClaudeURL = "https://claude.ai/oauth/authorize?response_type=code&code_challenge=synthetic-challenge&state=synthetic-state"

func TestClaudeAuthorizationOutputValidation(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"http://claude.ai/oauth/authorize?response_type=code&code_challenge=x", "https://attacker.invalid/oauth/authorize?response_type=code&code_challenge=x", "https://claude.ai.evil.invalid/oauth/authorize?response_type=code&code_challenge=x", "https://user@claude.ai/oauth/authorize?response_type=code&code_challenge=x", "https://claude.ai:444/oauth/authorize?response_type=code&code_challenge=x", "https://claude.ai/oauth/authorize?response_type=token&code_challenge=x", "https://claude.ai/oauth/authorize?response_type=code"} {
		if validClaudeAuthorizationURL(value) {
			t.Fatal("unsafe authorization URL admitted")
		}
	}
	ch := make(chan deviceInstructions, 1)
	out := &deviceInstructionOutput{provider: "claude", loginID: strings.Repeat("a", 32), instructions: ch, cancel: func() {}}
	io.WriteString(out, "Open browser:\n"+syntheticClaudeURL)
	select {
	case <-ch:
		t.Fatal("partial URL emitted")
	default:
	}
	io.WriteString(out, "\nPaste your authorization code:\n")
	select {
	case event := <-ch:
		if event.Type != "authorization" || event.VerificationURL != syntheticClaudeURL || event.LoginID != strings.Repeat("a", 32) {
			t.Fatal(event)
		}
	default:
		t.Fatal("Claude authorization missing")
	}
}

// Real HTTP streaming, but every database/worker/provider dependency is synthetic.
func TestClaudeBrowserAuthorizationThroughHTTP(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	s, db, _ := mockServer(t)
	for _, cid := range []string{"claude-fixture", "other-fixture"} {
		db.seed("connection.created", room.Connection{ID: cid, Provider: "claude", Protocol: "CLI", Endpoint: "ssh://commonroom-server/claude"})
	}
	received := make(chan string, 1)
	release := make(chan struct{})

	s.deviceLoginRun = func(ctx context.Context, alias, encoded string, in io.Reader, out io.Writer) error {
		request, err := decodeWorkerAuth(encoded)
		if err != nil || request.Provider != "claude" || alias != "commonroom-server" || in == nil {
			return io.ErrUnexpectedEOF
		}
		io.WriteString(out, syntheticClaudeURL+"\n")
		code, err := bufio.NewReader(in).ReadString('\n')
		if err != nil {
			return err
		}
		received <- code
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	handler := s.auth.protect(s.actorGuard(http.HandlerFunc(s.providerAuth)))
	host := httptest.NewServer(handler)
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
		host.Close()
	}()
	post := func(body, csrf string) *http.Response {
		t.Helper()
		r, err := http.NewRequest("POST", host.URL+"/api/provider-auth", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		r.Host = "127.0.0.1:8787"
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
		r.Header.Set("X-CSRF-Token", csrf)
		response, err := host.Client().Do(r)
		if err != nil {
			t.Fatal(err)
		}
		return response
	}
	response := post(`{"action":"login","connectionId":"claude-fixture"}`, "csrf")
	defer response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal(response.StatusCode)
	}
	decoder := json.NewDecoder(response.Body)
	var authorization deviceInstructions
	for authorization.Type != "authorization" {
		if err := decoder.Decode(&authorization); err != nil {
			t.Fatal(err)
		}
	}
	if authorization.LoginID == "" || authorization.VerificationURL != syntheticClaudeURL {
		t.Fatal("missing bound instructions")
	}
	body := func(cid, login, code string) string {
		raw, _ := json.Marshal(map[string]string{"action": "login-code", "connectionId": cid, "loginId": login, "code": code})
		return string(raw)
	}
	expect := func(payload, csrf string, want int) {
		t.Helper()
		r := post(payload, csrf)
		defer r.Body.Close()
		if r.StatusCode != want {
			b, _ := io.ReadAll(r.Body)
			t.Fatal(r.StatusCode, string(b))
		}
	}
	code := "syntheticCode#syntheticState"
	expect(body("claude-fixture", strings.Repeat("b", 32), code), "csrf", 409)
	expect(body("other-fixture", authorization.LoginID, code), "csrf", 409)
	expect(body("claude-fixture", authorization.LoginID, "bad\ninput"), "csrf", 400)
	expect(body("claude-fixture", authorization.LoginID, code), "wrong", 403)
	s.auth.mu.Lock()
	ownerSessionBefore := s.auth.sessions["owner"]
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	s.auth.mu.Unlock()
	expect(body("claude-fixture", authorization.LoginID, code), "csrf", 403)
	s.auth.mu.Lock()
	s.auth.sessions["owner"] = ownerSessionBefore
	s.auth.mu.Unlock()

	db.seed("feature.flag.updated", map[string]any{"name": "connectionSignIn", "enabled": false})
	expect(body("claude-fixture", authorization.LoginID, code), "csrf", 403)
	db.seed("feature.flag.updated", map[string]any{"name": "connectionSignIn", "enabled": true})
	expect(`{"action":"login","connectionId":"claude-fixture"}`, "csrf", 409)
	expect(body("claude-fixture", authorization.LoginID, code), "csrf", 200)
	select {
	case value := <-received:
		if value != code+"\n" {
			t.Fatal("incorrect worker stdin")
		}
	case <-time.After(time.Second):
		t.Fatal("code not delivered")
	}
	expect(body("claude-fixture", authorization.LoginID, code), "csrf", 409)
	close(release)
	var complete map[string]any
	if err := decoder.Decode(&complete); err != nil || complete["type"] != "complete" {
		t.Fatal(complete, err)
	}
	var trailer any
	if err := decoder.Decode(&trailer); err != io.EOF {
		t.Fatal("login stream did not finish", err)
	}
	s.mu.Lock()
	remaining := len(s.providerLoginSessions)
	s.mu.Unlock()
	if remaining != 0 {
		t.Fatal("completed login session retained")
	}

	for _, event := range db.snapshot() {
		if strings.Contains(event.payload, code) || strings.Contains(event.payload, "synthetic-challenge") {
			t.Fatal("authorization material audited")
		}
	}
}

func TestClaudeLoginCancellationClosesCodeInput(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "claude-fixture", Protocol: "CLI", Endpoint: "ssh://commonroom-server/claude"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopped := make(chan struct{})
	s.deviceLoginRun = func(ctx context.Context, _ string, _ string, in io.Reader, out io.Writer) error {
		io.WriteString(out, syntheticClaudeURL+"\n")
		cancel()
		_, err := bufio.NewReader(in).ReadString('\n')
		close(stopped)
		return err
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/provider-auth", strings.NewReader(`{"action":"login","connectionId":"claude-fixture"}`)).WithContext(ctx)
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	r.Header.Set("X-CSRF-Token", "csrf")
	s.auth.protect(s.actorGuard(http.HandlerFunc(s.providerAuth))).ServeHTTP(httptest.NewRecorder(), r)
	select {
	case <-stopped:
	default:
		t.Fatal("cancelled Claude input retained")
	}
	s.mu.Lock()
	remaining := len(s.providerLoginSessions)
	s.mu.Unlock()
	if remaining != 0 {
		t.Fatal("cancelled login retained")
	}
}
