package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func eventsOfKind(db *mockDB, kind string) []mockEvent {
	var out []mockEvent
	for _, e := range db.snapshot() {
		if e.kind == kind {
			out = append(out, e)
		}
	}
	return out
}

func membershipRecords(t *testing.T, db *mockDB) []membershipRecord {
	t.Helper()
	var out []membershipRecord
	for _, e := range eventsOfKind(db, "room.membership") {
		var r membershipRecord
		if err := json.Unmarshal([]byte(e.payload), &r); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

// No client operation may edit, delete, hide or rewrite room messages or
// membership records. Unknown kinds are refused as unsupported.
func TestNoEventKindEditsOrDeletesRoomLog(t *testing.T) {
	t.Parallel()
	_, db, h := mockServer(t)
	db.seed("message.created", room.Message{ID: "m", RoomID: "engineering", Text: "keep"})
	before := len(db.snapshot())
	for _, kind := range []string{"message.deleted", "message.updated", "message.edited", "message.redacted", "message.hidden", "room.messages.cleared", "room.membership.deleted", "room.membership", "participant.deleted", "room.deleted", "events.purged"} {
		if w := mockRequest(h, "POST", "/api/events", `{"kind":"`+kind+`","payload":{"id":"m"}}`); w.Code != 400 {
			t.Fatal("log-altering kind accepted", kind, w.Code)
		}
	}
	if len(db.snapshot()) != before {
		t.Fatal("rejected operations wrote events")
	}
}

func TestAddingAnAgentIsLoggedWithActorAndRoom(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "c", Provider: "openai", Name: "ChatGPT via Codex", Protocol: "CLI", Endpoint: "codex"})
	w := mockRequest(h, "POST", "/api/events", `{"kind":"participant.created","payload":{"connectionId":"c","name":"Agent","role":"Developer","instructions":"x","rank":3}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	recs := membershipRecords(t, db)
	if len(recs) != 1 || recs[0].Action != "added" || recs[0].Name != "Agent" || recs[0].RoomID != "engineering" || recs[0].ActorID != ownerID || recs[0].ParticipantID == "" {
		t.Fatalf("add not logged completely: %+v", recs)
	}
	if st, err := s.load(t.Context()); err != nil || len(st.Participants) != 1 {
		t.Fatal("agent not added", err)
	}
}

func TestRemovingAnAgentIsLoggedAndHistoryRetained(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("participant.created", room.Participant{ID: "p", Name: "Alex", RoomID: "engineering"})
	db.seed("message.created", room.Message{ID: "m", RoomID: "engineering", Text: "Retain me"})
	w := mockRequest(http.HandlerFunc(s.agents), "POST", "/api/agents", `{"action":"remove","participantId":"p"}`)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	recs := membershipRecords(t, db)
	if len(recs) != 1 || recs[0].Action != "removed" || recs[0].ParticipantID != "p" || recs[0].RoomID != "engineering" || recs[0].Name != "Alex" {
		t.Fatalf("removal not logged: %+v", recs)
	}
	removed := eventsOfKind(db, "participant.removed")
	var detail struct{ RoomID, ActorID string }
	if len(removed) != 1 || json.Unmarshal([]byte(removed[0].payload), &detail) != nil || detail.RoomID != "engineering" {
		t.Fatal("removal event lacks the room")
	}
	if st, _ := s.load(context.Background()); len(st.Messages) != 1 {
		t.Fatal("message lost on removal")
	}
}

func TestDeletingAConnectionLogsEachAgentItRemoves(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "c", Provider: "openai", Name: "Codex", Protocol: "CLI", Endpoint: "codex"})
	db.seed("participant.created", room.Participant{ID: "p1", Name: "One", RoomID: "engineering", ConnectionID: "c"})
	db.seed("participant.created", room.Participant{ID: "p2", Name: "Two", RoomID: "engineering", ConnectionID: "c"})
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"connection.deleted","payload":{"id":"c"}}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	recs := membershipRecords(t, db)
	if len(recs) != 2 || recs[0].Action != "removed" || recs[0].Reason != "connection removed" || recs[0].ActorID != ownerID {
		t.Fatalf("implicit removals not logged: %+v", recs)
	}
	if st, _ := s.load(t.Context()); len(st.Participants) != 0 {
		t.Fatal("agents kept after connection deletion")
	}
}

// Replaying the log after every kind of operation never loses a message or
// changes its text: history only grows.
func TestMessagesNeverShrinkAcrossOperations(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "c", Provider: "openai", Name: "Codex", Protocol: "CLI", Endpoint: "codex"})
	db.seed("participant.created", room.Participant{ID: "p", Name: "A", RoomID: "engineering", ConnectionID: "c", ThreadID: "t"})
	steps := []func(){
		func() { mockRequest(h, "POST", "/api/events", `{"kind":"message.created","payload":{"text":"first"}}`) },
		func() { mockRequest(h, "POST", "/api/events", clearBody) },
		func() {
			mockRequest(h, "POST", "/api/events", `{"kind":"room.context_export.updated","payload":{"policy":{"providers":[]}}}`)
		},
		func() {
			mockRequest(http.HandlerFunc(s.agents), "POST", "/api/agents", `{"action":"remove","participantId":"p"}`)
		},
		func() { mockRequest(h, "POST", "/api/events", `{"kind":"connection.deleted","payload":{"id":"c"}}`) },
		func() {
			mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"idleStopNotices","enabled":false}}`)
		},
		func() {
			mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"contextClear","enabled":false}}`)
		},
		func() {
			mockRequest(h, "POST", "/api/events", `{"kind":"message.created","payload":{"text":"second"}}`)
		},
	}
	seen := map[string]string{}
	for i, step := range steps {
		step()
		st, err := s.load(t.Context())
		if err != nil {
			t.Fatal(i, err)
		}
		now := map[string]string{}
		for _, m := range st.Messages {
			now[m.ID] = m.Text
		}
		for id, text := range seen {
			if now[id] != text {
				t.Fatalf("step %d lost or altered message %s", i, id)
			}
		}
		seen = now
	}
	if len(seen) < 3 {
		t.Fatal("expected two messages and a clear notice", len(seen))
	}
}
