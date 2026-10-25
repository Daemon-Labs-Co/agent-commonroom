package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func acceptanceFixture(t *testing.T) (*server, *mockDB, http.Handler, room.Task) {
	t.Helper()
	s, db, h, task := structuredTaskFixture(t)
	d := *task.Definition
	d.AcceptanceCriteria = append(append([]taskcontract.Criterion(nil), d.AcceptanceCriteria...), taskcontract.Criterion{ID: "shape", Description: "Returned result matches its finite schema", Method: "schema", Output: "result"})
	task.Definition = &d
	reviseValueTask(t, h, task)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task = st.Tasks[0]
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		return `{"result":"<exact synthetic value>"}`, nil
	}
	if w := dispatchRequest(h, "acceptance-source-work", definedDispatchBody(task)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	task = waitDispatch(t, s).Tasks[0]
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskAcceptanceReview","enabled":true}}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-acceptance", s.taskAcceptance)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/task-dispatch", s.taskDispatch)
	return s, db, s.auth.protect(s.actorGuard(mux)), task
}
func readTaskAcceptance(t *testing.T, h http.Handler) map[string]json.RawMessage {
	t.Helper()
	w := mockRequest(h, "GET", "/api/task-acceptance?room=engineering&taskId=task", "")
	var data map[string]json.RawMessage
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &data) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	return data
}
func taskAcceptanceBody(t *testing.T, h http.Handler, action string) taskAcceptanceRequest {
	t.Helper()
	data := readTaskAcceptance(t, h)
	var task room.Task
	var review room.TaskAcceptanceReview
	var snapshot string
	_ = json.Unmarshal(data["task"], &task)
	_ = json.Unmarshal(data["review"], &review)
	_ = json.Unmarshal(data["snapshot"], &snapshot)
	return taskAcceptanceRequest{TaskID: task.ID, TaskRevision: task.Revision, ReviewRevision: review.Revision, Snapshot: snapshot, Action: action, Note: "Owner compared the returned value with the declared source", Decisions: []taskCriterionDecision{{CriterionID: "matchesSource", Decision: "passed", Note: "Compared the exact synthetic input and returned text"}}}
}
func submitTaskAcceptance(h http.Handler, key string, body taskAcceptanceRequest) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/task-acceptance?room=engineering", strings.NewReader(string(raw)))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	r.Header.Set("X-CSRF-Token", "csrf")
	r.Header.Set("Idempotency-Key", key)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestTaskCriterionAcceptanceIsAtomicExplicitAndReplayable(t *testing.T) {
	t.Parallel()
	s, db, h, task := acceptanceFixture(t)
	// Generic task state cannot substitute for criterion-specific human review.
	task.Status = "Accepted"
	raw, _ := json.Marshal(task)
	if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 409 {
		t.Fatal("generic acceptance bypassed criterion review", w.Code, w.Body.String())
	}
	body := taskAcceptanceBody(t, h, "accept")
	before := len(db.snapshot())
	w := submitTaskAcceptance(h, "criterion-accept", body)
	if w.Code != 200 || len(db.snapshot()) != before+3 {
		t.Fatal("review/task/receipt not atomic", w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task = st.Tasks[0]
	review := latestTaskAcceptance(st, task.ID)
	if task.Status != "Accepted" || !taskHasCurrentAcceptance(st, task) || review.TaskRevision+1 != task.Revision || len(review.Assessments) != 2 || review.Assessments[1].Method != "schema" || review.Assessments[1].Decision != "passed" || review.Assessments[1].OutputDigest != st.TaskExecutions[task.DispatchID].Outputs[0].Digest {
		t.Fatal("incorrect explicit acceptance", task, review)
	}
	before = len(db.snapshot())
	w = submitTaskAcceptance(h, "criterion-accept", body)
	if w.Code != 200 || w.Header().Get("Idempotent-Replay") != "true" || len(db.snapshot()) != before {
		t.Fatal("review replay duplicated acceptance", w.Code)
	}
	data := readTaskAcceptance(t, h)
	if string(data["status"]) != `"Accepted"` {
		t.Fatal("acceptance invalidated itself", string(data["status"]))
	}
	restored, err := (&server{db: db}).load(t.Context())
	if err != nil || !taskHasCurrentAcceptance(restored, restored.Tasks[0]) {
		t.Fatal("replay lost immutable review", err)
	}
	// Human changes request remains an explicit terminal review, not a retry.
	changes := taskAcceptanceBody(t, h, "changes")
	changes.Decisions[0].Decision = "failed"
	changes.Decisions[0].Note = "The wording needs correction"
	if w = submitTaskAcceptance(h, "criterion-changes", changes); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if st.Tasks[0].Status != "Changes requested" || len(st.TaskAcceptanceReviews[task.ID]) != 2 || len(s.active) != 0 {
		t.Fatal("changes request launched work or lost history")
	}
}
func TestTaskCriterionReviewRefusals(t *testing.T) {
	for _, scenario := range []string{"missing human", "duplicate human", "unknown criterion", "spoof schema", "empty rationale", "failed criterion", "stale snapshot", "stale task", "stale review", "flag off", "definitions off", "member", "personal", "foreign", "persistence", "concurrent"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			s, db, h, task := acceptanceFixture(t)
			body := taskAcceptanceBody(t, h, "accept")
			want := 400
			switch scenario {
			case "missing human":
				body.Decisions = []taskCriterionDecision{}
			case "duplicate human":
				body.Decisions = append(body.Decisions, body.Decisions[0])
			case "unknown criterion":
				body.Decisions[0].CriterionID = "invented"
			case "spoof schema":
				body.Decisions = append(body.Decisions, taskCriterionDecision{CriterionID: "shape", Decision: "passed", Note: "I claim the schema passed"})
			case "empty rationale":
				body.Decisions[0].Note = " "
			case "failed criterion":
				body.Decisions[0].Decision = "failed"
				want = 409
			case "stale snapshot":
				body.Snapshot = strings.Repeat("a", 64)
				want = 409
			case "stale task":
				body.TaskRevision--
				want = 409
			case "stale review":
				body.ReviewRevision++
				want = 409
			case "flag off":
				db.seed("feature.flag.updated", map[string]any{"name": "taskAcceptanceReview", "enabled": false})
				want = 403
			case "definitions off":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
				want = 403
			case "member":
				db.seed("actor.revised", room.Actor{ID: "member", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
				s.auth.sessions["owner"] = ownerSession{ActorID: "member", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
				want = 403
			case "personal":
				db.seed("room.updated", room.Room{ID: "engineering", Name: "Synthetic personal", PersonalActorID: ownerID, Revision: 1})
				want = 404
			case "foreign":
				db.seed("actor.revised", room.Actor{ID: "foreign", Kind: "human", Enabled: true, RoomIDs: []string{}})
				s.auth.sessions["owner"] = ownerSession{ActorID: "foreign", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
				want = 403
			case "persistence":
				db.failInsertAt = 2
				want = 500
			case "concurrent":
				s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(context.Context) {
					db.seed("feature.flag.updated", map[string]any{"name": "taskAcceptanceReview", "enabled": false})
				}}
				want = 409
			}
			before := len(db.snapshot())
			w := submitTaskAcceptance(h, "refuse-criterion-review", body)
			if w.Code != want {
				t.Fatal(scenario, w.Code, w.Body.String())
			}
			expected := before
			if scenario == "concurrent" {
				expected++
			}
			if len(db.snapshot()) != expected {
				t.Fatal("refused review partially committed", len(db.snapshot()), expected)
			}
			if scenario == "persistence" && !s.integrityFault {
				t.Fatal("failed review commit did not latch integrity stop")
			}
			_ = task
		})
	}
}

func TestCriterionReviewFlagsStalenessHistoryAndFiniteSchemaFailure(t *testing.T) {
	t.Parallel()
	s, db, h, task := acceptanceFixture(t)
	body := taskAcceptanceBody(t, h, "accept")
	// A policy off/on cycle makes an already-open review stale.
	for _, enabled := range []bool{false, true} {
		if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"feature.flag.updated","payload":{"name":"taskAcceptanceReview","enabled":%t}}`, enabled)); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := submitTaskAcceptance(h, "stale-policy-review", body); w.Code != 409 {
		t.Fatal("stale policy review admitted", w.Code)
	}
	if w := submitTaskAcceptance(h, "fresh-policy-review", taskAcceptanceBody(t, h, "accept")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	// Source availability now requires the explicit current owner criterion review.
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task = st.Tasks[0]
	if !taskHasCurrentAcceptance(st, task) {
		t.Fatal("current explicit review unavailable")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "taskAcceptanceReview", "enabled": false})
	history := readTaskAcceptance(t, h)
	if string(history["enabled"]) != "false" || !strings.Contains(string(history["reviews"]), "Compared the exact synthetic input") {
		t.Fatal("flag off erased review history")
	}
	// Ordinary owner acceptance remains available when this optional feature is off.
	task.Status = "Changes requested"
	reviseValueTask(t, h, task)
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task = st.Tasks[0]
	task.Status = "Accepted"
	reviseValueTask(t, h, task)
	db.seed("feature.flag.updated", map[string]any{"name": "taskAcceptanceReview", "enabled": true})
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if taskHasCurrentAcceptance(st, st.Tasks[0]) {
		t.Fatal("old review silently accepted changed task")
	}
	// A schema criterion targeting an omitted optional output cannot be passed by
	// a human payload, even when all required work outputs validated successfully.
	definition := syntheticDefinition()
	definition.DesiredOutputs = append(definition.DesiredOutputs, taskcontract.DataContract{Name: "optionalReport", Type: "value", SchemaID: "text.v1", MaxBytes: 128})
	definition.AcceptanceCriteria = append(definition.AcceptanceCriteria, taskcontract.Criterion{ID: "optionalShape", Method: "schema", Output: "optionalReport", Description: "Report must be present and typed"})
	execution := room.TaskExecution{Outputs: []taskcontract.BoundValue{}}
	assessments, err := assessTaskCriteria(definition, execution, []taskCriterionDecision{{CriterionID: "matchesSource", Decision: "passed", Note: "Owner checked this condition"}})
	if err != nil || assessments[1].Decision != "failed" || taskReviewAllPassed(assessments) {
		t.Fatal("missing optional schema output inferred present", assessments, err)
	}
}
func TestCriterionReviewHistoryPaginationShowsOriginalReviewedAttempts(t *testing.T) {
	t.Parallel()
	s, _, h, _ := acceptanceFixture(t)
	original := taskAcceptanceBody(t, h, "changes")
	for i := 0; i < 12; i++ {
		s.mu.Lock()
		s.requestBudgets = nil
		s.mu.Unlock()
		if i == 6 {
			st, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			task := st.Tasks[0]
			s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
				return `{"result":"<later work value>"}`, nil
			}
			if w := dispatchRequest(h, "later-reviewed-work", definedDispatchBody(task)); w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			_ = waitDispatch(t, s)
		}
		body := taskAcceptanceBody(t, h, "changes")
		body.Decisions[0].Decision = "failed"
		if w := submitTaskAcceptance(h, fmt.Sprintf("review-page-%02d", i), body); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	data := readTaskAcceptance(t, h)
	var reviews []room.TaskAcceptanceReview
	var cursor string
	var executions map[string]room.TaskExecution
	_ = json.Unmarshal(data["reviews"], &reviews)
	_ = json.Unmarshal(data["nextCursor"], &cursor)
	_ = json.Unmarshal(data["reviewedExecutions"], &executions)
	if len(reviews) != 10 || cursor != "3" || reviews[0].Revision != 12 || len(executions) != 2 {
		t.Fatal("history page lost reviewed attempt identity", string(data["reviews"]), cursor)
	}
	w := mockRequest(h, "GET", "/api/task-acceptance?room=engineering&taskId=task&before="+cursor, "")
	var page map[string]json.RawMessage
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	_ = json.Unmarshal(page["reviews"], &reviews)
	_ = json.Unmarshal(page["reviewedExecutions"], &executions)
	if len(reviews) != 2 || reviews[0].Revision != 2 || reviews[1].Revision != 1 || reviews[1].Snapshot != original.Snapshot {
		t.Fatal("older review page lost original snapshot")
	}
	if got := executions[reviews[1].ExecutionID].Outputs[0].ValueJSON; got != `"\u003cexact synthetic value\u003e"` {
		t.Fatal("historical review showed newer values", got)
	}
}
func TestCriterionReviewReplayRejectsForgedDecisionsAndGenericInjection(t *testing.T) {
	t.Parallel()
	s, db, h, _ := acceptanceFixture(t)
	before, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if w := submitTaskAcceptance(h, "review-replay-base", taskAcceptanceBody(t, h, "accept")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var raw []byte
	for _, e := range db.snapshot() {
		if e.kind == "task.acceptance.reviewed" {
			raw = []byte(e.payload)
		}
	}
	if len(raw) == 0 {
		t.Fatal("review event absent")
	}
	for _, change := range []string{"actor", "snapshot", "digest", "schema decision", "missing human", "wrong execution", "revision"} {
		t.Run(change, func(t *testing.T) {
			var v room.TaskAcceptanceReview
			_ = json.Unmarshal(raw, &v)
			actor := ownerID
			switch change {
			case "actor":
				actor = "bot"
			case "snapshot":
				v.Snapshot = strings.Repeat("f", 64)
			case "digest":
				v.Assessments[0].OutputDigest = strings.Repeat("e", 64)
			case "schema decision":
				v.Assessments[1].Decision = "failed"
			case "missing human":
				v.Assessments = v.Assessments[1:]
			case "wrong execution":
				v.ExecutionID = "unknown"
			case "revision":
				v.Revision++
			}
			bytes, _ := json.Marshal(v)
			var fields map[string]json.RawMessage
			_ = json.Unmarshal(bytes, &fields)
			fields["performedBy"], _ = json.Marshal(actor)
			bytes, _ = json.Marshal(fields)
			candidate := before
			candidate.TaskAcceptanceReviews = map[string][]room.TaskAcceptanceReview{}
			if projectTaskAcceptanceReview(&candidate, bytes) == nil {
				t.Fatal("forged review accepted")
			}
		})
	}
	count := len(db.snapshot())
	w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.acceptance.reviewed","payload":%s}`, raw))
	if w.Code == 200 || len(db.snapshot()) != count {
		t.Fatal("generic event injected human review", w.Code)
	}
}

func TestCriterionPolicyCannotReviveDependentExecutionAuthority(t *testing.T) {
	t.Parallel()
	_, _, _, st, target, b := valueDependencyFixture(t)
	allocation, err := prepareTaskExecution(st, target, "work", []taskcontract.Binding{b}, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	st.FeatureFlags["taskAcceptanceReview"] = true
	if taskExecutionInputsCurrent(st, *allocation) {
		t.Fatal("unreviewed source retained authority when explicit review required")
	}
	st.FeatureFlags["taskAcceptanceReview"] = false
	st.TaskAcceptanceEpoch++
	if taskExecutionInputsCurrent(st, *allocation) {
		t.Fatal("disabling explicit review revived an old dependent allocation")
	}
	fresh, err := prepareTaskExecution(st, target, "work", []taskcontract.Binding{b}, ownerID)
	if err != nil || fresh.InputAcceptanceEpoch != st.TaskAcceptanceEpoch {
		t.Fatal("explicit owner redispatch could not bind current baseline policy", err)
	}
}

func TestCriterionReviewPolicyMakesRoomAndSourceReadinessStale(t *testing.T) {
	t.Parallel()
	s, db, h, _ := acceptanceFixture(t)
	if w := submitTaskAcceptance(h, "task-before-room", taskAcceptanceBody(t, h, "accept")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomes", "enabled": true})
	db.seed("feature.flag.updated", map[string]any{"name": "roomOutcomeReview", "enabled": true})
	db.seed("room.updated", room.Room{ID: "engineering", Name: "Synthetic delivery", Outcome: "Return a finite value", AcceptanceCriteria: "Owner reviews exact values", Revision: 1})
	mux := http.NewServeMux()
	mux.HandleFunc("/api/room-outcome-review", s.roomOutcomeReview)
	reviewHandler := s.auth.protect(s.actorGuard(mux))
	if w := submitRoomReview(reviewHandler, "room-after-task", roomReviewBody(t, reviewHandler, "accept")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task := st.Tasks[0]
	if _, _, _, ok := taskValueSource(st, task.RoomID, task.DispatchID, "result"); !ok {
		t.Fatal("reviewed source unavailable")
	}
	for _, enabled := range []bool{false, true} {
		db.seed("feature.flag.updated", map[string]any{"name": "taskAcceptanceReview", "enabled": enabled})
	}
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, ok := taskValueSource(st, task.RoomID, task.DispatchID, "result"); ok {
		t.Fatal("stale criterion review authorized upstream value")
	}
	data := readRoomReview(t, reviewHandler)
	if string(data["ready"]) != "false" || string(data["status"]) != `"Needs review"` {
		t.Fatal("stale structured acceptance kept room ready", string(data["status"]))
	}
	if w := submitRoomReview(reviewHandler, "room-stale-criteria", roomReviewBody(t, reviewHandler, "accept")); w.Code != 409 {
		t.Fatal("room accepted stale criterion review", w.Code)
	}
	if w := submitTaskAcceptance(h, "task-policy-rereview", taskAcceptanceBody(t, h, "accept")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := submitRoomReview(reviewHandler, "room-after-rereview", roomReviewBody(t, reviewHandler, "accept")); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestCriterionReviewCannotAcceptOmittedSchemaOutput(t *testing.T) {
	t.Parallel()
	s, _, h, task := acceptanceFixture(t)
	d := *task.Definition
	d.DesiredOutputs = append(append([]taskcontract.DataContract(nil), d.DesiredOutputs...), taskcontract.DataContract{Name: "report", Type: "value", SchemaID: "text.v1", MaxBytes: 128})
	d.AcceptanceCriteria = append(append([]taskcontract.Criterion(nil), d.AcceptanceCriteria...), taskcontract.Criterion{ID: "reportShape", Method: "schema", Output: "report", Description: "Optional report must be returned to pass this criterion"})
	task.Definition = &d
	reviseValueTask(t, h, task)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	task = st.Tasks[0]
	if w := dispatchRequest(h, "work-omits-report", definedDispatchBody(task)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	_ = waitDispatch(t, s)
	data := readTaskAcceptance(t, h)
	if !strings.Contains(string(data["schemaAssessments"]), `"decision":"failed"`) {
		t.Fatal("omitted schema output not shown as failed")
	}
	body := taskAcceptanceBody(t, h, "accept")
	if w := submitTaskAcceptance(h, "reject-missing-schema", body); w.Code != 409 {
		t.Fatal("acceptance bypassed missing schema output", w.Code, w.Body.String())
	}
	body.Action = "changes"
	if w := submitTaskAcceptance(h, "request-missing-schema", body); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}
