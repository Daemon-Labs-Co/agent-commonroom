package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"
)

func recoveryEnabled(flags map[string]bool) bool {
	enabled, exists := flags["agentRecovery"]
	return !exists || enabled
}

// agentRecovery never launches a model or replays a task. Quiescence is required
// even when an old worker job's identity was not retained. This deliberately
// blocks recovery while any other managed job runs on the shared worker.
func (s *server) agentRecovery(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
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
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	var req struct {
		Action        string `json:"action"`
		ParticipantID string `json:"participantId"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || (req.Action != "check" && req.Action != "resume" && req.Action != "fresh") {
		http.Error(w, "Choose check, resume or fresh", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(ctx)
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	if !recoveryEnabled(st.FeatureFlags) {
		http.Error(w, "Agent recovery is disabled", 403)
		return
	}
	rid := requestedRoom(r)
	p, endpoint, err := workflowAgent(st, rid, req.ParticipantID)
	if err != nil {
		http.Error(w, "Agent not found in this room or connection unavailable", 404)
		return
	}
	for _, active := range s.active {
		if active {
			http.Error(w, "A managed turn is still active. Monitor it or stop it before recovery", 409)
			return
		}
	}
	if err = s.workflowSafety(ctx, st, rid); err != nil {
		http.Error(w, "Resolve integrity or emergency stop before recovery", 409)
		return
	}
	probe := s.recoveryIdle
	if probe == nil {
		probe = remoteWorkerIdle
	}
	idle, err := probe(ctx, endpoint)
	if err != nil {
		http.Error(w, "Worker state could not be confirmed. Update/check the approved worker; recovery was not performed", 503)
		return
	}
	if !idle {
		reply(w, map[string]any{"idle": false, "recovered": false, "detail": "The shared worker has active or unresolved jobs. This does not prove this particular agent is running. Monitor or stop jobs before recovery."})
		return
	}
	if req.Action == "check" {
		reply(w, map[string]any{"idle": true, "recovered": false, "detail": "The worker confirmed no managed containers or nonterminal job records. You may resume the saved session or start a fresh one."})
		return
	}
	if req.Action == "resume" && p.ThreadID == "" {
		http.Error(w, "No saved provider session; choose Start fresh session", 409)
		return
	}
	if req.Action == "resume" {
		repository, e := controllerDevelopmentRepository(p.RoomID, os.Getenv("COMMONROOM_DEVELOPMENT_WORKSPACES"), os.Getenv("COMMONROOM_DEVELOPMENT_REPOSITORIES"))
		if e != nil {
			http.Error(w, "Workspace configuration unavailable", 409)
			return
		}
		_, thread, e := agentExecutionSession(p.ExecutionMode, p.ThreadID, repository)
		if e != nil || thread != p.ThreadID {
			http.Error(w, "Session does not match the current isolated workspace; choose Start fresh session", 409)
			return
		}
	}
	if req.Action == "fresh" {
		p.ThreadID = ""
		p.ExecutionMode = ""
	}
	p.Status = "Idle"
	p.Background = nil
	p.StatusDetails = nil
	if err = s.append(ctx, "agent.recovered", map[string]any{"participantId": p.ID, "roomId": rid, "action": req.Action, "actorId": ownerID, "participant": p}); err != nil {
		http.Error(w, "Recovery audit failed", 503)
		return
	}
	reply(w, map[string]any{"idle": true, "recovered": true, "detail": "Session is ready. No task was replayed. Compose a new request after inspecting partial work."})
}

func remoteWorkerIdle(ctx context.Context, endpoint string) (bool, error) {
	alias, _, err := remoteWorkerEndpoint(endpoint, os.Getenv("COMMONROOM_REMOTE_WORKER_ALIAS"), os.Getenv("COMMONROOM_REMOTE_WORKER"))
	if err != nil {
		return false, err
	}
	var output limitedAgentOutput
	cmd := workerSSHCommand(ctx, alias, "idle-v1", "")
	cmd.Stdout = &output
	if cmd.Run() != nil {
		return false, errors.New("worker idle probe failed")
	}
	var response struct {
		Version int   `json:"version"`
		Idle    *bool `json:"idle"`
	}
	if json.Unmarshal(output.Bytes(), &response) != nil || response.Version != 1 || response.Idle == nil {
		return false, errors.New("invalid worker idle response")
	}
	return *response.Idle, nil
}
