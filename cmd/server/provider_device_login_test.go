package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCodexSignInDoesNotRequireDesktopTerminal(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	s, db, _ := mockServer(t)
	s.deviceLoginRun = func(_ context.Context, alias, encoded string, _ io.Reader, out io.Writer) error {
		_, err := io.WriteString(out, "https://auth.openai.com/codex/device\nABCD-1234\naccess_token=synthetic-private-token\n")
		return err
	}
	db.seed("connection.created", room.Connection{ID: "device-fixture", Name: "Fixture", Provider: "openai", Protocol: "CLI", Endpoint: "ssh://commonroom-server"})
	h := s.auth.protect(s.actorGuard(http.HandlerFunc(s.providerAuth)))
	w := mockRequest(h, "POST", "/api/provider-auth", `{"action":"login","connectionId":"device-fixture"}`)
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "application/x-ndjson") {
		t.Fatal("Browser sign-in must stream provider instructions, not launch a desktop Terminal", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"type":"instructions"`) || !strings.Contains(w.Body.String(), `"type":"complete"`) || strings.Contains(w.Body.String(), "synthetic-private-token") {
		t.Fatal("Missing instructions/completion or leaked raw output")
	}
}

func TestDeviceInstructionsAreBoundedAndProviderOnly(t *testing.T) {
	t.Parallel()
	cancelled := false
	ch := make(chan deviceInstructions, 1)
	out := &deviceInstructionOutput{instructions: ch, cancel: func() { cancelled = true }}
	for _, part := range []string{"https://attacker.invalid/\nABCD-1234\n", "\x1b[32mhttps://auth.openai.com/codex/de", "vice\x1b[0m\r\nABCD-1234\r\n"} {
		if _, err := io.WriteString(out, part); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case event := <-ch:
		if event.VerificationURL != codexDeviceURL || event.UserCode != "ABCD-1234" {
			t.Fatal(event)
		}
	default:
		t.Fatal("chunked device instructions missing")
	}
	io.WriteString(out, "ABCD-1234\n")
	select {
	case <-ch:
		t.Fatal("duplicate instructions")
	default:
	}
	if _, err := io.WriteString(out, strings.Repeat("x", 16385)); err == nil || !cancelled {
		t.Fatal("unbounded output")
	}
}
func TestDeviceLoginAPIRejectsPermissionFlagAndBusyWork(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	for _, mode := range []string{"nonowner", "disabled", "busy", "stop", "csrf", "worker-failure"} {
		t.Run(mode, func(t *testing.T) {
			s, db, _ := mockServer(t)
			calls := 0
			s.deviceLoginRun = func(context.Context, string, string, io.Reader, io.Writer) error {
				calls++
				return errors.New("synthetic-secret-detail")
			}
			db.seed("connection.created", room.Connection{ID: "device-fixture", Protocol: "CLI", Endpoint: "ssh://commonroom-server"})
			if mode == "disabled" {
				db.seed("feature.flag.updated", map[string]any{"name": "connectionSignIn", "enabled": false})
			}
			if mode == "nonowner" {
				s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
			}
			if mode == "busy" {
				s.active[verificationParticipant("", "device-fixture").ID] = true
			}
			if mode == "stop" {
				s.stopAll = true
			}
			h := s.auth.protect(s.actorGuard(http.HandlerFunc(s.providerAuth)))
			body := `{"action":"login","connectionId":"device-fixture"}`
			var w *httptest.ResponseRecorder
			if mode == "csrf" {
				r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/provider-auth", strings.NewReader(body))
				r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
				w = httptest.NewRecorder()
				h.ServeHTTP(w, r)
			} else {
				w = mockRequest(h, "POST", "/api/provider-auth", body)
			}
			want := 403
			if mode == "busy" || mode == "stop" {
				want = 409
			}
			if mode == "worker-failure" {
				want = 200
			}
			if w.Code != want {
				t.Fatal(mode, w.Code, w.Body.String())
			}
			if mode == "worker-failure" {
				if calls != 1 || !strings.Contains(w.Body.String(), `"type":"error"`) || strings.Contains(w.Body.String(), "synthetic-secret-detail") {
					t.Fatal("failure exposure", w.Body.String())
				}
			} else if calls != 0 {
				t.Fatal("unauthorized worker launch")
			}
		})
	}
}
func TestDeviceLoginDisconnectCancelsWorker(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "device-fixture", Protocol: "CLI", Endpoint: "ssh://commonroom-server"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopped := make(chan struct{})
	s.deviceLoginRun = func(ctx context.Context, _ string, _ string, _ io.Reader, out io.Writer) error {
		cancel()
		<-ctx.Done()
		close(stopped)
		return ctx.Err()
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/provider-auth", strings.NewReader(`{"action":"login","connectionId":"device-fixture"}`)).WithContext(ctx)
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	r.Header.Set("X-CSRF-Token", "csrf")
	s.auth.protect(s.actorGuard(http.HandlerFunc(s.providerAuth))).ServeHTTP(httptest.NewRecorder(), r)
	select {
	case <-stopped:
	default:
		t.Fatal("worker was not cancelled and joined")
	}
}
func TestDeviceLoginNeverAuditsInstructions(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "device-fixture", Protocol: "CLI", Endpoint: "ssh://commonroom-server"})
	s.deviceLoginRun = func(_ context.Context, _ string, _ string, _ io.Reader, out io.Writer) error {
		_, err := io.WriteString(out, "https://auth.openai.com/codex/device\nABCD-1234\n")
		return err
	}
	mockRequest(s.auth.protect(s.actorGuard(http.HandlerFunc(s.providerAuth))), "POST", "/api/provider-auth", `{"action":"login","connectionId":"device-fixture"}`)
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(st)
	if strings.Contains(string(raw), "ABCD-1234") {
		t.Fatal("device code entered persisted state")
	}
}

func TestDeviceLoginStartupTimeoutCancelsAndJoinsWorker(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	s, db, _ := mockServer(t)
	s.deviceLoginStartupTimeout = time.Millisecond
	db.seed("connection.created", room.Connection{ID: "device-fixture", Protocol: "CLI", Endpoint: "ssh://commonroom-server"})
	stopped := make(chan struct{})
	s.deviceLoginRun = func(ctx context.Context, _ string, _ string, _ io.Reader, _ io.Writer) error {
		<-ctx.Done()
		close(stopped)
		return ctx.Err()
	}
	w := mockRequest(s.auth.protect(s.actorGuard(http.HandlerFunc(s.providerAuth))), "POST", "/api/provider-auth", `{"action":"login","connectionId":"device-fixture"}`)
	if !strings.Contains(w.Body.String(), "within 45 seconds") || strings.Contains(w.Body.String(), `"type":"complete"`) {
		t.Fatal(w.Body.String())
	}
	select {
	case <-stopped:
	default:
		t.Fatal("startup timeout left worker running")
	}
}

func TestDeviceLoginDoesNotPublishPartialCode(t *testing.T) {
	t.Parallel()
	ch := make(chan deviceInstructions, 1)
	out := &deviceInstructionOutput{instructions: ch, cancel: func() {}}
	io.WriteString(out, "https://auth.openai.com/codex/device\nABCD-1234")
	select {
	case <-ch:
		t.Fatal("partial pipe chunk published as complete code")
	default:
	}
	io.WriteString(out, "5\n")
	select {
	case event := <-ch:
		if event.UserCode != "ABCD-12345" {
			t.Fatal("truncated device code")
		}
	default:
		t.Fatal("complete code not published")
	}
}
