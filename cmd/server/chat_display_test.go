package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func chatDisplayFixture(t *testing.T) (*server, *mockDB, func(method, path, body string) (int, string)) {
	t.Helper()
	s, db, h := mockServer(t)
	db.seed("actor.revised", room.Actor{ID: "developer", Name: "Dev", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	db.seed("actor.revised", room.Actor{ID: "bot-ca", Name: "Alex", Kind: "agent", Role: "Agent", Enabled: true})
	return s, db, func(method, path, body string) (int, string) {
		w := mockRequest(h, method, path, body)
		return w.Code, w.Body.String()
	}
}

func TestChatStyleIsPersonalAndSurvivesProfileEdits(t *testing.T) {
	t.Parallel()
	s, db, request := chatDisplayFixture(t)
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	// A supplied id is ignored: the session actor is always the one changed.
	if code, body := request("POST", "/api/events", `{"kind":"actor.chat_style.updated","payload":{"id":"local-owner","enabled":true}}`); code != 200 {
		t.Fatal(code, body)
	}
	if code, body := request("POST", "/api/events", `{"kind":"actor.agent_highlight.updated","payload":{"id":"local-owner","agentId":"bot-ca","enabled":true}}`); code != 200 {
		t.Fatal(code, body)
	}
	db.seed("actor.revised", room.Actor{ID: "developer", Name: "Renamed", Kind: "human", Role: "QA", Enabled: true, RoomIDs: []string{"engineering"}})
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	dev, _ := member(st, "developer")
	owner, _ := member(st, ownerID)
	if !dev.ChatStyle || len(dev.HighlightedAgents) != 1 || dev.HighlightedAgents[0] != "bot-ca" || dev.Role != "QA" {
		t.Fatal("preference lost or permissions changed", dev)
	}
	if owner.ChatStyle || len(owner.HighlightedAgents) != 0 {
		t.Fatal("another person's display changed", owner)
	}
	// Each viewer receives only their own choices.
	for viewer, want := range map[string]bool{"developer": true, ownerID: false} {
		seen := visibleState(st, viewer, "engineering")
		for _, a := range seen.Actors {
			if a.ID == "developer" && (a.ChatStyle != want || (len(a.HighlightedAgents) == 1) != want) {
				t.Fatal("display choices visible to the wrong viewer", viewer, a)
			}
		}
	}
	if code, _ := request("POST", "/api/events", `{"kind":"actor.chat_style.updated","payload":{"enabled":false}}`); code != 200 {
		t.Fatal(code)
	}
	st, _ = s.load(context.Background())
	if dev, _ = member(st, "developer"); dev.ChatStyle {
		t.Fatal("chat style stayed on")
	}
}

func TestAgentHighlightValidationAndRemoval(t *testing.T) {
	t.Parallel()
	s, db, request := chatDisplayFixture(t)
	for body, want := range map[string]int{
		`{"kind":"actor.chat_style.updated","payload":{}}`:                                             400,
		`{"kind":"actor.chat_style.updated","payload":{"enabled":"yes"}}`:                              400,
		`{"kind":"actor.agent_highlight.updated","payload":{"enabled":true}}`:                          400,
		`{"kind":"actor.agent_highlight.updated","payload":{"agentId":"bot-ca"}}`:                      400,
		`{"kind":"actor.agent_highlight.updated","payload":{"agentId":"bot-missing","enabled":true}}`:  404,
		`{"kind":"actor.agent_highlight.updated","payload":{"agentId":"developer","enabled":true}}`:    404,
		`{"kind":"actor.agent_highlight.updated","payload":{"agentId":"bot-ca","enabled":true}}`:       200,
		`{"kind":"actor.agent_highlight.updated","payload":{"agentId":"bot-missing","enabled":false}}`: 200,
	} {
		if code, text := request("POST", "/api/events", body); code != want {
			t.Fatal(body, code, text)
		}
	}
	// Repeating a highlight keeps one entry; removing clears it.
	request("POST", "/api/events", `{"kind":"actor.agent_highlight.updated","payload":{"agentId":"bot-ca","enabled":true}}`)
	st, _ := s.load(context.Background())
	if owner, _ := member(st, ownerID); len(owner.HighlightedAgents) != 1 {
		t.Fatal(owner.HighlightedAgents)
	}
	request("POST", "/api/events", `{"kind":"actor.agent_highlight.updated","payload":{"agentId":"bot-ca","enabled":false}}`)
	st, _ = s.load(context.Background())
	if owner, _ := member(st, ownerID); owner.HighlightedAgents != nil {
		t.Fatal(owner.HighlightedAgents)
	}
	// The list is bounded.
	for i := 0; i < maxHighlightedAgents; i++ {
		payload, _ := json.Marshal(agentHighlightUpdate{ID: ownerID, AgentID: "bot-x" + string(rune('A'+i%26)) + string(rune('a'+i/26)), Enabled: true})
		db.seed("actor.agent_highlight.updated", json.RawMessage(payload))
	}
	if code, text := request("POST", "/api/events", `{"kind":"actor.agent_highlight.updated","payload":{"agentId":"bot-ca","enabled":true}}`); code != 409 {
		t.Fatal(code, text)
	}
}

func TestChatDisplayFlagStates(t *testing.T) {
	t.Parallel()
	s, db, request := chatDisplayFixture(t)
	if code, _ := request("POST", "/api/events", `{"kind":"actor.chat_style.updated","payload":{"enabled":true}}`); code != 200 {
		t.Fatal("enabled by default", code)
	}
	if code, body := request("POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"chatDisplayChoices","enabled":false}}`); code != 200 {
		t.Fatal(code, body)
	}
	for _, body := range []string{
		`{"kind":"actor.chat_style.updated","payload":{"enabled":false}}`,
		`{"kind":"actor.agent_highlight.updated","payload":{"agentId":"bot-ca","enabled":true}}`,
	} {
		if code, text := request("POST", "/api/events", body); code != 403 || !strings.Contains(text, "disabled") {
			t.Fatal(body, code, text)
		}
	}
	// Saved choices are kept while the flag is off.
	st, _ := s.load(context.Background())
	if owner, _ := member(st, ownerID); !owner.ChatStyle {
		t.Fatal("disabling the flag erased a saved choice")
	}
	// Only the owner may change the flag.
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if code, _ := request("POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"chatDisplayChoices","enabled":true}}`); code != 403 {
		t.Fatal("non-owner changed the flag", code)
	}
	_ = db
}
