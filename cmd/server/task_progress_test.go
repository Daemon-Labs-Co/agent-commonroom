package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func progressHandler(s *server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-progress", s.taskProgress)
	mux.HandleFunc("/api/task-acceptance", s.taskAcceptance)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/task-dispatch", s.taskDispatch)
	return s.auth.protect(s.actorGuard(mux))
}
func readTaskProgress(t *testing.T, h http.Handler) []taskProgressView {
	t.Helper()
	w := mockRequest(h, "GET", "/api/task-progress?room=engineering", "")
	var data struct {
		Tasks        []taskProgressView `json:"tasks"`
		Observations string             `json:"observations"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &data) != nil || data.Observations != "controller-recorded-state-only" {
		t.Fatal(w.Code, w.Body.String())
	}
	return data.Tasks
}
func TestTaskProgressDescribesDurableLifecycleWithoutProbingWorkers(t *testing.T) {
	t.Parallel()
	s, db, _, task := structuredTaskFixture(t)
	h := progressHandler(s)
	if w := mockRequest(h, "GET", "/api/task-progress?room=engineering", ""); w.Code != 403 {
		t.Fatal("progress not default-off", w.Code)
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskProgress","enabled":true}}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if v := readTaskProgress(t, h)[0]; v.Phase != "not-run" || v.InputState != "not-bound" || v.ExecutionID != "" {
		t.Fatal(v)
	}
	ready, release := make(chan struct{}), make(chan struct{})
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		close(ready)
		<-release
		return `{"result":"<private synthetic value>"}`, nil
	}
	if w := dispatchRequest(h, "progress-work", definedDispatchBody(task)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("mock turn not admitted")
	}
	before := len(db.snapshot())
	v := readTaskProgress(t, h)[0]
	if v.Phase != "work-admitted" || v.ExecutionStatus != "running" || !strings.Contains(v.Detail, "not an observation") || v.ReturnedValueCount != 0 || len(db.snapshot()) != before {
		t.Fatal("progress inferred worker activity or wrote events", v)
	}
	close(release)
	st := waitDispatch(t, s)
	task = st.Tasks[0]
	if v = readTaskProgress(t, h)[0]; v.Phase != "owner-review-required" || v.ReturnedValueCount != 1 || v.WorkExecutionID == "" {
		t.Fatal(v)
	}
	w := mockRequest(h, "GET", "/api/task-progress?room=engineering", "")
	if strings.Contains(w.Body.String(), "private synthetic value") || strings.Contains(w.Body.String(), "valueJson") || strings.Contains(w.Body.String(), "expectedInput") {
		t.Fatal("metadata endpoint disclosed raw contracts or values")
	}
	task.Status = "Accepted"
	reviseValueTask(t, h, task)
	if v = readTaskProgress(t, h)[0]; v.Phase != "accepted" {
		t.Fatal("owner acceptance not shown", v)
	}
	if w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskAcceptanceReview","enabled":true}}`); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if v = readTaskProgress(t, h)[0]; v.Phase != "criterion-review-required" || v.TaskStatus != "Accepted" {
		t.Fatal("saved task status confused with current criterion review", v)
	}
	if w = submitTaskAcceptance(h, "progress-criterion-accept", taskAcceptanceBody(t, h, "accept")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if v = readTaskProgress(t, h)[0]; v.Phase != "accepted" || v.CriterionReviewRevision != 1 {
		t.Fatal(v)
	}
	beforeState, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskProgress","enabled":false}}`); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w = mockRequest(h, "GET", "/api/task-progress?room=engineering", ""); w.Code != 403 {
		t.Fatal("disabled progress readable", w.Code)
	}
	afterState, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	beforeRecords, _ := json.Marshal([]any{beforeState.Tasks, beforeState.TaskExecutions, beforeState.TaskAcceptanceReviews})
	afterRecords, _ := json.Marshal([]any{afterState.Tasks, afterState.TaskExecutions, afterState.TaskAcceptanceReviews})
	if string(beforeRecords) != string(afterRecords) || beforeState.TaskAcceptanceEpoch != afterState.TaskAcceptanceEpoch || beforeState.TaskDefinitionEpoch != afterState.TaskDefinitionEpoch {
		t.Fatal("progress flag changed work or authority")
	}
}
func TestTaskProgressRecordsStaleDependenciesAndDistinctAttempts(t *testing.T) {
	t.Parallel()
	s, _, h, st, target, b := valueDependencyFixture(t)
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		return `{"result":"dependent value"}`, nil
	}
	if w := dispatchRequest(h, "progress-dependent", valueDispatchBody(t, target, b)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	st = waitDispatch(t, s)
	for _, v := range st.Tasks {
		if v.ID == target.ID {
			target = v
		}
	}
	v := describeTaskProgress(st, target)
	if v.Phase != "owner-review-required" || v.InputState != "current" || len(v.Dependencies) != 1 || !v.Dependencies[0].Current {
		t.Fatal(v)
	}
	source := st.Tasks[0]
	source.Status = "Changes requested"
	reviseValueTask(t, h, source)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	v = describeTaskProgress(st, target)
	if v.Phase != "inputs-unusable" || v.InputState != "unusable" || v.Dependencies[0].Current || v.ReturnedValueCount != 1 {
		t.Fatal("stale progress lost history or retained source authority", v)
	}
	// Derivation never mutates source state or immutable inputs.
	before, _ := json.Marshal(st.TaskExecutions)
	_ = describeTaskProgress(st, target)
	after, _ := json.Marshal(st.TaskExecutions)
	if string(before) != string(after) {
		t.Fatal("progress mutated execution inputs")
	}
}
func TestTaskProgressFiniteStatesLegacyDefinitionChangeAndFailedReview(t *testing.T) {
	t.Parallel()
	s, _, _, task := structuredTaskFixture(t)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	legacy := task
	legacy.Definition = nil
	if got := describeTaskProgress(st, legacy).Phase; got != "definition-required" {
		t.Fatal(got)
	}
	legacy.DispatchStatus = "pending"
	if got := describeTaskProgress(st, legacy).Phase; got != "legacy-admitted" {
		t.Fatal(got)
	}
	legacy.DispatchStatus = "completed"
	if got := describeTaskProgress(st, legacy).Phase; got != "legacy-result-recorded" {
		t.Fatal(got)
	}
	legacy.DispatchStatus = ""
	st.FeatureFlags["taskDefinitions"] = false
	if got := describeTaskProgress(st, legacy).Phase; got != "legacy" {
		t.Fatal(got)
	}
	st.FeatureFlags["taskDefinitions"] = true
	task.DispatchID = "attempt"
	task.DispatchStatus = "pending"
	if got := describeTaskProgress(st, task).Phase; got != "reconciliation-required" {
		t.Fatal(got)
	}
	execution, err := prepareTaskExecution(st, task, "work", []taskcontract.Binding{{Name: "source", Value: json.RawMessage(`"synthetic"`), SourceVersion: "v1"}}, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	execution.ID = "work"
	execution.Status = "validated"
	execution.Outputs, err = taskcontract.ReturnedValues(execution.Definition, []byte(`{"result":"synthetic"}`), execution.ID)
	if err != nil {
		t.Fatal(err)
	}
	review := *execution
	review.ID = "attempt"
	review.Mode = "review"
	review.ParentExecutionID = execution.ID
	review.Status = "running"
	review.Outputs = []taskcontract.BoundValue{}
	st.TaskExecutions = map[string]room.TaskExecution{execution.ID: *execution, review.ID: review}
	if v := describeTaskProgress(st, task); v.Phase != "review-admitted" || v.WorkExecutionID != "work" || v.ReturnedValueCount != 1 {
		t.Fatal("admitted review hid original work", v)
	}
	review.Status = "failed"
	st.TaskExecutions[review.ID] = review
	if v := describeTaskProgress(st, task); v.Phase != "retry-required" || v.WorkExecutionID != "work" {
		t.Fatal(v)
	}
	d := *task.Definition
	d.Task = "Revised task"
	d.Revision++
	task.Definition = &d
	if v := describeTaskProgress(st, task); v.Phase != "new-work-required" {
		t.Fatal(v)
	}
}
func TestTaskProgressScopePermissionsFlagsAndBoundedMetadata(t *testing.T) {
	t.Parallel()
	s, db, h, task := structuredTaskFixture(t)
	h = progressHandler(s)
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskProgress","enabled":true}}`); w.Code != 200 {
		t.Fatal(w.Code)
	}
	// Synthetic replay fixtures are legacy metadata, not execution evidence.
	for i := 0; i < 30; i++ {
		db.seed("task.revised", room.Task{ID: fmt.Sprintf("page-task-%02d", i), RoomID: "engineering", Title: "Synthetic legacy task", Status: "Draft", Revision: 1})
	}
	db.seed("task.revised", room.Task{ID: "foreign-task", RoomID: "marketing", Title: "PRIVATE FOREIGN TITLE", Status: "Draft", Revision: 1})
	w := mockRequest(h, "GET", "/api/task-progress?room=engineering", "")
	var page struct {
		Tasks  []taskProgressView `json:"tasks"`
		Counts map[string]int     `json:"pageCounts"`
		Total  int                `json:"totalTasks"`
		Next   string             `json:"nextCursor"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Tasks) != 25 || page.Total != 31 || page.Next == "" || strings.Contains(w.Body.String(), "PRIVATE FOREIGN TITLE") {
		t.Fatal(w.Code, w.Body.String())
	}
	sum := 0
	seen := map[string]bool{}
	for _, n := range page.Counts {
		sum += n
	}
	for _, v := range page.Tasks {
		seen[v.TaskID] = true
	}
	if sum != 25 {
		t.Fatal("page counts claimed room-wide progress", sum)
	}
	w = mockRequest(h, "GET", "/api/task-progress?room=engineering&after="+page.Next, "")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Tasks) != 6 || page.Next != "" {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, v := range page.Tasks {
		if seen[v.TaskID] {
			t.Fatal("progress page duplicate")
		}
	}
	if w = mockRequest(h, "GET", "/api/task-progress?room=engineering&after=missing", ""); w.Code != 409 {
		t.Fatal(w.Code)
	}
	if w = mockRequest(h, "POST", "/api/task-progress?room=engineering", `{}`); w.Code != 405 {
		t.Fatal(w.Code)
	}
	db.seed("actor.revised", room.Actor{ID: "member", Name: "Synthetic member", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "member", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w = mockRequest(h, "GET", "/api/task-progress?room=engineering", ""); w.Code != 200 {
		t.Fatal("member could not read shared progress", w.Code)
	}
	if w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskProgress","enabled":false}}`); w.Code != 403 {
		t.Fatal("member changed progress flag", w.Code)
	}
	if w = mockRequest(h, "GET", "/api/task-progress?room=marketing", ""); w.Code != 404 || strings.Contains(w.Body.String(), "PRIVATE FOREIGN TITLE") {
		t.Fatal("foreign progress leaked", w.Code)
	}
	s.auth.sessions["owner"] = ownerSession{ActorID: ownerID, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	db.seed("room.updated", room.Room{ID: "engineering", Name: "Synthetic personal", PersonalActorID: ownerID, Revision: 1})
	if w = mockRequest(h, "GET", "/api/task-progress?room=engineering", ""); w.Code != 404 {
		t.Fatal("personal progress bypassed history audit", w.Code)
	}
	_ = task
}
