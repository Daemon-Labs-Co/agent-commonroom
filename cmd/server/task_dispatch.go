package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// taskDispatch commits assignment, message and intent before any managed turn.
// The optional capability can be disabled, but never falls back to client-owned
// split writes or uncontained execution. Caller identities are human-only.
func (s *server) taskDispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if requestActor(r) != ownerID || !s.validCSRF(r) {
		http.Error(w, "Owner permission and request token required", 403)
		return
	}
	var req struct {
		TaskID   string `json:"taskId"`
		Revision int    `json:"revision"`
		Mode     string `json:"mode"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if dec.Decode(&req) != nil || dec.Decode(new(any)) != io.EOF || req.TaskID == "" || (req.Mode != "work" && req.Mode != "review" && req.Mode != "reconcile") {
		http.Error(w, "Task, revision and work/review/reconcile mode required", 400)
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
	if !canAccess(st, requestActor(r), rid) {
		http.Error(w, "Room access denied", 403)
		return
	}
	if !hasRoom(st, rid) {
		http.Error(w, "Room not found", 404)
		return
	}
	if !taskDispatchEnabled(st) {
		http.Error(w, "Task dispatch is disabled", 403)
		return
	}
	ctx, replay := prepareOperation(w, r, st, req)
	if replay {
		return
	}
	var task room.Task
	for _, t := range st.Tasks {
		if t.ID == req.TaskID && t.RoomID == rid {
			task = t
		}
	}
	if task.ID == "" {
		http.Error(w, "Task not found", 404)
		return
	}
	if task.Revision != req.Revision {
		http.Error(w, "Task changed or dispatch pending; reload", 409)
		return
	}
	if err = s.workflowSafety(ctx, st, rid); err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	if req.Mode == "reconcile" {
		if task.DispatchStatus != "pending" {
			http.Error(w, "No pending task dispatch", 409)
			return
		}
		for _, active := range s.active {
			if active {
				http.Error(w, "Wait for active agents to finish", 409)
				return
			}
		}
		probe := s.recoveryIdle
		if probe == nil {
			probe = remoteWorkerIdle
		}
		probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		idle, probeErr := probe(probeCtx, "ssh://"+os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"))
		if probeErr != nil || !idle {
			http.Error(w, "Worker shutdown could not be verified; task remains pending", 503)
			return
		}
		effects := taskOutcome(st, task.ID, task.DispatchID, "Stopped")
		response := map[string]any{"reconciled": true, "taskId": task.ID}
		if err = s.appendBatch(withOperationResponse(ctx, 200, response), effects); err != nil {
			operationError(w, err)
			return
		}
		reply(w, response)
		return
	}
	if task.DispatchStatus == "pending" {
		http.Error(w, "Task dispatch pending; wait or reconcile the worker", 409)
		return
	}
	if task.Status == "Accepted" {
		http.Error(w, "Accepted tasks cannot be dispatched", 409)
		return
	}
	pid := task.Assignee
	status := "Working"
	instruction := "Work on this task within your authorised capabilities. Report blockers and results."
	if req.Mode == "review" {
		pid = task.Reviewer
		status = "Reviewing"
		instruction = "Review this task and its evidence. Report findings; do not accept it automatically."
	}
	p, endpoint, err := workflowAgent(st, rid, pid)
	if err != nil || s.active[pid] {
		http.Error(w, "Assigned agent unavailable or already working", 409)
		return
	}
	if !strings.HasPrefix(endpoint, "ssh://") {
		http.Error(w, "Managed task dispatch requires an approved isolated worker", 503)
		return
	}
	if err = s.agentAdmission(st, rid, 1); err != nil {
		http.Error(w, err.Error(), 429)
		return
	}
	ready := isolationAvailable
	if s.taskReady != nil {
		ready = s.taskReady
	}
	if err = ready(endpoint); err != nil {
		http.Error(w, "Isolated worker unavailable; check connection diagnostics", 503)
		return
	}
	text := instruction + " Task: " + task.Title + "\nAcceptance criteria: " + task.Criteria + "\nEvidence (human-provided):\n"
	if len(text) > maxChatContextBytes {
		http.Error(w, "Task context exceeds 128 KiB", 413)
		return
	}
	for _, e := range st.Evidence {
		if e.TaskID == task.ID && e.RoomID == rid {
			part := e.Kind + ": " + e.Title + "\n" + e.Body + "\n"
			if len(part) > maxChatContextBytes-len(text) {
				http.Error(w, "Task evidence exceeds 128 KiB", 413)
				return
			}
			text += part
		}
	}
	prompt, err := chatPrompt(scoped(st, rid), p, text)
	if err != nil {
		http.Error(w, err.Error(), 413)
		return
	}
	actor, _ := member(st, requestActor(r))
	dispatchID := id()
	task.Revision++
	task.Status = status
	task.DispatchID = dispatchID
	task.DispatchStatus = "pending"
	task.DispatchDetail = "Awaiting managed worker result"
	p.Status = "Working"
	message := room.Message{ID: id(), RoomID: rid, ActorID: requestActor(r), ActorType: "human", Author: actor.Name, Text: "@" + p.Name + " " + text, Created: time.Now().UTC()}
	response := map[string]any{"queued": 1, "taskId": task.ID, "revision": task.Revision, "dispatchId": dispatchID}
	changes := []auditChange{{Kind: "task.revised", Payload: task}, {Kind: "message.created", Payload: message}, {Kind: "agent.activity", Payload: map[string]any{"participantId": p.ID, "activity": "Task: " + task.Title, "startedAt": time.Now().UTC()}}, {Kind: "participant.session", Payload: p}}
	if err = s.appendBatch(withOperationResponse(ctx, 202, response), changes); err != nil {
		operationError(w, err)
		return
	}
	s.active[p.ID] = true
	turnCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	s.cancels[p.ID] = cancel
	go func() {
		defer cancel()
		turn := s.executeAgentTurn
		if s.taskTurn != nil {
			turn = s.taskTurn
		}
		output, turnErr := turn(turnCtx, p, prompt, endpoint)
		if turnErr == nil && strings.TrimSpace(output) == "" {
			turnErr = errors.New("Agent returned no task result")
		}
		s.finishAgentTurnWithTask(turnCtx, p, output, endpoint, turnErr, task.ID, dispatchID)
	}()
	w.WriteHeader(202)
	reply(w, response)
}

func taskDispatchEnabled(st room.State) bool {
	enabled, ok := st.FeatureFlags["taskDispatch"]
	return !ok || enabled
}

// taskOutcome never auto-accepts work or overwrites a different dispatch.
// Pending intents survive restart and require explicit owner reconciliation;
// they are not automatically replayed following uncertain external effects.
func taskOutcome(st room.State, taskID, dispatchID, status string) []auditChange {
	if taskID == "" {
		return nil
	}
	for _, task := range st.Tasks {
		if task.ID != taskID || task.DispatchID != dispatchID || task.DispatchStatus != "pending" {
			continue
		}
		task.Revision++
		if status == "Idle" {
			task.DispatchStatus = "completed"
			task.DispatchDetail = "Agent replied; inspect conversation and evidence before accepting work"
		} else {
			task.DispatchStatus = "failed"
			task.DispatchDetail = "Agent stopped or failed; inspect agent diagnostics before retrying"
		}
		return []auditChange{{Kind: "task.revised", Payload: task}}
	}
	return nil
}
