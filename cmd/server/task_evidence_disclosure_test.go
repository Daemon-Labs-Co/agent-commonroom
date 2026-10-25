package main

import (
	"agentroom/internal/room"
	"context"
	"strings"
	"testing"
	"time"
)

func TestTaskEvidenceDisclosureHonorsReferencePolicyIndependentlyOfOptionalMetadata(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		for _, metadata := range []bool{false, true} {
			t.Run(strings.Join([]string{map[bool]string{true: "excluded", false: "included"}[exclude], map[bool]string{true: "metadata", false: "ordinary"}[metadata]}, "/"), func(t *testing.T) {
				t.Parallel()
				st := room.State{Rooms: []room.Room{{ID: "engineering", ContextExport: &room.ContextExportPolicy{Providers: []string{"openai"}, ExcludeReferences: exclude}}}, Connections: []room.Connection{{ID: "c", Protocol: "CLI", Endpoint: "codex"}}, FeatureFlags: map[string]bool{"roomOutcomes": metadata, "roomOutcomeContext": metadata, "contextPolicyEditor": false}, Evidence: []room.Evidence{{ID: "e", RoomID: "engineering", TaskID: "task", Title: "SYNTHETIC_EVIDENCE", Body: "<untrusted body>", Source: "SYNTHETIC_SOURCE"}, {RoomID: "foreign", TaskID: "task", Body: "FOREIGN_ROOM"}, {RoomID: "engineering", TaskID: "other", Body: "FOREIGN_TASK"}}}
				prompt, err := taskDispatchPrompt(st, room.Participant{ID: "dev", RoomID: "engineering", ConnectionID: "c"}, "EXPLICIT_TASK_INPUT", "task")
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(prompt, "EXPLICIT_TASK_INPUT") || strings.Contains(prompt, "SYNTHETIC_EVIDENCE") == exclude || strings.Contains(prompt, "FOREIGN_ROOM") || strings.Contains(prompt, "FOREIGN_TASK") {
					t.Fatal("evidence scope/exclusion mismatch")
				}
				if !exclude && (!strings.Contains(prompt, "untrusted human-provided reference") || !strings.Contains(prompt, `"evidenceId":"e"`) || strings.Contains(prompt, "<untrusted body>")) {
					t.Fatal("unbounded/trusted raw evidence context")
				}
			})
		}
	}
}
func TestExcludedOversizedEvidenceDoesNotBlockButIncludedOversizedEvidenceRefuses(t *testing.T) {
	t.Parallel()
	st := room.State{Rooms: []room.Room{{ID: "engineering", ContextExport: &room.ContextExportPolicy{Providers: []string{"openai"}, ExcludeReferences: true}}}, Connections: []room.Connection{{ID: "c", Protocol: "CLI", Endpoint: "codex"}}, Evidence: []room.Evidence{{RoomID: "engineering", TaskID: "task", Body: strings.Repeat("x", maxChatContextBytes+1)}}}
	p := room.Participant{RoomID: "engineering", ConnectionID: "c"}
	if _, err := taskDispatchPrompt(st, p, "explicit input", "task"); err != nil {
		t.Fatal("excluded data blocked dispatch", err)
	}
	st.Rooms[0].ContextExport.ExcludeReferences = false
	if _, err := taskDispatchPrompt(st, p, "explicit input", "task"); err == nil {
		t.Fatal("oversized evidence exported")
	}
	st.Evidence[0].Body = strings.Repeat("<", maxChatContextBytes/4)
	if _, err := taskDispatchPrompt(st, p, "explicit input", "task"); err == nil {
		t.Fatal("JSON escaping exceeded prompt bound")
	}
	st.Evidence[0].Body = string([]byte{255})
	if _, err := taskDispatchPrompt(st, p, "explicit input", "task"); err == nil {
		t.Fatal("invalid UTF-8 evidence exported")
	}

}
func TestActualTypedDispatchExcludesEvidenceButRetainsExplicitBindings(t *testing.T) {
	t.Parallel()
	s, db, h, task := structuredTaskFixture(t)
	db.seed("room.context_export.updated", map[string]any{"roomId": "engineering", "policy": room.ContextExportPolicy{Providers: []string{"openai"}, ExcludeReferences: true}})
	db.seed("evidence.created", room.Evidence{ID: "e", RoomID: "engineering", TaskID: task.ID, Title: "EXCLUDED_AUTOMATIC_EVIDENCE", Body: "SYNTHETIC_REFERENCE_SECRET"})
	seen := make(chan string, 1)
	s.taskTurn = func(_ context.Context, _ room.Participant, prompt, _ string) (string, error) {
		seen <- prompt
		return `{"result":"synthetic"}`, nil
	}
	if w := dispatchRequest(h, "evidence-excluded", definedDispatchBody(task)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case prompt := <-seen:
		if strings.Contains(prompt, "EXCLUDED_AUTOMATIC_EVIDENCE") || strings.Contains(prompt, "SYNTHETIC_REFERENCE_SECRET") || !strings.Contains(prompt, "synthetic input") {
			t.Fatal("dispatch disclosure mismatch")
		}
	case <-time.After(time.Second):
		t.Fatal("mock turn did not start")
	}
	_ = waitDispatch(t, s)
}
