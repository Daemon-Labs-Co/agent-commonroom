package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func syntheticDefinition() taskcontract.Definition {
	return taskcontract.Definition{SchemaVersion: 1, Revision: 1, Task: "Return a synthetic summary", ExpectedInputs: []taskcontract.DataContract{{Name: "source", Type: "value", Required: true, SchemaID: "text.v1", MaxBytes: 512}}, Process: taskcontract.Process{Mode: "objective", Instructions: "Summarize the supplied source without inventing facts", Constraints: []string{}}, DesiredOutputs: []taskcontract.DataContract{{Name: "result", Type: "value", Required: true, SchemaID: "text.v1", MaxBytes: 1024}}, AcceptanceCriteria: []taskcontract.Criterion{{ID: "matchesSource", Description: "Matches the source", Method: "human", Output: "result"}}, ExecutionConstraints: taskcontract.Limits{AllowedOperations: []string{"read"}, MaxDurationSeconds: 60, MaxResponseBytes: 2048, MaxToolCalls: 1, CompletionAuthority: "owner", ResourceScope: "assigned-task-workspace"}, FailureContract: taskcontract.Failure{MissingInput: "block", ExecutionFailure: "record-and-stop", Retry: "human-only"}}
}
func structuredTaskFixture(t *testing.T) (*server, *mockDB, http.Handler, room.Task) {
	t.Helper()
	s, db, h, _ := dispatchFixture(t)
	w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskDefinitions","enabled":true}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	d := syntheticDefinition()
	task := room.Task{ID: "task", RoomID: "engineering", Title: "Synthetic feature", Assignee: "dev", Reviewer: "dev", Status: "Assigned", Revision: 1, Definition: &d}
	raw, _ := json.Marshal(task)
	w = mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return s, db, h, st.Tasks[0]
}
func definedDispatchBody(task room.Task) string {
	return fmt.Sprintf(`{"taskId":"task","revision":%d,"mode":"work","inputs":[{"name":"source","value":"<synthetic input>","sourceVersion":"synthetic-v1"}]}`, task.Revision)
}
func TestStructuredTaskDispatchOutputsAndAuthority(t *testing.T) {
	t.Parallel()
	s, _, h, task := structuredTaskFixture(t)
	repo := strings.Repeat("b", 64)
	checked := make(chan error, 1)
	s.taskTurn = func(ctx context.Context, p room.Participant, prompt, _ string) (string, error) {
		if !strings.Contains(prompt, "Structured task contract") || !strings.Contains(prompt, "sourceVersion") || !strings.Contains(prompt, "desiredOutput") {
			t.Error("contract or bindings absent from prompt")
		}
		deadline, _ := ctx.Deadline()
		if time.Until(deadline) > 60*time.Second {
			t.Error("declared duration not enforced")
		}
		jc := s.managedExecutionJobContext(ctx, p, repo, func() (string, error) { return repo, nil })
		reg := jc.Value(managedJobContextKey{}).(managedJobRegistration)
		if err := reg.Register("synthetic-worker", "codex"); err != nil {
			checked <- err
			return "", err
		}
		if err := reg.Update("submitted"); err != nil {
			checked <- err
			return "", err
		}
		scope := toolbroker.Scope{Version: 1, JobID: reg.JobID, Room: p.RoomID, Participant: p.ID, Connection: p.ConnectionID, Provider: "codex", RepositoryDigest: repo, Deadline: reg.Deadline.UTC()}
		audit, err := newControllerToolAuditor(ctx, s, scope, ownerID, func() (string, error) { return repo, nil })
		if err != nil {
			checked <- err
			return "", err
		}
		intent := toolbroker.Intent{Scope: scope, CallID: "write-refusal", Operation: "write", PayloadDigest: strings.Repeat("c", 64)}
		if audit.RecordIntent(ctx, intent) == nil {
			t.Error("read-only contract granted write")
		}
		intent.CallID = "read-allowed"
		intent.Operation = "read"
		if err = audit.RecordIntent(ctx, intent); err != nil {
			checked <- err
			return "", err
		}
		if err = audit.RecordOutcome(ctx, intent, toolbroker.Outcome{Status: "succeeded", OutputDigest: strings.Repeat("d", 64)}); err != nil {
			checked <- err
			return "", err
		}
		intent.CallID = "read-over-budget"
		if audit.RecordIntent(ctx, intent) == nil {
			t.Error("contract tool call budget bypassed")
		}
		if err = reg.Update("shutdown_confirmed"); err != nil {
			checked <- err
			return "", err
		}
		checked <- nil
		return `{"result":"<synthetic result>"}`, nil
	}
	before := 0
	_ = before
	w := dispatchRequest(h, "structured-work", definedDispatchBody(task))
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case err := <-checked:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("synthetic turn did not run")
	}
	st := waitDispatch(t, s)
	task = st.Tasks[0]
	execution := st.TaskExecutions[task.DispatchID]
	if execution.Status != "validated" || len(execution.Outputs) != 1 || execution.Outputs[0].ValueJSON != `"\u003csynthetic result\u003e"` || len(execution.Inputs) != 1 || execution.Inputs[0].SourceVersion != "synthetic-v1" {
		t.Fatal("immutable definition/input/output record wrong", execution)
	}
	if task.Status == "Accepted" {
		t.Fatal("schema validation auto-accepted work")
	}
	task.Status = "Accepted"
	raw, _ := json.Marshal(task)
	w = mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw))
	if w.Code != 200 {
		t.Fatal("owner could not accept valid output", w.Code, w.Body.String())
	}
}
func TestTaskDefinitionMutationPermissionsVersionsAndLegacy(t *testing.T) {
	t.Parallel()
	s, db, h, task := structuredTaskFixture(t)
	if task.Definition.Revision != 1 || task.Criteria != taskcontract.CriteriaText(*task.Definition) {
		t.Fatal("definition version or derived criteria wrong")
	}
	// Older callers cannot erase a saved contract by omitting it.
	body := fmt.Sprintf(`{"kind":"task.revised","payload":{"id":"task","revision":%d,"title":"Synthetic feature","assignee":"dev","reviewer":"dev","status":"Assigned"}}`, task.Revision)
	if w := mockRequest(h, "POST", "/api/events", body); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task = st.Tasks[0]
	if task.Definition == nil || task.Definition.Revision != 1 {
		t.Fatal("old client removed definition")
	}
	previous := task
	changed := *task.Definition
	changed.Task = "Revised synthetic task"
	task.Definition = &changed
	raw, _ := json.Marshal(task)
	if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	raw, _ = json.Marshal(previous)
	if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 409 {
		t.Fatal("stale definition update accepted", w.Code)
	}
	st, err = (&server{db: db}).load(context.Background())
	if err != nil || st.Tasks[0].Definition.Revision != 2 {
		t.Fatal("definition replay/version failed", err)
	}
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	task = st.Tasks[0]
	changed = *task.Definition
	changed.Task = "Member changed scope"
	task.Definition = &changed
	raw, _ = json.Marshal(task)
	if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 403 {
		t.Fatal("member changed canonical contract", w.Code, w.Body.String())
	}
}
func TestStructuredTaskFailureRefusalsAndFlagEpoch(t *testing.T) {
	for _, scenario := range []string{"missing input", "wrong input", "file output", "flag off", "bad return", "too much response", "revoke off-on"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			s, db, h, task := structuredTaskFixture(t)
			body := definedDispatchBody(task)
			want := 202
			switch scenario {
			case "missing input":
				body = fmt.Sprintf(`{"taskId":"task","revision":%d,"mode":"work"}`, task.Revision)
				want = 409
			case "wrong input":
				body = strings.Replace(body, `"<synthetic input>"`, `42`, 1)
				want = 409
			case "file output":
				changed := *task.Definition
				changed.DesiredOutputs[0] = taskcontract.DataContract{Name: "result", Type: "file", Required: true, MediaType: "text/plain", FileName: "result.txt", MaxBytes: 1024}
				task.Definition = &changed
				raw, _ := json.Marshal(task)
				if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 200 {
					t.Fatal(w.Code, w.Body.String())
				}
				st, err := s.load(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				task = st.Tasks[0]
				body = definedDispatchBody(task)
				want = 409
			case "flag off":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
				want = 409
			}
			s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
				if scenario == "bad return" {
					return "Agent says the work is complete", nil
				}
				if scenario == "too much response" {
					return strings.Repeat("x", 2049), nil
				}
				if scenario == "revoke off-on" {
					db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
					db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
				}
				return `{"result":"Synthetic result"}`, nil
			}
			before := len(db.snapshot())
			w := dispatchRequest(h, "structured-refusal", body)
			if w.Code != want {
				t.Fatal(scenario, w.Code, w.Body.String())
			}
			if want != 202 {
				if len(db.snapshot()) != before || len(s.active) != 0 {
					t.Fatal("refused dispatch had effects")
				}
				return
			}
			st := waitDispatch(t, s)
			execution := st.TaskExecutions[st.Tasks[0].DispatchID]
			if scenario == "revoke off-on" {
				if execution.Status != "stopped" {
					t.Fatal("authority revived after off/on", execution.Status)
				}
			} else if execution.Status != "failed" {
				t.Fatal("invalid return became deliverable", execution.Status)
			}
			if len(execution.Outputs) != 0 || st.Tasks[0].DispatchStatus == "completed" {
				t.Fatal("unvalidated output published as success")
			}
		})
	}
}

func TestStructuredDelegationKeepsExactParentInputs(t *testing.T) {
	t.Parallel()
	d := syntheticDefinition()
	inputs, err := taskcontract.Bind(d, []taskcontract.Binding{{Name: "source", Value: json.RawMessage(`"original"`), SourceVersion: "original-v1"}})
	if err != nil {
		t.Fatal(err)
	}
	task := room.Task{ID: "task", RoomID: "engineering", Revision: 2, Assignee: "child", Status: "Assigned", Definition: &d, DelegationID: "delegation"}
	parent := room.TaskExecution{ID: "parent", TaskID: "task", RoomID: "engineering", Definition: d, DefinitionDigest: taskDefinitionDigest(task), Inputs: inputs, Mode: "work", Status: "failed"}
	st := room.State{Company: room.Company{ID: "company"}, FeatureFlags: map[string]bool{"taskDefinitions": true}, Tasks: []room.Task{task}, TaskExecutions: map[string]room.TaskExecution{"parent": parent}, Delegations: map[string]room.TaskDelegation{"delegation": {DispatchID: "parent"}}}
	execution, err := prepareTaskExecution(st, task, "work", nil, ownerID)
	if err != nil || execution.Inputs[0].Digest != inputs[0].Digest {
		t.Fatal("parent binding was not inherited", err)
	}
	if _, err = prepareTaskExecution(st, task, "work", []taskcontract.Binding{{Name: "source", Value: json.RawMessage(`"expanded scope"`), SourceVersion: "new-v2"}}, ownerID); err == nil {
		t.Fatal("delegated input scope expanded")
	}
	execution.ID = "child-dispatch"
	execution.ParticipantID = "child"
	st.Tasks[0].DispatchID = execution.ID
	st.Tasks[0].DispatchStatus = "pending"
	st.Tasks[0].Status = "Working"
	st.Tasks[0].Revision++
	raw, _ := json.Marshal(execution)
	var payload map[string]any
	_ = json.Unmarshal(raw, &payload)
	payload["performedBy"] = ownerID
	raw, _ = json.Marshal(payload)
	if err = projectTaskExecution(&st, "task.execution.started", raw); err != nil {
		t.Fatal("legitimate delegated execution replay refused", err)
	}
}
func TestStructuredReviewerUsesValidatedWorkSnapshot(t *testing.T) {
	t.Parallel()
	s, db, h, task := structuredTaskFixture(t)
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		return `{"result":"original work"}`, nil
	}
	if w := dispatchRequest(h, "work-before-review", definedDispatchBody(task)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	st := waitDispatch(t, s)
	task = st.Tasks[0]
	workID := task.DispatchID
	db.seed("connection.created", room.Connection{ID: "review-connection", Protocol: "CLI", Endpoint: "ssh://review-worker"})
	db.seed("participant.created", room.Participant{ID: "reviewer", RoomID: "engineering", ConnectionID: "review-connection", Status: "Idle"})
	task.Reviewer = "reviewer"
	task.Status = "Submitted"
	raw, _ := json.Marshal(task)
	if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task = st.Tasks[0]
	seen := make(chan struct{}, 1)
	s.taskTurn = func(ctx context.Context, p room.Participant, prompt, _ string) (string, error) {
		if p.ID != "reviewer" || !strings.Contains(prompt, "returnedOutputs") || !strings.Contains(prompt, "original work") {
			t.Error("review did not receive actual work values")
		}
		repo := strings.Repeat("b", 64)
		jobCtx := s.managedExecutionJobContext(ctx, p, repo, func() (string, error) { return repo, nil })
		reg := jobCtx.Value(managedJobContextKey{}).(managedJobRegistration)
		if err := reg.Register("review-worker", "codex"); err != nil {
			t.Error(err)
		} else {
			if err = reg.Update("submitted"); err != nil {
				t.Error(err)
			}
			if err = reg.Update("shutdown_confirmed"); err != nil {
				t.Error(err)
			}
		}
		seen <- struct{}{}
		return "Synthetic review findings", nil
	}
	body := fmt.Sprintf(`{"taskId":"task","revision":%d,"mode":"review"}`, task.Revision)
	if w := dispatchRequest(h, "review-current-work", body); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-seen:
	case <-time.After(time.Second):
		t.Fatal("review did not run")
	}
	deadline := time.Now().Add(time.Second)
	for {
		s.mu.Lock()
		active := s.active["reviewer"]
		st, err = s.load(context.Background())
		s.mu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
		if !active {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("review did not finish")
		}
		time.Sleep(time.Millisecond)
	}
	task = st.Tasks[0]
	review := st.TaskExecutions[task.DispatchID]
	if review.Status != "reviewed" || review.ParentExecutionID != workID || len(review.Outputs) != 0 || !taskHasValidCurrentOutput(st, task) {
		t.Fatal("review lost work/definition binding", review)
	}
	task.Status = "Accepted"
	raw, _ = json.Marshal(task)
	if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 200 {
		t.Fatal("owner acceptance after distinct review failed", w.Code, w.Body.String())
	}
}
