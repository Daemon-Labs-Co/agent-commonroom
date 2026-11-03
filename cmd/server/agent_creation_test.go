package main

import (
	"agentroom/internal/room"
	"strings"
	"testing"
)

func TestLegacyAgentCreationRequiresConfiguredConnection(t *testing.T) {
	t.Parallel()
	_, db, h := governanceFixture(t, false)
	before := len(db.snapshot())
	w := mockRequest(h, "POST", "/api/agents", `{"action":"create"}`)
	if w.Code != 409 || !strings.Contains(w.Body.String(), "Add agent") || strings.Contains(w.Body.String(), "CLI unavailable") {
		t.Fatalf("legacy create must direct user to configured connection: %d %s", w.Code, w.Body.String())
	}
	if len(db.snapshot()) != before {
		t.Fatal("legacy shortcut created state")
	}
}

func TestAddRoomAgentUsesRegisteredWorkerConnection(t *testing.T) {
	t.Parallel()
	s, db, h := governanceFixture(t, false)
	db.seed("connection.created", room.Connection{ID: "registered-worker", Provider: "anthropic", Protocol: "CLI", Endpoint: "ssh://commonroom-server/claude", Name: "Claude"})
	w := mockRequest(h, "POST", "/api/events", `{"kind":"participant.created","payload":{"connectionId":"registered-worker","name":"Room builder","role":"Engineer","rank":1,"accept":false}}`)
	if w.Code != 200 {
		t.Fatalf("saved connection invitation failed: %d %s", w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range st.Participants {
		if p.Name == "Room builder" {
			found = true
			if p.ConnectionID != "registered-worker" || p.RoomID != "engineering" || p.Status != "Idle" || p.ThreadID != "" {
				t.Fatal("wrong saved agent identity", p)
			}
		}
	}
	if !found {
		t.Fatal("invitation did not save participant")
	}
	if len(st.Connections) != 2 {
		t.Fatal("invitation created an extra connection")
	}
}
