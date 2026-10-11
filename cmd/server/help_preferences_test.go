package main

import (
	"agentroom/internal/room"
	"context"
	"strings"
	"testing"
	"time"
)

func TestHelpPreferenceSelfOnly(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("actor.revised", room.Actor{ID: "developer", Name: "Dev", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	w := mockRequest(h, "POST", "/api/events", `{"kind":"actor.help.updated","payload":{"id":"local-owner","mode":"separate-window","role":"Owner"}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	dev, _ := member(st, "developer")
	owner, _ := member(st, ownerID)
	if dev.HelpDisplay != "separate-window" || dev.Role != "Developer" || owner.HelpDisplay != "" {
		t.Fatal("preference changed someone else's profile or permissions")
	}
	db.seed("actor.revised", room.Actor{ID: "developer", Name: "Changed", Kind: "human", Role: "QA", Enabled: true, RoomIDs: []string{"engineering"}})
	st, err = s.load(context.Background())
	dev, _ = member(st, "developer")
	if err != nil || dev.HelpDisplay != "separate-window" {
		t.Fatal("member edits erased help preference")
	}
}
func TestHelpPreferenceValidationAndFlag(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	for _, mode := range []string{"", "popup", "https://evil.example"} {
		w := mockRequest(h, "POST", "/api/events", `{"kind":"actor.help.updated","payload":{"mode":"`+mode+`"}}`)
		if w.Code != 400 {
			t.Fatal(mode, w.Code)
		}
	}
	db.seed("feature.flag.updated", map[string]any{"name": "helpWindowChoice", "enabled": false})
	w := mockRequest(h, "POST", "/api/events", `{"kind":"actor.help.updated","payload":{"mode":"separate-window"}}`)
	if w.Code != 403 || !strings.Contains(w.Body.String(), "disabled") {
		t.Fatal(w.Code, w.Body.String())
	}
	db.seed("feature.flag.updated", map[string]any{"name": "helpWindowChoice", "enabled": true})
	for _, mode := range []string{"separate-window", "same-window"} {
		w = mockRequest(h, "POST", "/api/events", `{"kind":"actor.help.updated","payload":{"mode":"`+mode+`"}}`)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		st, err := s.load(context.Background())
		a, _ := member(st, ownerID)
		if err != nil || a.HelpDisplay != mode {
			t.Fatal(a, err)
		}
	}
}
