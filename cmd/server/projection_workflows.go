package main

// Event reducers for workflow definitions and runs. Each reducer applies one kind of recorded event to
// the projection and returns an error only for an event that must never have been recorded.

import (
	"agentroom/internal/room"
	"encoding/json"
	"fmt"
)

// reduceWorkflowRunUpdated reduces workflow.run.updated.
func reduceWorkflowRunUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var run room.WorkflowRun
	if json.Unmarshal(data, &run) != nil {
		return fmt.Errorf("Invalid workflow run event")
	}
	found := false
	for i, old := range st.WorkflowRuns {
		if old.ID == run.ID {
			st.WorkflowRuns[i] = run
			found = true
			break
		}
	}
	if !found {
		st.WorkflowRuns = append(st.WorkflowRuns, run)
	}
	return nil
}

// reduceWorkflowRevised reduces workflow.revised.
func reduceWorkflowRevised(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Workflow
	json.Unmarshal(data, &v)
	found := false
	for i, old := range st.Workflows {
		if old.ID == v.ID {
			st.Workflows[i] = v
			found = true
		}
	}
	if !found {
		st.Workflows = append(st.Workflows, v)
	}
	return nil
}
