package main

import (
	"agentroom/internal/room"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func localClaudeConnection(stConnectionProtocol, endpoint string) bool {
	return stConnectionProtocol == "CLI" && endpoint == "claude"
}
func shellArgument(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

type cappedStatusOutput struct{ bytes.Buffer }

func (b *cappedStatusOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 16384 {
		return 0, io.ErrShortBuffer
	}
	return b.Buffer.Write(p)
}

// providerAuth exposes owner-authorized connection login/status through the
// approved isolated SSH worker. Passwords and OAuth codes remain in Terminal.
// Disabling its feature flag cannot disable execution isolation or restore a
// host CLI fallback. Source forks CAN bypass this, BUT expose host credentials.
func connectionHasActiveAgent(st room.State, active map[string]bool, connection string) bool {
	if active[verificationParticipant("", connection).ID] {
		return true
	}
	for _, p := range st.Participants {
		if p.ConnectionID == connection && active[p.ID] {
			return true
		}
	}
	return false
}

func (s *server) providerAuth(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req struct {
		Action       string `json:"action"`
		ConnectionID string `json:"connectionId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Invalid request", 400)
		return
	}
	st, e := s.load(r.Context())
	if e != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	endpoint := ""
	for _, c := range st.Connections {
		if c.ID == req.ConnectionID && c.Protocol == "CLI" {
			endpoint = c.Endpoint
		}
	}
	if endpoint == "" {
		http.Error(w, "Choose a registered CLI connection", 400)
		return
	}
	if !connectionSignInEnabled(st.FeatureFlags) {
		http.Error(w, "Connection sign-in is disabled in Settings", 403)
		return
	}
	alias, provider, err := remoteWorkerEndpoint(endpoint, os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"), os.Getenv("COMMONROOM_REMOTE_WORKER"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.mu.Lock()
	stopped := s.agentsBlockedLocked(st, requestedRoom(r))
	s.mu.Unlock()
	if st.Incident || stopped {
		http.Error(w, "Provider operations blocked during an integrity incident", 409)
		return
	}
	encoded, e := encodeWorkerAuth(workerAuthRequest{Version: 1, Connection: req.ConnectionID, Provider: provider})
	if e != nil {
		http.Error(w, "Invalid connection identity", 400)
		return
	}
	switch req.Action {
	case "status":
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		cmd := workerSSHCommand(ctx, alias, "auth-status-v1", encoded)
		var out cappedStatusOutput
		cmd.Stdout = &out
		cmd.Stderr = io.Discard
		if cmd.Run() != nil {
			http.Error(w, "Worker unavailable or not configured; check remote worker setup", 503)
			return
		}
		present, recovery, statusErr := decodeWorkerCredentialStatus(out.Bytes())
		if statusErr != nil {
			http.Error(w, "Worker returned invalid status", 502)
			return
		}
		reply(w, map[string]any{"credentialPresent": present, "recoveryRequired": recovery, "provider": provider})
	case "logout":
		// Serialize against dispatch; already-running turns must finish or be stopped first.
		s.mu.Lock()
		defer s.mu.Unlock()
		current, err := s.load(r.Context())
		if err != nil {
			http.Error(w, "State unavailable", 503)
			return
		}
		if connectionHasActiveAgent(current, s.active, req.ConnectionID) {
			http.Error(w, "Stop or finish this connection's active agents before signing out", 409)
			return
		}
		if s.append(r.Context(), "provider.logout_requested", map[string]string{"actorId": ownerID, "connectionId": req.ConnectionID, "provider": provider}) != nil {
			http.Error(w, "Unable to record sign-out request", 500)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		cmd := workerSSHCommand(ctx, alias, "auth-logout-v1", encoded)
		var output cappedStatusOutput
		cmd.Stdout = &output
		cmd.Stderr = io.Discard
		if cmd.Run() != nil {
			http.Error(w, "Sign-out unconfirmed; worker unavailable or authentication busy. Check sign-in status before retrying", 503)
			return
		}
		var result struct {
			SignedOut bool `json:"signedOut"`
		}
		if json.Unmarshal(output.Bytes(), &result) != nil || !result.SignedOut {
			http.Error(w, "Worker returned invalid sign-out result", 502)
			return
		}
		if s.append(r.Context(), "provider.signed_out", map[string]string{"actorId": ownerID, "connectionId": req.ConnectionID, "provider": provider}) != nil {
			http.Error(w, "Credentials cleared but audit completion failed; check sign-in status", 500)
			return
		}
		reply(w, map[string]bool{"signedOut": true})
	case "login":
		if runtime.GOOS != "darwin" {
			http.Error(w, "Automatic Terminal launch requires macOS; use the documented worker sign-in command", 501)
			return
		}
		if e = os.MkdirAll(".data", 0700); e != nil {
			http.Error(w, "Unable to prepare sign-in", 500)
			return
		}
		script, e := os.CreateTemp(".data", "connection-login-*.command")
		if e != nil {
			http.Error(w, "Unable to prepare sign-in", 500)
			return
		}
		path, e := filepath.Abs(script.Name())
		if e != nil {
			script.Close()
			os.Remove(script.Name())
			http.Error(w, "Unable to prepare sign-in", 500)
			return
		}
		command := interactiveWorkerLogin(alias, encoded)
		if e = script.Chmod(0700); e == nil {
			_, e = script.WriteString("#!/bin/sh\ntrap 'rm -- \"$0\"' EXIT\n" + command + "\n")
		}
		closeErr := script.Close()
		if e != nil || closeErr != nil {
			os.Remove(path)
			http.Error(w, "Unable to prepare sign-in", 500)
			return
		}
		// Audit intent before launching; provider URLs, codes and credentials never enter the log.
		if e = s.append(r.Context(), "provider.login_launched", map[string]string{"actorId": ownerID, "connectionId": req.ConnectionID, "provider": provider}); e != nil {
			os.Remove(path)
			http.Error(w, "Unable to record sign-in request", 500)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if exec.CommandContext(ctx, "/usr/bin/open", "-a", "Terminal", path).Run() != nil {
			os.Remove(path)
			http.Error(w, "Unable to open sign-in Terminal", 502)
			return
		}
		reply(w, map[string]bool{"launched": true})
	default:
		http.Error(w, "Unsupported sign-in action", 400)
	}
}

// decodeWorkerCredentialStatus never treats uncertain custody as a usable login,
// including contradictory worker replies. Older workers omit the optional bit.
func decodeWorkerCredentialStatus(raw []byte) (bool, bool, error) {
	var status struct {
		CredentialPresent *bool `json:"credentialPresent"`
		RecoveryRequired  bool  `json:"recoveryRequired"`
	}
	if json.Unmarshal(raw, &status) != nil || status.CredentialPresent == nil {
		return false, false, errors.New("Worker returned invalid status")
	}
	return *status.CredentialPresent && !status.RecoveryRequired, status.RecoveryRequired, nil
}
