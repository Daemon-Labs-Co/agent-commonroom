package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"agentroom/internal/room"
)

func TestAgentDispatchBoundariesWithMockDB(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, body     string
		incident, busy bool
		code           int
	}{
		{"empty message", `{"action":"send","text":" "}`, false, false, 400},
		{"missing participant", `{"action":"send","participantId":"missing","text":"Hi"}`, false, false, 409},
		{"unsupported adapter", `{"action":"send","participantId":"p","text":"Hi"}`, false, false, 409},
		{"already working", `{"action":"send","participantId":"p","text":"Hi"}`, false, true, 409},
		{"incident", `{"action":"send","participantId":"p","text":"Hi"}`, true, false, 409},
		{"invalid attachment", `{"action":"attach","mode":"live"}`, false, false, 400},
		{"missing context ID", `{"action":"read-context"}`, false, false, 400},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s, db, _ := mockServer(t)
			endpoint := "https://example.invalid"
			protocol := "HTTP"
			if c.busy {
				endpoint = "codex"
				protocol = "CLI"
				s.active["p"] = true
			}
			db.seed("connection.created", room.Connection{ID: "c", Name: "Agent", Protocol: protocol, Endpoint: endpoint})
			db.seed("participant.created", room.Participant{ID: "p", RoomID: "engineering", Name: "Alex", ConnectionID: "c", Rank: 5})
			if c.incident {
				db.seed("incident.opened", map[string]string{"reason": "test"})
			}
			before := len(db.snapshot())
			w := mockRequest(http.HandlerFunc(s.agents), "POST", "/api/agents", c.body)
			if w.Code != c.code {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			if len(db.snapshot()) != before {
				t.Fatal("rejected dispatch wrote history")
			}
		})
	}
	t.Run("remove cancels and retains history", func(t *testing.T) {
		t.Parallel()
		s, db, _ := mockServer(t)
		db.seed("participant.created", room.Participant{ID: "p", Name: "Alex", RoomID: "engineering"})
		db.seed("message.created", room.Message{Author: "Alex", Text: "Retain me", RoomID: "engineering"})
		ctx, cancel := context.WithCancel(context.Background())
		s.cancels["p"] = cancel
		s.active["p"] = true
		w := mockRequest(http.HandlerFunc(s.agents), "POST", "/api/agents", `{"action":"remove","participantId":"p"}`)
		if w.Code != 200 || ctx.Err() != context.Canceled {
			t.Fatal("removal did not cancel work", w.Code)
		}
		st, err := s.load(context.Background())
		if err != nil || len(st.Participants) != 0 || len(st.Messages) != 1 || st.Messages[0].Text != "Retain me" {
			t.Fatal("removal lost history or retained agent")
		}
	})
	t.Run("broadcast reports skipped adapter", func(t *testing.T) {
		t.Parallel()
		s, db, _ := mockServer(t)
		db.seed("participant.created", room.Participant{ID: "p", Name: "Alex", RoomID: "engineering"})
		w := mockRequest(http.HandlerFunc(s.agents), "POST", "/api/agents", `{"action":"broadcast","text":"Hello everyone"}`)
		if w.Code != 202 || !strings.Contains(w.Body.String(), `"queued":0`) || !strings.Contains(w.Body.String(), "No executable adapter") {
			t.Fatal("skipped participant not reported", w.Body.String())
		}
	})
	t.Run("tampering blocks dispatch", func(t *testing.T) {
		t.Parallel()
		s, db, _ := mockServer(t)
		db.seed("message.created", room.Message{Text: "original"})
		db.events[0].hash = "tampered"
		w := mockRequest(http.HandlerFunc(s.agents), "POST", "/api/agents", `{"action":"send","text":"Hello"}`)
		if w.Code != 503 {
			t.Fatal("tampered chain dispatched", w.Code)
		}
	})
}
func TestChallengeCapacityAndExpiry(t *testing.T) {
	t.Parallel()
	s, _, _ := mockServer(t)
	for i := 0; i < 128; i++ {
		w := httptest.NewRecorder()
		s.challenge(w, httptest.NewRequest("POST", "/api/challenge", nil))
		if w.Code != 200 {
			t.Fatal("challenge rejected too early")
		}
	}
	w := httptest.NewRecorder()
	s.challenge(w, httptest.NewRequest("POST", "/api/challenge", nil))
	if w.Code != 429 {
		t.Fatal("challenge capacity not enforced")
	}
	for id, challenge := range s.auth.challenges {
		challenge.Expires = time.Now().Add(-time.Minute)
		s.auth.challenges[id] = challenge
	}
	w = httptest.NewRecorder()
	s.challenge(w, httptest.NewRequest("POST", "/api/challenge", nil))
	if w.Code != 200 || len(s.auth.challenges) != 1 {
		t.Fatal("expired challenges were not cleared")
	}
}
