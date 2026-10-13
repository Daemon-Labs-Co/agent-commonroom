package main

// Authenticated session and state HTTP projections.

import (
	"encoding/json"
	"net/http"
)

func reply(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *server) session(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	session, ok := s.auth.session(r)
	if !ok {
		http.Error(w, "Sign in required", 401)
		return
	}
	s.auth.mu.Lock()
	pending := s.auth.setupPending
	s.auth.mu.Unlock()
	reply(w, map[string]any{"csrf": session.CSRF, "actorId": requestActor(r), "actorType": "human", "keySetupPending": pending && requestActor(r) == ownerID})
}

func (s *server) state(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "Unable to load state", 500)
		return
	}
	rid := requestedRoom(r)
	if requestActor(r) == ownerID {
		for _, room := range st.Rooms {
			if room.ID == rid && room.PersonalActorID != "" && room.PersonalActorID != ownerID {
				if e := s.append(r.Context(), "room.owner_access", map[string]string{"actorId": ownerID, "roomId": rid, "reason": "owner viewed personal room"}); e != nil {
					http.Error(w, "Unable to audit owner access", 503)
					return
				}
			}
		}
	}
	s.mu.Lock()
	if st.FeatureFlags == nil {
		st.FeatureFlags = map[string]bool{}
	}
	st.FeatureFlags["agentSafetyControls"] = s.agentSafetyControlsEnabled(st)
	st.FeatureFlags["humanProfileImages"] = humanProfileImagesEnabled(st)
	st.AgentRoomStopped = s.agentsBlockedLocked(st, rid)
	st.AgentsStopped = st.AgentsStopped || s.stopAll
	st.Incident = st.Incident || s.integrityFault
	if st.Incident {
		for i := range st.Participants {
			if d := st.Participants[i].StatusDetails; d != nil {
				d.LongRunning = false
				d.Reason = "Audit integrity is failed or unverified."
				d.Next = "The owner must inspect History and explicitly authorize verified recovery."
			}
		}
	}
	if st.AgentRoomStopped && !st.Incident {
		for i := range st.Participants {
			if st.Participants[i].RoomID == rid && st.Participants[i].StatusDetails != nil {
				st.Participants[i].StatusDetails.Reason = "Emergency stop is active."
				st.Participants[i].StatusDetails.Next = "The owner must explicitly resume execution after investigating."
			}
		}
	}
	s.mu.Unlock()
	reply(w, visibleState(st, requestActor(r), rid))
}
