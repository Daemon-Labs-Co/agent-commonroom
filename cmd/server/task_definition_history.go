package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"net/http"
	"strconv"
)

// Called only after replay validates the complete acknowledged task event.
// Status/assignment-only revisions do not create fictional contract versions.
func recordTaskDefinitionRevision(st *room.State, task room.Task) {
	if task.Definition == nil {
		return
	}
	if st.TaskDefinitionHistory == nil {
		st.TaskDefinitionHistory = map[string][]room.TaskDefinitionRevision{}
	}
	records := st.TaskDefinitionHistory[task.ID]
	digest := taskcontract.Digest(*task.Definition)
	if len(records) > 0 && records[len(records)-1].Digest == digest && records[len(records)-1].RoomID == task.RoomID {
		return
	}
	st.TaskDefinitionHistory[task.ID] = append(records, room.TaskDefinitionRevision{CompanyID: st.Company.ID, RoomID: task.RoomID, TaskID: task.ID, Definition: *task.Definition, Digest: digest, EventPosition: st.EventCount})
}

// Reading retained versions remains available with taskDefinitions off, just as
// execution history does. It grants no mutation, execution or acceptance power.
func (s *server) taskDefinitionHistory(w http.ResponseWriter, r *http.Request) {
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
	// Preserve the owner-access audit boundary for every private history read.
	if requestActor(r) == ownerID {
		for _, v := range st.Rooms {
			if v.ID == rid && v.PersonalActorID != "" && v.PersonalActorID != ownerID {
				if s.append(r.Context(), "room.owner_access", map[string]string{"actorId": ownerID, "roomId": rid, "reason": "owner viewed personal task definition history"}) != nil {
					http.Error(w, "Personal-room access audit unavailable", 503)
					return
				}
			}
		}
	}
	records := []room.TaskDefinitionRevision{}
	all := st.TaskDefinitionHistory[tid]
	for i := len(all) - 1; i >= 0; i-- {
		v := all[i]
		if v.CompanyID == st.Company.ID && v.RoomID == rid && v.TaskID == tid {
			records = append(records, v)
		}
	}
	start := 0
	if cursor := r.URL.Query().Get("before"); cursor != "" {
		found = false
		for i, v := range records {
			if strconv.Itoa(v.Definition.Revision) == cursor {
				start = i + 1
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Definition cursor unavailable", 400)
			return
		}
	}
	end := start + 10
	if end > len(records) {
		end = len(records)
	}
	next := ""
	if end < len(records) {
		next = strconv.Itoa(records[end-1].Definition.Revision)
	}
	reply(w, map[string]any{"definitions": records[start:end], "nextCursor": next})
}
