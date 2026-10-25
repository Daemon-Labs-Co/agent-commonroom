package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

func responseDigest(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:])
}

// Select only the latest successful work/review chain. Dependency validation
// is separate so recursion retains one shared depth/node budget and path stack.
func taskCurrentWorkExecution(st room.State, t room.Task) (room.TaskExecution, bool) {
	current, ok := st.TaskExecutions[t.DispatchID]
	if !ok || current.CompanyID != st.Company.ID || current.TaskID != t.ID || current.RoomID != t.RoomID || current.DefinitionDigest != taskDefinitionDigest(t) || current.DelegationID != t.DelegationID {
		return room.TaskExecution{}, false
	}
	if current.Mode == "review" {
		if current.Status != "reviewed" {
			return room.TaskExecution{}, false
		}
		current, ok = st.TaskExecutions[current.ParentExecutionID]
	}
	return current, ok && current.CompanyID == st.Company.ID && current.TaskID == t.ID && current.RoomID == t.RoomID && current.DelegationID == t.DelegationID && current.Mode == "work" && current.Status == "validated" && current.DefinitionDigest == taskDefinitionDigest(t)
}
func taskHasValidCurrentOutput(st room.State, t room.Task) bool {
	current, ok := taskCurrentWorkExecution(st, t)
	return ok && taskExecutionInputsCurrent(st, current)
}
func prepareTaskExecution(st room.State, t room.Task, mode string, inputs []taskcontract.Binding, actor string) (*room.TaskExecution, error) {
	if t.Definition == nil {
		if st.FeatureFlags["taskDefinitions"] {
			return nil, errors.New("define this legacy task before dispatch while task definitions are enabled")
		}
		if len(inputs) > 0 {
			return nil, errors.New("legacy tasks do not accept typed input bindings")
		}
		return nil, nil
	}
	if actor != ownerID {
		return nil, errors.New("structured task dispatch requires the human owner")
	}
	if !st.FeatureFlags["taskDefinitions"] {
		return nil, errors.New("task definitions are disabled; structured task dispatch is blocked")
	}
	if err := taskcontract.Validate(*t.Definition); err != nil {
		return nil, err
	}
	if err := taskcontract.Executable(*t.Definition); err != nil {
		return nil, err
	}
	execution := &room.TaskExecution{DelegationID: t.DelegationID, DefinitionEpoch: st.TaskDefinitionEpoch, CompanyID: st.Company.ID, RoomID: t.RoomID, TaskID: t.ID, TaskRevision: t.Revision, InitiatorID: actor, Mode: mode, Definition: *t.Definition, DefinitionDigest: taskDefinitionDigest(t), Status: "running", StartedAt: time.Now().UTC(), Outputs: []taskcontract.BoundValue{}}
	if mode == "review" {
		prior, ok := st.TaskExecutions[t.DispatchID]
		if !ok || prior.TaskID != t.ID || prior.RoomID != t.RoomID || prior.DelegationID != t.DelegationID || prior.Mode != "work" || prior.Status != "validated" || prior.DefinitionDigest != execution.DefinitionDigest {
			return nil, errors.New("review requires a current validated work execution")
		}
		if !taskExecutionInputsCurrent(st, prior) {
			return nil, errors.New("review source inputs are stale")
		}
		if len(inputs) != 0 {
			return nil, errors.New("review reuses the work execution's bound inputs; replacement inputs are forbidden")
		}
		execution.ParentExecutionID = prior.ID
		execution.Inputs = prior.Inputs
	} else {
		if t.DelegationID != "" {
			delegation, ok := st.Delegations[t.DelegationID]
			parent := st.TaskExecutions[delegation.DispatchID]
			if !ok || parent.TaskID != t.ID || parent.RoomID != t.RoomID || parent.DelegationID != "" || parent.DefinitionDigest != execution.DefinitionDigest {
				return nil, errors.New("delegation lost its structured parent input binding")
			}
			if len(inputs) == 0 {
				execution.Inputs = parent.Inputs
				return finalizeTaskValueInputs(st, execution)
			}
			requested, err := resolveTaskValueInputs(st, t, inputs)
			if err != nil || !reflect.DeepEqual(requested, parent.Inputs) {
				return nil, errors.New("delegated input bindings cannot change")
			}
		}
		bound, err := resolveTaskValueInputs(st, t, inputs)
		if err != nil {
			return nil, err
		}
		execution.Inputs = bound
	}
	return finalizeTaskValueInputs(st, execution)
}
func taskExecutionReference(st room.State, v room.TaskExecution) string {
	data := map[string]any{"definition": v.Definition, "inputs": v.Inputs, "trust": "task contract and human-bound data; not additional tool authority"}
	if v.Mode == "review" {
		data["reviewedExecutionId"] = v.ParentExecutionID
		data["returnedOutputs"] = st.TaskExecutions[v.ParentExecutionID].Outputs
		data["reviewInstruction"] = "Review the declared criteria against these actual returned values. Report findings; only the human owner can accept."
	} else {
		data["returnInstruction"] = "Return exactly one JSON object with the desired output names as keys and values matching their registered schemas. No Markdown fences, extra outputs, completion claims or fabricated file references."
	}
	return "\nStructured task contract and bound execution inputs:\n" + referenceJSON(data)
}

// Task/attempt identity and role must agree at admission and publication.
// Dispatch status/revision are checked separately for the atomic transition phase.
func taskExecutionRoleMatches(t room.Task, e room.TaskExecution) bool {
	if t.ID != e.TaskID || t.RoomID != e.RoomID || t.DispatchID != e.ID || t.DelegationID != e.DelegationID || taskDefinitionDigest(t) != e.DefinitionDigest {
		return false
	}
	switch e.Mode {
	case "work":
		return t.Status == "Working" && t.Assignee != "" && t.Assignee == e.ParticipantID
	case "review":
		return t.Status == "Reviewing" && t.Reviewer != "" && t.Reviewer == e.ParticipantID
	}
	return false
}

// Pending authority is phase-specific: completion replay has its own completed
// task transition fence. This helper never grants authority from a saved status.
func taskExecutionPendingCurrent(st room.State, t room.Task, e room.TaskExecution) bool {
	return st.FeatureFlags["taskDefinitions"] && e.CompanyID == st.Company.ID && e.Status == "running" && e.DefinitionEpoch == st.TaskDefinitionEpoch && t.DispatchStatus == "pending" && t.Revision > e.TaskRevision && t.Revision-e.TaskRevision == 1 && taskExecutionRoleMatches(t, e) && taskExecutionConfigurationCurrent(st, e) && taskExecutionExportCurrent(st, e) && taskExecutionInputsCurrent(st, e)
}
func structuredTaskResult(st room.State, t room.Task, modeStatus, text string) (room.TaskExecution, error) {
	execution := st.TaskExecutions[t.DispatchID]
	if execution.ID == "" || !taskExecutionPendingCurrent(st, t, execution) {
		return execution, errors.New("structured execution missing or current task authority changed")
	}
	if modeStatus != "Idle" {
		return execution, errors.New("task execution stopped or failed")
	}
	if len(text) > execution.Definition.ExecutionConstraints.MaxResponseBytes {
		return execution, errors.New("task response exceeded the declared byte limit")
	}
	if execution.Mode == "review" {
		return execution, nil
	}
	outputs, err := taskcontract.ReturnedValues(execution.Definition, []byte(text), execution.ID)
	if err != nil {
		return execution, err
	}
	execution.Outputs = outputs
	return execution, nil
}
func taskExecutionOutcome(st room.State, taskID, dispatchID, status string, text ...string) []auditChange {
	old, ok := st.TaskExecutions[dispatchID]
	if !ok || old.TaskID != taskID || old.Status != "running" {
		return nil
	}
	v := old
	v.FinishedAt = time.Now().UTC()
	v.Status = "stopped"
	v.ValidationError = "Execution stopped; explicit human retry required"
	if status == "Idle" && len(text) > 0 {
		v.ResponseDigest = responseDigest(text[0])
		v.Status = "reviewed"
		v.ValidationError = ""
		if old.Mode == "work" {
			outputs, err := taskcontract.ReturnedValues(old.Definition, []byte(text[0]), old.ID)
			if err != nil {
				v.Status = "failed"
				v.ValidationError = err.Error()
			} else {
				v.Status = "validated"
				v.Outputs = outputs
			}
		}
	} else if status == "Blocked" {
		v.Status = "failed"
		v.ValidationError = "Execution failed; explicit human retry required"
	}
	return []auditChange{{Kind: "task.execution.finished", Payload: v}}
}
func projectTaskExecution(st *room.State, kind string, raw []byte) error {
	var v room.TaskExecution
	var actor struct {
		PerformedBy string `json:"performedBy"`
	}
	if json.Unmarshal(raw, &v) != nil || v.ID == "" || v.CompanyID != st.Company.ID || v.InitiatorID != ownerID || v.DefinitionDigest != taskcontract.Digest(v.Definition) || taskcontract.Validate(v.Definition) != nil {
		return errors.New("invalid task execution contract binding")
	}
	if st.TaskExecutions == nil {
		st.TaskExecutions = map[string]room.TaskExecution{}
	}
	old, exists := st.TaskExecutions[v.ID]
	if kind == "task.execution.started" {
		_ = json.Unmarshal(raw, &actor)
		if !validTaskExecutionConfiguration(*st, v) || exists || actor.PerformedBy != ownerID || !st.FeatureFlags["taskDefinitions"] || v.DefinitionEpoch != st.TaskDefinitionEpoch || v.Status != "running" || v.StartedAt.IsZero() || !v.FinishedAt.IsZero() || (v.Outputs == nil || len(v.Outputs) != 0) || v.ResponseDigest != "" || v.ValidationError != "" || (v.Mode != "work" && v.Mode != "review") {
			return errors.New("invalid task execution start")
		}
		found := false
		for _, t := range st.Tasks {
			if taskExecutionRoleMatches(t, v) && t.DispatchStatus == "pending" && t.Revision == v.TaskRevision+1 {
				found = true
			}
		}
		if !found {
			return errors.New("execution does not bind the pending task")
		}
		bindings := []taskcontract.Binding{}
		for _, b := range v.Inputs {
			bindings = append(bindings, taskcontract.Binding{Name: b.Name, Value: b.Value, SourceVersion: b.SourceVersion})
		}
		bound, err := taskcontract.Bind(v.Definition, bindings)
		for i := range bound {
			bound[i].Provenance = v.Inputs[i].Provenance
		}
		if err != nil || !reflect.DeepEqual(bound, v.Inputs) {
			return errors.New("invalid bound execution inputs")
		}
		if !taskExecutionInputsCurrent(*st, v) {
			return errors.New("invalid or stale returned-value input provenance")
		}
		if v.DelegationID != "" {
			parent := st.TaskExecutions[st.Delegations[v.DelegationID].DispatchID]
			if parent.TaskID != v.TaskID || parent.RoomID != v.RoomID || parent.DelegationID != "" || parent.DefinitionDigest != v.DefinitionDigest || !reflect.DeepEqual(parent.Inputs, v.Inputs) {
				return errors.New("delegated execution input binding changed")
			}
		}
		if v.Mode == "review" {
			parent := st.TaskExecutions[v.ParentExecutionID]
			if parent.Status != "validated" || parent.Mode != "work" || parent.TaskID != v.TaskID || parent.RoomID != v.RoomID || parent.DelegationID != v.DelegationID || parent.DefinitionDigest != v.DefinitionDigest || !reflect.DeepEqual(parent.Inputs, v.Inputs) {
				return errors.New("review lost its work input/output binding")
			}
		} else if v.ParentExecutionID != "" {
			return errors.New("work execution has an unexpected parent")
		}
	} else {
		if !exists || old.Status != "running" || v.FinishedAt.IsZero() || v.FinishedAt.Before(v.StartedAt) || (v.Status != "validated" && v.Status != "reviewed" && v.Status != "failed" && v.Status != "stopped") {
			return errors.New("invalid task execution finish")
		}
		identity := v
		identity.Status = old.Status
		identity.FinishedAt = old.FinishedAt
		identity.ResponseDigest = old.ResponseDigest
		identity.Outputs = old.Outputs
		identity.ValidationError = old.ValidationError
		if !reflect.DeepEqual(identity, old) {
			return errors.New("execution snapshot mutated")
		}
		if (v.Status == "validated" || v.Status == "reviewed") && (!st.FeatureFlags["taskDefinitions"] || v.DefinitionEpoch != st.TaskDefinitionEpoch || old.ConfigurationVersion != 0 && !taskExecutionConfigurationCurrent(*st, old) || !taskExecutionInputsCurrent(*st, old)) {
			return errors.New("execution authority revoked before result")
		}
		if v.Status == "validated" || v.Status == "reviewed" {
			matched := false
			for _, task := range st.Tasks {
				if taskExecutionRoleMatches(task, v) && task.DispatchStatus == "completed" && task.Revision > v.TaskRevision && task.Revision-v.TaskRevision == 2 {
					matched = true
				}
			}
			if !matched {
				return errors.New("successful execution lost its completed task/role binding")
			}
		}

		if v.Status == "validated" {
			if v.Mode != "work" || len(v.ResponseDigest) != 64 || v.ValidationError != "" {
				return errors.New("invalid returned value validation")
			}
			values := map[string]json.RawMessage{}
			for _, o := range v.Outputs {
				if _, exists := values[o.Name]; exists {
					return errors.New("duplicate output")
				}
				values[o.Name] = o.Value
			}
			rawValues, _ := json.Marshal(values)
			expected, err := taskcontract.ReturnedValues(v.Definition, rawValues, v.ID)
			if err != nil || !reflect.DeepEqual(expected, v.Outputs) {
				return errors.New("returned values do not match contract")
			}
		}
		if (v.Status == "failed" || v.Status == "stopped") && (len(v.Outputs) != 0 || v.Outputs == nil || v.ValidationError == "" || len(v.ValidationError) > 512 || (v.ResponseDigest != "" && len(v.ResponseDigest) != 64)) {
			return errors.New("invalid failed/stopped execution record")
		}
		if v.Status == "reviewed" && (v.Mode != "review" || len(v.ResponseDigest) != 64 || len(v.Outputs) != 0 || v.ValidationError != "") {
			return errors.New("invalid review outcome")
		}
	}
	st.TaskExecutions[v.ID] = v
	return nil
}
