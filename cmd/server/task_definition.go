package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func taskDefinitionDigest(t room.Task) string {
	if t.Definition == nil {
		return ""
	}
	return taskcontract.Digest(*t.Definition)
}
func prepareTaskDefinition(w http.ResponseWriter, r *http.Request, st room.State, req eventRequest, v *room.Task) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(req.Payload, &fields) != nil {
		http.Error(w, "Invalid task", 400)
		return false
	}
	var old room.Task
	for _, t := range st.Tasks {
		if t.ID == v.ID && t.RoomID == v.RoomID {
			old = t
		}
	}
	raw, present := fields["definition"]
	if !present {
		v.Definition = old.Definition
	} else {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			http.Error(w, "A saved task definition cannot be removed", 400)
			return false
		}
		d, err := taskcontract.Decode(raw)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return false
		}
		v.Definition = &d
	}
	changed := v.Definition != nil && (old.Definition == nil || !taskcontract.Same(*old.Definition, *v.Definition))
	if old.Definition != nil && v.Definition != nil && v.Definition.Revision != old.Definition.Revision {
		http.Error(w, "Task definition changed; reload", 409)
		return false
	}
	if changed {
		if v.Status == "Accepted" {
			http.Error(w, "Revise the definition in an unaccepted task state", 409)
			return false
		}
		if requestActor(r) != ownerID {
			http.Error(w, "Only the owner may define or revise task contracts", 403)
			return false
		}
		if !st.FeatureFlags["taskDefinitions"] {
			http.Error(w, "Task definitions are disabled", 403)
			return false
		}
		if old.DelegationID != "" {
			http.Error(w, "Delegated task definition cannot change", 409)
			return false
		}
		next := 1
		if old.Definition != nil {
			if old.Definition.Revision == int(^uint(0)>>1) {
				http.Error(w, "Definition revision exhausted", 409)
				return false
			}
			next = old.Definition.Revision + 1
		}
		v.Definition.Revision = next
	}
	if old.ID == "" && st.FeatureFlags["taskDefinitions"] && v.Definition == nil {
		http.Error(w, "New tasks require all seven definition parts", 400)
		return false
	}
	if v.Definition != nil {
		if requestActor(r) != ownerID && v.Title != old.Title {
			http.Error(w, "Only the owner may change a defined task's scope", 403)
			return false
		}
		v.Criteria = taskcontract.CriteriaText(*v.Definition)
		if v.Status == "Accepted" && st.FeatureFlags["taskAcceptanceReview"] {
			http.Error(w, "Review each criterion against returned values using task acceptance review", 409)
			return false
		}
		if v.Status == "Accepted" && old.Status != "Accepted" {
			valid := taskHasValidCurrentOutput(st, old)
			if !valid {
				http.Error(w, "A current structurally valid returned output is required before human acceptance", 409)
				return false
			}
		}
	}
	return true
}
func taskDefinitionContextAllowed(ctx context.Context, st room.State, p room.Participant) bool {
	binding, ok := ctx.Value(taskJobContextKey{}).(taskJobBinding)
	if !ok || binding.DefinitionDigest == "" {
		return true
	}
	if !st.FeatureFlags["taskDefinitions"] || binding.DefinitionEpoch != st.TaskDefinitionEpoch {
		return false
	}
	for _, t := range st.Tasks {
		if t.ID == binding.TaskID && t.RoomID == p.RoomID && t.DispatchID == binding.DispatchID && taskDefinitionDigest(t) == binding.DefinitionDigest {
			execution, exists := st.TaskExecutions[binding.DispatchID]
			return exists && execution.ParticipantID == p.ID && execution.DefinitionEpoch == binding.DefinitionEpoch && binding.DelegationID == t.DelegationID && taskExecutionPendingCurrent(st, t, execution)
		}
	}
	return false
}
func validateTaskDefinitionProjection(st room.State, t room.Task, raw []byte) error {
	var old room.Task
	for _, item := range st.Tasks {
		if item.ID == t.ID {
			old = item
		}
	}
	if t.Definition == nil {
		if old.Definition != nil {
			return errors.New("task definition removed")
		}
		return nil
	}
	if err := taskcontract.Validate(*t.Definition); err != nil {
		return err
	}
	changed := old.Definition == nil || !taskcontract.Same(*old.Definition, *t.Definition)
	if changed {
		if old.DispatchStatus == "pending" || old.DelegationID != "" {
			return errors.New("task definition changed while authority was bound")
		}
		var actor struct {
			PerformedBy string `json:"performedBy"`
		}
		if json.Unmarshal(raw, &actor) != nil || actor.PerformedBy != ownerID || !st.FeatureFlags["taskDefinitions"] {
			return errors.New("invalid task definition provenance")
		}
		expected := 1
		if old.Definition != nil {
			expected = old.Definition.Revision + 1
		}
		if t.Definition.Revision != expected {
			return errors.New("invalid task definition revision")
		}
	}
	if !changed && old.Definition != nil && t.Definition.Revision != old.Definition.Revision {
		return errors.New("task definition revision changed without contract")
	}
	if t.Status == "Accepted" && (old.Status != "Accepted" || changed) {
		var actor struct {
			PerformedBy string `json:"performedBy"`
		}
		_ = json.Unmarshal(raw, &actor)
		if changed || actor.PerformedBy != ownerID || !taskHasValidCurrentOutput(st, old) {
			return errors.New("typed task accepted without a current valid output and human decision")
		}
	}
	if t.Status == "Accepted" && st.FeatureFlags["taskAcceptanceReview"] && !taskHasCurrentAcceptance(st, t) {
		return errors.New("task accepted without a current criterion review")
	}
	if t.Criteria != taskcontract.CriteriaText(*t.Definition) {
		return errors.New("task criteria disagree with definition")
	}
	return nil
}
