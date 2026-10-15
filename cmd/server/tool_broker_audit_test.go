package main

import (
	"agentroom/internal/toolbroker"
	"strings"
	"testing"
	"time"
)

func sampleToolScope(t *testing.T) toolbroker.Scope {
	t.Helper()
	return toolbroker.Scope{Version: 1, JobID: strings.Repeat("a", 32), Room: "room-1", Participant: "participant-1", Connection: "connection-1", Provider: "codex", RepositoryDigest: strings.Repeat("b", 64), Deadline: time.Now().Add(2 * time.Minute).UTC().Round(0)}
}
func sampleIntent(t *testing.T) toolbroker.Intent {
	t.Helper()
	return toolbroker.Intent{Scope: sampleToolScope(t), CallID: "call-1", Operation: "read", PayloadDigest: strings.Repeat("c", 64)}
}

func TestLegacyToolCallHistoryRemainsReadable(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	intent := sampleIntent(t)
	outcome := toolbroker.Outcome{Status: "succeeded", OutputBytes: 12, OutputDigest: strings.Repeat("d", 64)}
	db.seed("tool.call.intent", intent)
	db.seed("tool.call.outcome", toolCallOutcomeRecord{Intent: intent, Outcome: outcome})
	st, err := s.load(t.Context())
	if err != nil || st.ToolCalls[intent.CallID].Status != "succeeded" {
		t.Fatal("legacy evidence lost", err)
	}
	if len(st.ToolAudits) != 0 {
		t.Fatal("legacy evidence gained scoped authority")
	}
}

func TestToolCallOutcomeWithoutIntentFailsReplay(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	intent := sampleIntent(t)
	outcome := toolbroker.Outcome{Status: "succeeded", OutputBytes: 1, OutputDigest: strings.Repeat("f", 64)}
	db.seed("tool.call.outcome", toolCallOutcomeRecord{Intent: intent, Outcome: outcome})
	if _, err := s.load(t.Context()); err == nil {
		t.Fatal("orphaned historical outcome accepted")
	}
}
