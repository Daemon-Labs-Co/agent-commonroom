package main

import (
	"strings"
	"testing"
)

func TestAuditSummary(t *testing.T) {
	t.Parallel()
	cases := []struct{ kind, payload, want string }{
		{"participant.created", `{"name":"Sofia","role":"Staff Engineer"}`, "Sofia joined a room as Staff Engineer"},
		{"message.created", `{"author":"Mark","text":"private discussion"}`, "Mark sent a room message"},
		{"connection.updated", `{"name":"Claude","endpoint":"https://secret:password@host"}`, "Connection settings updated: Claude"},
		{"agent.error", `{"error":"secret-token"}`, "Agent turn failed; check the agent connection and provider diagnostics"},
		{"future.event", `{}`, "Activity recorded: future event"},
		{"room.created", `broken`, "Room created: Room"},
	}
	for _, c := range cases {
		if got := auditSummary(c.kind, []byte(c.payload)); got != c.want {
			t.Errorf("%s: got %q, want %q", c.kind, got, c.want)
		}
	}
	got := auditSummary("room.created", []byte(`{"name":"a\nb"}`))
	if strings.Contains(got, "\n") {
		t.Fatal("control character not removed")
	}
}
