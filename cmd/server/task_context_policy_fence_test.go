package main

import (
	"agentroom/internal/room"
	"context"
	"strings"
	"testing"
)

func TestTypedTaskContextPolicyRevokesPendingAuthority(t *testing.T) {
	for _, scenario := range []string{"healthy", "provider denied", "exclusion changed", "policy cycle"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			st, e, p, ctx := pendingResultState(t, "work")
			st.TaskContextPolicyEpochs = map[string]int{}
			if scenario == "provider denied" {
				st.Rooms[0].ContextExport = &room.ContextExportPolicy{Providers: []string{}}
				st.TaskContextPolicyEpochs[e.RoomID] = 1
			}
			if scenario == "exclusion changed" {
				st.Rooms[0].ContextExport = &room.ContextExportPolicy{Providers: []string{"openai"}, ExcludeMessages: true}
				st.TaskContextPolicyEpochs[e.RoomID] = 1
			}
			if scenario == "policy cycle" {
				st.Rooms[0].ContextExport = &room.ContextExportPolicy{Providers: []string{"openai"}}
				st.TaskContextPolicyEpochs[e.RoomID] = 2
			}
			if taskDefinitionContextAllowed(ctx, st, p) != (scenario == "healthy") {
				t.Fatal("stale context policy authority")
			}
		})
	}
}
func TestContextPolicyCycleWithholdsTypedResultAndNativeSession(t *testing.T) {
	for _, cycle := range []bool{false, true} {
		t.Run(strings.Join([]string{"policy", map[bool]string{true: "cycle", false: "deny"}[cycle]}, "-"), func(t *testing.T) {
			t.Parallel()
			s, db, h, task := structuredTaskFixture(t)
			s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
				db.seed("room.context_export.updated", map[string]any{"roomId": "engineering", "policy": room.ContextExportPolicy{Providers: []string{}}})
				if cycle {
					db.seed("room.context_export.updated", map[string]any{"roomId": "engineering", "policy": room.ContextExportPolicy{Providers: []string{"openai"}}})
				}
				if s.recordAgentSession(ctx, p, "development", "MUST_WITHHOLD_POLICY_SESSION") == nil {
					t.Error("stale policy session persisted")
				}
				return `{"result":"MUST_WITHHOLD_POLICY_RESULT"}`, nil
			}
			if w := dispatchRequest(h, map[bool]string{true: "context-cycle", false: "context-deny"}[cycle], definedDispatchBody(task)); w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			st := waitDispatch(t, s)
			for _, e := range st.TaskExecutions {
				if e.Status != "stopped" || len(e.Outputs) > 0 {
					t.Fatal("policy-revoked result released", e)
				}
			}
			for _, m := range st.Messages {
				if strings.Contains(m.Text, "MUST_WITHHOLD_POLICY_RESULT") {
					t.Fatal("policy-revoked message released")
				}
			}
			for _, p := range st.Participants {
				if p.ThreadID == "MUST_WITHHOLD_POLICY_SESSION" {
					t.Fatal("policy-revoked thread persisted")
				}
			}
		})
	}
}

func TestPersonalRoomContextPolicySurvivesReplay(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("actor.revised", room.Actor{ID: "synthetic-human", Name: "Synthetic human", Kind: "human", Enabled: true})
	db.seed("feature.flag.updated", map[string]any{"name": "contextPolicyEditor", "enabled": true})
	w := mockRequest(h, "POST", "/api/events?room=personal-synthetic-human", `{"kind":"room.context_export.updated","payload":{"policy":{"providers":[]}}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range st.Rooms {
		if r.ID == "personal-synthetic-human" {
			found = r.ContextExport != nil && len(r.ContextExport.Providers) == 0 && st.TaskContextPolicyEpochs[r.ID] == 1
		}
	}
	if !found {
		t.Fatal("acknowledged personal policy lost during replay")
	}
}
