package main

import (
	"agentroom/internal/room"
	"strings"
	"testing"
)

const clearBody = `{"kind":"room.context.cleared","payload":{"reason":"Starting a new topic"}}`

func asHuman(s *server, db *mockDB, actorID string) {
	db.seed("actor.revised", room.Actor{ID: actorID, Name: "Dev " + actorID, Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	session := s.auth.sessions["owner"]
	session.ActorID = actorID
	s.auth.sessions["owner"] = session
}

func TestClearContextRecordsBoundaryNoticeAndResetsSessions(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("participant.created", room.Participant{ID: "p", RoomID: "engineering", ConnectionID: "c", ThreadID: "old-session", Name: "Agent"})
	db.seed("message.created", room.Message{ID: "old", RoomID: "engineering", ActorType: "human", Author: "You", Text: "historic detail"})
	if w := mockRequest(h, "POST", "/api/events", clearBody); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if st.Participants[0].ThreadID != "" {
		t.Fatal("native session survived a clear")
	}
	if roomContextBoundary(st, "engineering") == "" {
		t.Fatal("boundary not recorded")
	}
	last := st.Messages[len(st.Messages)-1]
	if last.ID != roomContextBoundary(st, "engineering") || !strings.Contains(last.Text, "Starting a new topic") || last.ActorType != "human" {
		t.Fatalf("reason not visible to the room: %+v", last)
	}
	if len(st.Messages) != 2 {
		t.Fatal("history must be preserved, not deleted:", len(st.Messages))
	}
}

func TestClearContextRequiresReason(t *testing.T) {
	t.Parallel()
	_, _, h := mockServer(t)
	for _, body := range []string{
		`{"kind":"room.context.cleared","payload":{}}`,
		`{"kind":"room.context.cleared","payload":{"reason":"   "}}`,
		`{"kind":"room.context.cleared","payload":{"reason":"` + strings.Repeat("x", maxContextClearReasonRunes+1) + `"}}`,
	} {
		if w := mockRequest(h, "POST", "/api/events", body); w.Code != 400 {
			t.Fatal("accepted invalid reason", w.Code, w.Body.String())
		}
	}
}

func TestClearContextAllowedForRoomMembersNotAgentsOrOutsiders(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	asHuman(s, db, "dev")
	if w := mockRequest(h, "POST", "/api/events", clearBody); w.Code != 200 {
		t.Fatal("room member refused", w.Code, w.Body.String())
	}
	db.seed("actor.revised", room.Actor{ID: "bot", Name: "Bot", Kind: "agent", Enabled: true, RoomIDs: []string{"engineering"}})
	session := s.auth.sessions["owner"]
	session.ActorID = "bot"
	s.auth.sessions["owner"] = session
	if w := mockRequest(h, "POST", "/api/events", clearBody); w.Code != 403 {
		t.Fatal("non-human cleared context", w.Code)
	}
}

func TestClearContextFlagBothStatesAndBoundarySurvivesDisable(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	// Default (flag undefined) is on.
	if w := mockRequest(h, "POST", "/api/events", clearBody); w.Code != 200 {
		t.Fatal("default state should allow clear", w.Code)
	}
	st, _ := s.load(t.Context())
	boundary := roomContextBoundary(st, "engineering")
	db.seed("feature.flag.updated", map[string]any{"name": "contextClear", "enabled": false})
	if w := mockRequest(h, "POST", "/api/events", clearBody); w.Code != 403 {
		t.Fatal("flag off must reject new clears", w.Code)
	}
	st, _ = s.load(t.Context())
	if roomContextBoundary(st, "engineering") != boundary {
		t.Fatal("disabling the flag removed an existing boundary")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "contextClear", "enabled": true})
	if w := mockRequest(h, "POST", "/api/events", clearBody); w.Code != 200 {
		t.Fatal("flag on must allow clear", w.Code)
	}
}

func TestClearContextRefusedWhileRoomAgentsRun(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("participant.created", room.Participant{ID: "p", RoomID: "engineering", ConnectionID: "c", Name: "Agent"})
	s.active["p"] = true
	if w := mockRequest(h, "POST", "/api/events", clearBody); w.Code != 409 {
		t.Fatal("cleared during an active turn", w.Code)
	}
}

func TestChatPromptHonoursContextBoundary(t *testing.T) {
	t.Parallel()
	st := disclosureFixture()
	p := room.Participant{ID: "p", RoomID: "engineering", ConnectionID: "c"}
	st.Messages = []room.Message{
		{ID: "m1", RoomID: "engineering", Text: "stale-secret-topic"},
		{ID: "m2", RoomID: "engineering", Text: "Cleared agent context. Reason: new topic"},
		{ID: "m3", RoomID: "engineering", Text: "fresh-question"},
	}
	before, err := chatPrompt(st, p, "go")
	if err != nil || !strings.Contains(before, "stale-secret-topic") {
		t.Fatal("fixture should include history before a clear", err)
	}
	st.Rooms = []room.Room{{ID: "engineering", ContextBoundaryID: "m2"}}
	after, err := chatPrompt(st, p, "go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(after, "stale-secret-topic") {
		t.Fatal("pre-boundary message reached the agent")
	}
	if !strings.Contains(after, "Reason: new topic") || !strings.Contains(after, "fresh-question") {
		t.Fatal("reason and later messages must remain visible")
	}
}
