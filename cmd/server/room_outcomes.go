package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"
)

// Outcome metadata defines intent, never execution permission or completion.
// Historical events remain readable regardless of the current feature setting.
func prepareRoomOutcome(w http.ResponseWriter, st room.State, req eventRequest, v *room.Room) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(req.Payload, &fields) != nil {
		http.Error(w, "Invalid room", 400)
		return false
	}
	old := room.Room{}
	for _, r := range st.Rooms {
		if r.ID == v.ID {
			old = r
			break
		}
	}
	_, hasOutcome := fields["outcome"]
	_, hasCriteria := fields["acceptanceCriteria"]
	if !hasOutcome {
		v.Outcome = old.Outcome
	}
	if !hasCriteria {
		v.AcceptanceCriteria = old.AcceptanceCriteria
	}
	v.Outcome = strings.TrimSpace(v.Outcome)
	v.AcceptanceCriteria = strings.TrimSpace(v.AcceptanceCriteria)
	if utf8.RuneCountInString(v.Outcome) > 2000 || utf8.RuneCountInString(v.AcceptanceCriteria) > 8000 {
		http.Error(w, "Outcome maximum 2000 characters; acceptance criteria maximum 8000", 400)
		return false
	}
	changed := v.Outcome != old.Outcome || v.AcceptanceCriteria != old.AcceptanceCriteria
	if !st.FeatureFlags["roomOutcomes"] {
		if changed {
			http.Error(w, "Room outcomes are disabled", 403)
			return false
		}
		return true
	}
	if old.PersonalActorID != "" {
		if changed {
			http.Error(w, "Outcomes belong to shared work rooms", 400)
			return false
		}
		return true
	}
	// Existing rooms can continue without a deliverable until explicitly defined.
	// Once defined, ordinary updates cannot silently clear either field.
	if req.Kind == "room.created" || changed || old.Outcome != "" || old.AcceptanceCriteria != "" {
		if v.Outcome == "" || v.AcceptanceCriteria == "" {
			http.Error(w, "Shared room outcome and acceptance criteria required", 400)
			return false
		}
	}
	return true
}
