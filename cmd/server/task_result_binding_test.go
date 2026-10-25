package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func pendingResultState(t *testing.T, mode string) (room.State, room.TaskExecution, room.Participant, context.Context) {
	t.Helper()
	d := syntheticDefinition()
	inputs, err := taskcontract.Bind(d, []taskcontract.Binding{{Name: "source", ValueJSON: `"synthetic"`, SourceVersion: "manual-v1"}})
	if err != nil {
		t.Fatal(err)
	}
	e := room.TaskExecution{ID: "attempt", CompanyID: "company", TaskID: "task", RoomID: "engineering", TaskRevision: 1, ParticipantID: "worker", InitiatorID: ownerID, Mode: mode, Definition: d, DefinitionDigest: taskcontract.Digest(d), Inputs: inputs, Outputs: []taskcontract.BoundValue{}, Status: "running", StartedAt: time.Now().UTC()}
	task := room.Task{ID: e.TaskID, RoomID: e.RoomID, Revision: 2, Title: "Synthetic task", Status: "Working", Assignee: "worker", Reviewer: "reviewer", DispatchID: e.ID, DispatchStatus: "pending", Definition: &d}
	if mode == "review" {
		task.Status = "Reviewing"
		e.ParticipantID = "reviewer"
		e.ParentExecutionID = "parent"
	}
	st := room.State{Company: room.Company{ID: "company"}, FeatureFlags: map[string]bool{"taskDefinitions": true}, Tasks: []room.Task{task}, TaskExecutions: map[string]room.TaskExecution{e.ID: e}}
	p := room.Participant{ID: e.ParticipantID, RoomID: e.RoomID, ConnectionID: "connection-" + e.ParticipantID, Revision: 1}
	st.Rooms = []room.Room{{ID: e.RoomID}}
	st.Participants = []room.Participant{p}
	st.Connections = []room.Connection{{ID: p.ConnectionID, Protocol: "CLI", Endpoint: "ssh://synthetic-worker", Revision: 1}}
	if err := bindTaskExecutionConfiguration(st, p, &e); err != nil {
		t.Fatal(err)
	}
	st.TaskExecutions[e.ID] = e
	ctx := context.WithValue(context.Background(), taskJobContextKey{}, taskJobBinding{TaskID: task.ID, DispatchID: e.ID, DefinitionDigest: e.DefinitionDigest})
	return st, e, p, ctx
}
func TestTypedResultContextRequiresExactExecutionRole(t *testing.T) {
	for _, mode := range []string{"work", "review"} {
		for _, scenario := range []string{"healthy", "wrong role", "wrong task state", "reconciled", "foreign execution company", "foreign execution task", "foreign execution room", "wrong execution id", "wrong execution participant", "wrong execution epoch", "wrong delegation", "wrong definition", "wrong task revision", "definitions off"} {
			t.Run(mode+"/"+scenario, func(t *testing.T) {
				t.Parallel()
				st, e, p, ctx := pendingResultState(t, mode)
				switch scenario {
				case "wrong role":
					if mode == "work" {
						st.Tasks[0].Assignee = "other"
					} else {
						st.Tasks[0].Reviewer = "other"
					}
				case "wrong task state":
					st.Tasks[0].Status = "Draft"
				case "reconciled":
					st.Tasks[0].DispatchStatus = "failed"
				case "foreign execution company":
					e.CompanyID = "foreign"
				case "foreign execution task":
					e.TaskID = "foreign"
				case "foreign execution room":
					e.RoomID = "foreign"
				case "wrong execution id":
					e.ID = "other"
				case "wrong execution participant":
					e.ParticipantID = "other"
				case "wrong execution epoch":
					e.DefinitionEpoch++
				case "wrong delegation":
					e.DelegationID = "other"
				case "wrong definition":
					e.DefinitionDigest = strings.Repeat("f", 64)
				case "wrong task revision":
					st.Tasks[0].Revision++
				case "definitions off":
					st.FeatureFlags["taskDefinitions"] = false
				}
				st.TaskExecutions["attempt"] = e
				allowed := taskDefinitionContextAllowed(ctx, st, p)
				if allowed != (scenario == "healthy") {
					t.Fatal("context authority", allowed)
				}
				{
					_, err := structuredTaskResult(st, st.Tasks[0], "Idle", `{"result":"synthetic"}`)
					if (err == nil) != (scenario == "healthy") {
						t.Fatal("result binding", err)
					}
				}
			})
		}
	}
}
func TestSuccessfulExecutionReplayRequiresCompletedTaskBinding(t *testing.T) {
	for _, mode := range []string{"work", "review"} {
		for _, scenario := range []string{"healthy", "pending task", "wrong revision", "wrong role", "wrong dispatch", "wrong task state"} {
			t.Run(mode+"/"+scenario, func(t *testing.T) {
				t.Parallel()
				st, e, _, _ := pendingResultState(t, mode)
				st.Tasks[0].DispatchStatus = "completed"
				st.Tasks[0].Revision = 3
				e.FinishedAt = e.StartedAt.Add(time.Second)
				e.ResponseDigest = responseDigest("synthetic response")
				e.Status = "reviewed"
				if mode == "work" {
					e.Status = "validated"
					var err error
					e.Outputs, err = taskcontract.ReturnedValues(e.Definition, []byte(`{"result":"synthetic"}`), e.ID)
					if err != nil {
						t.Fatal(err)
					}
				}
				switch scenario {
				case "pending task":
					st.Tasks[0].DispatchStatus = "pending"
				case "wrong revision":
					st.Tasks[0].Revision++
				case "wrong role":
					if mode == "work" {
						st.Tasks[0].Assignee = "other"
					} else {
						st.Tasks[0].Reviewer = "other"
					}
				case "wrong dispatch":
					st.Tasks[0].DispatchID = "other"
				case "wrong task state":
					st.Tasks[0].Status = "Assigned"
				}
				raw, _ := json.Marshal(e)
				err := projectTaskExecution(&st, "task.execution.finished", raw)
				if (err == nil) != (scenario == "healthy") {
					t.Fatal("replay binding", err)
				}
			})
		}
	}
}
func TestChangedTaskRoleWithholdsActualTurnResult(t *testing.T) {
	for _, scenario := range []string{"role", "state"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			s, db, h, task := structuredTaskFixture(t)
			s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
				st, err := s.load(t.Context())
				if err != nil {
					return "", err
				}
				changed := st.Tasks[0]
				if scenario == "role" {
					changed.Assignee = "other"
				} else {
					changed.Status = "Draft"
				}
				db.seed("task.revised", changed)
				return `{"result":"MUST_WITHHOLD_SYNTHETIC"}`, nil
			}
			if w := dispatchRequest(h, "result-binding-"+scenario, definedDispatchBody(task)); w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			st := waitDispatch(t, s)
			for _, e := range st.TaskExecutions {
				if e.Status != "stopped" || len(e.Outputs) > 0 {
					t.Fatal("revoked role released typed result", e)
				}
			}
			for _, m := range st.Messages {
				if strings.Contains(m.Text, "MUST_WITHHOLD_SYNTHETIC") {
					t.Fatal("revoked role published message")
				}
			}
		})
	}
}
