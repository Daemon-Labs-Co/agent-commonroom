package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTypedExecutionConfigurationFencesStableProfileAndConnection(t *testing.T) {
	for _, scenario := range []string{"healthy", "profile revision", "profile instructions", "participant reassigned", "connection revision", "connection endpoint", "connection provider", "connection removed", "participant removed", "session only", "derived only", "legacy pending", "unknown version", "partial legacy"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			st, e, p, ctx := pendingResultState(t, "work")
			expected := scenario == "healthy" || scenario == "session only" || scenario == "derived only"
			switch scenario {
			case "profile revision":
				st.Participants[0].Revision++
			case "profile instructions":
				st.Participants[0].Instructions = "Changed instructions"
			case "participant reassigned":
				st.Participants[0].ConnectionID = "other"
			case "connection revision":
				st.Connections[0].Revision++
			case "connection endpoint":
				st.Connections[0].Endpoint = "ssh://other"
			case "connection provider":
				st.Connections[0].Provider = "other"
			case "connection removed":
				st.Connections = nil
			case "participant removed":
				st.Participants = nil
			case "session only":
				st.Participants[0].Status = "Idle"
				st.Participants[0].ThreadID = "new-session"
				st.Participants[0].ExecutionMode = "development"
				st.Participants[0].DisclosureVersion = 9
				st.Participants[0].Background = []room.BackgroundActivity{{Stage: "running"}}
				st.Participants[0].StatusDetails = &room.AgentStatusDetails{}
			case "derived only":
				st.Participants[0].CompanyID = "company"
				st.Participants[0].ActorID = "bot-derived"
				st.Connections[0].CompanyID = "company"
			case "legacy pending":
				e.ConfigurationVersion = 0
				e.ParticipantRevision = 0
				e.ParticipantConfigurationDigest = ""
				e.ConnectionID = ""
				e.ConnectionRevision = 0
				e.ConnectionConfigurationDigest = ""
			case "unknown version":
				e.ConfigurationVersion = 2
			case "partial legacy":
				e.ConfigurationVersion = 0
			}
			st.TaskExecutions[e.ID] = e
			if taskDefinitionContextAllowed(ctx, st, p) != expected || taskExecutionPendingCurrent(st, st.Tasks[0], e) != expected {
				t.Fatal("configuration fence", scenario)
			}
			_, err := structuredTaskResult(st, st.Tasks[0], "Idle", `{"result":"synthetic"}`)
			if (err == nil) != expected {
				t.Fatal("result configuration", scenario, err)
			}
			if scenario == "legacy pending" {
				if !validTaskExecutionConfiguration(st, e) {
					t.Fatal("retained pre-configuration history rejected")
				}
			} else if !expected && validTaskExecutionConfiguration(st, e) {
				t.Fatal("changed/unknown configuration accepted")
			}
		})
	}
}
func TestConfigurationChangesWithholdRealMockedTaskResult(t *testing.T) {
	for _, scenario := range []string{"profile", "connection"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			s, db, h, task := structuredTaskFixture(t)
			s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
				st, err := s.load(t.Context())
				if err != nil {
					return "", err
				}
				if scenario == "profile" {
					p := st.Participants[0]
					p.Instructions = "Changed synthetic instructions"
					p.Revision++
					db.seed("participant.updated", p)
				} else {
					c := st.Connections[0]
					c.Endpoint = "ssh://changed-worker"
					c.Revision++
					db.seed("connection.updated", c)
				}
				return `{"result":"MUST_WITHHOLD_CONFIGURATION_RESULT"}`, nil
			}
			if w := dispatchRequest(h, "configuration-result-"+scenario, definedDispatchBody(task)); w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			st := waitDispatch(t, s)
			for _, e := range st.TaskExecutions {
				if e.ConfigurationVersion != 1 || e.ConnectionID == "" || len(e.ParticipantConfigurationDigest) != 64 || e.Status != "stopped" || len(e.Outputs) > 0 {
					t.Fatal("configuration result released", e)
				}
			}
			for _, m := range st.Messages {
				if strings.Contains(m.Text, "MUST_WITHHOLD_CONFIGURATION_RESULT") {
					t.Fatal("configuration result message released")
				}
			}
			// Full replay must preserve the exact original metadata despite later edits.
			reloaded, err := s.load(t.Context())
			if err != nil || len(reloaded.TaskExecutions) != 1 {
				t.Fatal("immutable metadata lost", err)
			}
		})
	}
}
func TestSuccessfulConfigurationReplayRejectsChangedCurrentConfig(t *testing.T) {
	t.Parallel()
	st, e, _, _ := pendingResultState(t, "review")
	st.Tasks[0].DispatchStatus = "completed"
	st.Tasks[0].Revision = 3
	e.Status = "reviewed"
	e.FinishedAt = e.StartedAt
	e.ResponseDigest = responseDigest("synthetic review")
	st.Connections[0].Revision++
	raw, _ := json.Marshal(e)
	if projectTaskExecution(&st, "task.execution.finished", raw) == nil {
		t.Fatal("successful stale configuration replayed")
	}
}

func TestCompletedTypedValuesRemainHistoryAfterConfigurationChanges(t *testing.T) {
	t.Parallel()
	st, e, _, _ := pendingResultState(t, "work")
	e.Status = "validated"
	values, err := taskcontract.ReturnedValues(e.Definition, []byte(`{"result":"synthetic completed value"}`), e.ID)
	if err != nil {
		t.Fatal(err)
	}
	e.Outputs = values
	st.TaskExecutions[e.ID] = e
	st.Tasks[0].Status = "Accepted"
	st.Tasks[0].DispatchStatus = "completed"
	st.Tasks[0].Revision = 3
	st.Participants[0].Revision++
	st.Connections[0].Revision++
	if !taskHasValidCurrentOutput(st, st.Tasks[0]) {
		t.Fatal("completed values erased by later profile settings")
	}
	e.ConfigurationVersion = 0
	e.ParticipantRevision = 0
	e.ParticipantConfigurationDigest = ""
	e.ConnectionID = ""
	e.ConnectionRevision = 0
	e.ConnectionConfigurationDigest = ""
	st.TaskExecutions[e.ID] = e
	if !taskHasValidCurrentOutput(st, st.Tasks[0]) {
		t.Fatal("pre-configuration completed history erased")
	}
}

func TestOwnerProfileSaveCancelsLocalTypedTurnAfterAcknowledgment(t *testing.T) {
	t.Parallel()
	s, _, h, task := structuredTaskFixture(t)
	started := make(chan struct{})
	cancelled := make(chan struct{})
	s.taskTurn = func(ctx context.Context, _ room.Participant, _, _ string) (string, error) {
		close(started)
		<-ctx.Done()
		close(cancelled)
		return `{"result":"MUST_WITHHOLD_AFTER_PROFILE_SAVE"}`, nil
	}
	if w := dispatchRequest(h, "profile-save-cancel", definedDispatchBody(task)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("mock turn did not start")
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	p := st.Participants[0]
	p.Name = "Updated synthetic profile"
	p.Rank = 1
	raw, _ := json.Marshal(p)
	w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"participant.updated","payload":%s}`, raw))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("acknowledged profile change did not cancel local typed turn")
	}
	st = waitDispatch(t, s)
	for _, e := range st.TaskExecutions {
		if e.Status != "stopped" || len(e.Outputs) > 0 {
			t.Fatal("cancelled turn published values")
		}
	}
}

func TestFailedProfileSaveDoesNotCancelAdmittedTurn(t *testing.T) {
	t.Parallel()
	s, db, h, task := structuredTaskFixture(t)
	started := make(chan context.Context, 1)
	s.taskTurn = func(ctx context.Context, _ room.Participant, _, _ string) (string, error) {
		started <- ctx
		<-ctx.Done()
		return "", ctx.Err()
	}
	if w := dispatchRequest(h, "profile-save-failure", definedDispatchBody(task)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	var turn context.Context
	select {
	case turn = <-started:
	case <-time.After(time.Second):
		t.Fatal("mock turn did not start")
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	p := st.Participants[0]
	p.Name = "Unacknowledged profile"
	p.Rank = 1
	raw, _ := json.Marshal(p)
	db.failInsertAt = 1
	w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"participant.updated","payload":%s}`, raw))
	db.failInsertAt = 0
	if w.Code != 500 || turn.Err() != nil {
		t.Fatal("failed profile write revoked existing authority", w.Code, turn.Err())
	}
	s.mu.Lock()
	s.cancels[p.ID]()
	s.mu.Unlock()
	_ = waitDispatch(t, s)
}

func TestTypedSessionCallbackRechecksConfiguration(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprint(changed), func(t *testing.T) {
			t.Parallel()
			s, db, h, task := structuredTaskFixture(t)
			checked := make(chan error, 1)
			s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
				if changed {
					st, err := s.load(t.Context())
					if err != nil {
						checked <- err
						return "", err
					}
					current := st.Participants[0]
					current.Instructions = "Changed synthetic profile"
					current.Revision++
					db.seed("participant.updated", current)
				}
				err := s.recordAgentSession(ctx, p, "development", "SYNTHETIC_NATIVE_THREAD")
				if (err != nil) != changed {
					checked <- fmt.Errorf("session configuration refusal: %v", err)
				} else {
					checked <- nil
				}
				return `{"result":"synthetic"}`, err
			}
			if w := dispatchRequest(h, fmt.Sprint("session-configuration-", changed), definedDispatchBody(task)); w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			select {
			case err := <-checked:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("session callback did not finish")
			}
			st := waitDispatch(t, s)
			found := false
			for _, p := range st.Participants {
				if p.ID == "dev" {
					found = p.ThreadID == "SYNTHETIC_NATIVE_THREAD"
				}
			}
			if found == changed {
				t.Fatal("session callback lost healthy session or saved stale session")
			}
		})
	}
}
