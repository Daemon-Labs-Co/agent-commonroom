package main

import (
	"agentroom/internal/room"
	"context"
	"testing"
	"time"
)

func TestAutomaticAgentRoomCatalogAndHistory(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "unused", Name: "Synthetic agent"})
	db.seed("connection.created", room.Connection{ID: "used", Name: "Historical agent"})
	db.seed("message.created", room.Message{ID: "historical", RoomID: "personal-bot-used", Text: "Synthetic retained history"})
	load := func() room.State {
		t.Helper()
		st, err := s.load(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		return st
	}
	st := load()
	if hasRoom(st, "personal-bot-unused") || !hasRoom(st, "personal-bot-used") || !hasRoom(st, "personal-local-owner") {
		t.Fatal("default catalog lost history or generated unused agent room")
	}
	if _, ok := member(st, "bot-unused"); !ok {
		t.Fatal("agent identity removed with room")
	}
	if len(st.Messages) != 1 || st.Messages[0].Text != "Synthetic retained history" {
		t.Fatal("history changed")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "automaticAgentRooms", "enabled": true})
	if !hasRoom(load(), "personal-bot-unused") {
		t.Fatal("compatibility flag did not restore automatic rooms")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "automaticAgentRooms", "enabled": false})
	if hasRoom(load(), "personal-bot-unused") || !hasRoom(load(), "personal-bot-used") {
		t.Fatal("flag off changed retained room policy")
	}
	restarted := &server{db: db}
	after, err := restarted.load(context.Background())
	if err != nil || hasRoom(after, "personal-bot-unused") || !hasRoom(after, "personal-bot-used") {
		t.Fatal("restart changed catalog", err)
	}
}

func TestAgentRoomReferenceRetention(t *testing.T) {
	for _, payload := range []string{
		`{"id":"personal-bot-agent","description":"Saved settings"}`,
		`{"roomId":"personal-bot-agent","taskId":"synthetic"}`,
		`{"scope":{"room":"personal-bot-agent"}}`,
		`{"nested":[{"roomId":"personal-bot-agent"}]}`,
		`{"roomId":"\u0070ersonal-bot-agent"}`,
		`{"roomId":"personal-\u0062ot-agent"}`,
	} {
		t.Run(payload, func(t *testing.T) {
			t.Parallel()
			referenced := map[string]bool{}
			rememberAgentRoomReferences([]byte(payload), referenced)
			st := room.State{Actors: []room.Actor{{ID: "bot-agent", Kind: "agent"}}, Rooms: []room.Room{{ID: "personal-bot-agent", PersonalActorID: "bot-agent"}, {ID: "shared"}}}
			omitUnusedAgentRooms(&st, referenced)
			if !hasRoom(st, "personal-bot-agent") || !hasRoom(st, "shared") {
				t.Fatal("historical reference lost room")
			}
		})
	}
}

func TestAutomaticAgentRoomFlagOwnerOnly(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"automaticAgentRooms","enabled":true}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	before := len(db.snapshot())
	w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"automaticAgentRooms","enabled":false}}`)
	if w.Code != 403 || len(db.snapshot()) != before {
		t.Fatal("member changed catalog flag", w.Code)
	}
}

func TestAgentRoomReferenceFastPathPreservesExistingReferences(t *testing.T) {
	t.Parallel()
	referenced := map[string]bool{"personal-bot-existing": true}
	rememberAgentRoomReferences([]byte(`{"roomId":"engineering","nested":[{"text":"synthetic ordinary event\nwith escaped newline"}]}`), referenced)
	if len(referenced) != 1 || !referenced["personal-bot-existing"] {
		t.Fatal("unrelated replay changed retained history")
	}
}
