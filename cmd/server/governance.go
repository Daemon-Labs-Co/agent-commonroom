package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Governance records coordinate work; they are never execution capabilities.
// Every effect still requires the ordinary authenticated, scoped broker grant.
func eligibleGovernor(st room.State, rid, pid string) bool {
	for _, p := range st.Participants {
		if p.ID != pid || p.RoomID != rid || (p.Status != "Idle" && p.Status != "Working") {
			continue
		}
		for _, c := range st.Connections {
			if c.ID == p.ConnectionID && room.SupportsAgent(c) {
				return true
			}
		}
	}
	return false
}

func projectGovernance(st *room.State, data []byte) error {
	var v room.GovernanceAppointment
	var actor struct {
		PerformedBy string `json:"performedBy"`
	}
	if json.Unmarshal(data, &v) != nil || json.Unmarshal(data, &actor) != nil || actor.PerformedBy != ownerID || !st.FeatureFlags["agentCoordination"] || v.CompanyID != st.Company.ID || !hasRoom(*st, v.RoomID) || !eligibleGovernor(*st, v.RoomID, v.ParticipantID) || v.Revision < 1 || v.Revision-1 != st.Governance[v.RoomID].Revision {
		return errors.New("invalid governance appointment")
	}
	if st.Governance == nil {
		st.Governance = map[string]room.GovernanceAppointment{}
	}
	st.Governance[v.RoomID] = v
	return nil
}

func (s *server) governance(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method == "POST" && !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
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
	if !hasRoom(st, rid) || !canAccess(st, requestActor(r), rid) {
		http.Error(w, "Room unavailable", 404)
		return
	}
	current := st.Governance[rid]
	enabled := st.FeatureFlags["agentCoordination"]
	if r.Method == "GET" {
		reply(w, map[string]any{"enabled": enabled, "appointment": current, "available": enabled && !st.Incident && !s.agentsBlockedLocked(st, rid) && eligibleGovernor(st, rid, current.ParticipantID)})
		return
	}
	if !enabled {
		http.Error(w, "Agent coordination is disabled", 403)
		return
	}
	var req struct {
		ParticipantID string `json:"participantId"`
		Revision      int    `json:"revision"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	raw, readErr := io.ReadAll(r.Body)
	fields, fieldErr := fenceObjectFields(raw, "participantId", "revision")
	if readErr != nil || fieldErr != nil || len(fields) != 2 || json.Unmarshal(raw, &req) != nil || req.ParticipantID == "" || req.Revision < 0 || r.Header.Get("Idempotency-Key") == "" {
		http.Error(w, "Participant, current revision and idempotency key required", 400)
		return
	}
	ctx, replayed := prepareOperation(w, r, st, req)
	if replayed {
		return
	}
	if st.Incident || s.agentsBlockedLocked(st, rid) {
		http.Error(w, "Coordination blocked by safety stop", 409)
		return
	}
	if req.Revision != current.Revision || current.Revision == int(^uint(0)>>1) {
		http.Error(w, "Governance changed; reload before selecting", 409)
		return
	}
	if !eligibleGovernor(st, rid, req.ParticipantID) {
		http.Error(w, "Choose an eligible agent in this room", 400)
		return
	}
	if err = s.workflowSafety(ctx, st, rid); err != nil {
		http.Error(w, "Governance blocked: audit integrity unavailable or failed", 409)
		return
	}
	next := room.GovernanceAppointment{CompanyID: st.Company.ID, RoomID: rid, ParticipantID: req.ParticipantID, Revision: current.Revision + 1}
	response := map[string]any{"committed": true, "appointment": next}
	if err = s.append(withOperationResponse(ctx, 200, response), "governance.appointed", next); err != nil {
		if !errors.Is(err, errAuditConflict) {
			s.tripIntegrityLocked()
		}
		operationError(w, err)
		return
	}
	reply(w, response)
}

func initialGovernanceChanges(st room.State, p room.Participant, c room.Connection) []auditChange {
	if !st.FeatureFlags["agentCoordination"] || st.Governance[p.RoomID].Revision != 0 {
		return nil
	}
	st.Participants = append(append([]room.Participant(nil), st.Participants...), p)
	if c.ID != "" {
		st.Connections = append(append([]room.Connection(nil), st.Connections...), c)
	}
	if !eligibleGovernor(st, p.RoomID, p.ID) {
		return nil
	}
	return []auditChange{{Kind: "governance.appointed", Payload: room.GovernanceAppointment{CompanyID: st.Company.ID, RoomID: p.RoomID, ParticipantID: p.ID, Revision: 1}}}
}

// appendAppointmentBatch requires s.mu, like all registration callers. A failed
// required appointment audit must stop admission even if no events committed.
func (s *server) appendAppointmentBatch(ctx context.Context, changes []auditChange) error {
	err := s.appendBatch(ctx, changes)
	if err != nil && !errors.Is(err, errAuditConflict) {
		for _, change := range changes {
			if change.Kind == "governance.appointed" {
				s.tripIntegrityLocked()
				break
			}
		}
	}
	return err
}
