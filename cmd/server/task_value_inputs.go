package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"bytes"
	"errors"
)

func hasTaskValueSources(inputs []taskcontract.BoundValue) bool {
	for _, input := range inputs {
		if input.Provenance != nil {
			return true
		}
	}
	return false
}
func sharedTaskValueRoom(st room.State, rid string) bool {
	for _, r := range st.Rooms {
		if r.ID == rid {
			return r.PersonalActorID == ""
		}
	}
	return false
}

// Select source metadata only. Callers must validate complete lineage using one
// shared path/depth/node budget; acceptance metadata must not start another walk.
func taskValueSource(st room.State, rid, executionID, outputName string) (room.Task, room.TaskExecution, taskcontract.BoundValue, bool) {
	execution, ok := st.TaskExecutions[executionID]
	if !ok || execution.CompanyID != st.Company.ID || execution.RoomID != rid || !sharedTaskValueRoom(st, rid) {
		return room.Task{}, room.TaskExecution{}, taskcontract.BoundValue{}, false
	}
	for _, t := range st.Tasks {
		if t.ID != execution.TaskID || t.RoomID != rid || !taskHasRecordedAcceptance(st, t) || t.DispatchStatus == "pending" {
			continue
		}
		current, valid := taskCurrentWorkExecution(st, t)
		if !valid || current.ID != executionID {
			break
		}
		for _, output := range current.Outputs {
			if output.Name == outputName {
				return t, current, output, true
			}
		}
	}
	return room.Task{}, room.TaskExecution{}, taskcontract.BoundValue{}, false
}

// Resolution uses only stored returned bytes. A client-supplied value or version
// accompanying a reference is refused, never promoted into trusted provenance.
func resolveTaskValueInputs(st room.State, t room.Task, inputs []taskcontract.Binding) ([]taskcontract.BoundValue, error) {
	resolved := make([]taskcontract.Binding, len(inputs))
	provenance := map[string]*taskcontract.ValueProvenance{}
	for i, input := range inputs {
		resolved[i] = input
		if input.Source == nil {
			continue
		}
		if !st.FeatureFlags["taskDefinitions"] || !st.FeatureFlags["taskValueDependencies"] || !sharedTaskValueRoom(st, t.RoomID) {
			return nil, errors.New("verified value dependencies are disabled or unavailable in this room")
		}
		if len(input.Value) != 0 || input.ValueJSON != "" || input.SourceVersion != "" {
			return nil, errors.New("returned-value references cannot supply replacement values or labels")
		}
		source, execution, output, ok := taskValueSource(st, t.RoomID, input.Source.ExecutionID, input.Source.OutputName)
		if !ok || source.ID == t.ID {
			return nil, errors.New("source must be another accepted task's current returned value in this shared room")
		}
		path := map[string]bool{t.ID: true}
		nodes := 0
		if !taskValueLineageCurrent(st, execution, path, 0, &nodes) {
			return nil, errors.New("source value has stale, cyclic or excessive input dependencies")
		}
		resolved[i] = taskcontract.Binding{Name: input.Name, Value: output.Value, SourceVersion: execution.ID + "/" + output.Name}
		provenance[input.Name] = &taskcontract.ValueProvenance{TaskID: source.ID, TaskRevision: source.Revision, ExecutionID: execution.ID, OutputName: output.Name, DefinitionDigest: execution.DefinitionDigest, OutputDigest: output.Digest}
	}
	bound, err := taskcontract.Bind(*t.Definition, resolved)
	if err != nil {
		return nil, err
	}
	for i := range bound {
		bound[i].Provenance = provenance[bound[i].Name]
	}
	return bound, nil
}
func finalizeTaskValueInputs(st room.State, v *room.TaskExecution) (*room.TaskExecution, error) {
	if hasTaskValueSources(v.Inputs) {
		v.ValueDependencyEpoch = st.ValueDependencyEpoch
		v.InputAcceptanceEpoch = st.TaskAcceptanceEpoch
	}
	if !taskExecutionInputsCurrent(st, *v) {
		return nil, errors.New("task input dependency is stale or revoked")
	}
	return v, nil
}

// Every use retains one path and one budget across the whole dependency DAG.
// Completed history survives flag cycling; in-flight authority never revives.
func taskExecutionInputsCurrent(st room.State, v room.TaskExecution) bool {
	if !hasTaskValueSources(v.Inputs) {
		return v.ValueDependencyEpoch == 0 && v.InputAcceptanceEpoch == 0
	}
	if !st.FeatureFlags["taskValueDependencies"] || v.Status == "running" && (v.ValueDependencyEpoch != st.ValueDependencyEpoch || v.InputAcceptanceEpoch != st.TaskAcceptanceEpoch) {
		return false
	}
	nodes := 0
	return taskValueLineageCurrent(st, v, map[string]bool{}, 0, &nodes)
}
func taskValueLineageCurrent(st room.State, v room.TaskExecution, path map[string]bool, depth int, nodes *int) bool {
	*nodes++
	if depth > 16 || *nodes > 128 || path[v.TaskID] {
		return false
	}
	path[v.TaskID] = true
	defer delete(path, v.TaskID)
	for _, input := range v.Inputs {
		p := input.Provenance
		if p == nil {
			continue
		}
		if !st.FeatureFlags["taskValueDependencies"] {
			return false
		}
		t, source, output, ok := taskValueSource(st, v.RoomID, p.ExecutionID, p.OutputName)
		if !ok || t.ID != p.TaskID || t.Revision != p.TaskRevision || source.DefinitionDigest != p.DefinitionDigest || output.Digest != p.OutputDigest || input.Digest != output.Digest || input.SourceVersion != source.ID+"/"+output.Name || !bytes.Equal(input.Value, output.Value) || input.ValueJSON != output.ValueJSON {
			return false
		}
		if !taskValueLineageCurrent(st, source, path, depth+1, nodes) {
			return false
		}
	}
	return true
}
