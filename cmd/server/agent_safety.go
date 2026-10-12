package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"agentroom/internal/room"
)

// agentsBlockedLocked requires s.mu. Persisted stops survive restart; volatile
// latches stop dispatch locally even when recording the stop fails.
func (s *server) agentsBlockedLocked(st room.State, rid string) bool {
	if st.AgentsStopped || s.stopAll || s.stopRooms[rid] {
		return true
	}
	for _, id := range st.AgentStopRooms {
		if id == rid {
			return true
		}
	}
	return false
}
func (s *server) agentSafety(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	var req struct {
		Action string `json:"action"`
		Scope  string `json:"scope"`
		Reason string `json:"reason"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || (req.Action != "stop" && req.Action != "resume") || (req.Scope != "room" && req.Scope != "application") || len(req.Reason) > 500 {
		http.Error(w, "Choose stop/resume and room/application scope; reason maximum 500 characters", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	rid := requestedRoom(r)
	var st room.State
	var err error
	// Application stop does not depend on a state read: cancel first, audit second.
	if req.Action == "resume" || req.Scope == "room" {
		st, err = s.load(ctx)
		if err != nil {
			http.Error(w, "State unavailable; use Stop all agents if stopping", 503)
			return
		}
	}
	if req.Scope == "room" && !hasRoom(st, rid) {
		http.Error(w, "Room not found", 404)
		return
	}
	if req.Action == "resume" && (st.Incident || (req.Scope == "room" && (st.AgentsStopped || s.stopAll))) {
		http.Error(w, "Resolve the integrity incident or application stop first", 409)
		return
	}
	if req.Action == "stop" {
		if s.stopRooms == nil {
			s.stopRooms = map[string]bool{}
		}
		if req.Scope == "application" {
			s.stopAll = true
			for _, cancel := range s.cancels {
				cancel()
			}
		} else {
			s.stopRooms[rid] = true
			for _, p := range st.Participants {
				if p.RoomID == rid {
					if cancel := s.cancels[p.ID]; cancel != nil {
						cancel()
					}
				}
			}
		}
	}
	if req.Action == "stop" {
		for _, run := range st.WorkflowRuns {
			if req.Scope == "application" || run.RoomID == rid {
				if cancel := s.workflowLive[run.ID]; cancel != nil {
					cancel()
				}
			}
		}
	}
	event := map[string]string{"actorId": ownerID, "scope": req.Scope, "reason": strings.TrimSpace(req.Reason)}
	if req.Scope == "room" {
		event["roomId"] = rid
	}
	if err = s.append(ctx, "agents."+req.Action, event); err != nil {
		http.Error(w, "Audit recording failed; local stop remains active. Do not restart until investigated.", 503)
		return
	}
	if req.Action == "resume" {
		if req.Scope == "application" {
			s.stopAll = false
		} else {
			delete(s.stopRooms, rid)
		}
	}
	reply(w, map[string]any{"action": req.Action, "scope": req.Scope, "recorded": true, "remoteTerminationConfirmed": false})
}

func agentSafetyUIEnabled() (bool, error) {
	value := os.Getenv("COMMONROOM_AGENT_SAFETY_UI")
	if value == "" {
		return true, nil
	}
	return strconv.ParseBool(value)
}

// The stored owner setting overrides the environment default. Neither can
// disable cutoff enforcement; only the optional presentation script is gated.
func (s *server) agentSafetyControlsEnabled(st room.State) bool {
	if enabled, ok := st.FeatureFlags["agentSafetyControls"]; ok {
		return enabled
	}
	return s.safetyUIDefault
}
func (s *server) agentSafetyScript(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		st, err := s.load(ctx)
		if err != nil {
			http.Error(w, "Feature settings unavailable", 503)
			return
		}
		optionalScriptHandler(s.agentSafetyControlsEnabled(st), root, "agent-safety.js").ServeHTTP(w, r)
	})
}
