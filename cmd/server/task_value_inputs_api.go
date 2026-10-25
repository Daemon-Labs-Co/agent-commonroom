package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"net/http"
	"sort"
)

type taskInputSource struct {
	TaskID       string `json:"taskId"`
	TaskTitle    string `json:"taskTitle"`
	TaskRevision int    `json:"taskRevision"`
	ExecutionID  string `json:"executionId"`
	OutputName   string `json:"outputName"`
	ValueJSON    string `json:"valueJson"`
	Digest       string `json:"digest"`
}

// Discovery exposes only accepted, current, compatible shared-room values.
// The eventual dispatch repeats resolution; discovery is not authorization.
func (s *server) taskInputSources(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	rid := requestedRoom(r)
	if !canAccess(st, requestActor(r), rid) || !sharedTaskValueRoom(st, rid) {
		http.Error(w, "Room unavailable", 404)
		return
	}
	if !st.FeatureFlags["taskDefinitions"] || !st.FeatureFlags["taskValueDependencies"] {
		http.Error(w, "Verified task value dependencies are disabled", 403)
		return
	}
	var target room.Task
	for _, t := range st.Tasks {
		if t.ID == r.URL.Query().Get("taskId") && t.RoomID == rid {
			target = t
			break
		}
	}
	if target.Definition == nil {
		http.Error(w, "Defined task unavailable", 404)
		return
	}
	var contract taskcontract.DataContract
	for _, c := range target.Definition.ExpectedInputs {
		if c.Name == r.URL.Query().Get("input") {
			contract = c
			break
		}
	}
	if contract.Type != "value" {
		http.Error(w, "Value input unavailable", 404)
		return
	}
	// Discovery checks one input independently of other required bindings.
	single := target
	definition := *target.Definition
	definition.ExpectedInputs = []taskcontract.DataContract{contract}
	single.Definition = &definition
	list := []taskInputSource{}
	for _, source := range st.Tasks {
		if source.ID == target.ID || source.RoomID != rid || source.Status != "Accepted" || source.DispatchStatus == "pending" {
			continue
		}
		execution, ok := taskCurrentWorkExecution(st, source)
		if !ok {
			continue
		}
		for _, output := range execution.Outputs {
			if _, err := resolveTaskValueInputs(st, single, []taskcontract.Binding{{Name: contract.Name, Source: &taskcontract.ValueReference{ExecutionID: execution.ID, OutputName: output.Name}}}); err != nil {
				continue
			}
			list = append(list, taskInputSource{TaskID: source.ID, TaskTitle: source.Title, TaskRevision: source.Revision, ExecutionID: execution.ID, OutputName: output.Name, ValueJSON: output.ValueJSON, Digest: output.Digest})
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].ExecutionID == list[j].ExecutionID {
			return list[i].OutputName < list[j].OutputName
		}
		return list[i].ExecutionID < list[j].ExecutionID
	})
	start := 0
	if cursor := r.URL.Query().Get("after"); cursor != "" {
		found := false
		for i, v := range list {
			if v.ExecutionID+"/"+v.OutputName == cursor {
				start = i + 1
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Source cursor unavailable; reload", 409)
			return
		}
	}
	end := start + 25
	if end > len(list) {
		end = len(list)
	}
	next := ""
	if end < len(list) {
		v := list[end-1]
		next = v.ExecutionID + "/" + v.OutputName
	}
	reply(w, map[string]any{"sources": list[start:end], "nextCursor": next})
}
