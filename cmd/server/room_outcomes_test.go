package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRoomOutcomeLifecycle(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	call := func(kind string, p any) int {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{"kind": kind, "payload": p})
		return mockRequest(h, "POST", "/api/events", string(raw)).Code
	}
	if code := call("room.created", map[string]any{"name": "Disabled scope", "outcome": "Cannot edit", "acceptanceCriteria": "No grant"}); code != 403 {
		t.Fatal(code)
	}
	if code := call("feature.flag.updated", map[string]any{"name": "roomOutcomes", "enabled": true}); code != 200 {
		t.Fatal(code)
	}
	for _, p := range []map[string]any{{"name": "Missing"}, {"name": "Partial", "outcome": "A result"}, {"name": "Long", "outcome": strings.Repeat("a", 2001), "acceptanceCriteria": "Verify"}} {
		before := len(db.snapshot())
		if code := call("room.created", p); code != 400 || len(db.snapshot()) != before {
			t.Fatal("invalid room persisted", code)
		}
	}
	if code := call("room.created", map[string]any{"name": "Release", "outcome": "  Publish synthetic release  ", "acceptanceCriteria": "  Tests pass; owner verifies package  "}); code != 200 {
		t.Fatal(code)
	}
	load := func() room.State {
		t.Helper()
		st, err := s.load(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		return st
	}
	st := load()
	r := st.Rooms[len(st.Rooms)-1]
	// Find explicitly created room rather than relying on derived catalog order.
	for _, candidate := range st.Rooms {
		if candidate.Name == "Release" {
			r = candidate
		}
	}
	if r.Outcome != "Publish synthetic release" || r.AcceptanceCriteria != "Tests pass; owner verifies package" {
		t.Fatal("outcome not saved", r)
	}
	if code := call("room.updated", map[string]any{"id": r.ID, "revision": r.Revision, "name": r.Name, "description": "Unrelated edit"}); code != 200 {
		t.Fatal(code)
	}
	for _, candidate := range load().Rooms {
		if candidate.ID == r.ID {
			r = candidate
		}
	}
	if r.Outcome == "" || r.AcceptanceCriteria == "" {
		t.Fatal("old client erased outcome")
	}
	if code := call("room.updated", map[string]any{"id": r.ID, "revision": r.Revision, "name": r.Name, "outcome": ""}); code != 400 {
		t.Fatal("cleared outcome", code)
	}
	stale := r
	r.Outcome = "Updated release"
	if code := call("room.updated", r); code != 200 {
		t.Fatal(code)
	}
	if code := call("room.updated", stale); code != 409 {
		t.Fatal("stale edit accepted", code)
	}
	if code := call("feature.flag.updated", map[string]any{"name": "roomOutcomes", "enabled": false}); code != 200 {
		t.Fatal(code)
	}
	for _, candidate := range load().Rooms {
		if candidate.ID == r.ID {
			r = candidate
		}
	}
	changed := r
	changed.Outcome = "Bypass disabled flag"
	if code := call("room.updated", changed); code != 403 {
		t.Fatal(code)
	}
	if code := call("room.updated", map[string]any{"id": r.ID, "revision": r.Revision, "name": r.Name, "description": "Off preserves intent"}); code != 200 {
		t.Fatal(code)
	}
	restarted := &server{db: db}
	after, err := restarted.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, candidate := range after.Rooms {
		if candidate.ID == r.ID {
			found = candidate.Outcome == "Updated release" && candidate.AcceptanceCriteria == r.AcceptanceCriteria
		}
	}
	if !found {
		t.Fatal("disabled restart lost outcome")
	}
}

func TestRoomOutcomePermissionsAndLegacy(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomes", "enabled": true})
	// Legacy rooms need not be rewritten before ordinary updates can continue.
	w := mockRequest(h, "POST", "/api/events", `{"kind":"room.updated","payload":{"id":"engineering","name":"Engineering","revision":0,"description":"Legacy work"}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	// The legacy browser form explicitly submits both blank textareas.
	w = mockRequest(h, "POST", "/api/events", `{"kind":"room.updated","payload":{"id":"engineering","name":"Renamed legacy room","revision":1,"description":"Unrelated description","outcome":"","acceptanceCriteria":""}}`)
	if w.Code != 200 {
		t.Fatal("legacy blank-definition save blocked", w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range st.Rooms {
		if r.ID == "engineering" && (r.Name != "Renamed legacy room" || r.Description != "Unrelated description" || r.Outcome != "" || r.AcceptanceCriteria != "") {
			t.Fatal("legacy save changed definition", r)
		}
		if r.PersonalActorID == ownerID {
			r.Outcome = "Not a work room"
			r.AcceptanceCriteria = "Not applicable"
			raw, _ := json.Marshal(r)
			w = mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"room.updated","payload":%s}`, raw))
			if w.Code != 400 {
				t.Fatal("personal outcome accepted", w.Code)
			}
		}
	}
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	before := len(db.snapshot())
	for _, body := range []string{`{"kind":"feature.flag.updated","payload":{"name":"roomOutcomes","enabled":false}}`, `{"kind":"room.updated","payload":{"id":"engineering","revision":1,"name":"Engineering","outcome":"Member expansion","acceptanceCriteria":"No approval"}}`} {
		w = mockRequest(h, "POST", "/api/events", body)
		if w.Code != 403 {
			t.Fatal("member changed outcome policy", w.Code, w.Body.String())
		}
	}
	if len(db.snapshot()) != before {
		t.Fatal("unauthorized writes")
	}
}
