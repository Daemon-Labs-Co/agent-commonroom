package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"strings"
	"testing"
	"time"
)

func scopedAuditFixture(t *testing.T) (*server, *mockDB, *controllerToolAuditor, toolbroker.Intent) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "connection", Name: "Synthetic provider", Protocol: "CLI", Provider: "openai", Endpoint: "ssh://worker", Revision: 1})
	db.seed("participant.created", room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection", Revision: 1})
	db.seed("actor.revised", room.Actor{ID: "human", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}, Revision: 1})
	now := time.Now().UTC().Round(0)
	scope := toolbroker.Scope{Version: 1, JobID: strings.Repeat("a", 32), Room: "engineering", Participant: "agent", Connection: "connection", Provider: "codex", RepositoryDigest: strings.Repeat("b", 64), Deadline: now.Add(time.Minute)}
	db.seed("worker.job.updated", room.ManagedJob{Version: 2, JobID: scope.JobID, RoomID: scope.Room, ParticipantID: scope.Participant, ConnectionID: scope.Connection, Worker: "worker", Provider: "codex", Status: "submitted", AllocatedAt: now, UpdatedAt: now, Deadline: scope.Deadline})
	auditor, err := newControllerToolAuditor(context.Background(), s, scope, "human", func() (string, error) { return scope.RepositoryDigest, nil })
	if err != nil {
		t.Fatal("synthetic auditor refused", err)
	}
	return s, db, auditor, toolbroker.Intent{Scope: scope, CallID: "call", Operation: "read", PayloadDigest: strings.Repeat("c", 64)}
}

func TestControllerToolAuditPersistsScopeRealActorAndExactlyOneOutcome(t *testing.T) {
	t.Parallel()
	s, db, auditor, intent := scopedAuditFixture(t)
	if auditor.RecordIntent(context.Background(), intent) != nil {
		t.Fatal("intent refused")
	}
	if auditor.RecordIntent(context.Background(), intent) != toolbroker.ErrAudit {
		t.Fatal("intent replay accepted")
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
		if event.kind == "worker.tool.intent" && (!strings.Contains(event.payload, `"performedBy":"human"`) || strings.Contains(event.payload, "capability")) {
			t.Fatal("missing real attribution or leaked capability")
		}
	}
	outcome := toolbroker.Outcome{Status: "succeeded", OutputBytes: 2, OutputDigest: strings.Repeat("d", 64)}
	if auditor.RecordOutcome(context.Background(), intent, outcome) != nil {
		t.Fatal("outcome refused")
	}
	if auditor.RecordOutcome(context.Background(), intent, outcome) != toolbroker.ErrAudit {
		t.Fatal("outcome replay accepted")
	}
	st, err = s.load(context.Background())
	if err != nil || st.ToolAudits[toolAuditKey(intent.Scope.JobID, intent.CallID)].Status != "succeeded" {
		t.Fatal("outcome not projected", err)
	}
}

func TestControllerToolAuditRevalidatesBeforeEveryAcknowledgement(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"disabled-human", "room-access", "role-revision", "participant-revision", "connection-revision", "removed-participant", "repository-disabled", "repository-changed", "job-ended", "stop", "context-policy", "wrong-scope"} {
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
