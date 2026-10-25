package main

import (
	"agentroom/internal/room"
	"net/http"
	"sort"
)

func (s *server) taskProgress(w http.ResponseWriter, r *http.Request) {
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
	rid := requestedRoom(r)
	if !canAccess(st, requestActor(r), rid) || !sharedTaskValueRoom(st, rid) {
		http.Error(w, "Shared room unavailable", 404)
		return
	}
	if !st.FeatureFlags["taskProgress"] {
		http.Error(w, "Task progress is disabled", 403)
		return
	}
	list := []room.Task{}
	for _, t := range st.Tasks {
		if t.RoomID == rid {
			list = append(list, t)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	start := 0
	if cursor := r.URL.Query().Get("after"); cursor != "" {
		found := false
		for i, t := range list {
			if t.ID == cursor {
				start = i + 1
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Task cursor unavailable; reload progress", 409)
			return
		}
	}
	end := start + 25
	if end > len(list) {
		end = len(list)
	}
	next := ""
	if end < len(list) {
		next = list[end-1].ID
	}
	views := []taskProgressView{}
	counts := map[string]int{}
	for _, t := range list[start:end] {
		v := describeTaskProgress(st, t)
		views = append(views, v)
		counts[v.Phase]++
	}
	reply(w, map[string]any{"tasks": views, "pageCounts": counts, "totalTasks": len(list), "nextCursor": next, "structuredDispatchEnabled": st.FeatureFlags["taskDefinitions"] && taskDispatchEnabled(st), "criterionReviewEnabled": st.FeatureFlags["taskDefinitions"] && st.FeatureFlags["taskAcceptanceReview"], "observations": "controller-recorded-state-only"})
}
