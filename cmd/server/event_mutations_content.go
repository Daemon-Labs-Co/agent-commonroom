package main

// Content event validation and typed payload preparation; persistence stays in the handler.

import (
	"agentroom/internal/room"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *server) prepareContentEvent(w http.ResponseWriter, r *http.Request, st room.State, roomID string, req eventRequest) (any, bool) {
	var p any
	switch req.Kind {
	case "message.attention.acknowledged":
		actor, ok := member(st, requestActor(r))
		if !ok || actor.Kind != "human" {
			http.Error(w, "Human account required", 403)
			return nil, false
		}
		if !humanAttentionEnabled(st) {
			http.Error(w, "Human attention is disabled", 403)
			return nil, false
		}
		var reqAck struct {
			MessageID string `json:"messageId"`
		}
		if json.Unmarshal(req.Payload, &reqAck) != nil {
			http.Error(w, "Invalid acknowledgment", 400)
			return nil, false
		}
		found := false
		for _, message := range st.Messages {
			if message.ID == reqAck.MessageID && message.RoomID == roomID && message.ActorType == "agent" && message.HumanRequest != "" {
				found = true
				if message.AttentionAcknowledgedBy != "" {
					reply(w, map[string]bool{"committed": true})
					return nil, false
				}
			}
		}
		if !found {
			http.Error(w, "Agent request not found in this room", 404)
			return nil, false
		}
		p = map[string]string{"messageId": reqAck.MessageID, "roomId": roomID, "actorId": requestActor(r)}
	case "message.created":
		var v room.Message
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Text) == "" {
			http.Error(w, "Message required", 400)
			return nil, false
		}
		v.RoomID = roomID
		v.ID = id()
		v.ActorID = requestActor(r)
		v.ActorType = "human"
		v.HumanRequest = ""
		v.AttentionAcknowledgedBy = ""
		actor, _ := member(st, v.ActorID)
		v.Author = actor.Name
		v.Created = time.Now().UTC()
		p = v
	case "room.context_export.updated":
		if !contextPolicyEditorEnabled(st) {
			http.Error(w, "Context policy editor is disabled", 403)
			return nil, false
		}
		var v struct {
			Revision int                      `json:"revision"`
			RoomID   string                   `json:"roomId"`
			Policy   room.ContextExportPolicy `json:"policy"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || !validExportProviders(v.Policy.Providers) {
			http.Error(w, "Choose supported export providers", 400)
			return nil, false
		}
		for _, r := range st.Rooms {
			if r.ID == roomID && v.Revision != r.Revision {
				http.Error(w, "Room policy changed; reload", 409)
				return nil, false
			}
		}
		v.RoomID = roomID
		for _, member := range st.Participants {
			if member.RoomID == roomID && s.active[member.ID] {
				http.Error(w, "Stop room agents before changing disclosure policy", 409)
				return nil, false
			}
		}
		p = v
	case "memory.revised":
		var v room.Memory
		if json.Unmarshal(req.Payload, &v) != nil || v.Title == "" {
			http.Error(w, "Title required", 400)
			return nil, false
		}
		v.RoomID = roomID
		v.SourceActorID = requestActor(r)
		if !validExportProviders(v.ExportProviders) {
			http.Error(w, "Invalid memory export providers", 400)
			return nil, false
		}
		if requestActor(r) != ownerID {
			var old room.Memory
			for _, m := range st.Memories {
				if m.ID == v.ID {
					old = m
				}
			}
			if referenceJSON(v.AgentExport) != referenceJSON(old.AgentExport) || referenceJSON(v.ExportProviders) != referenceJSON(old.ExportProviders) {
				http.Error(w, "Export permissions require owner", 403)
				return nil, false
			}
		}
		for _, r := range st.Rooms {
			if r.ID == roomID && r.PersonalActorID != "" && v.Scope != "Room" {
				http.Error(w, "Personal-room memory must remain room scoped", 403)
				return nil, false
			}
		}
		if v.Scope != "Room" && v.Scope != "Project" && v.Scope != "Application" {
			http.Error(w, "Invalid scope", 400)
			return nil, false
		}
		if v.ID == "" {
			v.ID = id()
			v.Revision = 1
		} else {
			found := false
			for _, m := range st.Memories {
				if m.ID == v.ID && (m.RoomID == roomID || m.Scope != "Room") {
					v.RoomID = m.RoomID
					if requestActor(r) != ownerID {
						v.AgentExport = m.AgentExport
						v.ExportProviders = m.ExportProviders
					}
					found = true
					if v.Revision != m.Revision {
						http.Error(w, "Memory conflict: reload current revision", 409)
						return nil, false
					}
					v.Revision++
				}
			}
			if !found {
				http.Error(w, "Memory not found", 404)
				return nil, false
			}
		}
		p = v
	}
	return p, true
}
