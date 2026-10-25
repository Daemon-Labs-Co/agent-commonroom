package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"fmt"
	"testing"
)

func TestConnectionGuardRetainsDurableWorkAcrossControllerRestart(t *testing.T) {
	for _, status := range []string{"allocated", "submitted", "shutdown_unconfirmed"} {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			st := room.State{ManagedJobs: map[string]room.ManagedJob{"job": {ParticipantID: "worker", ConnectionID: "connection", Status: status}}, FeatureFlags: map[string]bool{"jobDiagnostics": false, "taskDispatch": false, "taskDefinitions": false}}
			if !connectionHasActiveAgent(st, nil, "connection") || !participantHasUnsettledWork(st, nil, "worker") {
				t.Fatal("durable job lost after restart")
			}
			if connectionHasActiveAgent(st, nil, "other") || participantHasUnsettledWork(st, nil, "other") {
				t.Fatal("foreign identity blocked")
			}
			job := st.ManagedJobs["job"]
			job.ShutdownConfirmed = true
			st.ManagedJobs["job"] = job
			if connectionHasActiveAgent(st, nil, "connection") || participantHasUnsettledWork(st, nil, "worker") {
				t.Fatal("verified cleanup did not release guard")
			}
		})
	}
}
func TestConnectionGuardFencesPendingTaskActualRole(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for _, mode := range []string{"work", "review"} {
			t.Run(fmt.Sprint(typed)+mode, func(t *testing.T) {
				t.Parallel()
				task := room.Task{ID: "task", Assignee: "worker", Reviewer: "reviewer", Status: "Working", DispatchID: "attempt", DispatchStatus: "pending"}
				pid, cid := "worker", "work-connection"
				if mode == "review" {
					task.Status = "Reviewing"
					pid, cid = "reviewer", "review-connection"
				}
				st := room.State{Tasks: []room.Task{task}, Participants: []room.Participant{{ID: "worker", ConnectionID: "work-connection"}, {ID: "reviewer", ConnectionID: "review-connection"}}}
				if typed {
					st.TaskExecutions = map[string]room.TaskExecution{"attempt": {ParticipantID: pid}}
				}
				if !connectionHasActiveAgent(st, nil, cid) || !participantHasUnsettledWork(st, nil, pid) {
					t.Fatal("pending task lost after restart")
				}
				other := "reviewer"
				if pid == other {
					other = "worker"
				}
				if participantHasUnsettledWork(st, nil, other) {
					t.Fatal("inactive other role blocked")
				}
				st.Tasks[0].DispatchStatus = "failed"
				if connectionHasActiveAgent(st, nil, cid) {
					t.Fatal("reconciled task still blocked")
				}
			})
		}
	}
}
func TestConnectionMutationsWaitForDurableCleanup(t *testing.T) {
	for _, operation := range []string{"connection.updated", "connection.deleted", "participant.updated"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			s, db, h, reg := syntheticManagedJob(t)
			db.seed("connection.created", room.Connection{ID: "connection", Name: "Synthetic original", Protocol: "CLI", Endpoint: "ssh://worker/codex", Revision: 1})
			db.seed("connection.created", room.Connection{ID: "other", Name: "Synthetic replacement", Protocol: "CLI", Endpoint: "ssh://worker/codex", Revision: 1})
			// Simulate loss of process-local bookkeeping while retaining only the journal.
			s.active = map[string]bool{}
			s.activeJobs = map[string]string{}
			var payload any = room.Connection{ID: "connection", Name: "Synthetic edited", Protocol: "CLI", Endpoint: "ssh://worker/codex", Revision: 1}
			if operation == "connection.deleted" {
				payload = map[string]string{"id": "connection"}
			} else if operation == "participant.updated" {
				payload = room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "other", Name: "Synthetic", Rank: 1}
			}
			raw, _ := json.Marshal(payload)
			body := fmt.Sprintf(`{"kind":%q,"payload":%s}`, operation, raw)
			before := len(db.snapshot())
			w := mockRequest(h, "POST", "/api/events", body)
			if w.Code != 409 || len(db.snapshot()) != before {
				t.Fatal("unconfirmed binding mutated", w.Code, w.Body.String())
			}
			if err := reg.Update("shutdown_confirmed"); err != nil {
				t.Fatal(err)
			}
			w = mockRequest(h, "POST", "/api/events", body)
			if w.Code != 200 {
				t.Fatal("verified cleanup did not unblock safe edit", w.Code, w.Body.String())
			}
		})
	}
}
