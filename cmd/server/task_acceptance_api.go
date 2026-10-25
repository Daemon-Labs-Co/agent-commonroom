package main

import (
	"agentroom/internal/room"
	"agentroom/internal/taskcontract"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type taskAcceptanceRequest struct {
	TaskID         string                  `json:"taskId"`
	TaskRevision   int                     `json:"taskRevision"`
	ReviewRevision int                     `json:"reviewRevision"`
	Snapshot       string                  `json:"snapshot"`
	Action         string                  `json:"action"`
	Note           string                  `json:"note"`
	Decisions      []taskCriterionDecision `json:"decisions"`
}

func (s *server) taskAcceptance(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Method == "POST" && (requestActor(r) != ownerID || !s.validCSRF(r)) {
		http.Error(w, "Owner permission and request token required", 403)
		return
	}
	tid := r.URL.Query().Get("taskId")
	var req taskAcceptanceRequest
	if r.Method == "POST" {
		raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 65536))
		if err != nil || !taskcontract.ValidJSON(raw, 65536) {
			http.Error(w, "Invalid bounded task review JSON", 400)
			return
		}
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if dec.Decode(&req) != nil || dec.Decode(new(any)) != io.EOF || req.TaskID == "" || req.TaskRevision < 1 || req.ReviewRevision < 0 || len(req.Snapshot) != 64 || (req.Action != "accept" && req.Action != "changes") || strings.TrimSpace(req.Note) == "" || utf8.RuneCountInString(req.Note) > 2000 || req.Decisions == nil || len(req.Decisions) > 32 || r.Header.Get("Idempotency-Key") == "" || (tid != "" && tid != req.TaskID) {
			http.Error(w, "Task/review revisions, snapshot, action, rationale, criterion decisions and operation key required", 400)
			return
		}
		tid = req.TaskID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	rid := requestedRoom(r)
	// A separate unaudited path into personal execution history is forbidden.
	if !canAccess(st, requestActor(r), rid) || !sharedTaskValueRoom(st, rid) {
		http.Error(w, "Shared room unavailable", 404)
		return
	}
	var task room.Task
	for _, t := range st.Tasks {
		if t.ID == tid && t.RoomID == rid {
			task = t
			break
		}
	}
	if task.Definition == nil {
		http.Error(w, "Defined task unavailable", 404)
		return
	}
	snapshot, execution, ready := taskAcceptanceSnapshot(st, task)
	current := latestTaskAcceptance(st, tid)
	enabled := st.FeatureFlags["taskDefinitions"] && st.FeatureFlags["taskAcceptanceReview"]
	if r.Method == "GET" {
		list := st.TaskAcceptanceReviews[tid]
		end := len(list)
		if before := r.URL.Query().Get("before"); before != "" {
			revision, e := strconv.Atoi(before)
			if e != nil || revision < 1 || revision > len(list) {
				http.Error(w, "Review cursor unavailable", 400)
				return
			}
			end = revision - 1
		}
		start := end - 10
		if start < 0 {
			start = 0
		}
		history := []room.TaskAcceptanceReview{}
		for i := end - 1; i >= start; i-- {
			history = append(history, list[i])
		}
		next := ""
		if start > 0 {
			next = strconv.Itoa(list[start].Revision)
		}
		reviewedWork := map[string]room.TaskExecution{}
		for _, review := range history {
			if e, ok := st.TaskExecutions[review.ExecutionID]; ok && e.CompanyID == st.Company.ID && e.RoomID == rid && e.TaskID == tid {
				reviewedWork[e.ID] = e
			}
		}
		status := "Not accepted"
		if task.Status == "Accepted" {
			status = "Needs review"
			if taskHasCurrentAcceptance(st, task) && ready {
				status = "Accepted"
			}
		}
		reply(w, map[string]any{"enabled": enabled, "snapshot": snapshot, "ready": ready, "task": task, "workExecution": execution, "schemaAssessments": taskSchemaAssessments(*task.Definition, execution), "review": current, "reviews": history, "reviewedExecutions": reviewedWork, "nextCursor": next, "status": status})
		return
	}
	ctx, replay := prepareOperation(w, r, st, req)
	if replay {
		return
	}
	if !enabled {
		http.Error(w, "Task acceptance review is disabled", 403)
		return
	}
	if !ready || task.Revision != req.TaskRevision || req.ReviewRevision != current.Revision || current.Revision == int(^uint(0)>>1) || task.Revision == int(^uint(0)>>1) || req.Snapshot != snapshot {
		http.Error(w, "Task, outputs or review policy changed; reload before reviewing", 409)
		return
	}
	assessments, e := assessTaskCriteria(*task.Definition, execution, req.Decisions)
	if e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	next := room.TaskAcceptanceReview{CompanyID: st.Company.ID, RoomID: rid, TaskID: tid, Revision: current.Revision + 1, TaskRevision: task.Revision, ExecutionID: execution.ID, DispatchID: task.DispatchID, Snapshot: snapshot, Action: req.Action, Note: strings.TrimSpace(req.Note), Assessments: assessments, ReviewedBy: ownerID, ReviewedAt: time.Now().UTC()}
	if e = validateTaskAcceptanceReview(st, next); e != nil {
		http.Error(w, e.Error(), 409)
		return
	}
	if e = s.workflowSafety(ctx, st, rid); e != nil {
		http.Error(w, "Review blocked by safety or audit state", 409)
		return
	}
	task.Revision++
	task.Status = "Changes requested"
	if req.Action == "accept" {
		task.Status = "Accepted"
	}
	response := map[string]any{"committed": true, "review": next, "task": task}
	changes := []auditChange{{Kind: "task.acceptance.reviewed", Payload: next}, {Kind: "task.revised", Payload: task}}
	if err = s.appendBatch(withOperationResponse(ctx, 200, response), changes); err != nil {
		if !errors.Is(err, errAuditConflict) {
			s.tripIntegrityLocked()
		}
		operationError(w, err)
		return
	}
	reply(w, response)
}
