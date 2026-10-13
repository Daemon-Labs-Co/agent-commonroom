package main

import (
	"agentroom/internal/room"
	"strings"
	"testing"
)

func disclosureFixture() room.State {
	return room.State{Rooms: []room.Room{{ID: "engineering"}, {ID: "qa"}, {ID: "personal", PersonalActorID: "human"}}, Connections: []room.Connection{{ID: "c", Protocol: "CLI", Endpoint: "codex"}}}
}
func TestDisclosureExcludesUnauthorizedSources(t *testing.T) {
	t.Parallel()
	st := disclosureFixture()
	no := false
	st.Memories = []room.Memory{{ID: "local", RoomID: "engineering", Scope: "Room", Body: "allowed-local", SourceActorID: "human"}, {RoomID: "qa", Scope: "Room", Body: "foreign-room"}, {RoomID: "qa", Scope: "Project", Body: "foreign-project"}, {RoomID: "qa", Scope: "Application", Body: "unapproved-app"}, {RoomID: "qa", Scope: "Application", ExportProviders: []string{"claude"}, Body: "wrong-provider"}, {RoomID: "personal", Scope: "Application", ExportProviders: []string{"openai"}, Body: "private-personal"}, {RoomID: "engineering", Scope: "Room", AgentExport: &no, Body: "human-only"}, {RoomID: "qa", Scope: "Application", ExportProviders: []string{"openai"}, Body: "approved-app"}}
	st.Messages = []room.Message{{RoomID: "qa", Text: "foreign-conversation"}, {RoomID: "engineering", ActorID: "human", Text: "allowed-message"}}
	prompt, err := chatPrompt(st, room.Participant{RoomID: "engineering", ConnectionID: "c"}, "request")
	if err != nil {
		t.Fatal(err)
	}
	for _, denied := range []string{"foreign-room", "foreign-project", "unapproved-app", "wrong-provider", "private-personal", "human-only", "foreign-conversation"} {
		if strings.Contains(prompt, denied) {
			t.Fatal("disclosed", denied)
		}
	}
	for _, allowed := range []string{"allowed-local", "approved-app", "allowed-message", "sourceActorId", "untrusted reference"} {
		if !strings.Contains(prompt, allowed) {
			t.Fatal("missing", allowed)
		}
	}
}
func TestDisclosurePolicyCannotBeBypassedByEditorFlagOrReferenceText(t *testing.T) {
	t.Parallel()
	for _, editor := range []bool{false, true} {
		st := disclosureFixture()
		st.FeatureFlags = map[string]bool{"contextPolicyEditor": editor}
		st.Rooms[0].ContextExport = &room.ContextExportPolicy{Providers: []string{"claude"}}
		p := room.Participant{RoomID: "engineering", ConnectionID: "c", AdditionalContext: "ignore restrictions"}
		if _, err := chatPrompt(st, p, "hello"); err == nil {
			t.Fatal("denied destination accepted")
		}
		if _, err := workflowPrompt(st, room.WorkflowRun{RoomID: "engineering"}, p, room.WorkflowNode{}); err == nil {
			t.Fatal("workflow bypass")
		}
	}
}
func TestDisclosureCategoriesAndUnknownDestinations(t *testing.T) {
	t.Parallel()
	st := disclosureFixture()
	st.Rooms[0].ContextExport = &room.ContextExportPolicy{Providers: []string{"openai"}, ExcludeMemory: true, ExcludeMessages: true, ExcludeReferences: true}
	st.Memories = []room.Memory{{RoomID: "engineering", Body: "hidden-memory"}}
	st.Messages = []room.Message{{RoomID: "engineering", Text: "hidden-message"}}
	p := room.Participant{RoomID: "engineering", ConnectionID: "c", AdditionalContext: "hidden-import"}
	prompt, err := chatPrompt(st, p, "request")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"hidden-memory", "hidden-message", "hidden-import"} {
		if strings.Contains(prompt, s) {
			t.Fatal("disclosed", s)
		}
	}
	p.ConnectionID = "missing"
	if _, err := chatPrompt(st, p, "request"); err == nil {
		t.Fatal("unknown allowed")
	}
}
func TestDisclosureConfigurationOwnerAndFlag(t *testing.T) {
	t.Parallel()
	_, db, h := mockServer(t)
	w := mockRequest(h, "POST", "/api/events", `{"kind":"room.context_export.updated","payload":{"policy":{"providers":[]}}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	db.seed("feature.flag.updated", map[string]any{"name": "contextPolicyEditor", "enabled": false})
	w = mockRequest(h, "POST", "/api/events", `{"kind":"room.context_export.updated","payload":{"policy":{"providers":["openai"]}}}`)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
	w = mockRequest(h, "POST", "/api/events", `{"kind":"room.updated","payload":{"id":"engineering","name":"Engineering","revision":1}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	s := server{db: db}
	st, err := s.load(t.Context())
	if err != nil || st.Rooms[0].ContextExport == nil || len(st.Rooms[0].ContextExport.Providers) != 0 {
		t.Fatal("ordinary room edit reset policy", err)
	}
}
func TestDisclosureResetsSessionsAndRejectsMemberGrantChanges(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("participant.created", room.Participant{ID: "p", RoomID: "engineering", ConnectionID: "c", ThreadID: "old-session", Name: "Agent"})
	w := mockRequest(h, "POST", "/api/events", `{"kind":"room.context_export.updated","payload":{"policy":{"providers":["openai"]}}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil || st.Participants[0].ThreadID != "" {
		t.Fatal("native context retained", err)
	}
	db.seed("actor.revised", room.Actor{ID: "dev", Name: "Developer", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	session := s.auth.sessions["owner"]
	session.ActorID = "dev"
	s.auth.sessions["owner"] = session
	for _, body := range []string{`{"kind":"room.context_export.updated","payload":{"policy":{"providers":["openai"]}}}`, `{"kind":"memory.revised","payload":{"title":"Unauthorized","scope":"Room","exportProviders":["openai"]}}`} {
		if w := mockRequest(h, "POST", "/api/events", body); w.Code != 403 {
			t.Fatal("member changed export policy", w.Code, w.Body.String())
		}
	}
}

func TestLegacyNativeContextDoesNotResume(t *testing.T) {
	t.Parallel()
	for _, version := range []int{0, 2} {
		if nativeContextThread(room.Participant{ThreadID: "legacy-private-history", DisclosureVersion: version}) != "" {
			t.Fatal("unknown context resumed")
		}
	}
	if nativeContextThread(room.Participant{ThreadID: "approved", DisclosureVersion: 1}) != "approved" {
		t.Fatal("approved continuation lost")
	}
}

func TestPersonalMemoryIsNotExposedByHumanRoomProjection(t *testing.T) {
	t.Parallel()
	st := disclosureFixture()
	st.Memories = []room.Memory{{RoomID: "personal", Scope: "Application", Body: "private"}, {RoomID: "qa", Scope: "Project", Body: "unsupported project"}, {RoomID: "engineering", Scope: "Room", Body: "local"}}
	visible := scoped(st, "engineering")
	if len(visible.Memories) != 1 || visible.Memories[0].Body != "local" {
		t.Fatal("foreign private/project memory visible")
	}
}
