package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"fmt"
	"testing"
)

// Synthetic projected records deliberately include a forged cycle/overlong
// graph: even invalid historical state must refuse within one traversal budget.
// No DB, provider, filesystem or live identity participates.
func acceptanceLineageState(t *testing.T, n int, cycle bool) room.State {
	t.Helper()
	d := syntheticDefinition()
	st := room.State{Company: room.Company{ID: "company"}, Rooms: []room.Room{{ID: "engineering"}}, FeatureFlags: map[string]bool{"taskDefinitions": true, "taskValueDependencies": true, "taskAcceptanceReview": true}, TaskExecutions: map[string]room.TaskExecution{}, TaskAcceptanceReviews: map[string][]room.TaskAcceptanceReview{}}
	for i := 0; i < n; i++ {
		id, eid := fmt.Sprintf("task-%d", i), fmt.Sprintf("execution-%d", i)
		inputs, err := taskcontract.Bind(d, []taskcontract.Binding{{Name: "source", ValueJSON: `"synthetic"`, SourceVersion: "manual-v1"}})
		if err != nil {
			t.Fatal(err)
		}
		outputs, err := taskcontract.ReturnedValues(d, []byte(`{"result":"synthetic"}`), eid)
		if err != nil {
			t.Fatal(err)
		}
		e := room.TaskExecution{ID: eid, CompanyID: "company", RoomID: "engineering", TaskID: id, Mode: "work", Status: "validated", Definition: d, DefinitionDigest: taskcontract.Digest(d), Inputs: inputs, Outputs: outputs}
		st.TaskExecutions[eid] = e
		st.Tasks = append(st.Tasks, room.Task{ID: id, RoomID: "engineering", Title: id, Revision: 2, Status: "Accepted", DispatchID: eid, DispatchStatus: "done", Definition: &d})
	}
	for i, task := range st.Tasks {
		if i == n-1 && !cycle {
			continue
		}
		next := (i + 1) % n
		sourceTask := st.Tasks[next]
		source := st.TaskExecutions[sourceTask.DispatchID]
		output := source.Outputs[0]
		e := st.TaskExecutions[task.DispatchID]
		e.Inputs = []taskcontract.BoundValue{{Name: "source", Value: output.Value, ValueJSON: output.ValueJSON, Digest: output.Digest, SourceVersion: source.ID + "/result", Provenance: &taskcontract.ValueProvenance{TaskID: sourceTask.ID, TaskRevision: sourceTask.Revision, ExecutionID: source.ID, OutputName: "result", DefinitionDigest: source.DefinitionDigest, OutputDigest: output.Digest}}}
		st.TaskExecutions[e.ID] = e
	}
	for _, task := range st.Tasks {
		snapshot, execution, ready := taskAcceptanceContentSnapshot(st, task)
		if !ready {
			t.Fatal("bad synthetic work")
		}
		st.TaskAcceptanceReviews[task.ID] = []room.TaskAcceptanceReview{{CompanyID: "company", RoomID: task.RoomID, TaskID: task.ID, TaskRevision: task.Revision - 1, ExecutionID: execution.ID, DispatchID: task.DispatchID, Action: "accept", Snapshot: snapshot}}
	}
	return st
}
func TestAcceptanceUsesOneBoundedLineageWalk(t *testing.T) {
	for _, tc := range []struct {
		name      string
		n         int
		cycle, ok bool
	}{{"healthy chain", 6, false, true}, {"maximum depth", 17, false, true}, {"over depth", 18, false, false}, {"cycle", 6, true, false}, {"self cycle", 1, true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			st := acceptanceLineageState(t, tc.n, tc.cycle)
			task := st.Tasks[0]
			// Matching accepted metadata alone must not recursively restart the walk.
			if !taskHasRecordedAcceptance(st, task) {
				t.Fatal("metadata unexpectedly stale")
			}
			if got := taskHasCurrentAcceptance(st, task); got != tc.ok {
				t.Fatal("current acceptance", got)
			}
			_, _, ready := taskAcceptanceSnapshot(st, task)
			if ready != tc.ok {
				t.Fatal("acceptance readiness", ready)
			}
			if taskHasValidCurrentOutput(st, task) != tc.ok {
				t.Fatal("output eligibility disagrees")
			}
			if tc.ok {
				nodes := 128
				e := st.TaskExecutions[task.DispatchID]
				if taskValueLineageCurrent(st, e, map[string]bool{}, 0, &nodes) {
					t.Fatal("shared node budget reset")
				}
			}
		})
	}
}
func TestAcceptanceRevokesStaleLineageWithoutErasingRecordedDecision(t *testing.T) {
	t.Parallel()
	st := acceptanceLineageState(t, 3, false)
	task := st.Tasks[0]
	st.Tasks[2].Revision++
	if !taskHasRecordedAcceptance(st, task) || taskHasCurrentAcceptance(st, task) {
		t.Fatal("stale source acceptance survived")
	}
	st = acceptanceLineageState(t, 3, false)
	st.FeatureFlags["taskValueDependencies"] = false
	if taskHasCurrentAcceptance(st, st.Tasks[0]) {
		t.Fatal("disabled lineage survived")
	}
}

func TestAcceptanceSharedBudgetBoundsRepeatedDAGBranches(t *testing.T) {
	t.Parallel()
	st := acceptanceLineageState(t, 9, false)
	task := st.Tasks[0]
	execution := st.TaskExecutions[task.DispatchID]
	d := *task.Definition
	d.ExpectedInputs = nil
	input := execution.Inputs[0]
	execution.Inputs = nil
	for i := 0; i < 16; i++ {
		name := fmt.Sprintf("input%d", i)
		d.ExpectedInputs = append(d.ExpectedInputs, taskcontract.DataContract{Name: name, Type: "value", Required: true, SchemaID: "text.v1", MaxBytes: 512})
		v := input
		v.Name = name
		execution.Inputs = append(execution.Inputs, v)
	}
	if err := taskcontract.Validate(d); err != nil {
		t.Fatal(err)
	}
	task.Definition = &d
	execution.Definition = d
	execution.DefinitionDigest = taskcontract.Digest(d)
	st.Tasks[0] = task
	st.TaskExecutions[execution.ID] = execution
	snapshot, _, ready := taskAcceptanceContentSnapshot(st, task)
	if !ready {
		t.Fatal("bad synthetic snapshot")
	}
	review := st.TaskAcceptanceReviews[task.ID][0]
	review.Snapshot = snapshot
	st.TaskAcceptanceReviews[task.ID] = []room.TaskAcceptanceReview{review}
	if !taskHasRecordedAcceptance(st, task) || taskHasCurrentAcceptance(st, task) {
		t.Fatal("repeated branches restarted total budget")
	}
	nodes := 0
	if taskValueLineageCurrent(st, execution, map[string]bool{}, 0, &nodes) || nodes != 129 {
		t.Fatal("unexpected bounded traversal", nodes)
	}
	// Disabling criterion review must not disable dependency integrity bounds.
	st.FeatureFlags["taskAcceptanceReview"] = false
	if taskHasCurrentAcceptance(st, task) {
		t.Fatal("criterion flag bypassed lineage bound")
	}
}
