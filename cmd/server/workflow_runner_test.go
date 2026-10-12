package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func executionGraph() room.Workflow {
	w := room.Workflow{ID: "flow", RoomID: "engineering", Name: "Delivery", Revision: 1, MaxIterations: 3}
	for i, kind := range []string{"start", "agent", "review", "evidence", "approval", "finish"} {
		n := room.WorkflowNode{ID: kind, Type: kind, Title: kind, Timeout: 5, X: float64(60 + (i%3)*290), Y: float64(80 + (i/3)*230)}
		if kind == "agent" {
			n.AgentID = "dev"
		}
		if kind == "review" {
			n.AgentID = "qa"
		}
		w.Nodes = append(w.Nodes, n)
		if i > 0 {
			from := w.Nodes[i-1]
			port := "next"
			if from.Type == "review" {
				port = "passed"
			}
			w.Edges = append(w.Edges, room.WorkflowEdge{ID: kind, From: from.ID, To: n.ID, Port: port})
		}
	}
	w.Edges = append(w.Edges, room.WorkflowEdge{ID: "rework", From: "review", To: "agent", Port: "changes"})
	return w
}

func executionServer(t *testing.T) (*server, *mockDB, http.Handler) {
	s, db, _ := mockServer(t)
	db.seed("feature.flag.updated", map[string]any{"name": "workflowExecution", "enabled": true})
	for _, pid := range []string{"dev", "qa"} {
		db.seed("connection.created", room.Connection{ID: pid, Provider: "openai", Protocol: "CLI", Endpoint: "codex"})
		db.seed("participant.created", room.Participant{ID: pid, RoomID: "engineering", ConnectionID: pid, Name: pid})
	}
	db.seed("workflow.revised", executionGraph())
	s.workflowReady = func(string) error { return nil }
	s.workflowTurn = func(_ context.Context, p room.Participant, prompt, endpoint string) (string, error) {
		if p.ID == "qa" {
			return `{"outcome":"passed","summary":"Reviewed supplied evidence; implementation is acceptable"}`, nil
		}
		return "Implementation summary with manually inspectable patch references", nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/workflow-runs", s.workflowRuns)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/agent-safety", s.agentSafety)
	return s, db, s.auth.protect(s.actorGuard(mux))
}

func startExecution(t *testing.T, h http.Handler) room.WorkflowRun {
	t.Helper()
	res := mockRequest(h, "POST", "/api/workflow-runs", `{"action":"start","workflowId":"flow","revision":1,"input":"Approved feature"}`)
	if res.Code != 202 {
		t.Fatalf("start: %d %s", res.Code, res.Body.String())
	}
	var run room.WorkflowRun
	if err := json.Unmarshal(res.Body.Bytes(), &run); err != nil {
		t.Fatal(err)
	}
	return run
}
func awaitExecution(t *testing.T, s *server, status string) room.WorkflowRun {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		st, err := s.load(context.Background())
		var run room.WorkflowRun
		if len(st.WorkflowRuns) > 0 {
			run = st.WorkflowRuns[len(st.WorkflowRuns)-1]
		}
		live := s.workflowLive[run.ID] != nil
		s.mu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
		if run.Status == status && !live {
			return run
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("did not reach %s", status)
	return room.WorkflowRun{}
}
func gateExecution(t *testing.T, h http.Handler, run room.WorkflowRun, action, port string) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"action": action, "runId": run.ID, "revision": run.Revision, "port": port, "evidence": "Inspected patch and actual test results; acceptance rationale"})
	w := mockRequest(h, "POST", "/api/workflow-runs", string(body))
	if w.Code != 200 {
		t.Fatalf("gate: %d %s", w.Code, w.Body.String())
	}
}

func TestWorkflowExecutesReviewReworkAndHumanAcceptance(t *testing.T) {
	t.Parallel()
	s, _, h := executionServer(t)
	var reviews atomic.Int32
	var developments atomic.Int32
	s.workflowTurn = func(_ context.Context, p room.Participant, prompt, _ string) (string, error) {
		if !strings.Contains(prompt, "Approved feature") {
			t.Error("approved task omitted")
		}
		if p.ID == "dev" {
			developments.Add(1)
			return "Patch and test report", nil
		}
		if !strings.Contains(prompt, "Patch and test report") {
			t.Error("reviewer lacks prior output")
		}
		if reviews.Add(1) == 1 {
			return `{"outcome":"changes","summary":"Add a failure-path test"}`, nil
		}
		return `{"outcome":"passed","summary":"Required test now supplied"}`, nil
	}
	startExecution(t, h)
	run := awaitExecution(t, s, "awaiting_human")
	if run.NodeID != "evidence" || developments.Load() != 2 {
		t.Fatal("review did not return work", run)
	}
	if len(run.Steps) != 5 || len(run.Steps[1].SHA256) != 64 {
		t.Fatal("missing pinned output attempts", run.Steps)
	}
	gateExecution(t, h, run, "advance", "next")
	run = awaitExecution(t, s, "awaiting_human")
	if run.NodeID != "approval" {
		t.Fatal(run)
	}
	gateExecution(t, h, run, "reject", "changes")
	run = awaitExecution(t, s, "awaiting_human")
	if developments.Load() != 3 || run.NodeID != "evidence" {
		t.Fatal("human cannot return work", run)
	}
	gateExecution(t, h, run, "advance", "next")
	run = awaitExecution(t, s, "awaiting_human")
	gateExecution(t, h, run, "advance", "next")
	run = awaitExecution(t, s, "completed")
	if run.Workflow.Revision != 1 || run.Steps[len(run.Steps)-1].Port != "next" {
		t.Fatal("acceptance not pinned", run)
	}
}

func TestWorkflowRunnableRejectsUnsupportedAndUnsafeGraphs(t *testing.T) {
	t.Parallel()
	s, _, _ := executionServer(t)
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		change func(*room.Workflow)
	}{
		{"parallel", func(w *room.Workflow) { w.Nodes[1].Type = "parallel" }},
		{"retry", func(w *room.Workflow) { w.Nodes[1].Retries = 1 }},
		{"timeout", func(w *room.Workflow) { w.Nodes[1].Timeout = 181 }},
		{"same reviewer", func(w *room.Workflow) { w.Nodes[2].AgentID = "dev" }},
		{"foreign assignee", func(w *room.Workflow) { w.Nodes[1].AgentID = "outsider" }},
		{"missing changes", func(w *room.Workflow) { w.Edges = w.Edges[:len(w.Edges)-1] }},
		{"approval bypass", func(w *room.Workflow) { w.Edges[3].To = "finish" }},
		{"approval before implementation", func(w *room.Workflow) {
			w.Nodes[3].Type = "agent"
			w.Nodes[3].AgentID = "dev"
			w.Nodes[4].Type = "evidence"
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			w := executionGraph()
			c.change(&w)
			if validateRunnable(w, st) == nil {
				t.Fatal("unsafe graph accepted")
			}
		})
	}
}

func TestWorkflowBoundariesAndExplicitRecovery(t *testing.T) {
	t.Parallel()
	t.Run("off", func(t *testing.T) {
		t.Parallel()
		s, db, h := executionServer(t)
		db.seed("feature.flag.updated", map[string]any{"name": "workflowExecution", "enabled": false})
		w := mockRequest(h, "POST", "/api/workflow-runs", `{"action":"start"}`)
		if w.Code != 403 {
			t.Fatal(w.Code)
		}
		st, _ := s.load(context.Background())
		if len(st.WorkflowRuns) != 0 {
			t.Fatal("off dispatched")
		}
	})
	t.Run("restart", func(t *testing.T) {
		t.Parallel()
		s, db, h := executionServer(t)
		db.seed("workflow.run.updated", room.WorkflowRun{ID: "interrupted", RoomID: "engineering", Workflow: executionGraph(), Status: "running", NodeID: "agent", Revision: 1, Visits: map[string]int{}, Steps: []room.WorkflowStep{}, Input: "Approved"})
		w := mockRequest(h, "GET", "/api/workflow-runs", "")
		if w.Code != 200 {
			t.Fatal(w.Code)
		}
		run := awaitExecution(t, s, "interrupted")
		gateExecution(t, h, run, "resume", "")
		awaitExecution(t, s, "awaiting_human")
	})
	t.Run("invalid verdict", func(t *testing.T) {
		t.Parallel()
		s, _, h := executionServer(t)
		s.workflowTurn = func(context.Context, room.Participant, string, string) (string, error) {
			return "Everything passed!", nil
		}
		startExecution(t, h)
		run := awaitExecution(t, s, "failed")
		if !strings.Contains(run.Detail, "explicit passed/changes") {
			t.Fatal(run.Detail)
		}
		st, _ := s.load(context.Background())
		for _, participant := range st.Participants {
			if participant.ID == "qa" && (participant.Status != "Blocked" || participant.StatusDetails == nil || !strings.Contains(participant.StatusDetails.Reason, "explicit verdict")) {
				t.Fatal("review failure lacks blocked diagnostic", participant)
			}
		}
	})
	t.Run("provider failure", func(t *testing.T) {
		t.Parallel()
		s, _, h := executionServer(t)
		s.workflowTurn = func(context.Context, room.Participant, string, string) (string, error) {
			return "", context.DeadlineExceeded
		}
		startExecution(t, h)
		run := awaitExecution(t, s, "failed")
		if run.NodeID != "agent" {
			t.Fatal(run)
		}
	})
	t.Run("isolation unavailable", func(t *testing.T) {
		t.Parallel()
		s, _, h := executionServer(t)
		s.workflowReady = func(string) error { return errors.New("No isolated worker") }
		w := mockRequest(h, "POST", "/api/workflow-runs", `{"action":"start","workflowId":"flow","revision":1,"input":"Task"}`)
		if w.Code != 409 || !strings.Contains(w.Body.String(), "isolated") {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("iteration budget", func(t *testing.T) {
		t.Parallel()
		s, _, h := executionServer(t)
		s.workflowTurn = func(_ context.Context, p room.Participant, _, _ string) (string, error) {
			if p.ID == "qa" {
				return `{"outcome":"changes","summary":"Not ready"}`, nil
			}
			return "attempt", nil
		}
		startExecution(t, h)
		run := awaitExecution(t, s, "failed")
		if !strings.Contains(run.Detail, "iteration limit") {
			t.Fatal(run)
		}
	})
}

func TestWorkflowAuthStaleAndTamperedHistory(t *testing.T) {
	t.Parallel()
	t.Run("owner CSRF", func(t *testing.T) {
		t.Parallel()
		_, _, h := executionServer(t)
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/workflow-runs", strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 401 {
			t.Fatal(w.Code)
		}
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
		w = httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal(w.Code)
		}
	})
	t.Run("non owner", func(t *testing.T) {
		t.Parallel()
		s, db, h := executionServer(t)
		db.seed("actor.created", room.Actor{ID: "human", Kind: "human", Enabled: true, Role: "Developer", RoomIDs: []string{"engineering"}})
		s.auth.sessions["other"] = ownerSession{ActorID: "human", CSRF: "other", Expires: time.Now().Add(time.Hour)}
		r := httptest.NewRequest("GET", "http://127.0.0.1:8787/api/workflow-runs", nil)
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "other"})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatal(w.Code)
		}
	})
	t.Run("foreign and stale", func(t *testing.T) {
		t.Parallel()
		s, _, h := executionServer(t)
		startExecution(t, h)
		run := awaitExecution(t, s, "awaiting_human")
		body, _ := json.Marshal(map[string]any{"action": "advance", "runId": run.ID, "revision": run.Revision - 1, "port": "next", "evidence": "proof"})
		w := mockRequest(h, "POST", "/api/workflow-runs", string(body))
		if w.Code != 409 {
			t.Fatal(w.Code)
		}
		w = mockRequest(h, "POST", "/api/workflow-runs?room=personal-local-owner", string(body))
		if w.Code != 404 {
			t.Fatal(w.Code)
		}
	})
	t.Run("tamper", func(t *testing.T) {
		t.Parallel()
		s, db, h := executionServer(t)
		db.mu.Lock()
		db.events[0].hash = "tampered"
		db.mu.Unlock()
		w := mockRequest(h, "POST", "/api/workflow-runs", `{"action":"start","workflowId":"flow","revision":1,"input":"Task"}`)
		if w.Code != 409 {
			t.Fatal(w.Code)
		}
		st, _ := s.load(context.Background())
		if !st.Incident || len(st.WorkflowRuns) > 0 {
			t.Fatal("integrity bypass")
		}
	})
}

func TestWorkflowCutoffsDiscardLateOutputs(t *testing.T) {
	t.Parallel()
	for _, cutoff := range []string{"cancel", "disable", "stop", "remove"} {
		t.Run(cutoff, func(t *testing.T) {
			t.Parallel()
			s, db, h := executionServer(t)
			entered, release := make(chan struct{}), make(chan struct{})
			s.workflowTurn = func(context.Context, room.Participant, string, string) (string, error) {
				close(entered)
				<-release
				return "late output must not survive", nil
			}
			run := startExecution(t, h)
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("turn not started")
			}
			s.mu.Lock()
			st, _ := s.load(context.Background())
			run = st.WorkflowRuns[0]
			s.mu.Unlock()
			switch cutoff {
			case "cancel":
				gateExecution(t, h, run, "cancel", "")
			case "disable":
				w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"workflowExecution","enabled":false}}`)
				if w.Code != 200 {
					t.Fatal(w.Code, w.Body.String())
				}
			case "stop":
				w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"room"}`)
				if w.Code != 200 {
					t.Fatal(w.Code)
				}
			case "remove":
				db.seed("participant.removed", map[string]string{"id": "dev"})
			}
			close(release)
			status := "failed"
			if cutoff == "cancel" {
				status = "cancelled"
			}
			run = awaitExecution(t, s, status)
			for _, step := range run.Steps {
				if strings.Contains(step.Output, "late output") {
					t.Fatal("cancelled result committed")
				}
			}
		})
	}
}

func TestWorkflowWaitDecisionAndPreservedDisabledInspection(t *testing.T) {
	t.Parallel()
	s, db, h := executionServer(t)
	w := room.Workflow{ID: "gates", RoomID: "engineering", Name: "Human routing", Revision: 1, MaxIterations: 2,
		Nodes: []room.WorkflowNode{{ID: "s", Type: "start", Title: "Start"}, {ID: "w", Type: "wait", Title: "Wait"}, {ID: "d", Type: "decision", Title: "Choose"}, {ID: "a", Type: "approval", Title: "Accept"}, {ID: "f", Type: "finish", Title: "Done"}},
		Edges: []room.WorkflowEdge{{ID: "1", From: "s", To: "w", Port: "next"}, {ID: "2", From: "w", To: "d", Port: "next"}, {ID: "3", From: "d", To: "a", Port: "yes"}, {ID: "4", From: "d", To: "a", Port: "no"}, {ID: "5", From: "a", To: "f", Port: "next"}}}
	db.seed("workflow.revised", w)
	s.workflowTurn = func(context.Context, room.Participant, string, string) (string, error) {
		t.Error("human-only flow dispatched a model")
		return "", nil
	}
	wres := mockRequest(h, "POST", "/api/workflow-runs", `{"action":"start","workflowId":"gates","revision":1,"input":"Choose a direction"}`)
	if wres.Code != 202 {
		t.Fatal(wres.Code, wres.Body.String())
	}
	run := awaitExecution(t, s, "awaiting_human")
	if run.NodeID != "d" || len(run.Steps) != 2 {
		t.Fatal(run)
	}
	gateExecution(t, h, run, "advance", "no")
	run = awaitExecution(t, s, "awaiting_human")
	gateExecution(t, h, run, "advance", "next")
	run = awaitExecution(t, s, "completed")
	db.seed("feature.flag.updated", map[string]any{"name": "workflowExecution", "enabled": false})
	wres = mockRequest(h, "GET", "/api/workflow-runs", "")
	if wres.Code != 200 || !strings.Contains(wres.Body.String(), run.ID) {
		t.Fatal("disabled flag hides history")
	}
	if workflowExecutionEnabled(room.State{}) {
		t.Fatal("absent flag must be off")
	}
	st, _ := s.load(context.Background())
	if len(visibleState(st, "other", "engineering").WorkflowRuns) != 0 {
		t.Fatal("non-owner sees run outputs")
	}
}

func TestWorkflowRejectsForgedRunEventsAndDuplicateStarts(t *testing.T) {
	t.Parallel()
	s, _, h := executionServer(t)
	w := mockRequest(h, "POST", "/api/events", `{"kind":"workflow.run.updated","payload":{"status":"completed"}}`)
	if w.Code != 400 {
		t.Fatal("forged execution accepted", w.Code)
	}
	startExecution(t, h)
	run := awaitExecution(t, s, "awaiting_human")
	w = mockRequest(h, "POST", "/api/workflow-runs", `{"action":"start","workflowId":"flow","revision":1,"input":"Duplicate"}`)
	if w.Code != 409 {
		t.Fatal("duplicate run accepted", w.Code)
	}
	gateExecution(t, h, run, "cancel", "")
}
