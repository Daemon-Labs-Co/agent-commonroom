package main

import (
	"agentroom/internal/room"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// workflowExecutionEnabled defaults off. Owner-controlled orchestration never
// grants repository permissions or changes mandatory container isolation.
func workflowExecutionEnabled(st room.State) bool { return st.FeatureFlags["workflowExecution"] }

// validateRunnable rejects unsupported semantics instead of treating saved
// graphs as executable guarantees. Every finish path requires human acceptance.
func validateRunnable(w room.Workflow, st room.State) error {
	if err := validateGraph(w); err != nil {
		return err
	}
	starts, finishes := 0, 0
	developers := map[string]bool{}
	for _, n := range w.Nodes {
		if n.Type == "agent" {
			developers[n.AgentID] = true
		}
	}
	for _, n := range w.Nodes {
		if n.Type == "start" {
			starts++
		}
		if n.Type == "finish" {
			finishes++
		}
		if n.Type == "parallel" {
			return errors.New("Parallel nodes are not supported by the sequential runner")
		}
		if n.Retries != 0 {
			return errors.New("Automatic retries are not supported; explicitly resume failed work")
		}
		if n.Timeout > 180 {
			return errors.New("Execution timeouts and waits must be at most 180 seconds")
		}
		if n.Type == "agent" || n.Type == "review" {
			if !hasMember(st, w.RoomID, n.AgentID) {
				return fmt.Errorf("%s: choose a current room agent", n.Title)
			}
			if n.Type == "review" && developers[n.AgentID] {
				return errors.New("Review must use an agent independent of implementation")
			}
		}
		for _, port := range runnablePorts(n.Type) {
			if nextWorkflowNode(w, n.ID, port) == "" {
				return fmt.Errorf("%s: connect %s", n.Title, port)
			}
		}
	}
	if starts != 1 || finishes == 0 {
		return errors.New("Execution requires exactly one Start and at least one Finish")
	}
	type path struct {
		node     string
		accepted bool
	}
	queue := []path{{startWorkflowNode(w), false}}
	seen := map[path]bool{}
	reachable := map[string]bool{}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if seen[p] {
			continue
		}
		seen[p] = true
		reachable[p.node] = true
		n := workflowNode(w, p.node)
		if n.Type == "approval" {
			p.accepted = true
		}
		// A new implementation invalidates acceptance of previous outputs.
		if n.Type == "agent" || n.Type == "review" {
			p.accepted = false
		}
		if n.Type == "finish" && !p.accepted {
			return errors.New("Every Finish path must pass human approval after the last implementation")
		}
		for _, e := range w.Edges {
			if e.From == n.ID {
				queue = append(queue, path{e.To, p.accepted})
			}
		}
	}
	for _, n := range w.Nodes {
		if !reachable[n.ID] {
			return fmt.Errorf("%s: unreachable", n.Title)
		}
	}
	return nil
}

func runnablePorts(kind string) []string {
	switch kind {
	case "finish":
		return nil
	case "review":
		return []string{"passed", "changes"}
	case "decision":
		return []string{"yes", "no"}
	default:
		return []string{"next"}
	}
}
func workflowNode(w room.Workflow, id string) room.WorkflowNode {
	for _, n := range w.Nodes {
		if n.ID == id {
			return n
		}
	}
	return room.WorkflowNode{}
}
func startWorkflowNode(w room.Workflow) string {
	for _, n := range w.Nodes {
		if n.Type == "start" {
			return n.ID
		}
	}
	return ""
}
func nextWorkflowNode(w room.Workflow, id, port string) string {
	for _, e := range w.Edges {
		if e.From == id && e.Port == port {
			return e.To
		}
	}
	return ""
}
func workflowTerminal(status string) bool { return status == "completed" || status == "cancelled" }

// saveWorkflowRun creates an immutable revision; graph revisions remain pinned.
func (s *server) saveWorkflowRun(run *room.WorkflowRun) error {
	run.Revision++
	run.Updated = time.Now().UTC()
	return s.append(context.Background(), "workflow.run.updated", run)
}

// workflowRuns handles owner-authorized starts, gates, cancellation and explicit
// recovery. No provider call is made in a request or without safety checks.
func (s *server) workflowRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Method == "POST" && !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
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
	if !hasRoom(st, rid) {
		http.Error(w, "Room not found", 404)
		return
	}
	// A lost controller never silently replays a potentially mutating agent turn.
	for i := range st.WorkflowRuns {
		run := &st.WorkflowRuns[i]
		if run.Status == "running" && s.workflowLive[run.ID] == nil {
			run.Status = "interrupted"
			run.Detail = "Controller interrupted. Inspect workspace before explicit resume; the current node may have made changes."
			if err = s.saveWorkflowRun(run); err != nil {
				http.Error(w, "Recovery audit unavailable", 503)
				return
			}
		}
	}
	if r.Method == "GET" {
		reply(w, scoped(st, rid).WorkflowRuns)
		return
	}
	var req struct {
		Action     string `json:"action"`
		WorkflowID string `json:"workflowId"`
		RunID      string `json:"runId"`
		Revision   int    `json:"revision"`
		Input      string `json:"input"`
		Evidence   string `json:"evidence"`
		Port       string `json:"port"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024)).Decode(&req) != nil {
		http.Error(w, "Invalid request", 400)
		return
	}
	if len(req.Input) > 16000 || len(req.Evidence) > 16000 {
		http.Error(w, "Input or evidence exceeds 16000 bytes", 400)
		return
	}
	st, err = s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	opCtx, replayed := prepareOperation(w, r, st, req)
	if replayed {
		return
	}
	r = r.WithContext(opCtx)
	if req.Action == "start" {
		if !workflowExecutionEnabled(st) {
			http.Error(w, "Workflow execution is disabled", 403)
			return
		}
		var graph room.Workflow
		for _, v := range st.Workflows {
			if v.ID == req.WorkflowID && v.RoomID == rid {
				graph = v
			}
		}
		if graph.ID == "" {
			http.Error(w, "Saved workflow not found", 404)
			return
		}
		if graph.Revision != req.Revision {
			http.Error(w, "Workflow changed; reload", 409)
			return
		}
		if strings.TrimSpace(req.Input) == "" {
			http.Error(w, "Describe the approved task before starting", 400)
			return
		}
		if err = validateRunnable(graph, st); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		for _, run := range st.WorkflowRuns {
			if run.RoomID == rid && !workflowTerminal(run.Status) {
				http.Error(w, "Resolve the existing room run first", 409)
				return
			}
		}
		if err = s.workflowSafety(r.Context(), st, rid); err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
		if err = s.workflowAssignments(st, graph); err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
		run := room.WorkflowRun{ID: id(), RoomID: rid, Workflow: graph, Input: req.Input, ActorID: ownerID, NodeID: startWorkflowNode(graph), Status: "running", Visits: map[string]int{}, Steps: []room.WorkflowStep{}}
		if err = s.saveOperationRun(r.Context(), &run, 202); err != nil {
			operationError(w, err)
			return
		}
		s.launchWorkflow(run.ID)
		w.WriteHeader(202)
		reply(w, run)
		return
	}
	var run room.WorkflowRun
	for _, v := range st.WorkflowRuns {
		if v.ID == req.RunID && v.RoomID == rid {
			run = v
		}
	}
	if run.ID == "" {
		http.Error(w, "Run not found", 404)
		return
	}
	if req.Revision != run.Revision {
		http.Error(w, "Run changed; reload", 409)
		return
	}
	if workflowTerminal(run.Status) {
		http.Error(w, "Run has finished", 409)
		return
	}
	if req.Action == "cancel" {

		run.Status = "cancelled"
		run.Detail = "Cancelled by owner"
	} else {
		if !workflowExecutionEnabled(st) {
			http.Error(w, "Workflow execution is disabled", 403)
			return
		}
		if err = s.workflowSafety(r.Context(), st, rid); err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
		if req.Action != "resume" && s.workflowLive[run.ID] != nil {
			http.Error(w, "Current step is still finishing; reload", 409)
			return
		}
		n := workflowNode(run.Workflow, run.NodeID)
		switch req.Action {
		case "resume":
			if run.Visits[run.NodeID] >= run.Workflow.MaxIterations {
				http.Error(w, "Iteration budget exhausted; cancel and revise", 409)
				return
			}
			if run.Status != "failed" && run.Status != "interrupted" {
				http.Error(w, "Only failed or interrupted runs can resume", 409)
				return
			}
			if s.workflowLive[run.ID] != nil {
				http.Error(w, "Previous worker is still stopping", 409)
				return
			}
		case "reject":
			if run.Status != "awaiting_human" || n.Type != "approval" || strings.TrimSpace(req.Evidence) == "" {
				http.Error(w, "Only human approval can return work, with a reason", 409)
				return
			}
			destination := ""
			for i := len(run.Steps) - 1; i >= 0; i-- {
				if workflowNode(run.Workflow, run.Steps[i].NodeID).Type == "agent" {
					destination = run.Steps[i].NodeID
					break
				}
			}
			if destination == "" {
				http.Error(w, "No implementation step to return to; cancel and revise the workflow", 409)
				return
			}
			appendWorkflowStep(&run, n, "Changes requested by "+ownerID, req.Evidence, "changes")
			run.NodeID = destination
		case "advance":
			if run.Status != "awaiting_human" {
				http.Error(w, "Run is not awaiting a human", 409)
				return
			}
			port := req.Port
			if n.Type == "approval" || n.Type == "evidence" {
				if port != "next" || strings.TrimSpace(req.Evidence) == "" {
					http.Error(w, "Record evidence and approval rationale", 400)
					return
				}
			} else if n.Type != "decision" || (port != "yes" && port != "no") {
				http.Error(w, "Invalid gate decision", 400)
				return
			}
			if strings.TrimSpace(req.Evidence) == "" {
				http.Error(w, "Record a decision rationale", 400)
				return
			}
			appendWorkflowStep(&run, n, "Human decision by "+ownerID, req.Evidence, port)
			run.NodeID = nextWorkflowNode(run.Workflow, n.ID, port)
		default:
			http.Error(w, "Choose start, advance, reject, resume or cancel", 400)
			return
		}
		run.Status = "running"
		run.Detail = ""
	}
	if err = s.saveOperationRun(r.Context(), &run, 200); err != nil {
		operationError(w, err)
		return
	}
	if req.Action == "cancel" {
		if cancel := s.workflowLive[run.ID]; cancel != nil {
			cancel()
		}
	}
	if run.Status == "running" {
		s.launchWorkflow(run.ID)
	}
	reply(w, run)
}

// workflowSafety verifies the audit chain before every dispatch and human gate.
// Emergency-stop enforcement is shared with ordinary managed model turns.
func (s *server) workflowSafety(ctx context.Context, st room.State, rid string) error {
	if st.Incident || s.agentsBlockedLocked(st, rid) {
		return errors.New("Workflow blocked by integrity incident or emergency stop")
	}
	var body bytes.Buffer
	rec := &bufferResponse{header: http.Header{}, body: &body}
	s.integrityLocked(rec, (&http.Request{Method: "GET"}).WithContext(ctx))
	var result struct {
		Valid bool `json:"valid"`
	}
	if rec.status >= 400 || json.Unmarshal(body.Bytes(), &result) != nil || !result.Valid {
		return errors.New("Workflow blocked: audit integrity unavailable or failed")
	}
	return nil
}

func workflowAgent(st room.State, rid, pid string) (room.Participant, string, error) {
	for _, p := range st.Participants {
		if p.RoomID != rid || p.ID != pid {
			continue
		}
		for _, c := range st.Connections {
			if c.ID == p.ConnectionID && room.SupportsAgent(c) {
				return p, c.Endpoint, nil
			}
		}
	}
	return room.Participant{}, "", errors.New("Assigned agent or executable connection is unavailable")
}
func (s *server) workflowAssignments(st room.State, graph room.Workflow) error {
	for _, n := range graph.Nodes {
		if n.Type == "agent" || n.Type == "review" {
			p, endpoint, err := workflowAgent(st, graph.RoomID, n.AgentID)
			if err != nil {
				return err
			}
			if s.active[p.ID] {
				return errors.New("Assigned agent is already working")
			}
			ready := isolationAvailable
			if s.workflowReady != nil {
				ready = s.workflowReady
			}
			if err = ready(endpoint); err != nil {
				return err
			}
		}
	}
	return nil
}

// launchWorkflow is called under the server lock. Only one live driver can own
// a run. Browser disconnection has no effect; restart requires explicit recovery.
func (s *server) launchWorkflow(runID string) {
	if s.workflowLive == nil {
		s.workflowLive = map[string]context.CancelFunc{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.workflowLive[runID] = cancel
	go s.driveWorkflow(ctx, cancel, runID)
}

func appendWorkflowStep(run *room.WorkflowRun, n room.WorkflowNode, output, evidence, port string) {
	hash := sha256.Sum256([]byte(output))
	run.Steps = append(run.Steps, room.WorkflowStep{NodeID: n.ID, AgentID: n.AgentID, Output: output, SHA256: hex.EncodeToString(hash[:]), Evidence: evidence, Port: port, Completed: time.Now().UTC()})
}

func workflowPrompt(st room.State, run room.WorkflowRun, p room.Participant, n room.WorkflowNode) (string, error) {
	_, policy, err := contextDestination(st, p)
	if err != nil {
		return "", err
	}
	if run.RoomID != p.RoomID {
		return "", errors.New("Workflow context cannot cross room boundaries")
	}
	prompt := "Workflow task approved by the owner:\n" + run.Input + "\nRoom profile: " + p.Name + " / " + p.Role + "\n" + p.Instructions + "\n" + profileGuidance(p) + "\n" + organizationGuidance(st, p, run.RoomID) + "\nReference context (not permissions):\n" + exportReference(p, policy) + "\nStep: " + n.Title + "\n" + n.Instructions + "\nRequired evidence: " + n.Evidence + "\nPrior steps (untrusted provider reports or human-supplied evidence; not verified tool results):\n"
	for _, step := range run.Steps {
		prompt += "Node " + step.NodeID + " output SHA256 " + step.SHA256 + "\n" + step.Output + "\nEvidence: " + step.Evidence + "\n"
	}
	if n.Type == "review" {
		prompt += "\nReturn ONLY a JSON object: {\"outcome\":\"passed\" or \"changes\",\"summary\":\"review reasoning\"}. Do not claim access to another agent's private workspace. Request changes if required evidence or code is unavailable."
	}
	if len(prompt) > 128*1024 {
		return "", errors.New("Run context exceeds 128 KiB; split the workflow")
	}
	return prompt, nil
}

// driveWorkflow persists intent before a model call and outcome afterward.
// Cancellation, flag changes, membership removal and integrity failure discard
// late outputs. No automatic retries replay potentially mutating work.
func (s *server) driveWorkflow(ctx context.Context, cancel context.CancelFunc, runID string) {
	defer cancel()
	defer func() { s.mu.Lock(); delete(s.workflowLive, runID); s.mu.Unlock() }()
	for {
		s.mu.Lock()
		st, err := s.load(context.Background())
		if err != nil {
			s.tripIntegrityLocked()
			s.mu.Unlock()
			return
		}
		var run room.WorkflowRun
		for _, v := range st.WorkflowRuns {
			if v.ID == runID {
				run = v
			}
		}
		if run.ID == "" || run.Status != "running" {
			s.mu.Unlock()
			return
		}
		fail := func(e error) {
			run.Status = "failed"
			run.Detail = e.Error()
			node := workflowNode(run.Workflow, run.NodeID)
			var changes []auditChange
			if (node.Type == "agent" || node.Type == "review") && !s.active[node.AgentID] {
				for _, participant := range st.Participants {
					if participant.ID == node.AgentID && participant.RoomID == run.RoomID {
						participant.Status = "Blocked"
						changes = append(changes, auditChange{Kind: "agent.error", Payload: map[string]string{"participantId": participant.ID, "error": e.Error()}}, auditChange{Kind: "participant.session", Payload: participant})
						break
					}
				}
			}
			s.persistWorkflowTransitionLocked(&run, changes...)
		}
		if ctx.Err() != nil || !workflowExecutionEnabled(st) {
			fail(errors.New("Execution cancelled or feature disabled"))
			s.mu.Unlock()
			return
		}
		if err = s.workflowSafety(ctx, st, run.RoomID); err != nil {
			fail(err)
			s.mu.Unlock()
			return
		}
		n := workflowNode(run.Workflow, run.NodeID)
		if n.ID == "" {
			fail(errors.New("Missing execution node"))
			s.mu.Unlock()
			return
		}
		if run.Visits == nil {
			run.Visits = map[string]int{}
		}
		run.Visits[n.ID]++
		if run.Visits[n.ID] > run.Workflow.MaxIterations || len(run.Steps) >= 500 {
			fail(errors.New("Workflow iteration limit reached"))
			s.mu.Unlock()
			return
		}
		if n.Type == "approval" || n.Type == "evidence" || n.Type == "decision" {
			run.Status = "awaiting_human"
			run.Detail = n.Title + ": " + n.Instructions + "\n" + n.Evidence
			s.persistWorkflowTransitionLocked(&run)
			s.mu.Unlock()
			return
		}
		if n.Type == "finish" {
			run.Status = "completed"
			run.Detail = "Completed with human acceptance"
			s.persistWorkflowTransitionLocked(&run)
			s.mu.Unlock()
			return
		}
		if n.Type == "start" {
			appendWorkflowStep(&run, n, "Run started by owner", "", "next")
			run.NodeID = nextWorkflowNode(run.Workflow, n.ID, "next")
			if s.persistWorkflowTransitionLocked(&run) != nil {
				s.mu.Unlock()
				return
			}
			s.mu.Unlock()
			continue
		}
		var p room.Participant
		var endpoint, prompt string
		if n.Type == "agent" || n.Type == "review" {
			p, endpoint, err = workflowAgent(st, run.RoomID, n.AgentID)
			if err == nil && s.active[p.ID] {
				err = errors.New("Assigned agent is already working")
			}
			if err == nil {
				err = s.agentAdmission(st, run.RoomID, 1)
			}
			ready := isolationAvailable
			if s.workflowReady != nil {
				ready = s.workflowReady
			}
			if err == nil {
				err = ready(endpoint)
			}
			if err == nil {
				prompt, err = workflowPrompt(st, run, p, n)
			}
			if err != nil {
				fail(err)
				s.mu.Unlock()
				return
			}
		} else if n.Type != "wait" {
			fail(errors.New("Unsupported execution node"))
			s.mu.Unlock()
			return
		}
		if p.ID == "" {
			if err = s.persistWorkflowTransitionLocked(&run); err != nil {
				s.mu.Unlock()
				return
			}
		}
		seconds := n.Timeout
		if seconds == 0 {
			seconds = 180
		}
		turnCtx, stop := context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
		if p.ID != "" {
			p.Status = "Working"
			run.Revision++
			run.Updated = time.Now().UTC()
			changes := []auditChange{{Kind: "workflow.run.updated", Payload: run}, {Kind: "agent.activity", Payload: map[string]any{"participantId": p.ID, "activity": "Workflow: " + run.Workflow.Name + " / " + n.Title, "startedAt": time.Now().UTC()}}, {Kind: "participant.session", Payload: p}}
			if err = s.appendBatch(context.Background(), changes); err != nil {
				stop()
				s.tripIntegrityLocked()
				s.mu.Unlock()
				return
			}
			s.active[p.ID] = true
			s.cancels[p.ID] = stop
		}
		s.mu.Unlock()
		var output string
		if n.Type == "wait" {
			timer := time.NewTimer(time.Duration(n.Timeout) * time.Second)
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-timer.C:
				output = "Wait completed"
			}
			timer.Stop()
		} else {
			turn := s.executeAgentTurn
			if s.workflowTurn != nil {
				turn = s.workflowTurn
			}
			output, err = turn(turnCtx, p, prompt, endpoint)
			if err == nil {
				err = turnCtx.Err()
			}
		}
		stop()
		s.mu.Lock()
		delete(s.active, p.ID)
		delete(s.cancels, p.ID)
		current, loadErr := s.load(context.Background())
		if loadErr != nil {
			s.tripIntegrityLocked()
			s.mu.Unlock()
			return
		}
		for _, v := range current.WorkflowRuns {
			if v.ID == runID {
				run = v
			}
		}
		st = current
		var sessionChanges []auditChange
		if p.ID != "" {
			for _, member := range current.Participants {
				if member.ID == p.ID {
					p = member
					p.Status = "Idle"
					if err != nil {
						p.Status = "Blocked"
					}
					sessionChanges = append(sessionChanges, auditChange{Kind: "participant.session", Payload: p})
					break
				}
			}
		}
		if run.Status != "running" {
			if err := s.appendBatch(context.Background(), sessionChanges); err != nil {
				s.tripIntegrityLocked()
			}
			s.mu.Unlock()
			return
		}
		if ctx.Err() != nil || !workflowExecutionEnabled(current) {
			err = errors.New("Execution cancelled or feature disabled; late output discarded")
		}
		if p.ID != "" && !hasMember(current, run.RoomID, p.ID) {
			err = errors.New("Assigned agent was removed; output discarded")
		}
		if err == nil {
			err = s.workflowSafety(context.Background(), current, run.RoomID)
		}
		port := "next"
		if len(output) > 32768 {
			err = errors.New("Step output exceeds 32 KiB; request a shorter summary")
		}
		if err == nil && n.Type == "review" {
			var verdict struct {
				Outcome string `json:"outcome"`
				Summary string `json:"summary"`
			}
			if json.Unmarshal([]byte(output), &verdict) != nil || (verdict.Outcome != "passed" && verdict.Outcome != "changes") || strings.TrimSpace(verdict.Summary) == "" {
				err = errors.New("Review must return explicit passed/changes JSON and reasoning")
			} else {
				port = verdict.Outcome
			}
		}
		if err == nil && strings.TrimSpace(output) == "" {
			err = errors.New("Agent returned no output")
		}
		if err != nil {
			fail(err)
			s.mu.Unlock()
			return
		}
		appendWorkflowStep(&run, n, output, "", port)
		run.NodeID = nextWorkflowNode(run.Workflow, n.ID, port)
		if s.persistWorkflowTransitionLocked(&run, sessionChanges...) != nil {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
	}
}

func (s *server) saveOperationRun(ctx context.Context, run *room.WorkflowRun, status int) error {
	run.Revision++
	run.Updated = time.Now().UTC()
	return s.append(withOperationResponse(ctx, status, run), "workflow.run.updated", run)
}
