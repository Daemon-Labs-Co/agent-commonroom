package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"unicode/utf8"
)

type taskCriterionDecision struct {
	CriterionID string `json:"criterionId"`
	Decision    string `json:"decision"`
	Note        string `json:"note"`
}

func latestTaskAcceptance(st room.State, tid string) room.TaskAcceptanceReview {
	list := st.TaskAcceptanceReviews[tid]
	if len(list) == 0 {
		return room.TaskAcceptanceReview{}
	}
	return list[len(list)-1]
}

// Status/revision change in the same atomic acceptance operation. Request
// revision and post-decision revision are fenced separately from this content
// snapshot so that acceptance does not immediately invalidate its own record.
func taskAcceptanceContentSnapshot(st room.State, t room.Task) (string, room.TaskExecution, bool) {
	execution, ok := taskCurrentWorkExecution(st, t)
	normalized := t
	normalized.Status = ""
	normalized.Revision = 0
	dependencyEpoch := 0
	if hasTaskValueSources(execution.Inputs) {
		dependencyEpoch = st.ValueDependencyEpoch
	}
	raw, _ := json.Marshal([]any{st.Company.ID, normalized, execution, st.TaskAcceptanceEpoch, st.TaskDefinitionEpoch, dependencyEpoch})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), execution, ok && t.DispatchStatus != "pending"
}

// Snapshot construction never traverses lineage: source acceptance metadata is
// checked inside the single bounded dependency walk without resetting its budget.
func taskAcceptanceSnapshot(st room.State, t room.Task) (string, room.TaskExecution, bool) {
	snapshot, execution, ready := taskAcceptanceContentSnapshot(st, t)
	return snapshot, execution, ready && taskExecutionInputsCurrent(st, execution)
}
func taskHasRecordedAcceptance(st room.State, t room.Task) bool {
	if t.Status != "Accepted" {
		return false
	}
	if t.Definition == nil || !st.FeatureFlags["taskAcceptanceReview"] {
		return true
	}
	v := latestTaskAcceptance(st, t.ID)
	snapshot, _, ready := taskAcceptanceContentSnapshot(st, t)
	return ready && v.RoomID == t.RoomID && v.CompanyID == st.Company.ID && v.Action == "accept" && v.TaskRevision == t.Revision-1 && v.Snapshot == snapshot
}
func taskHasCurrentAcceptance(st room.State, t room.Task) bool {
	if !taskHasRecordedAcceptance(st, t) {
		return false
	}
	if t.Definition == nil {
		return true
	}
	execution, ok := taskCurrentWorkExecution(st, t)
	return ok && taskExecutionInputsCurrent(st, execution)
}
func assessTaskCriteria(d taskcontract.Definition, execution room.TaskExecution, decisions []taskCriterionDecision) ([]room.TaskCriterionAssessment, error) {
	choices := map[string]taskCriterionDecision{}
	for _, decision := range decisions {
		decision.Note = strings.TrimSpace(decision.Note)
		if _, exists := choices[decision.CriterionID]; exists || decision.CriterionID == "" || (decision.Decision != "passed" && decision.Decision != "failed") || decision.Note == "" || utf8.RuneCountInString(decision.Note) > 1000 {
			return nil, errors.New("each human criterion needs a unique passed/failed decision and a bounded rationale")
		}
		choices[decision.CriterionID] = decision
	}
	outputs := map[string]taskcontract.BoundValue{}
	for _, output := range execution.Outputs {
		outputs[output.Name] = output
	}
	contracts := map[string]taskcontract.DataContract{}
	for _, c := range d.DesiredOutputs {
		contracts[c.Name] = c
	}
	assessments := []room.TaskCriterionAssessment{}
	for _, criterion := range d.AcceptanceCriteria {
		output, present := outputs[criterion.Output]
		v := room.TaskCriterionAssessment{CriterionID: criterion.ID, Method: criterion.Method, OutputName: criterion.Output, OutputDigest: output.Digest}
		if criterion.Method == "schema" {
			v = taskSchemaAssessment(criterion, contracts[criterion.Output], output, present)
		} else {
			choice, ok := choices[criterion.ID]
			if !ok {
				return nil, errors.New("a human criterion decision is missing")
			}
			v.Decision = choice.Decision
			v.Note = choice.Note
			delete(choices, criterion.ID)
		}
		assessments = append(assessments, v)
	}
	if len(choices) != 0 {
		return nil, errors.New("unknown or server-assessed criterion supplied")
	}
	return assessments, nil
}
func taskSchemaAssessment(c taskcontract.Criterion, contract taskcontract.DataContract, output taskcontract.BoundValue, present bool) room.TaskCriterionAssessment {
	v := room.TaskCriterionAssessment{CriterionID: c.ID, Method: c.Method, OutputName: c.Output, OutputDigest: output.Digest, Decision: "failed", Note: "Declared output is absent or does not match its schema."}
	if present && int64(len(output.Value)) <= contract.MaxBytes && taskcontract.ValidContractValue(contract, output.Value) {
		v.Decision = "passed"
		v.Note = "Declared output matches its finite schema; this is a structural check only."
	}
	return v
}
func taskSchemaAssessments(d taskcontract.Definition, execution room.TaskExecution) []room.TaskCriterionAssessment {
	outputs := map[string]taskcontract.BoundValue{}
	for _, o := range execution.Outputs {
		outputs[o.Name] = o
	}
	contracts := map[string]taskcontract.DataContract{}
	for _, c := range d.DesiredOutputs {
		contracts[c.Name] = c
	}
	list := []room.TaskCriterionAssessment{}
	for _, c := range d.AcceptanceCriteria {
		if c.Method == "schema" {
			o, ok := outputs[c.Output]
			list = append(list, taskSchemaAssessment(c, contracts[c.Output], o, ok))
		}
	}
	return list
}
func taskReviewAllPassed(assessments []room.TaskCriterionAssessment) bool {
	if len(assessments) == 0 {
		return false
	}
	for _, a := range assessments {
		if a.Decision != "passed" {
			return false
		}
	}
	return true
}
func validateTaskAcceptanceReview(st room.State, v room.TaskAcceptanceReview) error {
	if !st.FeatureFlags["taskDefinitions"] || !st.FeatureFlags["taskAcceptanceReview"] || v.CompanyID != st.Company.ID || !sharedTaskValueRoom(st, v.RoomID) || v.ReviewedBy != ownerID || v.ReviewedAt.IsZero() || v.Note != strings.TrimSpace(v.Note) || v.Note == "" || utf8.RuneCountInString(v.Note) > 2000 || (v.Action != "accept" && v.Action != "changes") {
		return errors.New("invalid task review authority or decision")
	}
	var task room.Task
	for _, t := range st.Tasks {
		if t.ID == v.TaskID && t.RoomID == v.RoomID {
			task = t
			break
		}
	}
	if task.Definition == nil || task.Revision != v.TaskRevision || task.Revision == int(^uint(0)>>1) {
		return errors.New("task review does not bind the current definition/revision")
	}
	old := latestTaskAcceptance(st, task.ID)
	if old.Revision == int(^uint(0)>>1) || v.Revision != old.Revision+1 {
		return errors.New("task review revision changed")
	}
	snapshot, execution, ready := taskAcceptanceSnapshot(st, task)
	if !ready || v.Snapshot != snapshot || v.ExecutionID != execution.ID || v.DispatchID != task.DispatchID {
		return errors.New("task review has stale or missing returned values")
	}
	choices := []taskCriterionDecision{}
	for _, a := range v.Assessments {
		if a.Method == "human" {
			choices = append(choices, taskCriterionDecision{CriterionID: a.CriterionID, Decision: a.Decision, Note: a.Note})
		}
	}
	assessments, err := assessTaskCriteria(*task.Definition, execution, choices)
	if err != nil || !reflect.DeepEqual(assessments, v.Assessments) || v.Action == "accept" && !taskReviewAllPassed(assessments) {
		return errors.New("task review criterion decisions are incomplete or inconsistent")
	}
	return nil
}
func projectTaskAcceptanceReview(st *room.State, raw []byte) error {
	var v room.TaskAcceptanceReview
	var actor struct {
		PerformedBy string `json:"performedBy"`
	}
	if json.Unmarshal(raw, &v) != nil || json.Unmarshal(raw, &actor) != nil || actor.PerformedBy != ownerID {
		return errors.New("invalid task review provenance")
	}
	if err := validateTaskAcceptanceReview(*st, v); err != nil {
		return err
	}
	if st.TaskAcceptanceReviews == nil {
		st.TaskAcceptanceReviews = map[string][]room.TaskAcceptanceReview{}
	}
	st.TaskAcceptanceReviews[v.TaskID] = append(st.TaskAcceptanceReviews[v.TaskID], v)
	return nil
}
