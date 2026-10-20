package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func scopedAuditFixture(t *testing.T) (*server, *mockDB, *controllerToolAuditor, toolbroker.Intent) {
	t.Helper()
	return scopedAuditFixtureWithLifetime(t, time.Minute)
}

func scopedAuditFixtureWithLifetime(t *testing.T, lifetime time.Duration) (*server, *mockDB, *controllerToolAuditor, toolbroker.Intent) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "connection", Name: "Synthetic provider", Protocol: "CLI", Provider: "openai", Endpoint: "ssh://worker", Revision: 1})
	db.seed("participant.created", room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection", Revision: 1})
	db.seed("actor.revised", room.Actor{ID: "human", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}, Revision: 1})
	now := time.Now().UTC().Round(0)
	scope := toolbroker.Scope{Version: 1, JobID: strings.Repeat("a", 32), Room: "engineering", Participant: "agent", Connection: "connection", Provider: "codex", RepositoryDigest: strings.Repeat("b", 64), Deadline: now.Add(lifetime)}
	db.seed("worker.job.updated", room.ManagedJob{Version: 2, InitiatorID: "human", CompanyID: "company", JobID: scope.JobID, RoomID: scope.Room, ParticipantID: scope.Participant, ConnectionID: scope.Connection, Worker: "worker", Provider: "codex", Status: "submitted", AllocatedAt: now, UpdatedAt: now, Deadline: scope.Deadline})
	auditor, err := newControllerToolAuditor(context.Background(), s, scope, "human", func() (string, error) { return scope.RepositoryDigest, nil })
	if err != nil {
		t.Fatal("synthetic auditor refused", err)
	}
	return s, db, auditor, toolbroker.Intent{Scope: scope, CallID: "call", Operation: "read", PayloadDigest: strings.Repeat("c", 64)}
}

func TestControllerToolAuditPersistsScopeRealActorAndExactlyOneOutcome(t *testing.T) {
	t.Parallel()
	s, db, auditor, intent := scopedAuditFixture(t)
	for _, flag := range []string{"jobDiagnostics", "agentBackgroundWork", "contextPolicyEditor"} {
		db.seed("feature.flag.updated", map[string]any{"name": flag, "enabled": false})
	}
	if auditor.RecordIntent(context.Background(), intent) != nil {
		t.Fatal("intent refused")
	}
	if auditor.RecordIntent(context.Background(), intent) != toolbroker.ErrAudit {
		t.Fatal("intent replay accepted")
	}
	other := intent
	other.CallID = "another"
	if auditor.RecordIntent(context.Background(), other) != toolbroker.ErrAudit {
		t.Fatal("unresolved call permitted overlapping work")
	}
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	call := st.ToolAudits[toolAuditKey(intent.Scope.JobID, intent.CallID)]
	if call.Status != "pending" || call.ActorID != "human" || call.ActorID == ownerID || call.CompanyID != st.Company.ID {
		t.Fatal("wrong principal/scope", call)
	}
	for _, event := range db.snapshot() {
		if event.kind == "worker.tool.intent" {
			var metadata struct {
				PerformedBy string `json:"performedBy"`
			}
			if json.Unmarshal([]byte(event.payload), &metadata) != nil || metadata.PerformedBy != "human" || strings.Contains(event.payload, "capability") {
				t.Fatal("missing real attribution or leaked capability")
			}
		}
	}
	outcome := toolbroker.Outcome{Status: "succeeded", OutputBytes: 2, OutputDigest: strings.Repeat("d", 64)}
	if auditor.RecordOutcome(context.Background(), intent, outcome) != nil {
		t.Fatal("outcome refused")
	}
	if auditor.RecordOutcome(context.Background(), intent, outcome) != toolbroker.ErrAudit {
		t.Fatal("outcome replay accepted")
	}
	if auditor.RecordIntent(context.Background(), other) != nil {
		t.Fatal("completed call blocked the next explicit action")
	}
	st, err = s.load(context.Background())
	if err != nil || st.ToolAudits[toolAuditKey(intent.Scope.JobID, intent.CallID)].Status != "succeeded" {
		t.Fatal("outcome not projected", err)
	}
}

func TestControllerToolAuditRevalidatesBeforeEveryAcknowledgement(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"disabled-human", "room-access", "role-revision", "participant-revision", "connection-revision", "removed-participant", "repository-disabled", "repository-changed", "job-ended", "job-initiator", "company-binding", "stop", "context-policy", "wrong-scope"} {
		t.Run(mode, func(t *testing.T) {
			s, db, auditor, intent := scopedAuditFixture(t)
			if auditor.RecordIntent(context.Background(), intent) != nil {
				t.Fatal("initial intent refused")
			}
			st, _ := s.load(context.Background())
			switch mode {
			case "disabled-human", "room-access", "role-revision":
				actor, _ := member(st, "human")
				actor.Revision++
				if mode == "disabled-human" {
					actor.Enabled = false
				}
				if mode == "room-access" {
					actor.RoomIDs = nil
				}
				if mode == "role-revision" {
					actor.Role = "QA"
				}
				db.seed("actor.revised", actor)
			case "participant-revision":
				p := st.Participants[0]
				p.Revision++
				db.seed("participant.updated", p)
			case "connection-revision":
				c := st.Connections[0]
				c.Revision++
				db.seed("connection.updated", c)
			case "removed-participant":
				db.seed("participant.removed", map[string]string{"id": "agent"})
			case "repository-disabled":
				auditor.repository = func() (string, error) { return "", nil }
			case "repository-changed":
				auditor.repository = func() (string, error) { return strings.Repeat("f", 64), nil }
			case "job-ended":
				job := st.ManagedJobs[intent.Scope.JobID]
				job.Status = "shutdown_confirmed"
				job.ShutdownConfirmed = true
				db.seed("worker.job.updated", job)
			case "job-initiator":
				job := st.ManagedJobs[intent.Scope.JobID]
				job.InitiatorID = "another-human"
				db.seed("worker.job.updated", job)
			case "company-binding":
				company := st.Company
				company.ID = "another-company"
				db.seed("company.updated", company)
			case "stop":
				s.stopAll = true
			case "context-policy":
				for _, r := range st.Rooms {
					if r.ID == "engineering" {
						r.ContextExport = &room.ContextExportPolicy{Providers: []string{"claude"}}
						db.seed("room.updated", r)
					}
				}
			case "wrong-scope":
				intent.Scope.Connection = "another"
			}
			before := len(db.snapshot())
			if auditor.RecordOutcome(context.Background(), intent, toolbroker.Outcome{Status: "succeeded", OutputBytes: 2, OutputDigest: strings.Repeat("d", 64)}) != toolbroker.ErrAudit || len(db.snapshot()) != before {
				t.Fatal("revoked/stale scope acknowledged")
			}
			intent.CallID = "next"
			if auditor.RecordIntent(context.Background(), intent) != toolbroker.ErrAudit {
				t.Fatal("revoked/stale scope admitted new tool")
			}
		})
	}
}

type scopedAuditExecutor func(context.Context, toolbroker.Scope, toolbroker.Action, io.Writer) error

func (executor scopedAuditExecutor) Execute(ctx context.Context, scope toolbroker.Scope, action toolbroker.Action, out io.Writer) error {
	return executor(ctx, scope, action, out)
}

func TestScopedAuditTransportUsesRealStoreBeforeMockToolAndHidesPayloads(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			s, db, _, intent := scopedAuditFixture(t)
			t.Setenv("COMMONROOM_DEVELOPMENT_WORKSPACES", map[bool]string{false: "false", true: "true"}[enabled])
			t.Setenv("COMMONROOM_DEVELOPMENT_REPOSITORIES", `{"engineering":"`+intent.Scope.RepositoryDigest+`"}`)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx = context.WithValue(ctx, actorContextKey{}, "human")
			ctx = context.WithValue(ctx, managedJobContextKey{}, managedJobRegistration{JobID: intent.Scope.JobID})
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			done := make(chan error, 1)
			go func() { done <- s.serveToolAudit(ctx, intent.Scope, right) }()
			client, err := toolbroker.NewAuditClient(left, intent.Scope)
			if err != nil {
				t.Fatal(err)
			}
			authority, token, err := toolbroker.New(intent.Scope, []string{"write"}, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			executed := false
			dispatcher, err := toolbroker.NewDispatcher(authority, client, scopedAuditExecutor(func(_ context.Context, _ toolbroker.Scope, _ toolbroker.Action, out io.Writer) error {
				st, err := s.load(context.Background())
				if err != nil {
					return err
				}
				call := st.ToolAudits[toolAuditKey(intent.Scope.JobID, "call")]
				if call.Status != "pending" || call.ActorID != "human" {
					return errors.New("tool before durable human-bound intent")
				}
				executed = true
				_, err = io.WriteString(out, "synthetic-private-output")
				return err
			}))
			if err != nil {
				t.Fatal(err)
			}
			result, err := dispatcher.Dispatch(ctx, toolbroker.Request{Scope: intent.Scope, Capability: token, CallID: "call", Operation: "write", Path: "private-fixture.txt", Content: "synthetic-private-content"})
			if enabled {
				if err != nil || string(result) != "synthetic-private-output" || !executed {
					t.Fatal("enabled audit stream failed", err)
				}
			} else {
				if err != toolbroker.ErrAudit || result != nil || executed {
					t.Fatal("disabled feature ran tool", err)
				}
			}
			left.Close()
			if <-done != toolbroker.ErrAudit {
				t.Fatal("closed stream reported success")
			}
			for _, event := range db.snapshot() {
				for _, secret := range []string{token, "synthetic-private-content", "synthetic-private-output", "private-fixture.txt"} {
					if strings.Contains(event.payload, secret) {
						t.Fatal("raw payload/capability entered audit")
					}
				}
			}
			st, err := s.load(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if enabled && st.ToolAudits[toolAuditKey(intent.Scope.JobID, "call")].Status != "succeeded" {
				t.Fatal("result before committed outcome")
			}
		})
	}
}

func TestManagedJobRecordsRealInitiatorAndNeverPromotesLegacyJobsForTools(t *testing.T) {
	t.Parallel()
	s, _, _, intent := scopedAuditFixture(t)
	job := room.ManagedJob{}
	st, _ := s.load(context.Background())
	job = st.ManagedJobs[intent.Scope.JobID]
	job.InitiatorID = ""
	legacy, legacyDB, _ := mockServer(t)
	legacyDB.seed("worker.job.updated", job)
	if _, err := legacy.load(context.Background()); err != nil {
		t.Fatal("legacy metadata became unreadable", err)
	}
	if _, err := newControllerToolAuditor(context.Background(), legacy, intent.Scope, ownerID, func() (string, error) { return intent.Scope.RepositoryDigest, nil }); err != toolbroker.ErrAudit {
		t.Fatal("legacy job acquired a human implicitly")
	}
	fresh, db2, _ := mockServer(t)
	db2.seed("participant.created", room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection"})
	db2.seed("actor.revised", room.Actor{ID: "human", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}, Revision: 1})
	ctx := context.WithValue(context.Background(), actorContextKey{}, "human")
	registration := fresh.managedJobContext(ctx, room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection"}).Value(managedJobContextKey{}).(managedJobRegistration)
	if err := registration.Register("worker", "codex"); err != nil {
		t.Fatal(err)
	}
	st, _ = fresh.load(context.Background())
	record := st.ManagedJobs[registration.JobID]
	if record.InitiatorID != "human" || record.CompanyID != st.Company.ID {
		t.Fatal("initiator not captured durably")
	}
}

func TestControllerToolAuditFailuresLeaveNoSuccessAndNoOwnerFallback(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"intent", "outcome"} {
		t.Run(stage, func(t *testing.T) {
			s, db, auditor, intent := scopedAuditFixture(t)
			if stage == "outcome" && auditor.RecordIntent(context.Background(), intent) != nil {
				t.Fatal("setup intent failed")
			}
			db.failAt = "commit"
			var err error
			if stage == "intent" {
				err = auditor.RecordIntent(context.Background(), intent)
			} else {
				err = auditor.RecordOutcome(context.Background(), intent, toolbroker.Outcome{Status: "succeeded", OutputBytes: 2, OutputDigest: strings.Repeat("d", 64)})
			}
			if err != toolbroker.ErrAudit || !s.integrityFault {
				t.Fatal("persistence failure acknowledged", err)
			}
			for _, event := range db.snapshot() {
				if event.kind == "worker.tool.outcome" {
					t.Fatal("failed commit published success")
				}
			}
		})
	}
	s, _, auditor, intent := scopedAuditFixture(t)
	if _, err := newControllerToolAuditor(context.Background(), s, intent.Scope, "", auditor.repository); err != toolbroker.ErrAudit {
		t.Fatal("missing human inherited owner")
	}
	if _, err := s.toolAuditor(context.Background(), intent.Scope); err != toolbroker.ErrAudit {
		t.Fatal("unregistered callback admitted")
	}
}

func TestToolAuditProjectionRejectsOrphanAndMutatedOutcome(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"orphan", "scope", "replay"} {
		t.Run(mode, func(t *testing.T) {
			s, db, auditor, intent := scopedAuditFixture(t)
			if auditor.RecordIntent(context.Background(), intent) != nil {
				t.Fatal("setup intent failed")
			}
			st, _ := s.load(context.Background())
			call := st.ToolAudits[toolAuditKey(intent.Scope.JobID, intent.CallID)]
			call.Status = "succeeded"
			call.OutputBytes = 2
			call.OutputDigest = strings.Repeat("d", 64)
			if mode == "orphan" {
				call.CallID = "unknown"
			}
			if mode == "scope" {
				call.ActorID = ownerID
			}
			db.seed("worker.tool.outcome", call)
			if mode == "replay" {
				db.seed("worker.tool.outcome", call)
			}
			if _, err := s.load(context.Background()); err == nil {
				t.Fatal("invalid audit transition projected")
			}
		})
	}
}
