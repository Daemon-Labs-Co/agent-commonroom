package main

// Rooms event validation and typed payload preparation; persistence stays in the handler.

import (
	"agentroom/internal/room"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *server) prepareRoomsEvent(w http.ResponseWriter, r *http.Request, st room.State, roomID string, req eventRequest) (any, bool) {
	var p any
	switch req.Kind {
	case "workflow.revised":
		var v room.Workflow
		if json.Unmarshal(req.Payload, &v) != nil {
			http.Error(w, "Invalid workflow", 400)
			return nil, false
		}
		v.RoomID = roomID
		if e := validateGraph(v); e != nil {
			http.Error(w, e.Error(), 400)
			return nil, false
		}
		if v.ID == "" {
			v.ID = id()
			v.Revision = 1
		} else {
			found := false
			for _, old := range st.Workflows {
				if old.ID == v.ID && old.RoomID == roomID {
					found = true
					if old.Revision != v.Revision {
						http.Error(w, "Workflow changed; reload before saving", 409)
						return nil, false
					}
					v.Revision++
				}
			}
			if !found {
				http.Error(w, "Workflow not found", 404)
				return nil, false
			}
		}
		p = v
	case "room.created", "room.updated":
		var v room.Room
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Name) == "" {
			http.Error(w, "Room name required", 400)
			return nil, false
		}
		if req.Kind == "room.created" {
			v.Revision = 1
			v.ContextExport = nil
			v.ID = id()
		} else if !hasRoom(st, v.ID) {
			http.Error(w, "Room not found", 404)
			return nil, false
		}
		if v.HumanPresent == nil {
			present := true
			for _, old := range st.Rooms {
				if old.ID == v.ID && old.HumanPresent != nil {
					present = *old.HumanPresent
				}
			}
			v.HumanPresent = &present
		}
		if v.HumanRank == nil {
			rank := 11
			for _, old := range st.Rooms {
				if old.ID == v.ID && old.HumanRank != nil {
					rank = *old.HumanRank
				}
			}
			v.HumanRank = &rank
		}
		if *v.HumanRank < 1 || *v.HumanRank > 11 {
			http.Error(w, "Human seniority must be 1–11", 400)
			return nil, false
		}
		v.HumanName = strings.TrimSpace(v.HumanName)
		if v.HumanName == "" {
			for _, old := range st.Rooms {
				if old.ID == v.ID {
					v.HumanName = old.HumanName
				}
			}
		}
		for _, old := range st.Rooms {
			if old.ID == v.ID {
				if v.Revision != old.Revision {
					http.Error(w, "Room changed; reload", 409)
					return nil, false
				}
				v.Revision++
				v.ContextExport = old.ContextExport
			}
		}
		v.Name = strings.TrimSpace(v.Name)
		p = v
	case "task.revised":
		var v room.Task
		if json.Unmarshal(req.Payload, &v) != nil || strings.TrimSpace(v.Title) == "" {
			http.Error(w, "Task title required", 400)
			return nil, false
		}
		v.RoomID = roomID
		v.DispatchID = ""
		v.DispatchStatus = ""
		v.DispatchDetail = ""
		switch v.Status {
		case "Assigned", "Working", "Submitted", "Reviewing", "Changes requested", "Accepted":
		default:
			http.Error(w, "Invalid task status", 400)
			return nil, false
		}
		for _, pid := range []string{v.Assignee, v.Reviewer} {
			if pid != "" && !hasMember(st, roomID, pid) {
				http.Error(w, "Choose a member of this room", 400)
				return nil, false
			}
		}
		if v.ID == "" {
			v.ID = id()
			v.Revision = 1
		} else {
			found := false
			for _, old := range st.Tasks {
				if old.ID == v.ID && old.RoomID == roomID {
					found = true
					if old.DispatchStatus == "pending" {
						http.Error(w, "Task dispatch pending; wait for its result or reconcile the worker", 409)
						return nil, false
					}
					v.DispatchID = old.DispatchID
					v.DispatchStatus = old.DispatchStatus
					v.DispatchDetail = old.DispatchDetail
					if old.Revision != v.Revision {
						http.Error(w, "Task changed; reload", 409)
						return nil, false
					}
					v.Revision++
				}
			}
			if !found {
				http.Error(w, "Task not found", 404)
				return nil, false
			}
		}
		p = v
	case "evidence.created":
		var v room.Evidence
		if json.Unmarshal(req.Payload, &v) != nil || v.Title == "" || v.Body == "" {
			http.Error(w, "Evidence title and content required", 400)
			return nil, false
		}
		found := false
		for _, t := range st.Tasks {
			if t.ID == v.TaskID && t.RoomID == roomID {
				found = true
			}
		}
		if !found {
			http.Error(w, "Task not found", 404)
			return nil, false
		}
		switch v.Kind {
		case "Changes", "Checks", "Artifacts", "Review":
		default:
			http.Error(w, "Invalid evidence category", 400)
			return nil, false
		}
		v.ID = id()
		v.RoomID = roomID
		v.Source = "Human-provided"
		v.Created = time.Now().UTC()
		p = v
	}
	return p, true
}
