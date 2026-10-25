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

func reviseValueTask(t *testing.T, h http.Handler, task room.Task) {
	t.Helper()
	raw, _ := json.Marshal(task)
	if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}
func valueDependencyFixture(t *testing.T) (*server, *mockDB, http.Handler, room.State, room.Task, taskcontract.Binding) {
	t.Helper()
	s, db, h, source := structuredTaskFixture(t)
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskValueDependencies","enabled":true}}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		return `{"result":"<accepted source>"}`, nil
	}
	if w := dispatchRequest(h, "source-work", definedDispatchBody(source)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	st := waitDispatch(t, s)
	source = st.Tasks[0]
	source.Status = "Accepted"
	reviseValueTask(t, h, source)
	d := syntheticDefinition()
	target := room.Task{RoomID: "engineering", Title: "Dependent task", Status: "Assigned", Assignee: "dev", Reviewer: "dev", Definition: &d}
	reviseValueTask(t, h, target)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range st.Tasks {
		if v.Title == "Dependent task" {
			target = v
		}
	}
	return s, db, h, st, target, taskcontract.Binding{Name: "source", Source: &taskcontract.ValueReference{ExecutionID: source.DispatchID, OutputName: "result"}}
}
func valueDispatchBody(t *testing.T, target room.Task, b taskcontract.Binding) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"taskId": target.ID, "revision": target.Revision, "mode": "work", "inputs": []taskcontract.Binding{b}})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
func TestVerifiedTaskValueDispatchDiscoveryAndImmutableProvenance(t *testing.T) {
	t.Parallel()
	s, _, h, st, target, b := valueDependencyFixture(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-input-sources", s.taskInputSources)
	discovery := s.auth.protect(s.actorGuard(mux))
	w := mockRequest(discovery, "GET", "/api/task-input-sources?room=engineering&taskId="+target.ID+"&input=source", "")
	var result struct {
		Sources []taskInputSource `json:"sources"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || len(result.Sources) != 1 || result.Sources[0].ExecutionID != b.Source.ExecutionID || result.Sources[0].ValueJSON != `"\u003caccepted source\u003e"` {
		t.Fatal(w.Code, w.Body.String())
	}
	s.taskTurn = func(_ context.Context, _ room.Participant, prompt, _ string) (string, error) {
		if !strings.Contains(prompt, "provenance") || !strings.Contains(prompt, b.Source.ExecutionID) {
			t.Error("verified binding absent from prompt")
		}
		return `{"result":"dependent result"}`, nil
	}
	if w = dispatchRequest(h, "dependent-work", valueDispatchBody(t, target, b)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	st = waitDispatch(t, s)
	for _, task := range st.Tasks {
		if task.ID == target.ID {
			target = task
		}
	}
	e := st.TaskExecutions[target.DispatchID]
	if e.Status != "validated" || len(e.Inputs) != 1 || e.Inputs[0].Provenance == nil || e.Inputs[0].Provenance.TaskID != "task" || e.Inputs[0].Provenance.OutputDigest != e.Inputs[0].Digest || !taskHasValidCurrentOutput(st, target) {
		t.Fatal("verified output/provenance missing", e)
	}
	// Human acceptance remains distinct from dependency provenance and validity.
	if target.Status == "Accepted" {
		t.Fatal("dependency automatically accepted target")
	}
	target.Status = "Accepted"
	reviseValueTask(t, h, target)
	// Reopening a source leaves bytes/history intact but invalidates acceptance.
	source := st.Tasks[0]
	source.Status = "Changes requested"
	reviseValueTask(t, h, source)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if taskHasValidCurrentOutput(st, target) || len(st.TaskExecutions[e.ID].Outputs) != 1 {
		t.Fatal("stale lineage accepted or immutable history lost")
	}
	target.Status = "Assigned"
	target.Revision++ // accepted task revision advanced above
	reviseValueTask(t, h, target)
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range st.Tasks {
		if v.ID == target.ID {
			target = v
		}
	}
	target.Status = "Accepted"
	raw, _ := json.Marshal(target)
	if w = mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":%s}`, raw)); w.Code != 409 {
		t.Fatal("stale source allowed acceptance", w.Code, w.Body.String())
	}
}
func cloneValueState(st room.State) room.State {
	st.Tasks = append([]room.Task(nil), st.Tasks...)
	flags := map[string]bool{}
	for k, v := range st.FeatureFlags {
		flags[k] = v
	}
	st.FeatureFlags = flags
	executions := map[string]room.TaskExecution{}
	for k, v := range st.TaskExecutions {
		raw, _ := json.Marshal(v)
		var copy room.TaskExecution
		_ = json.Unmarshal(raw, &copy)
		executions[k] = copy
	}
	st.TaskExecutions = executions
	return st
}
func TestVerifiedTaskValueInputRefusalsAndLineageRevocation(t *testing.T) {
	t.Parallel()
	_, _, _, base, target, b := valueDependencyFixture(t)
	for _, tc := range []struct {
		name   string
		change func(*room.State, *room.Task, *taskcontract.Binding)
	}{
		{"disabled", func(st *room.State, _ *room.Task, _ *taskcontract.Binding) {
			st.FeatureFlags["taskValueDependencies"] = false
		}},
		{"definitions disabled", func(st *room.State, _ *room.Task, _ *taskcontract.Binding) {
			st.FeatureFlags["taskDefinitions"] = false
		}},
		{"replacement value", func(_ *room.State, _ *room.Task, b *taskcontract.Binding) { b.Value = json.RawMessage(`"spoof"`) }},
		{"replacement exact value", func(_ *room.State, _ *room.Task, b *taskcontract.Binding) { b.ValueJSON = `"spoof"` }},
		{"replacement label", func(_ *room.State, _ *room.Task, b *taskcontract.Binding) { b.SourceVersion = "claimed" }},
		{"missing output", func(_ *room.State, _ *room.Task, b *taskcontract.Binding) { b.Source.OutputName = "missing" }},
		{"missing execution", func(_ *room.State, _ *room.Task, b *taskcontract.Binding) { b.Source.ExecutionID = "missing" }},
		{"self", func(_ *room.State, t *room.Task, _ *taskcontract.Binding) { t.ID = "task" }},
		{"unaccepted", func(st *room.State, _ *room.Task, _ *taskcontract.Binding) { st.Tasks[0].Status = "Submitted" }},
		{"old attempt", func(st *room.State, _ *room.Task, _ *taskcontract.Binding) { st.Tasks[0].DispatchID = "new-attempt" }},
		{"old definition", func(st *room.State, _ *room.Task, b *taskcontract.Binding) {
			e := st.TaskExecutions[b.Source.ExecutionID]
			e.DefinitionDigest = "old"
			st.TaskExecutions[e.ID] = e
		}},
		{"foreign company", func(st *room.State, _ *room.Task, b *taskcontract.Binding) {
			e := st.TaskExecutions[b.Source.ExecutionID]
			e.CompanyID = "foreign"
			st.TaskExecutions[e.ID] = e
		}},
		{"foreign room", func(st *room.State, _ *room.Task, b *taskcontract.Binding) {
			e := st.TaskExecutions[b.Source.ExecutionID]
			e.RoomID = "elsewhere"
			st.TaskExecutions[e.ID] = e
		}},
		{"wrong target schema", func(_ *room.State, t *room.Task, _ *taskcontract.Binding) {
			d := *t.Definition
			d.ExpectedInputs = append([]taskcontract.DataContract(nil), d.ExpectedInputs...)
			d.ExpectedInputs[0].SchemaID = "integer.v1"
			t.Definition = &d
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := cloneValueState(base)
			v := target
			ref := *b.Source
			input := b
			input.Source = &ref
			tc.change(&st, &v, &input)
			if _, err := prepareTaskExecution(st, v, "work", []taskcontract.Binding{input}, ownerID); err == nil {
				t.Fatal("invalid reference admitted")
			}
		})
	}
	e, err := prepareTaskExecution(base, target, "work", []taskcontract.Binding{b}, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*room.State){func(st *room.State) { st.Tasks[0].Revision++ }, func(st *room.State) { st.Tasks[0].Status = "Changes requested" }, func(st *room.State) { st.ValueDependencyEpoch++; st.FeatureFlags["taskValueDependencies"] = true }} {
		st := cloneValueState(base)
		change(&st)
		if taskExecutionInputsCurrent(st, *e) {
			t.Fatal("stale or revoked authority revived")
		}
	}
	// Ordinary labels remain unverified, even when they resemble execution IDs.
	manual, err := prepareTaskExecution(base, target, "work", []taskcontract.Binding{{Name: "source", Value: json.RawMessage(`"manual"`), SourceVersion: b.Source.ExecutionID + "/result"}}, ownerID)
	if err != nil || manual.Inputs[0].Provenance != nil {
		t.Fatal("manual label promoted to provenance", err)
	}
	st := cloneValueState(base)
	st.ValueDependencyEpoch++
	st.FeatureFlags["taskValueDependencies"] = false
	if !taskExecutionInputsCurrent(st, *manual) {
		t.Fatal("dependency flag revoked manual execution")
	}
}
func TestTaskValueSourceChangeRevokesActualBrokerAndResult(t *testing.T) {
	t.Parallel()
	s, db, h, st, target, b := valueDependencyFixture(t)
	ready, release, checked := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
		repo := strings.Repeat("b", 64)
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
		close(ready)
		<-release
		intent := toolbroker.Intent{Scope: scope, CallID: "stale-source-read", Operation: "read", PayloadDigest: strings.Repeat("c", 64)}
		if audit.RecordIntent(ctx, intent) == nil {
			checked <- fmt.Errorf("broker admitted effect after source reopened")
		} else {
			checked <- nil
		}
		_ = reg.Update("shutdown_confirmed")
		return `{"result":"must be withheld"}`, nil
	}
	if w := dispatchRequest(h, "dependent-revocation", valueDispatchBody(t, target, b)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-ready:
	case err := <-checked:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatal("mock worker not ready")
	}
	source := st.Tasks[0]
	source.Status = "Changes requested"
	reviseValueTask(t, h, source)
	before := len(db.snapshot())
	close(release)
	select {
	case err := <-checked:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("broker check blocked")
	}
	st = waitDispatch(t, s)
	for _, v := range st.Tasks {
		if v.ID == target.ID {
			target = v
		}
	}
	e := st.TaskExecutions[target.DispatchID]
	if e.Status == "validated" || len(e.Outputs) != 0 || target.DispatchStatus == "completed" {
		t.Fatal("stale source published a result", e, target)
	}
	for _, event := range db.snapshot()[before:] {
		if event.kind == "worker.tool.intent" {
			t.Fatal("stale authority wrote a tool intent")
		}
	}
}

func TestTaskValueLineageCyclesDepthBudgetsAndHistoricalEpoch(t *testing.T) {
	t.Parallel()
	_, _, _, st, target, b := valueDependencyFixture(t)
	root, err := prepareTaskExecution(st, target, "work", []taskcontract.Binding{b}, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	// Finished history remains usable after a flag cycle if every source version
	// is still current. An in-flight allocation with the same epoch does not.
	root.Status = "validated"
	st.ValueDependencyEpoch++
	if !taskExecutionInputsCurrent(st, *root) {
		t.Fatal("completed lineage erased by flag cycle")
	}
	root.Status = "running"
	if taskExecutionInputsCurrent(st, *root) {
		t.Fatal("old active allocation revived")
	}
	st.ValueDependencyEpoch--
	// Install a synthetic chain using the real resolver and finite output binder.
	ref := b
	for i := 0; i < 18; i++ {
		task := target
		task.ID = fmt.Sprintf("chain-task-%02d", i)
		task.Revision = 1
		task.DispatchID = fmt.Sprintf("chain-execution-%02d", i)
		task.Status = "Accepted"
		if i >= 16 { // The final target would traverse too many dependent tasks.
			if _, err := prepareTaskExecution(st, task, "work", []taskcontract.Binding{ref}, ownerID); err == nil {
				t.Fatal("excessive dependency depth admitted")
			}
			break
		}
		execution, err := prepareTaskExecution(st, task, "work", []taskcontract.Binding{ref}, ownerID)
		if err != nil {
			t.Fatal(i, err)
		}
		execution.ID = task.DispatchID
		execution.Status = "validated"
		execution.Outputs, err = taskcontract.ReturnedValues(execution.Definition, []byte(`{"result":"<accepted source>"}`), execution.ID)
		if err != nil {
			t.Fatal(err)
		}
		st.Tasks = append(st.Tasks, task)
		st.TaskExecutions[execution.ID] = *execution
		ref = taskcontract.Binding{Name: "source", Source: &taskcontract.ValueReference{ExecutionID: execution.ID, OutputName: "result"}}
	}
	// A forged historical cycle cannot pass the same resolver or current check.
	cyc := cloneValueState(st)
	original := cyc.TaskExecutions[b.Source.ExecutionID]
	original.Inputs = root.Inputs
	// Point the source back to itself with otherwise matching version/digest.
	p := *original.Inputs[0].Provenance
	original.Inputs[0].Provenance = &p
	cyc.TaskExecutions[original.ID] = original
	if taskExecutionInputsCurrent(cyc, original) {
		t.Fatal("cyclic lineage accepted")
	}
	// Wide DAGs have a total traversal budget even when no path is deep.
	nodes := 128
	if taskValueLineageCurrent(st, *root, map[string]bool{}, 0, &nodes) {
		t.Fatal("dependency traversal exceeded node budget")
	}
}
func TestTaskValueSourceAPIFlagsPermissionsAndCompatibility(t *testing.T) {
	t.Parallel()
	s, db, h, _, target, _ := valueDependencyFixture(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-input-sources", s.taskInputSources)
	api := s.auth.protect(s.actorGuard(mux))
	path := "/api/task-input-sources?room=engineering&taskId=" + target.ID + "&input=source"
	if w := mockRequest(api, "GET", path+"&after=missing", ""); w.Code != 409 {
		t.Fatal("unknown cursor accepted", w.Code)
	}
	if w := mockRequest(api, "GET", strings.Replace(path, "input=source", "input=missing", 1), ""); w.Code != 404 {
		t.Fatal("missing input accepted", w.Code)
	}
	if w := mockRequest(api, "POST", path, `{}`); w.Code != 405 {
		t.Fatal(w.Code)
	}
	db.seed("actor.revised", room.Actor{ID: "foreign-member", Name: "Synthetic outsider", Kind: "human", Enabled: true, RoomIDs: []string{}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "foreign-member", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w := mockRequest(api, "GET", path, ""); w.Code != 404 || strings.Contains(w.Body.String(), "accepted source") {
		t.Fatal("foreign discovery disclosed values", w.Code)
	}
	s.auth.sessions["owner"] = ownerSession{ActorID: ownerID, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	// A known member still cannot change the server-enforced dependency setting.
	db.seed("actor.revised", room.Actor{ID: "member", Name: "Synthetic member", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "member", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskValueDependencies","enabled":false}}`); w.Code != 403 {
		t.Fatal("member changed dependency flag", w.Code)
	}
	s.auth.sessions["owner"] = ownerSession{ActorID: ownerID, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskValueDependencies","enabled":false}}`); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := mockRequest(api, "GET", path, ""); w.Code != 403 || strings.Contains(w.Body.String(), "accepted source") {
		t.Fatal("disabled discovery disclosed values", w.Code)
	}
}
func TestDependencyFlagCycleCancelsOnlyDependentTurns(t *testing.T) {
	t.Parallel()
	s, _, h, _, target, b := valueDependencyFixture(t)
	ready := make(chan struct{})
	s.taskTurn = func(ctx context.Context, _ room.Participant, _, _ string) (string, error) {
		close(ready)
		<-ctx.Done()
		return `{"result":"late"}`, nil
	}
	if w := dispatchRequest(h, "dependent-cycle", valueDispatchBody(t, target, b)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("dependent turn not ready")
	}
	// Also register a synthetic unrelated local turn. Disabling dependencies must
	// not cancel it; ordinary manual contracts retain their own authority.
	manualCanceled := false
	s.mu.Lock()
	s.cancels["manual-turn"] = func() { manualCanceled = true }
	s.taskDefinitionTurns["manual-turn"] = "manual-attempt"
	s.mu.Unlock()
	for _, enabled := range []bool{false, true} {
		if w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"feature.flag.updated","payload":{"name":"taskValueDependencies","enabled":%t}}`, enabled)); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	st := waitDispatch(t, s)
	if manualCanceled {
		t.Fatal("dependency flag canceled manual work")
	}
	for _, execution := range st.TaskExecutions {
		if execution.TaskID == target.ID && (execution.Status == "validated" || execution.Status == "running" || len(execution.Outputs) > 0) {
			t.Fatal("revoked dependent result survived cycle", execution)
		}
	}
}

func TestTaskValueSourceDiscoveryPaginatesAcceptedValuesAndRetainsReviewIdentity(t *testing.T) {
	t.Parallel()
	s, _, h, st, target, b := valueDependencyFixture(t)
	// Reviewing an already-produced result preserves its original output identity.
	source := st.Tasks[0]
	source.Status = "Submitted"
	reviseValueTask(t, h, source)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	source = st.Tasks[0]
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		return "Synthetic reviewer findings", nil
	}
	if w := dispatchRequest(h, "source-review", fmt.Sprintf(`{"taskId":"task","revision":%d,"mode":"review"}`, source.Revision)); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	source = waitDispatch(t, s).Tasks[0]
	source.Status = "Accepted"
	reviseValueTask(t, h, source)
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	bound, err := resolveTaskValueInputs(st, target, []taskcontract.Binding{b})
	if err != nil || bound[0].Provenance.ExecutionID != b.Source.ExecutionID {
		t.Fatal("review changed output identity", err)
	}
	s.taskTurn = func(context.Context, room.Participant, string, string) (string, error) {
		return `{"result":"<page value>"}`, nil
	}
	for i := 0; i < 26; i++ {
		// Pagination fixtures are independent owner interactions, not a burst-rate
		// test. Reset only this synthetic server's limiter rather than sleeping.
		s.mu.Lock()
		s.requestBudgets = nil
		s.mu.Unlock()
		d := syntheticDefinition()
		task := room.Task{RoomID: "engineering", Title: fmt.Sprintf("Synthetic source %02d", i), Status: "Assigned", Assignee: "dev", Reviewer: "dev", Definition: &d}
		reviseValueTask(t, h, task)
		st, err = s.load(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range st.Tasks {
			if v.Title == task.Title {
				task = v
			}
		}
		manual := taskcontract.Binding{Name: "source", Value: json.RawMessage(`"synthetic input"`), SourceVersion: "synthetic-v1"}
		if w := dispatchRequest(h, fmt.Sprintf("source-page-%02d", i), valueDispatchBody(t, task, manual)); w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		st = waitDispatch(t, s)
		for _, v := range st.Tasks {
			if v.ID == task.ID {
				task = v
			}
		}
		task.Status = "Accepted"
		reviseValueTask(t, h, task)
	}
	// Independent source discovery must not require the other input to be bound.
	d := *target.Definition
	d.ExpectedInputs = append(append([]taskcontract.DataContract(nil), d.ExpectedInputs...), taskcontract.DataContract{Name: "other", Type: "value", Required: true, SchemaID: "text.v1", MaxBytes: 512})
	target.Definition = &d
	reviseValueTask(t, h, target)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-input-sources", s.taskInputSources)
	api := s.auth.protect(s.actorGuard(mux))
	path := "/api/task-input-sources?room=engineering&taskId=" + target.ID + "&input=source"
	var page struct {
		Sources    []taskInputSource `json:"sources"`
		NextCursor string            `json:"nextCursor"`
	}
	w := mockRequest(api, "GET", path, "")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Sources) != 25 || page.NextCursor == "" {
		t.Fatal(w.Code, w.Body.String())
	}
	seen := map[string]bool{}
	for _, v := range page.Sources {
		seen[v.ExecutionID+"/"+v.OutputName] = true
	}
	w = mockRequest(api, "GET", path+"&after="+page.NextCursor, "")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Sources) != 2 || page.NextCursor != "" {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, v := range page.Sources {
		if seen[v.ExecutionID+"/"+v.OutputName] {
			t.Fatal("source pagination duplicated output")
		}
	}
	// A different finite target schema excludes all text outputs, despite names.
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range st.Tasks {
		if v.ID == target.ID {
			target = v
		}
	}
	d = *target.Definition
	d.ExpectedInputs = append([]taskcontract.DataContract(nil), d.ExpectedInputs...)
	d.ExpectedInputs[0].SchemaID = "integer.v1"
	target.Definition = &d
	reviseValueTask(t, h, target)
	w = mockRequest(api, "GET", path, "")
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || len(page.Sources) != 0 {
		t.Fatal("incompatible values listed", w.Code, w.Body.String())
	}
}
