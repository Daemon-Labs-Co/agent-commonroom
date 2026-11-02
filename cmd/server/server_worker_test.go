package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"agentroom/internal/room"
)

func TestServerWorkerRoutingAndFlag(t *testing.T) {
	for _, flag := range []string{"", "false", "true"} {
		t.Setenv("COMMONROOM_SERVER_WORKER", flag)
		for _, suffix := range []string{"", "/claude"} {
			alias, provider, err := remoteWorkerEndpoint("ssh://commonroom-server"+suffix, "", "false")
			if flag != "true" {
				if err == nil {
					t.Fatal("disabled server worker admitted")
				}
				continue
			}
			if err != nil || alias != "commonroom-server" || (suffix == "/claude" && provider != "claude") {
				t.Fatal(alias, provider, err)
			}
			command := workerSSHCommand(context.Background(), alias, "ready-v1", "")
			if strings.Contains(command.Path, "ssh") || len(command.Args) != 3 || command.Args[1] != "--agent-worker" || command.Args[2] != "ready-v1" {
				t.Fatal(command.Args)
			}
		}
	}
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	if _, _, err := remoteWorkerEndpoint("ssh://commonroom-server/sh", "", "false"); err == nil {
		t.Fatal("arbitrary provider admitted")
	}
	if _, _, err := remoteWorkerEndpoint("codex", "", "false"); err == nil {
		t.Fatal("host command admitted")
	}
}
func TestServerWorkerConnectionProbe(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	calls := 0
	result := checkConnectionDraft(context.Background(), room.Connection{Protocol: "CLI", Endpoint: "ssh://commonroom-server"}, "", "false", func(_ context.Context, alias, action, value string) ([]byte, error) {
		calls++
		if alias != "commonroom-server" || action != "ready-v1" {
			t.Fatal(alias, action)
		}
		return nil, errors.New("synthetic missing worker configuration")
	}, false)
	if calls != 1 || result.Status != "blocked" || !strings.Contains(result.Checks[len(result.Checks)-1].Detail, "missing worker configuration") {
		t.Fatal(result)
	}
}

func TestServerWorkerCredentialsAndHistoricalBinding(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "false")
	if alias, _, err := boundWorkerEndpoint("ssh://commonroom-server", "commonroom-server"); err != nil || alias != "commonroom-server" {
		t.Fatal("disabling execution must preserve historical validation", err)
	}
	if _, _, err := boundWorkerEndpoint("ssh://other", "commonroom-server"); err == nil {
		t.Fatal("foreign historical binding admitted")
	}
	command := workerSSHCommand(context.Background(), "commonroom-server", "run-v2", "fixture")
	if command.Path != "" {
		t.Fatal("disabled worker returned runnable command")
	}
	env := workerProcessEnvironment([]string{"DATABASE_URL=synthetic", "MIGRATION_DATABASE_URL=synthetic", "PGPASSWORD=synthetic", "POSTGRES_PASS=synthetic", "PATH=/synthetic", "HOME=/synthetic"})
	if strings.Join(env, ",") != "PATH=/synthetic,HOME=/synthetic" {
		t.Fatal("controller database credentials leaked")
	}
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	if approvedWorkerEndpoint() != "ssh://commonroom-server" || configuredWorkerIdentity() != "commonroom-server" {
		t.Fatal("startup/recovery worker mismatch")
	}
	result := checkConnectionDraft(context.Background(), room.Connection{Protocol: "CLI", Endpoint: "ssh://commonroom-server/claude"}, "", "false", func(_ context.Context, alias, action, value string) ([]byte, error) {
		if alias != "commonroom-server" || action != "ready-v1" {
			t.Fatal(alias, action)
		}
		return []byte(`{"version":1,"lifecycleVersion":2,"configured":true,"runtimeReady":true}`), nil
	}, false)
	if result.Status != "needs-registration" {
		t.Fatal(result)
	}
	login := interactiveWorkerLogin("commonroom-server", "fixture")
	if strings.Contains(login, "ssh ") || !strings.Contains(login, "--agent-worker auth-login-v1") {
		t.Fatal(login)
	}
}

func TestServerWorkerBrowserDraftThroughAuthenticatedAPI(t *testing.T) {
	t.Setenv("COMMONROOM_SERVER_WORKER", "true")
	s, db, _ := mockServer(t)
	calls := 0
	s.connectionProbe = func(_ context.Context, alias, action, value string) ([]byte, error) {
		calls++
		if alias != "commonroom-server" || action != "ready-v1" {
			t.Fatal(alias, action)
		}
		return []byte(`{"version":1,"lifecycleVersion":2,"configured":true,"runtimeReady":true}`), nil
	}
	h := s.auth.protect(s.actorGuard(http.HandlerFunc(s.connectionTest)))
	body := `{"draft":{"name":"Fixture","provider":"openai","protocol":"CLI","location":"Local","endpoint":"ssh://commonroom-server"}}`
	w := mockRequest(h, "POST", "/api/connections/test", body)
	var result connectionTestResult
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Status != "needs-registration" || calls != 1 {
		t.Fatal(w.Code, w.Body.String(), calls)
	}
	t.Setenv("COMMONROOM_SERVER_WORKER", "false")
	w = mockRequest(h, "POST", "/api/connections/test", body)
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Status != "blocked" || calls != 1 {
		t.Fatal("disabled worker probed", w.Body.String())
	}
	db.seed("feature.flag.updated", map[string]any{"name": "connectionTesting", "enabled": false})
	w = mockRequest(h, "POST", "/api/connections/test", body)
	if w.Code != 403 || calls != 1 {
		t.Fatal("feature flag bypass")
	}
}
