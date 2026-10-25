package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
)

type taskProgressDependency struct {
	InputName          string `json:"inputName"`
	SourceTaskID       string `json:"sourceTaskId"`
	SourceTaskRevision int    `json:"sourceTaskRevision"`
	ExecutionID        string `json:"executionId"`
	OutputName         string `json:"outputName"`
	Current            bool   `json:"current"`
}
type taskProgressView struct {
	DispatchStatus          string                   `json:"dispatchStatus,omitempty"`
	TaskID                  string                   `json:"taskId"`
	Title                   string                   `json:"title"`
	TaskRevision            int                      `json:"taskRevision"`
	TaskStatus              string                   `json:"taskStatus"`
	DefinitionRevision      int                      `json:"definitionRevision,omitempty"`
	Phase                   string                   `json:"phase"`
	Detail                  string                   `json:"detail"`
	ExecutionID             string                   `json:"executionId,omitempty"`
	ExecutionMode           string                   `json:"executionMode,omitempty"`
	ExecutionStatus         string                   `json:"executionStatus,omitempty"`
	WorkExecutionID         string                   `json:"workExecutionId,omitempty"`
	ReturnedValueCount      int                      `json:"returnedValueCount"`
	InputState              string                   `json:"inputState"`
	Dependencies            []taskProgressDependency `json:"dependencies"`
	CriterionReviewRevision int                      `json:"criterionReviewRevision,omitempty"`
}

// Progress derives only from durable controller records. It never probes a
// provider or infers worker activity, correctness, observed checks or acceptance.
func describeTaskProgress(st room.State, t room.Task) taskProgressView {
	v := taskProgressView{TaskID: t.ID, Title: t.Title, TaskRevision: t.Revision, TaskStatus: t.Status, DispatchStatus: t.DispatchStatus, InputState: "not-bound", Dependencies: []taskProgressDependency{}}
	if t.Definition == nil {
		v.Phase = "legacy"
		v.Detail = "Legacy task without a structured delivery contract. Task status alone does not prove returned work."
		if st.FeatureFlags["taskDefinitions"] {
			v.Phase = "definition-required"
			v.Detail = "Define the seven task parts before structured dispatch."
		}
		if t.DispatchStatus == "pending" {
			v.Phase = "legacy-admitted"
			v.Detail = "A legacy dispatch was durably recorded and awaits an outcome or reconciliation. Worker activity is not observed; define the task before future structured dispatch."
		} else if t.DispatchStatus == "completed" {
			v.Phase = "legacy-result-recorded"
			v.Detail = "A legacy dispatch completed with a message, without a typed returned-value contract. Saved task status and message claims do not prove structured delivery."
		}
		return v
	}
	v.DefinitionRevision = t.Definition.Revision
	v.CriterionReviewRevision = latestTaskAcceptance(st, t.ID).Revision
	if len(t.Definition.ExpectedInputs) == 0 {
		v.InputState = "none"
	}
	execution, exists := st.TaskExecutions[t.DispatchID]
	if !exists || execution.CompanyID != st.Company.ID || execution.RoomID != t.RoomID || execution.TaskID != t.ID {
		v.Phase = "not-run"
		v.Detail = "Definition is saved. Inputs must be bound at dispatch; provider and worker availability are not checked here."
		if t.DispatchStatus == "pending" {
			v.Phase = "reconciliation-required"
			v.Detail = "A pending dispatch has no structured attempt record. Inspect and reconcile before retrying."
		}
		return v
	}
	v.ExecutionID = execution.ID
	v.ExecutionMode = execution.Mode
	v.ExecutionStatus = execution.Status
	for _, input := range execution.Inputs {
		if p := input.Provenance; p != nil {
			v.Dependencies = append(v.Dependencies, taskProgressDependency{InputName: input.Name, SourceTaskID: p.TaskID, SourceTaskRevision: p.TaskRevision, ExecutionID: p.ExecutionID, OutputName: p.OutputName})
		}
	}
	inputsCurrent := taskExecutionInputsCurrent(st, execution)
	v.InputState = "current"
	if !inputsCurrent {
		v.InputState = "unusable"
	}
	for i := range v.Dependencies {
		dependency := &v.Dependencies[i]
		for _, input := range execution.Inputs {
			if input.Name == dependency.InputName {
				single := execution
				single.Inputs = []taskcontract.BoundValue{input}
				dependency.Current = taskExecutionInputsCurrent(st, single)
				break
			}
		}
	}
	if execution.DefinitionDigest != taskDefinitionDigest(t) {
		v.Phase = "new-work-required"
		v.Detail = "The saved definition differs from the last attempt. Old outputs remain history; dispatch new work explicitly."
		return v
	}
	work, workCurrent := taskCurrentWorkExecution(st, t)
	if workCurrent {
		v.WorkExecutionID = work.ID
		v.ReturnedValueCount = len(work.Outputs)
	} else if execution.Mode == "review" {
		parent, ok := st.TaskExecutions[execution.ParentExecutionID]
		if ok && parent.CompanyID == st.Company.ID && parent.RoomID == t.RoomID && parent.TaskID == t.ID && parent.Mode == "work" && parent.Status == "validated" && parent.DefinitionDigest == taskDefinitionDigest(t) {
			v.WorkExecutionID = parent.ID
			v.ReturnedValueCount = len(parent.Outputs)
		}
	}
	if !inputsCurrent {
		v.Phase = "inputs-unusable"
		v.Detail = "Bound input lineage is stale, disabled, revoked or exceeds verification limits. Tool and result authority are refused; inspect the attempt before redispatch."
		return v
	}
	if execution.Status == "running" {
		v.Phase = "work-admitted"
		v.Detail = "A work attempt was durably admitted and awaits an outcome. This is not an observation that the worker is running."
		if execution.Mode == "review" {
			v.Phase = "review-admitted"
			v.Detail = "An agent review was durably admitted and awaits an outcome. It cannot accept work."
		}
		return v
	}
	if execution.Status == "failed" || execution.Status == "stopped" {
		v.Phase = "retry-required"
		v.Detail = "The latest attempt failed or stopped. History is preserved; retry requires an explicit owner action."
		return v
	}
	if !workCurrent {
		v.Phase = "new-work-required"
		v.Detail = "The latest attempt does not establish current structurally valid work outputs. Inspect history before proceeding."
		return v
	}
	if t.Status == "Accepted" && taskHasCurrentAcceptance(st, t) {
		v.Phase = "accepted"
		v.Detail = "The human owner accepted the current returned values. Acceptance does not prove independent runtime checks or authorize new work."
		return v
	}
	v.Phase = "owner-review-required"
	v.Detail = "Current returned values passed structural validation. The owner must assess the criteria; an agent review is advice."
	if st.FeatureFlags["taskAcceptanceReview"] {
		v.Phase = "criterion-review-required"
		v.Detail = "Current returned values need an explicit owner criterion review. Previous task status is not a current criterion decision."
	}
	return v
}
