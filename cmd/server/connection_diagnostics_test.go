package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestConnectionReadinessStages(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, protocol, endpoint, ready, auth, want string
		failedAction                                string
		calls                                       int
	}{
		{"ready", "CLI", "ssh://worker", `{"version":1,"configured":true,"runtimeReady":true}`, `{"credentialPresent":true}`, "working", "", 3},

		{"provider denied", "CLI", "ssh://worker", `{"version":1,"configured":true,"runtimeReady":true}`, `{"credentialPresent":true}`, "failed", "", 3},
		{"provider unavailable", "CLI", "ssh://worker", `{"version":1,"configured":true,"runtimeReady":true}`, `{"credentialPresent":true}`, "failed", "verify-v1", 3},
		{"invalid verification", "CLI", "ssh://worker", `{"version":1,"configured":true,"runtimeReady":true}`, `{"credentialPresent":true}`, "failed", "", 3},
		{"signed out", "CLI", "ssh://worker/claude", `{"version":1,"configured":true,"runtimeReady":true}`, `{"credentialPresent":false}`, "needs-sign-in", "", 2},
		{"unreachable", "CLI", "ssh://worker", "", "", "blocked", "ready-v1", 1},
		{"invalid runtime", "CLI", "ssh://worker", `{"runtimeReady":true}`, "", "blocked", "", 1},
		{"invalid auth", "CLI", "ssh://worker", `{"version":1,"configured":true,"runtimeReady":true}`, `{}`, "blocked", "", 2},
		{"credentials unreachable", "CLI", "ssh://worker", `{"version":1,"configured":true,"runtimeReady":true}`, "", "blocked", "auth-status-v1", 2},
		{"unapproved SSH", "CLI", "ssh://elsewhere", "", "", "blocked", "", 0},
		{"direct host denied", "CLI", "codex", "", "", "blocked", "", 0},
		{"HTTP registration", "HTTP", "http://127.0.0.1/admin", "", "", "unsupported", "", 0},
		{"MCP registration", "MCP", "http://169.254.169.254/", "", "", "unsupported", "", 0},
		{"WS registration", "WebSocket", "ws://127.0.0.1/", "", "", "unsupported", "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			probe := func(_ context.Context, alias, action, envelope string) ([]byte, error) {
				calls++
				if alias != "worker" {
					t.Fatal("unapproved alias")
				}
				if action == tc.failedAction {
					return nil, errors.New("fixture unavailable")
				}
				if action == "ready-v1" {
					return []byte(tc.ready), nil
				}
				v, err := decodeWorkerAuth(envelope)
				if err != nil || v.Connection != "connection-a" {
					t.Fatal("wrong credential scope", v, err)
				}
				if tc.endpoint == "ssh://worker/claude" && v.Provider != "claude" {
					t.Fatal("wrong provider")
				}

				if action == "verify-v1" {
					if tc.name == "invalid verification" {
						return []byte(`{"verified":true}`), nil
					}
					raw, _ := json.Marshal(map[string]any{"version": 1, "connection": v.Connection, "provider": v.Provider, "verified": tc.name != "provider denied"})
					return raw, nil
				}
				return []byte(tc.auth), nil
			}
			got := checkConnection(context.Background(), room.Connection{ID: "connection-a", Protocol: tc.protocol, Endpoint: tc.endpoint}, "worker", "true", probe)
			if got.Status != tc.want || calls != tc.calls || got.ProviderResponseTested != (tc.name == "ready" || tc.name == "provider denied") {
				t.Fatal(got, calls)
			}
		})
	}
}
func TestConnectionChecksDisabledTransport(t *testing.T) {
	t.Parallel()
	got := checkConnection(context.Background(), room.Connection{ID: "a", Protocol: "CLI", Endpoint: "ssh://worker"}, "worker", "false", func(context.Context, string, string, string) ([]byte, error) {
		t.Fatal("disabled transport contacted worker")
		return nil, nil
	})
	if got.Status != "blocked" {
		t.Fatal(got)
	}
}
func TestConnectionTestAPIControls(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	h := s.auth.protect(s.actorGuard(http.HandlerFunc(s.connectionTest)))
	db.seed("connection.created", room.Connection{ID: "http-fixture", Protocol: "HTTP", Endpoint: "http://invalid.example", Name: "Fixture"})
	w := mockRequest(h, "POST", "/api/connections/test", `{"connectionId":"http-fixture"}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result connectionTestResult
	if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Status != "unsupported" {
		t.Fatal(w.Body.String())
	}
	w = mockRequest(h, "POST", "/api/connections/test", `{"connectionId":"missing"}`)
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
	w = mockRequest(h, "POST", "/api/connections/test", `{"connectionId":"http-fixture","endpoint":"ssh://attacker"}`)
	if w.Code != 400 {
		t.Fatal("endpoint override accepted", w.Code)
	}
	db.seed("feature.flag.updated", map[string]any{"name": "connectionTesting", "enabled": false})
	w = mockRequest(h, "POST", "/api/connections/test", `{"connectionId":"http-fixture"}`)
	if w.Code != 403 {
		t.Fatal("disabled check accepted", w.Code)
	}
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	w = mockRequest(h, "POST", "/api/connections/test", `{"connectionId":"http-fixture"}`)
	if w.Code != 403 {
		t.Fatal("non-owner check accepted", w.Code)
	}
	if !connectionTestingEnabled(nil) || connectionTestingEnabled(map[string]bool{"connectionTesting": false}) || !connectionTestingEnabled(map[string]bool{"connectionTesting": true}) {
		t.Fatal("flag defaults")
	}
}

func TestDraftConnectionScope(t *testing.T) {
	t.Parallel()
	connections := []room.Connection{{ID: "saved", Name: "Saved", Protocol: "CLI", Endpoint: "ssh://worker", Provider: "openai"}}
	for _, tc := range []struct {
		name, id, protocol, endpoint string
		registered                   bool
	}{
		{"new", "", "CLI", "ssh://worker", false}, {"unchanged", "saved", "CLI", "ssh://worker", true}, {"provider changed", "saved", "CLI", "ssh://worker/claude", false}, {"method changed", "saved", "HTTP", "https://service.example", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			draft := connectionDraft{Name: "Unsaved", Protocol: tc.protocol, Endpoint: tc.endpoint}
			got, registered, err := resolveConnectionTest(connections, tc.id, &draft)
			if err != nil || got == nil || registered != tc.registered || got.Name != "Unsaved" {
				t.Fatal(got, registered, err)
			}
			if connections[0].Name != "Saved" || connections[0].Endpoint != "ssh://worker" {
				t.Fatal("draft changed stored connection")
			}
		})
	}
	if got, _, _ := resolveConnectionTest(connections, "missing", &connectionDraft{Protocol: "CLI", Endpoint: "ssh://worker"}); got != nil {
		t.Fatal("unknown connection borrowed credentials")
	}
	for _, draft := range []connectionDraft{{Protocol: "CLI", Endpoint: " "}, {Protocol: "shell", Endpoint: "ssh://worker"}, {Protocol: "CLI", Endpoint: "ssh://worker", Location: "invalid"}} {
		if _, _, err := resolveConnectionTest(connections, "", &draft); err == nil {
			t.Fatal("invalid draft accepted")
		}
	}
}
func TestDraftRuntimeNeverReadsCredentials(t *testing.T) {
	t.Parallel()
	calls := 0
	result := checkConnectionDraft(context.Background(), room.Connection{Protocol: "CLI", Endpoint: "ssh://worker"}, "worker", "true", func(_ context.Context, _ string, action, _ string) ([]byte, error) {
		calls++
		if action != "ready-v1" {
			t.Fatal("draft read credentials or launched model", action)
		}
		return []byte(`{"version":1,"configured":true,"runtimeReady":true}`), nil
	}, false)
	if calls != 1 || result.Status != "needs-registration" || result.ProviderResponseTested {
		t.Fatal(result, calls)
	}
}
func TestDraftAPILeavesConnectionUnchanged(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	h := s.auth.protect(s.actorGuard(http.HandlerFunc(s.connectionTest)))
	db.seed("connection.created", room.Connection{ID: "saved", Name: "Before", Protocol: "HTTP", Endpoint: "https://old.example"})
	for _, body := range []string{`{"draft":{"name":"New","protocol":"HTTP","endpoint":"https://new.example","location":"LAN"}}`, `{"connectionId":"saved","draft":{"name":"Edited","protocol":"HTTP","endpoint":"https://edited.example"}}`} {
		w := mockRequest(h, "POST", "/api/connections/test", body)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	st, err := s.load(context.Background())
	if err != nil || len(st.Connections) != 1 || st.Connections[0].Name != "Before" || st.Connections[0].Endpoint != "https://old.example" {
		t.Fatal("testing persisted draft", st.Connections, err)
	}
	w := mockRequest(h, "POST", "/api/connections/test", `{"draft":{"id":"saved","protocol":"CLI","endpoint":"ssh://worker"}}`)
	if w.Code != 400 {
		t.Fatal("draft credential identity override accepted", w.Code)
	}
}
