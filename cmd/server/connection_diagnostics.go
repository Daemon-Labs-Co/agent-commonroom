package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type connectionCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}
type connectionTestResult struct {
	Status                 string            `json:"status"`
	Checks                 []connectionCheck `json:"checks"`
	CheckedAt              time.Time         `json:"checkedAt"`
	ProviderResponseTested bool              `json:"providerResponseTested"`
}
type workerProbe func(context.Context, string, string, string) ([]byte, error)

func connectionTestingEnabled(flags map[string]bool) bool {
	enabled, ok := flags["connectionTesting"]
	return !ok || enabled
}

// probeWorker performs bounded, verified SSH diagnostics only. It never starts
// unrestricted commands, signs in, forwards credentials, or follows arbitrary URLs.
func probeWorker(ctx context.Context, alias, action, value string) ([]byte, error) {
	cmd := workerSSHCommand(ctx, alias, action, value)
	cmd.WaitDelay = time.Second
	var out, diagnostics cappedStatusOutput
	cmd.Stdout = &out
	cmd.Stderr = &diagnostics
	if cmd.Run() != nil {
		if ctx.Err() != nil {
			return nil, errors.New("Check timed out; verify the worker is reachable")
		}
		switch {
		case strings.Contains(diagnostics.String(), "config.json"):
			return nil, errors.New("Worker configuration is missing or not private; follow the worker setup guide")
		case strings.Contains(diagnostics.String(), "image is unavailable"):
			return nil, errors.New("Configured container image is unavailable on the worker")
		case strings.Contains(diagnostics.String(), "Rootless Docker"):
			return nil, errors.New("Rootless Docker or required resource limits are unavailable")
		}
		return nil, errors.New("Worker check failed; verify SSH access, trusted host keys and the installed worker")
	}
	return out.Bytes(), nil
}

// checkConnection distinguishes execution readiness from provider validation.
// Registration-only adapters never receive a misleading successful result.
func checkConnection(ctx context.Context, c room.Connection, alias, enabled string, probe workerProbe) connectionTestResult {
	return checkConnectionDraft(ctx, c, alias, enabled, probe, true)
}
func checkConnectionDraft(ctx context.Context, c room.Connection, alias, enabled string, probe workerProbe, registered bool) connectionTestResult {
	result := connectionTestResult{Status: "blocked", Checks: []connectionCheck{}, CheckedAt: time.Now().UTC()}
	add := func(name, status, detail string) {
		result.Checks = append(result.Checks, connectionCheck{name, status, detail})
	}
	if c.Protocol != "CLI" {
		result.Status = "unsupported"
		add("Adapter", "unsupported", "This connection method can be registered, but its execution adapter is not implemented. No network request was made.")
		return result
	}
	if !strings.HasPrefix(c.Endpoint, "ssh://") {
		add("Execution environment", "blocked", "Managed CLI execution needs an approved Linux/rootless-Docker worker. Use ssh://<approved-worker> or ssh://<approved-worker>/claude.")
		return result
	}
	worker, provider, err := remoteWorkerEndpoint(c.Endpoint, alias, enabled)
	if err != nil {
		add("Worker configuration", "blocked", err.Error())
		return result
	}
	add("Worker configuration", "passed", "Connection targets the approved isolated worker")
	raw, err := probe(ctx, worker, "ready-v1", "")
	if err != nil {
		add("Worker runtime", "failed", err.Error())
		return result
	}
	var ready struct {
		Version      int  `json:"version"`
		Configured   bool `json:"configured"`
		RuntimeReady bool `json:"runtimeReady"`
	}
	if json.Unmarshal(raw, &ready) != nil || ready.Version != 1 || !ready.Configured || !ready.RuntimeReady {
		add("Worker runtime", "failed", "Worker returned an invalid or unready runtime response")
		return result
	}
	add("Worker runtime", "passed", "SSH transport, rootless Docker, required limits and pinned image are ready")
	if !registered {
		result.Status = "needs-registration"
		add("Connection sign-in", "required", "Worker setup passed. Register or save this configuration, then sign in and test again to verify a provider response. No credentials were read and no model prompt was sent.")
		return result
	}
	envelope, err := encodeWorkerAuth(workerAuthRequest{Version: 1, Connection: c.ID, Provider: provider})
	if err != nil {
		add("Connection identity", "failed", "Connection identity is invalid")
		return result
	}
	raw, err = probe(ctx, worker, "auth-status-v1", envelope)
	if err != nil {
		add("Connection credentials", "failed", err.Error())
		return result
	}
	var auth struct {
		CredentialPresent *bool `json:"credentialPresent"`
	}
	if json.Unmarshal(raw, &auth) != nil || auth.CredentialPresent == nil {
		add("Connection credentials", "failed", "Worker returned invalid credential status")
		return result
	}
	if !*auth.CredentialPresent {
		result.Status = "needs-sign-in"
		add("Connection credentials", "required", "Use this connection's Sign in button, finish authentication and test again")
		return result
	}
	add("Connection credentials", "passed", "Private credentials are saved for this connection; account validity has not been checked with the provider")
	raw, err = probe(ctx, worker, "verify-v1", envelope)
	if err != nil {
		result.Status = "failed"
		add("Provider response", "failed", err.Error())
		return result
	}
	var verified struct {
		Version    int    `json:"version"`
		Connection string `json:"connection"`
		Provider   string `json:"provider"`
		Verified   *bool  `json:"verified"`
	}
	if json.Unmarshal(raw, &verified) != nil || verified.Version != 1 || verified.Connection != c.ID || verified.Provider != provider || verified.Verified == nil {
		result.Status = "failed"
		add("Provider response", "failed", "Worker returned an invalid provider verification result")
		return result
	}
	result.ProviderResponseTested = true
	if !*verified.Verified {
		result.Status = "failed"
		add("Provider response", "failed", "Provider verification failed. Check sign-in, account access and worker connectivity; setup checks alone do not establish a working connection.")
		return result
	}
	result.Status = "working"
	add("Provider response", "passed", "The isolated agent returned the expected response to a fixed test prompt. This verifies the connection at the time of this check.")
	return result
}

// connectionTest requires owner authorization, CSRF and incident clearance.
// Results are transient in the UI; only safe status metadata enters the audit.
func (s *server) connectionTest(w http.ResponseWriter, r *http.Request) {
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
		ConnectionID string           `json:"connectionId"`
		Draft        *connectionDraft `json:"draft,omitempty"`
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(&req) != nil || d.Decode(new(any)) != io.EOF {
		http.Error(w, "Invalid request", 400)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	if !connectionTestingEnabled(st.FeatureFlags) {
		http.Error(w, "Connection testing is disabled in Settings", 403)
		return
	}
	connection, registered, err := resolveConnectionTest(st.Connections, req.ConnectionID, req.Draft)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if connection == nil {
		http.Error(w, "Connection not found", 404)
		return
	}
	s.mu.Lock()
	blocked := s.agentsBlockedLocked(st, requestedRoom(r))
	s.mu.Unlock()
	if st.Incident || blocked {
		http.Error(w, "Connection checks blocked during an integrity incident or agent stop", 409)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
	defer cancel()
	result := checkConnectionDraft(ctx, *connection, os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"), os.Getenv("COMMONROOM_REMOTE_WORKER"), probeWorker, registered)
	if s.append(r.Context(), "connection.tested", map[string]string{"actorId": ownerID, "connectionId": connection.ID, "status": result.Status}) != nil {
		http.Error(w, "Unable to record connection check", 500)
		return
	}
	reply(w, result)
}

// Draft tests never persist edits or borrow another connection's credentials.
// Only a registered unchanged execution configuration can verify a model turn.
type connectionDraft struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Endpoint string `json:"endpoint"`
	Provider string `json:"provider"`
	Location string `json:"location"`
}

func resolveConnectionTest(connections []room.Connection, connectionID string, draft *connectionDraft) (*room.Connection, bool, error) {
	var saved *room.Connection
	for i := range connections {
		if connections[i].ID == connectionID {
			v := connections[i]
			saved = &v
			break
		}
	}
	if connectionID != "" && saved == nil {
		return nil, false, nil
	}
	if draft == nil {
		return saved, true, nil
	}
	endpoint := strings.TrimSpace(draft.Endpoint)
	if endpoint == "" || len(endpoint) > 2048 {
		return nil, false, errors.New("Enter a worker address or service URL before testing")
	}
	switch draft.Protocol {
	case "CLI", "HTTP", "WebSocket", "MCP":
	default:
		return nil, false, errors.New("Choose a supported connection method")
	}
	if draft.Location != "" && draft.Location != "Local" && draft.Location != "LAN" && draft.Location != "Remote/Internet" {
		return nil, false, errors.New("Invalid connection location")
	}
	v := room.Connection{ID: connectionID, Name: draft.Name, Protocol: draft.Protocol, Endpoint: endpoint, Provider: draft.Provider, Location: draft.Location}
	unchanged := saved != nil && saved.Endpoint == endpoint && saved.Protocol == draft.Protocol
	return &v, unchanged, nil
}
