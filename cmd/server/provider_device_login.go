package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

const codexDeviceURL = "https://auth.openai.com/codex/device"

var deviceANSI = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

// Strip OSC terminal hyperlinks before extracting the visible authorization URL.
var deviceOSC = regexp.MustCompile(`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)
var deviceCode = regexp.MustCompile(`(?m)^[ \t]*([A-Z0-9]{4,8}-[A-Z0-9]{4,8})[ \t]*\r?$`)

type deviceInstructions struct {
	Type            string `json:"type"`
	VerificationURL string `json:"verificationUrl"`
	UserCode        string `json:"userCode,omitempty"`
	LoginID         string `json:"loginId,omitempty"`
}

// Only fixed-provider device instructions leave this bounded private collector.
// Raw provider stdout/stderr, tokens and credential exports are never forwarded.
type deviceInstructionOutput struct {
	mu           sync.Mutex
	text         string
	provider     string
	loginID      string
	emitted      bool
	instructions chan deviceInstructions
	cancel       context.CancelFunc
}

func (o *deviceInstructionOutput) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.text)+len(p) > 16384 {
		o.cancel()
		return 0, io.ErrShortBuffer
	}
	o.text += string(p)
	text := deviceANSI.ReplaceAllString(deviceOSC.ReplaceAllString(o.text, ""), "")
	// A pipe write may end partway through a code. Never publish that prefix.
	boundary := strings.LastIndex(text, "\n")
	if boundary < 0 {
		return len(p), nil
	}
	text = text[:boundary+1]
	if o.provider == "claude" {
		for _, candidate := range authorizationURL.FindAllString(text, -1) {
			if validClaudeAuthorizationURL(candidate) && !o.emitted {
				o.emitted = true
				o.instructions <- deviceInstructions{Type: "authorization", VerificationURL: candidate, LoginID: o.loginID}
				break
			}
		}
		return len(p), nil
	}
	code := deviceCode.FindStringSubmatch(text)
	if !o.emitted && strings.Contains(text, codexDeviceURL) && len(code) == 2 {
		o.emitted = true
		o.instructions <- deviceInstructions{Type: "instructions", VerificationURL: codexDeviceURL, UserCode: code[1]}
	}
	return len(p), nil
}
func runDeviceLogin(ctx context.Context, alias, encoded string, in io.Reader, out io.Writer) error {

	command := workerSSHCommand(ctx, alias, "auth-login-v1", encoded)
	command.Stdin = in
	command.Stdout = out
	command.Stderr = out
	// Give the constrained worker time to remove its container and temporary
	// volumes on cancellation. Never substitute a direct provider invocation.
	command.Cancel = func() error { return command.Process.Signal(syscall.SIGTERM) }
	command.WaitDelay = 30 * time.Second
	return command.Run()
}
func (s *server) providerDeviceLogin(w http.ResponseWriter, r *http.Request, connection, alias, provider, encoded string) {

	s.mu.Lock()
	current, err := s.load(r.Context())
	if err != nil {
		s.mu.Unlock()
		http.Error(w, "State unavailable", 503)
		return
	}
	if connectionHasActiveAgent(current, s.active, connection) {
		s.mu.Unlock()
		http.Error(w, "Finish active agents or reconcile unresolved connection work before signing in", 409)
		return
	}
	if s.providerLoginSessions == nil {
		s.providerLoginSessions = map[string]*providerLoginSession{}
	}
	if s.providerLoginSessions[connection] != nil {
		s.mu.Unlock()
		http.Error(w, "A sign-in is already running for this connection", 409)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Minute)
	var input *os.File
	session := &providerLoginSession{ID: id(), Alias: alias, Provider: provider, Context: ctx, Cancel: cancel}
	if provider == "claude" {
		input, session.Input, err = os.Pipe()
		if err != nil {
			cancel()
			s.mu.Unlock()
			http.Error(w, "Unable to prepare Claude sign-in", 503)
			return
		}
	}
	defer func() {
		cancel()
		if input != nil {
			input.Close()
		}
		s.mu.Lock()
		if session.Input != nil {
			session.Input.Close()
		}
		if s.providerLoginSessions[connection] == session {
			delete(s.providerLoginSessions, connection)
		}
		s.mu.Unlock()
	}()
	if err = s.append(r.Context(), "provider.login_launched", map[string]string{"actorId": ownerID, "connectionId": connection, "provider": provider}); err != nil {
		s.mu.Unlock()
		http.Error(w, "Unable to record sign-in request", 500)
		return
	}
	s.providerLoginSessions[connection] = session
	s.mu.Unlock()
	// This one bounded response is long lived. Other HTTP responses retain their
	// normal deadlines. Heartbeats prevent reverse-proxy idle disconnection.
	controller := http.NewResponseController(w)
	if err = controller.SetWriteDeadline(time.Now().Add(16 * time.Minute)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		http.Error(w, "Unable to start sign-in response", 503)
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	send := func(event any) bool {
		if json.NewEncoder(w).Encode(event) != nil {
			cancel()
			return false
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return true
	}
	if !send(map[string]string{"type": "starting"}) {
		return
	}
	instructions := make(chan deviceInstructions, 1)
	output := &deviceInstructionOutput{instructions: instructions, cancel: cancel, provider: provider, loginID: session.ID}
	run := s.deviceLoginRun
	if run == nil {
		run = runDeviceLogin
	}
	result := make(chan error, 1)
	done := make(chan struct{})
	var reader io.Reader
	if input != nil {
		reader = input
	}
	go func() { defer close(done); result <- run(ctx, alias, encoded, reader, output) }()
	defer func() {
		cancel()
		s.mu.Lock()
		if session.Input != nil {
			session.Input.Close()
		}
		s.mu.Unlock()
		select {
		case <-done:
		case <-time.After(35 * time.Second):
		}
	}()
	startupTimeout := s.deviceLoginStartupTimeout
	if startupTimeout <= 0 {
		startupTimeout = 45 * time.Second
	}
	startupTimer := time.NewTimer(startupTimeout)
	defer startupTimer.Stop()
	startupDeadline := startupTimer.C
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case event := <-instructions:
			startupTimer.Stop()
			startupDeadline = nil
			if event.Type == "authorization" {
				s.mu.Lock()
				session.Ready = true
				s.mu.Unlock()
			}
			if !send(event) {
				return
			}
		case err := <-result:
			// A fast worker may exit before the instructions event is consumed.
			select {
			case event := <-instructions:
				if !send(event) {
					return
				}
			default:
			}
			if err != nil {
				send(map[string]string{"type": "error", "message": "Sign-in did not complete. Check the worker configuration and try again; an expired authorization requires a new sign-in."})
				return
			}
			send(map[string]string{"type": "complete"})
			return
		case <-ticker.C:
			if !send(map[string]string{"type": "waiting"}) {
				return
			}
		case <-startupDeadline:
			cancel()
			send(map[string]string{"type": "error", "message": "The worker did not return provider sign-in instructions within 45 seconds. Sign-in was cancelled; check the worker and try again."})
			return
		case <-ctx.Done():
			send(map[string]string{"type": "error", "message": "Sign-in cancelled or timed out. A new sign-in is required."})
			return
		}
	}
}

var authorizationURL = regexp.MustCompile(`https://[^\s\x1b<>"']+`)

func validClaudeAuthorizationURL(raw string) bool {
	if len(raw) > 4096 {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" {
		return false
	}
	switch {
	case u.Host == "claude.com" && u.Path == "/cai/oauth/authorize":
	case u.Path == "/oauth/authorize" && (u.Host == "claude.ai" || u.Host == "console.anthropic.com" || u.Host == "platform.claude.com"):
	default:
		return false
	}
	return u.Query().Get("response_type") == "code" && u.Query().Get("code_challenge") != ""
}
