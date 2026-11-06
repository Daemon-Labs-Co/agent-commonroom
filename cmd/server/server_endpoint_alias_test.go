package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"agentroom/internal/room"
)

func setWorkerEnv(t *testing.T, server, remote, alias, via string) {
	t.Helper()
	t.Setenv("COMMONROOM_SERVER_WORKER", server)
	t.Setenv("COMMONROOM_REMOTE_WORKER", remote)
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", alias)
	t.Setenv("COMMONROOM_SERVER_ENDPOINT_VIA_ALIAS", via)
}

func TestServerEndpointStaysRefusedUnlessTheOperatorDeclaresTheAlias(t *testing.T) {
	for _, via := range []string{"", "false"} {
		setWorkerEnv(t, "false", "true", "cr-space", via)
		for _, endpoint := range []string{"ssh://commonroom-server", "ssh://commonroom-server/claude"} {
			if alias, _, err := remoteWorkerEndpoint(endpoint, "cr-space", "true"); err == nil {
				t.Fatalf("via=%q admitted %s as %q", via, endpoint, alias)
			}
		}
	}
}

func TestServerEndpointRoutesThroughTheDeclaredAlias(t *testing.T) {
	setWorkerEnv(t, "false", "true", "cr-space", "true")
	for endpoint, want := range map[string]string{"ssh://commonroom-server": "codex", "ssh://commonroom-server/claude": "claude"} {
		alias, provider, err := remoteWorkerEndpoint(endpoint, "cr-space", "true")
		if err != nil || alias != "cr-space" || provider != want {
			t.Fatalf("%s: %q %q %v", endpoint, alias, provider, err)
		}
		command := workerSSHCommand(context.Background(), alias, "ready-v1", "")
		if !strings.HasSuffix(command.Path, "ssh") && command.Args[0] != "ssh" {
			t.Fatalf("expected the SSH transport, got %v", command.Args)
		}
		joined := strings.Join(command.Args, " ")
		if !strings.Contains(joined, "BatchMode=yes") || !strings.Contains(joined, "StrictHostKeyChecking=yes") || !strings.Contains(joined, "cr-space") || strings.Contains(joined, "--agent-worker") {
			t.Fatalf("not the verified remote transport: %v", command.Args)
		}
	}
}

func TestServerEndpointAliasNeverEnablesATransportByItself(t *testing.T) {
	cases := []struct {
		name                           string
		server, remote, alias, enabled string
	}{
		{"remote worker disabled", "false", "false", "cr-space", "false"},
		{"missing alias", "false", "true", "", "true"},
		{"invalid alias", "false", "true", "bad alias;id", "true"},
	}
	for _, tc := range cases {
		setWorkerEnv(t, tc.server, tc.remote, tc.alias, "true")
		if alias, _, err := remoteWorkerEndpoint("ssh://commonroom-server", tc.alias, tc.enabled); err == nil {
			t.Errorf("%s: admitted as %q", tc.name, alias)
		}
	}
	setWorkerEnv(t, "false", "true", "cr-space", "true")
	for _, endpoint := range []string{"ssh://commonroom-server/sh", "ssh://commonroom-server/", "ssh://commonroom-server/claude/x", "ssh://commonroom-server:22", "codex"} {
		if _, _, err := remoteWorkerEndpoint(endpoint, "cr-space", "true"); err == nil {
			t.Errorf("admitted %q", endpoint)
		}
	}
}

func TestServerWorkerTransportTakesPrecedenceOverTheAliasSetting(t *testing.T) {
	setWorkerEnv(t, "true", "false", "", "true")
	alias, _, err := remoteWorkerEndpoint("ssh://commonroom-server", "", "false")
	if err != nil || alias != "commonroom-server" {
		t.Fatal(alias, err)
	}
}

func TestServerEndpointAliasStartupValidation(t *testing.T) {
	cases := []struct {
		name                       string
		server, remote, alias, via string
		wantErr                    bool
	}{
		{"unset", "", "", "", "", false},
		{"explicitly off", "false", "false", "", "false", false},
		{"valid", "false", "true", "cr-space", "true", false},
		{"typo value", "false", "true", "cr-space", "yes", true},
		{"with server worker", "true", "false", "", "true", true},
		{"without remote worker", "false", "false", "cr-space", "true", true},
		{"without alias", "false", "true", "", "true", true},
		{"invalid alias", "false", "true", "-bad", "true", true},
	}
	for _, tc := range cases {
		setWorkerEnv(t, tc.server, tc.remote, tc.alias, tc.via)
		if err := serverEndpointAliasConfigError(); (err != nil) != tc.wantErr {
			t.Errorf("%s: err=%v", tc.name, err)
		}
	}
}

func TestServerEndpointConnectionProbeUsesTheAlias(t *testing.T) {
	setWorkerEnv(t, "false", "true", "cr-space", "true")
	calls := 0
	result := checkConnectionDraft(context.Background(), room.Connection{Protocol: "CLI", Endpoint: "ssh://commonroom-server"}, "cr-space", "true", func(_ context.Context, alias, action, value string) ([]byte, error) {
		calls++
		if alias != "cr-space" || action != "ready-v1" {
			t.Fatal(alias, action)
		}
		return nil, errors.New("synthetic unavailable worker")
	}, false)
	if calls != 1 || result.Status != "blocked" {
		t.Fatal(calls, result)
	}
}

func TestServerEndpointHistoricalBindingFollowsTheAlias(t *testing.T) {
	setWorkerEnv(t, "false", "true", "cr-space", "true")
	if alias, provider, err := boundWorkerEndpoint("ssh://commonroom-server/claude", "cr-space"); err != nil || alias != "cr-space" || provider != "claude" {
		t.Fatal(alias, provider, err)
	}
	setWorkerEnv(t, "false", "true", "cr-space", "false")
	if _, _, err := boundWorkerEndpoint("ssh://commonroom-server", "cr-space"); err == nil {
		t.Fatal("binding admitted without the setting")
	}
}
