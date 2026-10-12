package main

import (
	"agentroom/internal/room"
	"context"
	"strings"
	"testing"
)

func TestBlockedStatusUsesSafeFailureAndRecordedActivity(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("participant.created", room.Participant{ID: "agent", RoomID: "engineering", Status: "Idle"})
	db.seed("agent.activity", map[string]string{"participantId": "agent", "activity": "Workflow: Search / Implement"})
	db.seed("agent.error", map[string]string{"participantId": "agent", "error": "context deadline exceeded; token=synthetic-do-not-display; /private/synthetic/path"})
	db.seed("participant.session", room.Participant{ID: "agent", RoomID: "engineering", Status: "Blocked"})
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	d := st.Participants[0].StatusDetails
	if d == nil || d.Activity != "Workflow: Search / Implement" || !strings.Contains(d.Reason, "time limit") {
		t.Fatal(d)
	}
	if strings.Contains(d.Reason+d.Next, "synthetic") || strings.Contains(d.Reason, "token") {
		t.Fatal("raw error exposed", d)
	}
	db.seed("participant.updated", room.Participant{ID: "agent", RoomID: "engineering", Status: "Blocked", Name: "New name"})
	st, _ = s.load(context.Background())
	if !strings.Contains(st.Participants[0].StatusDetails.Reason, "time limit") {
		t.Fatal("profile edit loses diagnostic")
	}
	db.seed("participant.session", room.Participant{ID: "agent", RoomID: "engineering", Status: "Working"})
	st, _ = s.load(context.Background())
	if strings.Contains(st.Participants[0].StatusDetails.Reason, "time limit") {
		t.Fatal("stale failure survives new turn")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "agentStatusDetails", "enabled": false})
	st, _ = s.load(context.Background())
	if st.Participants[0].StatusDetails != nil {
		t.Fatal("disabled server exposes details")
	}
}

func TestAgentStatusHistoricalFallbackScopeAndSafetyPrecedence(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("participant.created", room.Participant{ID: "one", RoomID: "engineering", Status: "Blocked"})
	db.seed("participant.created", room.Participant{ID: "two", RoomID: "private", Status: "Blocked"})
	st, _ := s.load(context.Background())
	if !strings.Contains(st.Participants[0].StatusDetails.Activity, "not recorded") || !strings.Contains(st.Participants[0].StatusDetails.Reason, "no detailed") {
		t.Fatal("invented historical cause")
	}
	if len(scoped(st, "engineering").Participants) != 1 {
		t.Fatal("foreign diagnostic visible")
	}
	db.seed("agents.stop", map[string]string{"scope": "room", "roomId": "engineering"})
	st, _ = s.load(context.Background())
	if st.Participants[0].StatusDetails.Reason != "Emergency stop is active." || st.Participants[1].StatusDetails.Reason == "Emergency stop is active." {
		t.Fatal("wrong stop scope")
	}
	db.seed("incident.opened", map[string]string{"reason": "mismatch"})
	st, _ = s.load(context.Background())
	if !strings.Contains(st.Participants[0].StatusDetails.Reason, "integrity incident") {
		t.Fatal("incident not explained")
	}
}

func TestAgentFailureCategoriesNeverReturnRawDiagnostics(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"bwrap: No permissions to create a new namespace", "read-only workspace", "Login failed: proxy refused the connection", "credential unavailable", "Docker not ready", "Review must return explicit passed/changes JSON", "Workflow iteration limit reached", "Step output exceeds 32 KiB", "feature disabled", "integrity unavailable", "unrecognized token=synthetic-secret"} {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			reason, next := safeAgentFailure(raw)
			if reason == "" || next == "" || strings.Contains(reason+next, "synthetic-secret") {
				t.Fatal(reason, next)
			}
		})
	}
	if len([]rune(statusActivity(strings.Repeat("A", 300)))) != 181 {
		t.Fatal("activity not bounded")
	}
	if strings.Contains(statusActivity("Task\nname"), "\n") {
		t.Fatal("control character retained")
	}
}
