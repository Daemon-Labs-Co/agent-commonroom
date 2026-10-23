package main

// Authenticated event command orchestration: permissions, preparation, atomic persistence and response.

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type eventRequest struct {
	TemplateID string          `json:"templateId"`
	Kind       string          `json:"kind"`
	Payload    json.RawMessage `json:"payload"`
}

func (s *server) events(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		ctx, cancel := s.storageContext(r.Context())
		defer cancel()
		r = r.WithContext(ctx)
		paged := r.URL.Query().Get("page") == "true"
		sql := "SELECT sequence,kind,committed_at,hash,payload FROM events ORDER BY sequence DESC LIMIT 200"
		var args []any
		if paged {
			var err error
			args, err = auditPageBounds(r.URL.Query())
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			sql = auditPageSQL
		}
		refined, flagErr := s.historyPresentationEnabled(r.Context())
		if flagErr != nil {
			http.Error(w, "History settings unavailable", 500)
			return
		}
		rows, err := s.db.Query(r.Context(), sql, args...)
		if err != nil {
			http.Error(w, "Unable to load events", 500)
			return
		}
		defer rows.Close()
		result := []map[string]any{}
		for rows.Next() {
			var seq int64
			var kind, hash string
			var payload []byte
			var at time.Time
			if err = rows.Scan(&seq, &kind, &at, &hash, &payload); err != nil {
				http.Error(w, "Unable to read event", 500)
				return
			}
			event := map[string]any{"sequence": seq, "kind": kind, "at": at, "hash": hash, "summary": auditSummary(kind, payload)}
			if refined {
				event["presentation"] = describeAuditEvent(kind, payload)
			}
			result = append(result, event)
		}
		if rows.Err() != nil || ctx.Err() != nil {
			http.Error(w, "Unable to read events", 500)
			return
		}
		if paged {
			var next any
			if len(result) > 100 {
				result = result[:100]
				next = result[99]["sequence"]
			}
			reply(w, map[string]any{"events": result, "nextBefore": next})
			return
		}
		reply(w, result)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	var req eventRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	if req.TemplateID != "" {
		if !s.templatesEnabled {
			http.Error(w, "Workflow templates are disabled", 403)
			return
		}
		if req.Kind != "workflow.revised" || !knownWorkflowTemplate(req.TemplateID) {
			http.Error(w, "Unknown workflow template", 400)
			return
		}
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	roomID := requestedRoom(r)
	if !hasRoom(st, roomID) {
		http.Error(w, "Room not found", 404)
		return
	}
	if requestActor(r) != ownerID {
		allowed := req.Kind == "actor.help.updated" || req.Kind == "actor.image.updated" || req.Kind == "actor.name.updated" || req.Kind == "message.created" || req.Kind == "memory.revised" || req.Kind == "task.revised" || req.Kind == "message.attention.acknowledged"
		a, _ := member(st, requestActor(r))
		if req.Kind == "workflow.revised" && a.Role == "Developer" {
			allowed = true
		}
		if !allowed {
			http.Error(w, "Owner permission required", 403)
			return
		}
		if req.Kind == "memory.revised" {
			var mem room.Memory
			json.Unmarshal(req.Payload, &mem)
			oldShared := false
			for _, old := range st.Memories {
				if old.ID == mem.ID && old.Scope != "Room" {
					oldShared = true
				}
			}
			if mem.Scope != "Room" || oldShared {
				http.Error(w, "Shared memory scope changes require owner", 403)
				return
			}
		}
		if req.Kind == "task.revised" {
			var task room.Task
			json.Unmarshal(req.Payload, &task)
			if task.Status == "Accepted" {
				http.Error(w, "Acceptance currently requires owner", 403)
				return
			}
		}
	}
	ctx, replayed := prepareOperation(w, r, st, req)
	if replayed {
		return
	}
	r = r.WithContext(ctx)
	var p any
	var ok bool
	switch req.Kind {
	case "agents.idle_policy.updated", "feature.flag.updated", "actor.help.updated", "actor.image.updated", "actor.name.updated", "company.organization.updated", "company.updated", "human.profile.updated":
		p, ok = s.prepareIdentityEvent(w, r, st, roomID, req)
	case "workflow.revised", "room.created", "room.updated", "task.revised", "evidence.created":
		p, ok = s.prepareRoomsEvent(w, r, st, roomID, req)
	case "connection.deleted", "connection.created", "connection.updated", "participant.created", "participant.updated":
		p, ok = s.prepareAgentsEvent(w, r, st, roomID, req)
	case "message.attention.acknowledged", "message.created", "room.context_export.updated", "memory.revised":
		p, ok = s.prepareContentEvent(w, r, st, roomID, req)
	default:
		http.Error(w, "Unsupported operation", 400)
		return
	}
	if !ok {
		return
	}

	result := map[string]any{"committed": true}
	if workflow, ok := p.(room.Workflow); ok {
		result["workflowId"] = workflow.ID
	}
	r = r.WithContext(withOperationResponse(r.Context(), 200, result))
	if req.Kind == "agents.idle_policy.updated" {
		r = r.WithContext(context.WithValue(r.Context(), expectedEventCountKey{}, st.EventCount))
	}
	if req.Kind == "memory.revised" {
		for _, member := range st.Participants {
			if s.active[member.ID] {
				http.Error(w, "Wait for affected agents to finish before changing memory", 409)
				return
			}
		}
	}
	if req.Kind == "room.context_export.updated" || req.Kind == "memory.revised" {
		changes := []auditChange{{Kind: req.Kind, Payload: p}}
		for _, member := range st.Participants {
			if member.RoomID == roomID || req.Kind == "memory.revised" {
				member.ThreadID = ""
				changes = append(changes, auditChange{Kind: "participant.session", Payload: member})
			}
		}
		err = s.appendBatch(r.Context(), changes)
	} else if req.Kind == "participant.created" && st.FeatureFlags["agentCoordination"] && st.Governance[roomID].Revision == 0 {
		if err = s.workflowSafety(r.Context(), st, roomID); err != nil {
			http.Error(w, "Governance blocked: audit integrity unavailable or failed", 409)
			return
		}
		participant := p.(room.Participant)
		candidate := st
		candidate.Participants = append(append([]room.Participant(nil), st.Participants...), participant)
		changes := []auditChange{{Kind: req.Kind, Payload: p}}
		if eligibleGovernor(candidate, roomID, participant.ID) {
			changes = append(changes, auditChange{Kind: "governance.appointed", Payload: room.GovernanceAppointment{CompanyID: st.Company.ID, RoomID: roomID, ParticipantID: participant.ID, Revision: 1}})
		}
		err = s.appendBatch(r.Context(), changes)
	} else {
		err = s.append(r.Context(), req.Kind, p)
	}
	if err != nil {
		log.Print(err)
		operationError(w, err)
		return
	}
	if req.Kind == "feature.flag.updated" {
		var flag struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		}
		json.Unmarshal(req.Payload, &flag)
		if flag.Name == "workflowExecution" && !flag.Enabled {
			for _, cancel := range s.workflowLive {
				cancel()
			}
		}
	}
	reply(w, result)
}
