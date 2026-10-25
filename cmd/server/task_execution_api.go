package main

import (
	"agentroom/internal/room"
	"net/http"
	"sort"
)

// Immutable execution records are room-scoped history, not agent capabilities.
func (s *server) taskExecutions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	rid, tid := requestedRoom(r), r.URL.Query().Get("taskId")
	if !hasRoom(st, rid) || !canAccess(st, requestActor(r), rid) {
		http.Error(w, "Room unavailable", 404)
		return
	}
	found := false
	for _, t := range st.Tasks {
		if t.ID == tid && t.RoomID == rid {
			found = true
		}
	}
	if !found {
		http.Error(w, "Task unavailable", 404)
		return
	}
	if requestActor(r) == ownerID {
		for _, room := range st.Rooms {
			if room.ID == rid && room.PersonalActorID != "" && room.PersonalActorID != ownerID {
				if err = s.append(r.Context(), "room.owner_access", map[string]string{"actorId": ownerID, "roomId": rid, "reason": "owner viewed personal task execution history"}); err != nil {
					http.Error(w, "Personal-room access audit unavailable", 503)
					return
				}
			}
		}
	}
	list := []room.TaskExecution{}
	for _, execution := range st.TaskExecutions {
		if execution.RoomID == rid && execution.TaskID == tid && execution.CompanyID == st.Company.ID {
			list = append(list, execution)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].StartedAt.Equal(list[j].StartedAt) {
			return list[i].ID > list[j].ID
		}
		return list[i].StartedAt.After(list[j].StartedAt)
	})
	start := 0
	if cursor := r.URL.Query().Get("before"); cursor != "" {
		found = false
		for i, item := range list {
			if item.ID == cursor {
				start = i + 1
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Execution cursor unavailable", 400)
			return
		}
	}
	end := start + 10
	if end > len(list) {
		end = len(list)
	}
	next := ""
	if end < len(list) {
		next = list[end-1].ID
	}
	type executionView struct {
		room.TaskExecution
		InputsCurrent bool `json:"inputsCurrent"`
	}
	views := []executionView{}
	for _, v := range list[start:end] {
		views = append(views, executionView{TaskExecution: v, InputsCurrent: taskExecutionInputsCurrent(st, v)})
	}
	reply(w, map[string]any{"executions": views, "nextCursor": next})
}
