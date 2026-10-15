package main

import (
	"strings"
	"testing"
	"time"

	"agentroom/internal/toolbroker"
)

func sampleToolScope(t *testing.T) toolbroker.Scope {
	t.Helper()
	return toolbroker.Scope{
		Version: 1, JobID: strings.Repeat("a", 32), Room: "room-1", Participant: "participant-1",
		Connection: "connection-1", Provider: "codex", RepositoryDigest: strings.Repeat("b", 64),
		Deadline: time.Now().Add(2 * time.Minute).UTC().Round(0),
	}
}
func sampleIntent(t *testing.T) toolbroker.Intent {
	t.Helper()
	return toolbroker.Intent{Scope: sampleToolScope(t), CallID: "call-1", Operation: "read", PayloadDigest: strings.Repeat("c", 64)}
}

func TestStoredToolAuditorRecordsIntentAndOutcome(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := &server{db: db}
	auditor := &storedToolAuditor{s: s}
	ctx := t.Context()
	intent := sampleIntent(t)

	if err := auditor.RecordIntent(ctx, intent); err != nil {
		t.Fatal(err)
	}
	st, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := st.ToolCalls[intent.CallID]
	if !ok || got.Status != "pending" || got.JobID != intent.Scope.JobID || got.Operation != "read" || got.PayloadDigest != intent.PayloadDigest {
		t.Fatalf("intent not replayed correctly: %+v", got)
	}

	outcome := toolbroker.Outcome{Status: "succeeded", OutputBytes: 12, OutputDigest: strings.Repeat("d", 64)}
	if err := auditor.RecordOutcome(ctx, intent, outcome); err != nil {
		t.Fatal(err)
	}
	st, err = s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got = st.ToolCalls[intent.CallID]
	if got.Status != "succeeded" || got.OutputBytes != 12 || got.OutputDigest != outcome.OutputDigest {
		t.Fatalf("outcome not merged into existing intent record: %+v", got)
	}
}

// Outcomes legitimately arrive at or after the intent's deadline (a call that
// ran right up to its time limit); RecordOutcome must not reject that.
func TestStoredToolAuditorRecordOutcomeAfterDeadline(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := &server{db: db}
	auditor := &storedToolAuditor{s: s}
	ctx := t.Context()
	scope := sampleToolScope(t)
	scope.Deadline = time.Now().Add(50 * time.Millisecond).UTC().Round(0)
	intent := toolbroker.Intent{Scope: scope, CallID: "call-2", Operation: "write", PayloadDigest: strings.Repeat("e", 64)}
	if err := auditor.RecordIntent(ctx, intent); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	outcome := toolbroker.Outcome{Status: "failed", OutputBytes: 0, OutputDigest: strings.Repeat("0", 64)}
	if err := auditor.RecordOutcome(ctx, intent, outcome); err != nil {
		t.Fatalf("outcome after deadline must still be recorded: %v", err)
	}
}

func TestStoredToolAuditorRejectsInvalidOutcome(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := &server{db: db}
	auditor := &storedToolAuditor{s: s}
	intent := sampleIntent(t)
	if err := auditor.RecordIntent(t.Context(), intent); err != nil {
		t.Fatal(err)
	}
	if err := auditor.RecordOutcome(t.Context(), intent, toolbroker.Outcome{Status: "unknown-status"}); err == nil {
		t.Fatal("expected rejection of an invalid outcome status")
	}
}

func TestStoredToolAuditorRejectsMalformedIntent(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := &server{db: db}
	auditor := &storedToolAuditor{s: s}
	bad := sampleIntent(t)
	bad.Operation = "delete-everything"
	if err := auditor.RecordIntent(t.Context(), bad); err == nil {
		t.Fatal("expected rejection of a malformed intent")
	}
	if len(db.snapshot()) != 0 {
		t.Fatal("malformed intent must not be persisted")
	}
}

func TestToolCallOutcomeWithoutIntentFailsReplay(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := &server{db: db}
	intent := sampleIntent(t)
	outcome := toolbroker.Outcome{Status: "succeeded", OutputBytes: 1, OutputDigest: strings.Repeat("f", 64)}
	if err := s.append(t.Context(), "tool.call.outcome", toolCallOutcomeRecord{Intent: intent, Outcome: outcome}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.load(t.Context()); err == nil {
		t.Fatal("expected replay to fail closed on an orphaned outcome")
	}
}
