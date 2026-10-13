package main

import (
	"agentroom/internal/room"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestWorkflowBrowserHarness is an opt-in manual UI fixture. It serves the real
// handlers with in-memory event storage and fake providers, never PostgreSQL,
// real credentials or real agents. Normal suites skip the bounded browser wait.
func TestWorkflowBrowserHarness(t *testing.T) {
	historyFixture := os.Getenv("COMMONROOM_BROWSER_HISTORY_TEST") == "1"
	attentionFixture := os.Getenv("COMMONROOM_BROWSER_ATTENTION_TEST") == "1"
	scrollFixture := os.Getenv("COMMONROOM_BROWSER_SCROLL_TEST") == "1"
	statusFixture := os.Getenv("COMMONROOM_BROWSER_AGENT_STATUS_TEST") == "1"
	if os.Getenv("COMMONROOM_BROWSER_WORKFLOW_TEST") != "1" && !statusFixture && !scrollFixture && !attentionFixture && !historyFixture {
		t.Skip("manual synthetic browser fixture")
	}
	s, db, _ := executionServer(t)
	s.recoveryIdle = func(context.Context, string) (bool, error) { return true, nil }
	s.auth.port = "8794"
	db.seed("human.profile.updated", map[string]string{"name": "Demo owner", "title": "Synthetic verification"})
	db.seed("room.updated", room.Room{ID: "engineering", Name: "Workflow verification"})
	if statusFixture {
		for _, pid := range []string{"dev", "qa"} {
			reason := "context deadline exceeded"
			if pid == "qa" {
				reason = "Login failed: proxy refused the connection"
			}
			db.seed("agent.activity", map[string]string{"participantId": pid, "activity": "Workflow: Example feature / Review changes"})
			db.seed("agent.error", map[string]string{"participantId": pid, "error": reason})
			db.seed("participant.session", room.Participant{ID: pid, RoomID: "engineering", ConnectionID: pid, Name: pid, Status: "Blocked", ThreadID: "synthetic-session", ExecutionMode: "container-v1"})
			for _, stage := range []string{"preparing", "running", "command_started", "command_finished", "timed_out", "shutdown_confirmed"} {
				db.seed("agent.background", room.BackgroundActivity{ParticipantID: pid, RoomID: "engineering", TurnID: "synthetic-turn", Stage: stage, Label: backgroundLabels[stage], ObservedAt: time.Now().UTC()})
			}
		}
	}
	if scrollFixture {
		db.seed("participant.session", room.Participant{ID: "dev", RoomID: "engineering", ConnectionID: "dev", Name: "dev", Status: "Working"})
		for i := 0; i < 25; i++ {
			db.seed("message.created", room.Message{ID: fmt.Sprintf("synthetic-%d", i), RoomID: "engineering", ActorID: ownerID, ActorType: "human", Author: "Demo owner", Text: fmt.Sprintf("Synthetic message %d.\n", i) + strings.Repeat("Synthetic scroll verification text. This is not a real conversation.\n", 4), Created: time.Unix(int64(i), 0).UTC()})
		}
	}
	if attentionFixture {
		db.seed("message.created", room.Message{ID: "explicit-request", RoomID: "engineering", ActorType: "agent", Author: "dev", Text: "[HUMAN ACTION REQUIRED] Please choose the design before implementation begins.", Created: time.Unix(1, 0).UTC()})
		db.seed("message.created", room.Message{ID: "suggested-request", RoomID: "engineering", ActorType: "agent", Author: "qa", Text: "Please confirm the acceptance criteria so I can write the tests.", Created: time.Unix(2, 0).UTC()})
		db.seed("message.created", room.Message{ID: "routine", RoomID: "engineering", ActorType: "agent", Author: "dev", Text: "Routine progress update: the design notes are ready.", Created: time.Unix(3, 0).UTC()})
	}
	if historyFixture {
		db.seed("agent.error", map[string]string{"name": "Demo developer", "participantId": "dev", "roomId": "engineering", "error": "context deadline exceeded; synthetic-raw-diagnostic"})
		db.seed("task.revised", room.Task{ID: "demo-task", RoomID: "engineering", Title: "Search-filter implementation", Status: "Accepted", Revision: 1})
		db.seed("workflow.run.updated", room.WorkflowRun{ID: "demo-run", RoomID: "engineering", Workflow: executionGraph(), ActorID: ownerID, Status: "awaiting_human", NodeID: "approval", Revision: 1})
		for i := 0; i < 4; i++ {
			db.seed("participant.session", room.Participant{ID: "dev", RoomID: "engineering", ConnectionID: "dev", Name: "Demo developer", Status: "Idle"})
		}
		db.seed("feature.flag.updated", map[string]any{"name": "humanAttention", "enabled": true})
	}
	var refreshes atomic.Int32
	var reviews atomic.Int32
	s.workflowTurn = func(_ context.Context, p room.Participant, _, _ string) (string, error) {
		if p.ID == "qa" {
			if reviews.Add(1) == 1 {
				return `{"outcome":"changes","summary":"Add a failure-path test (synthetic review)"}`, nil
			}
			return `{"outcome":"passed","summary":"Revised output accepted by synthetic reviewer"}`, nil
		}
		return "Synthetic implementation output; no provider or filesystem changes were made.", nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/workflow-runs", s.workflowRuns)
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		if !scrollFixture {
			s.state(w, r)
			return
		}
		count := refreshes.Add(1)
		if count == 2 {
			db.seed("message.created", room.Message{ID: "synthetic-new", RoomID: "engineering", ActorType: "human", ActorID: ownerID, Author: "Demo owner", Text: "A new synthetic message arrived.\n" + strings.Repeat("New-message scrolling verification.\n", 4), Created: time.Unix(30, 0).UTC()})
		}
		var body bytes.Buffer
		rec := &bufferResponse{header: http.Header{}, body: &body}
		s.state(rec, r)
		var st room.State
		if rec.status >= 400 || json.Unmarshal(body.Bytes(), &st) != nil {
			http.Error(w, "Synthetic fixture state failed", 500)
			return
		}
		for i := range st.Rooms {
			if st.Rooms[i].ID == "engineering" {
				st.Rooms[i].Name = fmt.Sprintf("Scroll verification · refresh %d", count)
			}
		}
		reply(w, st)
	})
	mux.HandleFunc("/api/session", s.session)
	mux.HandleFunc("/api/login", s.login)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/agent-recovery", s.agentRecovery)
	mux.Handle("/help/", documentationHandler("../.."))
	mux.Handle("/", http.FileServer(http.Dir("../../web")))
	host := &http.Server{Addr: "127.0.0.1:8794", Handler: s.auth.protect(s.actorGuard(mux)), ReadHeaderTimeout: 5 * time.Second}
	t.Cleanup(func() { _ = host.Close() })
	errs := make(chan error, 1)
	go func() { errs <- host.ListenAndServe() }()
	t.Log("Synthetic UI fixture on 127.0.0.1:8794; bootstrap uses the existing mockServer test key")
	select {
	case err := <-errs:
		t.Fatal(err)
	case <-time.After(func() time.Duration {
		if statusFixture || scrollFixture || attentionFixture || historyFixture {
			return time.Minute
		}
		return 3 * time.Minute
	}()):
	}
}
