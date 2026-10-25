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

func TestTaskDeliverableContextDisclosure(t *testing.T) {
	for _, tc := range []struct {
		name                                      string
		metadata, export, exclude, deny, personal bool
		want                                      bool
		wantError                                 bool
	}{
		{name: "enabled", metadata: true, export: true, want: true},
		{name: "export off", metadata: true}, {name: "metadata off", export: true}, {name: "default off"},
		{name: "references excluded", metadata: true, export: true, exclude: true},
		{name: "provider denied", metadata: true, export: true, deny: true, wantError: true},
		{name: "personal room", metadata: true, export: true, personal: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			st := room.State{FeatureFlags: map[string]bool{"roomOutcomes": tc.metadata, "roomOutcomeContext": tc.export}, Connections: []room.Connection{{ID: "c", Protocol: "CLI", Endpoint: "codex"}}, Rooms: []room.Room{{ID: "work", Revision: 7, Outcome: "SYNTHETIC_OUTCOME", AcceptanceCriteria: "SYNTHETIC_CRITERIA"}, {ID: "foreign", Outcome: "FOREIGN_SECRET"}}}
			if tc.exclude || tc.deny {
				st.Rooms[0].ContextExport = &room.ContextExportPolicy{Providers: []string{"openai"}, ExcludeReferences: tc.exclude}
				if tc.deny {
					st.Rooms[0].ContextExport.Providers = []string{"claude"}
				}
			}
			if tc.personal {
				st.Rooms[0].PersonalActorID = "human"
			}
			p := room.Participant{ID: "dev", ConnectionID: "c", RoomID: "work"}
			prompt, err := taskDispatchPrompt(st, p, "Bounded task instructions")
			if (err != nil) != tc.wantError {
				t.Fatal(err)
			}
			if strings.Contains(prompt, "SYNTHETIC_OUTCOME") != tc.want || strings.Contains(prompt, "SYNTHETIC_CRITERIA") != tc.want || strings.Contains(prompt, "FOREIGN_SECRET") {
				t.Fatal("disclosure mismatch", prompt)
			}
			if tc.want && (!strings.Contains(prompt, `"roomRevision":7`) || !strings.Contains(prompt, "not task scope")) {
				t.Fatal("missing provenance/boundary")
			}
			ordinary, err := chatPrompt(st, p, "Ordinary chat")
			if !tc.wantError && err != nil {
				t.Fatal(err)
			}
			if strings.Contains(ordinary, "SYNTHETIC_OUTCOME") {
				t.Fatal("ordinary chat exported deliverable")
			}
		})
	}
}

func TestTaskDeliverableDispatchWorkAndReview(t *testing.T) {
	for _, mode := range []string{"work", "review"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, h, _ := dispatchFixture(t)
			db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomes", "enabled": true})
			db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomeContext", "enabled": true})
			db.seed("room.updated", room.Room{ID: "engineering", Name: "Engineering", Outcome: "DELIVERABLE_SNAPSHOT", AcceptanceCriteria: "Owner reviews synthetic evidence", Revision: 1})
			seen := make(chan string, 1)
			release := make(chan struct{})
			finished := make(chan struct{})
			s.taskTurn = func(ctx context.Context, _ room.Participant, prompt, _ string) (string, error) {
				defer close(finished)
				seen <- prompt
				select {
				case <-release:
					return "Synthetic result", nil
				case <-ctx.Done():
					return "", ctx.Err()
				}
			}
			w := dispatchRequest(h, "deliverable-"+mode, fmt.Sprintf(`{"taskId":"task","revision":1,"mode":%q}`, mode))
			if w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			var prompt string
			select {
			case prompt = <-seen:
			case <-time.After(time.Second):
				t.Fatal("no dispatch")
			}
			if !strings.Contains(prompt, "DELIVERABLE_SNAPSHOT") || !strings.Contains(prompt, "Synthetic acceptance") {
				t.Fatal("lost room or task criteria", prompt)
			}
			// A repeated request must not start another provider turn.
			replay := dispatchRequest(h, "deliverable-"+mode, fmt.Sprintf(`{"taskId":"task","revision":1,"mode":%q}`, mode))
			if replay.Code != 202 || replay.Header().Get("Idempotent-Replay") != "true" {
				t.Fatal("dispatch replay failed")
			}
			db.seed("room.updated", room.Room{ID: "engineering", Name: "Engineering", Outcome: "CHANGED_AFTER_DISPATCH", AcceptanceCriteria: "Changed later", Revision: 2})
			if strings.Contains(prompt, "CHANGED_AFTER_DISPATCH") {
				t.Fatal("snapshot mutated")
			}
			close(release)
			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("turn did not stop")
			}
		})
	}
}

func TestTaskDeliverableBoundsAndNoEffectsOnRefusal(t *testing.T) {
	t.Parallel()
	s, db, h, _ := dispatchFixture(t)
	db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomes", "enabled": true})
	db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomeContext", "enabled": true})
	db.seed("room.updated", room.Room{ID: "engineering", Name: "Engineering", Outcome: strings.Repeat("x", 2001), AcceptanceCriteria: "Synthetic", Revision: 1})
	before := len(db.snapshot())
	w := dispatchRequest(h, "oversize", dispatchBody)
	if w.Code != 413 || len(db.snapshot()) != before || len(s.active) != 0 {
		t.Fatal("oversized historical outcome dispatched", w.Code, w.Body.String())
	}
	// Valid metadata can still exceed the complete 128 KiB prompt budget.
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	st.Rooms[0].Outcome = strings.Repeat("\x00", 2000)
	st.Rooms[0].AcceptanceCriteria = strings.Repeat("\x00", 8000)
	p := st.Participants[0]
	p.AdditionalContext = strings.Repeat("a", maxChatContextBytes-10000)
	if _, err = taskDispatchPrompt(st, p, "Synthetic task"); err == nil {
		t.Fatal("combined context overflow accepted")
	}
	// Owner-only flag changes use the real event admission path.
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
	raw, _ := json.Marshal(map[string]any{"kind": "feature.flag.updated", "payload": map[string]any{"name": "roomOutcomeContext", "enabled": false}})
	if w = mockRequest(h, "POST", "/api/events", string(raw)); w.Code != 403 {
		t.Fatal("member changed export flag", w.Code)
	}
}
