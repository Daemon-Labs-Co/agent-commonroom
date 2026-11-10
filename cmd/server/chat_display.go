package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"net/http"
	"slices"
)

// Personal chat display preferences. Both are display-only values on the
// authenticated human's own profile: they never reach agent prompts, grant
// nothing and cannot be set for another actor. The payload ID is always
// replaced with the session actor.
const (
	chatDisplayFlag      = "chatDisplayChoices"
	maxHighlightedAgents = 50
)

func chatDisplayEnabled(st room.State) bool {
	enabled, ok := st.FeatureFlags[chatDisplayFlag]
	return !ok || enabled
}

// chatDisplayActor returns the requesting human, or writes the refusal.
func chatDisplayActor(w http.ResponseWriter, r *http.Request, st room.State) (room.Actor, bool) {
	if !chatDisplayEnabled(st) {
		http.Error(w, "Chat display choices are disabled", 403)
		return room.Actor{}, false
	}
	a, ok := member(st, requestActor(r))
	if !ok || a.Kind != "human" {
		http.Error(w, "Human profile required", 403)
		return room.Actor{}, false
	}
	return a, true
}

type chatStyleUpdate struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type agentHighlightUpdate struct {
	ID      string `json:"id"`
	AgentID string `json:"agentId"`
	Enabled bool   `json:"enabled"`
}

func prepareChatStyle(w http.ResponseWriter, r *http.Request, st room.State, payload json.RawMessage) (any, bool) {
	var v struct {
		Enabled *bool `json:"enabled"`
	}
	if json.Unmarshal(payload, &v) != nil || v.Enabled == nil {
		http.Error(w, "Choose whether chat style display is on", 400)
		return nil, false
	}
	a, ok := chatDisplayActor(w, r, st)
	if !ok {
		return nil, false
	}
	return chatStyleUpdate{ID: a.ID, Enabled: *v.Enabled}, true
}

func prepareAgentHighlight(w http.ResponseWriter, r *http.Request, st room.State, payload json.RawMessage) (any, bool) {
	var v struct {
		AgentID string `json:"agentId"`
		Enabled *bool  `json:"enabled"`
	}
	if json.Unmarshal(payload, &v) != nil || v.Enabled == nil || v.AgentID == "" || len(v.AgentID) > 128 {
		http.Error(w, "Choose an agent and whether it is highlighted", 400)
		return nil, false
	}
	a, ok := chatDisplayActor(w, r, st)
	if !ok {
		return nil, false
	}
	// Removing is always allowed so a highlight for a deleted agent can be cleared.
	if *v.Enabled {
		if agent, found := member(st, v.AgentID); !found || agent.Kind != "agent" {
			http.Error(w, "Agent not found", 404)
			return nil, false
		}
		if !slices.Contains(a.HighlightedAgents, v.AgentID) && len(a.HighlightedAgents) >= maxHighlightedAgents {
			http.Error(w, "Highlight limit reached. Remove a highlight first", 409)
			return nil, false
		}
	}
	return agentHighlightUpdate{ID: a.ID, AgentID: v.AgentID, Enabled: *v.Enabled}, true
}

func projectChatStyle(st *room.State, data []byte) error {
	var v chatStyleUpdate
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	projectActors(st)
	for i := range st.Actors {
		if st.Actors[i].ID == v.ID && st.Actors[i].Kind == "human" {
			st.Actors[i].ChatStyle = v.Enabled
		}
	}
	return nil
}

func projectAgentHighlight(st *room.State, data []byte) error {
	var v agentHighlightUpdate
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	projectActors(st)
	for i := range st.Actors {
		if st.Actors[i].ID != v.ID || st.Actors[i].Kind != "human" {
			continue
		}
		kept := slices.DeleteFunc(slices.Clone(st.Actors[i].HighlightedAgents), func(id string) bool { return id == v.AgentID })
		if v.Enabled {
			kept = append(kept, v.AgentID)
		}
		if len(kept) == 0 {
			kept = nil
		}
		st.Actors[i].HighlightedAgents = kept
	}
	return nil
}
