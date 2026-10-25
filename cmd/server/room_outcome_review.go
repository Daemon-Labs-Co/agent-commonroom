package main

import (
	"agentroom/internal/room"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// Human acceptance is a review record, never a grant to dispatch, dispose of
// data or skip worker cleanup. Its snapshot becomes stale when room work changes.
func roomReviewSnapshot(st room.State, rid string) (string, bool) {
	var definition room.Room
	found := false
	for _, r := range st.Rooms {
		if r.ID == rid {
			definition = r
			found = true
			break
		}
	}
	ready := found && definition.PersonalActorID == "" && strings.TrimSpace(definition.Outcome) != "" && strings.TrimSpace(definition.AcceptanceCriteria) != ""
	tasks := []room.Task{}
	evidence := []room.Evidence{}
	for _, t := range st.Tasks {
		if t.RoomID == rid {
			tasks = append(tasks, t)
			if t.Status != "Accepted" || t.DispatchStatus == "pending" || t.Definition != nil && (!taskHasValidCurrentOutput(st, t) || !taskHasCurrentAcceptance(st, t)) {
				ready = false
			}
		}
	}
	if len(tasks) == 0 {
		ready = false
	}
	for _, e := range st.Evidence {
		if e.RoomID == rid {
			evidence = append(evidence, e)
		}
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	sort.Slice(evidence, func(i, j int) bool { return evidence[i].ID < evidence[j].ID })
	h := sha256.New()
	enc := json.NewEncoder(h)
	_ = enc.Encode([]any{st.Company.ID, rid, definition.Revision, definition.Outcome, definition.AcceptanceCriteria})
	_ = enc.Encode(tasks)
	_ = enc.Encode(evidence)
	return hex.EncodeToString(h.Sum(nil)), ready
}
func roomReviewStatus(current room.RoomOutcomeReview, snapshot string) string {
	if current.Status == "Accepted" {
		if current.Snapshot == snapshot {
			return "Accepted"
		}
		return "Needs review"
	}
	return "Open"
}
func validRoomReviewTransition(st room.State, v room.RoomOutcomeReview) bool {
	old := st.OutcomeReviews[v.RoomID]
	snapshot, ready := roomReviewSnapshot(st, v.RoomID)
	shared := false
	for _, r := range st.Rooms {
		if r.ID == v.RoomID && r.PersonalActorID == "" {
			shared = true
		}
	}
	if !shared || !st.FeatureFlags["roomOutcomeReview"] || !st.FeatureFlags["roomOutcomes"] || v.CompanyID != st.Company.ID || v.Revision < 1 || old.Revision != v.Revision-1 || v.ReviewedBy != ownerID || v.ReviewedAt.IsZero() || strings.TrimSpace(v.Note) == "" || utf8.RuneCountInString(v.Note) > 4000 || v.Snapshot != snapshot {
		return false
	}
	if v.Status == "Accepted" {
		return ready && roomReviewStatus(old, snapshot) != "Accepted"
	}
	return v.Status == "Reopened" && old.Status == "Accepted"
}
func projectRoomOutcomeReview(st *room.State, data []byte) error {
	var v room.RoomOutcomeReview
	var actor struct {
		PerformedBy string `json:"performedBy"`
	}
	if json.Unmarshal(data, &v) != nil || json.Unmarshal(data, &actor) != nil || actor.PerformedBy != ownerID || !validRoomReviewTransition(*st, v) {
		return errors.New("invalid room outcome review")
	}
	if st.OutcomeReviews == nil {
		st.OutcomeReviews = map[string]room.RoomOutcomeReview{}
	}
	st.OutcomeReviews[v.RoomID] = v
	return nil
}
func (s *server) roomOutcomeReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Method == "POST" && (requestActor(r) != ownerID || !s.validCSRF(r)) {
		http.Error(w, "Owner permission and request token required", 403)
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
	// This endpoint is for shared deliverables only. Do not create an alternate
	// unaudited read path into another person's personal-room history.
	for _, item := range st.Rooms {
		if item.ID == rid && item.PersonalActorID != "" {
			http.Error(w, "Shared room required", 404)
			return
		}
	}
	snapshot, ready := roomReviewSnapshot(st, rid)
	current := st.OutcomeReviews[rid]
	enabled := st.FeatureFlags["roomOutcomeReview"] && st.FeatureFlags["roomOutcomes"]
	if r.Method == "GET" {
		var definition room.Room
		tasks := []room.Task{}
		evidence := []room.Evidence{}
		for _, item := range st.Rooms {
			if item.ID == rid {
				definition = item
			}
		}
		for _, item := range st.Tasks {
			if item.RoomID == rid {
				tasks = append(tasks, item)
			}
		}
		for _, item := range st.Evidence {
			if item.RoomID == rid {
				evidence = append(evidence, item)
			}
		}
		status := roomReviewStatus(current, snapshot)
		if status == "Accepted" && !ready {
			status = "Needs review"
		}
		reply(w, map[string]any{"enabled": enabled, "snapshot": snapshot, "ready": ready, "status": status, "review": current, "definition": definition, "tasks": tasks, "evidence": evidence})
		return
	}
	var req struct {
		Action   string `json:"action"`
		Revision int    `json:"revision"`
		Snapshot string `json:"snapshot"`
		Note     string `json:"note"`
	}
	raw, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 32768))
	fields, ferr := fenceObjectFields(raw, "action", "revision", "snapshot", "note")
	if e != nil || ferr != nil || len(fields) != 4 || json.Unmarshal(raw, &req) != nil || (req.Action != "accept" && req.Action != "reopen") || req.Revision < 0 || len(req.Snapshot) != 64 || strings.TrimSpace(req.Note) == "" || utf8.RuneCountInString(req.Note) > 4000 || r.Header.Get("Idempotency-Key") == "" {
		http.Error(w, "Action, revision, snapshot, review note and operation key required", 400)
		return
	}
	ctx, replay := prepareOperation(w, r, st, req)
	if replay {
		return
	}
	if !enabled {
		http.Error(w, "Room outcome review is disabled", 403)
		return
	}
	if req.Revision != current.Revision || current.Revision == int(^uint(0)>>1) || req.Snapshot != snapshot {
		http.Error(w, "Room work changed; reload before reviewing", 409)
		return
	}
	next := room.RoomOutcomeReview{CompanyID: st.Company.ID, RoomID: rid, Revision: current.Revision + 1, Snapshot: snapshot, Note: strings.TrimSpace(req.Note), ReviewedBy: ownerID, ReviewedAt: time.Now().UTC(), Status: map[string]string{"accept": "Accepted", "reopen": "Reopened"}[req.Action]}
	if !validRoomReviewTransition(st, next) {
		http.Error(w, "Define the room and accept all tasks before acceptance; reopen requires prior acceptance", 409)
		return
	}
	if err = s.workflowSafety(ctx, st, rid); err != nil {
		http.Error(w, "Review blocked by safety or audit state", 409)
		return
	}
	response := map[string]any{"committed": true, "review": next}
	if err = s.append(withOperationResponse(ctx, 200, response), "room.outcome.reviewed", next); err != nil {
		if !errors.Is(err, errAuditConflict) {
			s.tripIntegrityLocked()
		}
		operationError(w, err)
		return
	}
	reply(w, response)
}
