package main

// Agents event validation and typed payload preparation; persistence stays in the handler.

import (
	"agentroom/internal/room"
	"encoding/json"
	"net/http"
	"strings"
)

func (s *server) prepareAgentsEvent(w http.ResponseWriter, r *http.Request, st room.State, roomID string, req eventRequest) (any, bool) {
	var p any
	switch req.Kind {
	case "connection.deleted":
		var v struct {
			ID    string `json:"id"`
			Actor string `json:"actor"`
		}
		if json.Unmarshal(req.Payload, &v) != nil || v.ID == "" {
			http.Error(w, "Connection ID required", 400)
			return nil, false
		}
		found := false
		for _, c := range st.Connections {
			if c.ID == v.ID {
				found = true
			}
		}
		if !found {
			http.Error(w, "Connection not found", 404)
			return nil, false
		}
		for _, member := range st.Participants {
			if member.ConnectionID == v.ID && s.active[member.ID] {
				http.Error(w, "Wait for connected agents to finish before deleting the connection", 409)
				return nil, false
			}
		}
		v.Actor = "human"
		p = v
	case "connection.created", "connection.updated":
		var v room.Connection
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" {
			http.Error(w, "Name required", 400)
			return nil, false
		}
		if v.Location != "" && v.Location != "Local" && v.Location != "LAN" && v.Location != "Remote/Internet" {
			http.Error(w, "Invalid connection location", 400)
			return nil, false
		}
		switch v.Protocol {
		case "CLI", "HTTP", "WebSocket", "MCP":
		default:
			http.Error(w, "Unsupported protocol", 400)
			return nil, false
		}

		v.Name = strings.TrimSpace(v.Name)
		v.Endpoint = strings.TrimSpace(v.Endpoint)
		if v.Endpoint == "" {
			http.Error(w, "Endpoint required", 400)
			return nil, false
		}
		if req.Kind == "connection.created" {
			v.ID = id()
			v.Revision = 1
		} else {
			found := false
			for _, old := range st.Connections {
				if old.ID == v.ID {
					if v.Revision != old.Revision {
						http.Error(w, "Connection changed; reload", 409)
						return nil, false
					}
					v.Revision++
					found = true
					if old.Endpoint != v.Endpoint || old.Protocol != v.Protocol {
						for _, member := range st.Participants {
							if member.ConnectionID == v.ID && s.active[member.ID] {
								http.Error(w, "Wait for connected agents to finish before changing the endpoint or protocol", 409)
								return nil, false
							}
						}
					}
				}
			}
			if !found {
				http.Error(w, "Connection not found", 404)
				return nil, false
			}
		}
		p = v
	case "participant.created", "participant.updated":
		if st.Incident || s.agentsBlockedLocked(st, roomID) {
			http.Error(w, "Invitations/profile activation blocked during an incident or emergency stop", 409)
			return nil, false
		}
		var v room.Participant
		if json.Unmarshal(req.Payload, &v) != nil || v.Name == "" {
			http.Error(w, "Name required", 400)
			return nil, false
		}
		v.RoomID = roomID
		if v.Rank < 1 || v.Rank > 10 {
			http.Error(w, "Agent seniority must be 1–10", 400)
			return nil, false
		}
		if v.Challenge == nil {
			n := 1
			v.Challenge = &n
		}
		if *v.Challenge < 0 || *v.Challenge > 4 {
			http.Error(w, "Challenge must be 0–4", 400)
			return nil, false
		}
		if len(v.AdditionalContext) > 40000 {
			http.Error(w, "Additional context exceeds 40,000 bytes; select an excerpt", 400)
			return nil, false
		}
		found := false
		for _, c := range st.Connections {
			if c.ID == v.ConnectionID {
				found = true
			}
		}
		if !found {
			http.Error(w, "Register a connection first", 400)
			return nil, false
		}
		if req.Kind == "participant.created" {
			v.DisclosureVersion = 0
			v.Revision = 1
			v.ThreadID = ""
			v.ExecutionMode = ""
			v.Status = ""
			for _, c := range st.Connections {
				if c.ID == v.ConnectionID && room.SupportsAgent(c) {
					v.Status = "Idle"
				}
			}
			v.ID = id()
		} else {
			found = false
			for _, b := range st.Participants {
				if b.ID == v.ID && b.RoomID == roomID {
					v.DisclosureVersion = b.DisclosureVersion
					if v.Revision != b.Revision {
						http.Error(w, "Participant changed; reload", 409)
						return nil, false
					}
					v.Revision++
					if v.ConnectionID != b.ConnectionID {
						if s.active[b.ID] {
							http.Error(w, "Cannot reassign a working agent", 409)
							return nil, false
						}
						v.ThreadID = ""
						v.ExecutionMode = ""
						v.Status = "Idle"
					} else {
						v.ThreadID = b.ThreadID
						v.ExecutionMode = b.ExecutionMode
						v.Status = b.Status
					}
					found = true
				}
			}
			if !found {
				http.Error(w, "Participant not found", 404)
				return nil, false
			}
		}
		p = v
	}
	return p, true
}
