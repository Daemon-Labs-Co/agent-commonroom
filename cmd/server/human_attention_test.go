package main

import (
	"agentroom/internal/room"
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHumanRequestDetectionAndProjection(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ text, kind string }{
		{"[HUMAN ACTION REQUIRED] Choose the deployment target.", "explicit"},
		{"Please approve the proposed design.", "suggested"},
		{"To unblock implementation, identify the staff engineers.", "suggested"},
		{"What does this function do?", ""},
		{"Tests passed; proceeding with the assigned task.", ""},
	} {
		if got := humanRequestKind(c.text); got != c.kind {
			t.Fatalf("%q: %s", c.text, got)
		}
	}
	s, db, _ := mockServer(t)
	db.seed("message.created", room.Message{ID: "agent", RoomID: "engineering", ActorType: "agent", Text: "[HUMAN ACTION REQUIRED] Please approve the design."})
	db.seed("message.created", room.Message{ID: "human", RoomID: "engineering", ActorType: "human", Text: "Please approve", HumanRequest: "explicit"})
	st, _ := s.load(context.Background())
	if st.Messages[0].HumanRequest != "explicit" || st.Messages[1].HumanRequest != "" || st.Rooms[0].PendingHumanRequests != 1 {
		t.Fatal(st.Messages, st.Rooms)
	}
	db.seed("feature.flag.updated", map[string]any{"name": "humanAttention", "enabled": false})
	st, _ = s.load(context.Background())
	if st.Messages[0].HumanRequest != "" || st.Rooms[0].PendingHumanRequests != 0 {
		t.Fatal("off flag exposes attention")
	}
}

func TestHumanAttentionAcknowledgmentPreservesMessageAndIsIdempotent(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	original := room.Message{ID: "request", RoomID: "engineering", ActorType: "agent", Author: "Synthetic agent", Text: "[HUMAN ACTION REQUIRED] Choose a target."}
	db.seed("message.created", original)
	body := `{"kind":"message.attention.acknowledged","payload":{"messageId":"request"}}`
	w := mockRequest(h, "POST", "/api/events", body)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ := s.load(context.Background())
	if st.Messages[0].Text != original.Text || st.Messages[0].AttentionAcknowledgedBy != ownerID || st.Rooms[0].PendingHumanRequests != 0 {
		t.Fatal(st)
	}
	count := len(db.snapshot())
	w = mockRequest(h, "POST", "/api/events", body)
	if w.Code != 200 || len(db.snapshot()) != count {
		t.Fatal("duplicate acknowledgment wrote another event")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "humanAttention", "enabled": false})
	w = mockRequest(h, "POST", "/api/events", body)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
	db.seed("feature.flag.updated", map[string]any{"name": "humanAttention", "enabled": true})
	st, _ = s.load(context.Background())
	if st.Messages[0].AttentionAcknowledgedBy != ownerID {
		t.Fatal("toggle loses acknowledgment")
	}
}

func TestHumanAttentionAcknowledgmentRoomAndAccountBoundaries(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("room.created", room.Room{ID: "elsewhere", Name: "Private room"})
	db.seed("message.created", room.Message{ID: "foreign", RoomID: "elsewhere", ActorType: "agent", Text: "Please approve"})
	db.seed("message.created", room.Message{ID: "human", RoomID: "engineering", ActorType: "human", Text: "Please approve"})
	for _, id := range []string{"foreign", "human", "missing"} {
		w := mockRequest(h, "POST", "/api/events", `{"kind":"message.attention.acknowledged","payload":{"messageId":"`+id+`"}}`)
		if w.Code != 404 {
			t.Fatal(id, w.Code)
		}
	}
	db.seed("message.created", room.Message{ID: "allowed", RoomID: "engineering", ActorType: "agent", Text: "Please approve"})
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true, Role: "Developer", RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	w := mockRequest(h, "POST", "/api/events", `{"kind":"message.attention.acknowledged","payload":{"messageId":"allowed","actorId":"forged-owner"}}`)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ := s.load(context.Background())
	if st.Messages[len(st.Messages)-1].AttentionAcknowledgedBy != "developer" {
		t.Fatal("forged identity used")
	}
	visible := visibleState(st, "developer", "engineering")
	for _, r := range visible.Rooms {
		if r.ID == "elsewhere" {
			t.Fatal("foreign-room attention visible")
		}
	}
}

func TestAgentAccountCannotAcknowledgeHumanRequest(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("message.created", room.Message{ID: "request", RoomID: "engineering", ActorType: "agent", Text: "[HUMAN ACTION REQUIRED] Choose."})
	db.seed("actor.revised", room.Actor{ID: "agent-account", Kind: "agent", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "agent-account", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	w := mockRequest(h, "POST", "/api/events", `{"kind":"message.attention.acknowledged","payload":{"messageId":"request"}}`)
	if w.Code != 403 {
		t.Fatal("agent acknowledged request", w.Code)
	}
}

func TestPublicMessageCannotForgeHumanAttention(t *testing.T) {
	t.Parallel()
	s, _, h := mockServer(t)
	w := mockRequest(h, "POST", "/api/events", `{"kind":"message.created","payload":{"text":"[HUMAN ACTION REQUIRED] Spoof","actorType":"agent","humanRequest":"explicit","attentionAcknowledgedBy":"owner"}}`)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	st, _ := s.load(context.Background())
	m := st.Messages[0]
	if m.ActorType != "human" || m.HumanRequest != "" || m.AttentionAcknowledgedBy != "" {
		t.Fatal("client forged agent request", m)
	}
}
