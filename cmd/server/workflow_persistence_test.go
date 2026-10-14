package main

import (
	"agentroom/internal/room"
	"context"
	"testing"
	"time"
)

func TestWorkflowTransitionRollbackStopsExecution(t *testing.T) {
	for _, stage := range []string{"begin", "first", "second", "third", "commit"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			s, db, _ := mockServer(t)
			baseline := len(db.snapshot())
			switch stage {
			case "first":
				db.failInsertAt = 1
			case "second":
				db.failInsertAt = 2
			case "third":
				db.failInsertAt = 3
			default:
				db.failAt = stage
			}
			cancelled := false
			s.workflowLive = map[string]context.CancelFunc{"run": func() { cancelled = true }}
			run := room.WorkflowRun{ID: "run", Revision: 4, Status: "failed"}
			s.mu.Lock()
			err := s.persistWorkflowTransitionLocked(&run,
				auditChange{Kind: "agent.error", Payload: map[string]string{"participantId": "dev", "error": "failure"}},
				auditChange{Kind: "participant.session", Payload: room.Participant{ID: "dev", Status: "Blocked"}})
			s.mu.Unlock()
			if err == nil || !s.integrityFault || !cancelled {
				t.Fatal("storage failure did not stop execution")
			}
			if len(db.snapshot()) != baseline || run.Revision != 4 {
				t.Fatal("partial transition persisted or local revision advanced")
			}
		})
	}
}

func TestWorkflowTransitionCommitsOutcomeTogether(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	baseline := len(db.snapshot())
	run := room.WorkflowRun{ID: "run", Revision: 4, Status: "running", NodeID: "review"}
	s.mu.Lock()
	err := s.persistWorkflowTransitionLocked(&run, auditChange{Kind: "participant.session", Payload: room.Participant{ID: "dev", Status: "Idle"}})
	s.mu.Unlock()
	events := db.snapshot()[baseline:]
	if err != nil || len(events) != 2 || events[0].kind != "workflow.run.updated" || events[1].kind != "participant.session" || run.Revision != 5 || s.integrityFault {
		t.Fatal("run and participant outcome did not commit together", err)
	}
}

func TestWorkflowOutcomeFailurePreventsNextAgentDispatch(t *testing.T) {
	t.Parallel()
	s, db, h := executionServer(t)
	calls := 0
	s.workflowTurn = func(_ context.Context, _ room.Participant, _, _ string) (string, error) {
		calls++
		db.failInsertAt = 2 // Fail the participant outcome after staging the run result.
		return "Implementation complete", nil
	}
	started := startExecution(t, h)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		live := s.workflowLive[started.ID] != nil
		fault := s.integrityFault
		s.mu.Unlock()
		if !live {
			if !fault || calls != 1 {
				t.Fatal("failed persistence did not fence next dispatch")
			}
			st, err := s.load(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			for _, run := range st.WorkflowRuns {
				if run.ID == started.ID && (len(run.Steps) != 1 || run.NodeID != "agent") {
					t.Fatal("partial result advanced workflow")
				}
			}
			for _, participant := range st.Participants {
				if participant.ID == "dev" && participant.Status != "Working" {
					t.Fatal("partial result changed participant")
				}
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("workflow did not stop")
}
