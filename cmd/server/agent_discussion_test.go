package main

import (
	"agentroom/internal/room"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type discussionScript struct {
	mu      sync.Mutex
	calls   []string
	prompts []string
	reply   func(name string, call int) string
	block   chan struct{}
}

func discussionFixture(t *testing.T, enabled bool) (*server, *mockDB, http.Handler, *discussionScript) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "ca", Provider: "openai", Protocol: "CLI", Endpoint: "ssh://worker", Name: "Codex"})
	db.seed("connection.created", room.Connection{ID: "cb", Provider: "anthropic", Protocol: "CLI", Endpoint: "ssh://worker/claude", Name: "Claude"})
	db.seed("participant.created", room.Participant{ID: "a", RoomID: "engineering", ConnectionID: "ca", Name: "Avery", ActorID: "bot-a", Rank: 1})
	db.seed("participant.created", room.Participant{ID: "b", RoomID: "engineering", ConnectionID: "cb", Name: "Blake", ActorID: "bot-b", Rank: 1})
	if enabled {
		db.seed("feature.flag.updated", map[string]any{"name": "agentDiscussion", "enabled": true})
	}
	script := &discussionScript{reply: func(name string, call int) string { return name + " reply " + fmt.Sprint(call) }}
	s.agentTurn = func(ctx context.Context, p room.Participant, prompt, _ string) (string, error) {
		script.mu.Lock()
		script.calls = append(script.calls, p.Name)
		script.prompts = append(script.prompts, prompt)
		call := len(script.calls)
		block := script.block
		script.mu.Unlock()
		if block != nil {
			select {
			case <-block:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		return script.reply(p.Name, call), nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/agents", s.agents)
	mux.HandleFunc("/api/events", s.events)
	return s, db, s.auth.protect(s.actorGuard(mux)), script
}

func awaitDiscussionEnd(t *testing.T, s *server) room.State {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		st, err := s.load(context.Background())
		running := len(s.discussions)
		s.mu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
		if running == 0 && activeDiscussion(st, "engineering") == "" {
			return st
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("discussion did not end")
	return room.State{}
}

func lastNotice(st room.State) string {
	for i := len(st.Messages) - 1; i >= 0; i-- {
		if st.Messages[i].ActorType == "system" {
			return st.Messages[i].Text
		}
	}
	return ""
}

func TestAgentDiscussions(t *testing.T) {
	// Synthetic worker configuration; the scripted turn never reaches a provider.
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	const start = `{"action":"discuss","text":"Which database should we use?"}`

	t.Run("agents reply to each other until all agree", func(t *testing.T) {
		s, _, h, script := discussionFixture(t, true)
		script.reply = func(name string, call int) string {
			if call > 2 {
				return "[AGREE] settled by " + name
			}
			return name + " prefers option " + fmt.Sprint(call)
		}
		if w := mockRequest(h, "POST", "/api/agents", start); w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		st := awaitDiscussionEnd(t, s)
		if strings.Join(script.calls, ",") != "Avery,Blake,Avery,Blake" {
			t.Fatal("turn order", script.calls)
		}
		if !strings.Contains(script.prompts[1], "Avery prefers option 1") || !strings.Contains(script.prompts[2], "Blake prefers option 2") {
			t.Fatal("an agent did not see the other agent's earlier reply")
		}
		if !strings.Contains(script.prompts[0], "Discussion round 1 of 3") || !strings.Contains(script.prompts[3], "Discussion round 2 of 3") {
			t.Fatal("round not stated to agents")
		}
		if notice := lastNotice(st); !strings.Contains(notice, "after 4 turn(s): every agent agreed") {
			t.Fatal("end notice:", notice)
		}
	})

	t.Run("round limit ends a discussion that never settles", func(t *testing.T) {
		s, _, h, script := discussionFixture(t, true)
		if w := mockRequest(h, "POST", "/api/agents", start); w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		st := awaitDiscussionEnd(t, s)
		if len(script.calls) != discussionRoundCeiling*2 || !strings.Contains(lastNotice(st), "round limit") {
			t.Fatal("round limit not enforced", len(script.calls), lastNotice(st))
		}
	})

	t.Run("escalation stops immediately", func(t *testing.T) {
		s, _, h, script := discussionFixture(t, true)
		script.reply = func(string, int) string { return "[ESCALATE] which budget applies?" }
		mockRequest(h, "POST", "/api/agents", start)
		st := awaitDiscussionEnd(t, s)
		if len(script.calls) != 1 || !strings.Contains(lastNotice(st), "escalated") {
			t.Fatal("escalation did not stop the discussion", script.calls, lastNotice(st))
		}
	})

	t.Run("owner lowers the room limit and cannot raise it", func(t *testing.T) {
		s, _, h, script := discussionFixture(t, true)
		if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.discussion_limit.updated","payload":{"rounds":4}}`); w.Code != 400 {
			t.Fatal("limit raised above the ceiling", w.Code)
		}
		if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.discussion_limit.updated","payload":{"rounds":1}}`); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		if w := mockRequest(h, "POST", "/api/agents", `{"action":"discuss","text":"topic","rounds":2}`); w.Code != 400 {
			t.Fatal("request exceeded the room limit", w.Code)
		}
		mockRequest(h, "POST", "/api/agents", start)
		awaitDiscussionEnd(t, s)
		if len(script.calls) != 2 {
			t.Fatal("lowered limit ignored", script.calls)
		}
	})

	t.Run("one at a time and a human can stop it", func(t *testing.T) {
		s, _, h, script := discussionFixture(t, true)
		script.block = make(chan struct{})
		if w := mockRequest(h, "POST", "/api/agents", start); w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		if w := mockRequest(h, "POST", "/api/agents", start); w.Code != 409 {
			t.Fatal("second discussion accepted", w.Code)
		}
		for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
			script.mu.Lock()
			started := len(script.calls)
			script.mu.Unlock()
			if started == 1 {
				break
			}
		}
		if w := mockRequest(h, "POST", "/api/agents", `{"action":"stop-discussion"}`); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		st := awaitDiscussionEnd(t, s)
		if !strings.Contains(lastNotice(st), "a human stopped it") || len(script.calls) != 1 {
			t.Fatal("stop not honoured", lastNotice(st), script.calls)
		}
		if w := mockRequest(h, "POST", "/api/agents", `{"action":"stop-discussion"}`); w.Code != 409 {
			t.Fatal("stopping nothing succeeded", w.Code)
		}
	})

	t.Run("refusals", func(t *testing.T) {
		_, db, h, script := discussionFixture(t, false)
		before := len(db.snapshot())
		if w := mockRequest(h, "POST", "/api/agents", start); w.Code != 403 {
			t.Fatal("disabled feature accepted", w.Code)
		}
		db.seed("feature.flag.updated", map[string]any{"name": "agentDiscussion", "enabled": true})
		if w := mockRequest(h, "POST", "/api/agents", `{"action":"discuss","text":"  "}`); w.Code != 400 {
			t.Fatal("empty topic accepted", w.Code)
		}
		db.seed("participant.removed", map[string]string{"id": "b", "roomId": "engineering"})
		if w := mockRequest(h, "POST", "/api/agents", start); w.Code != 409 {
			t.Fatal("single-agent discussion accepted", w.Code)
		}
		if len(script.calls) != 0 || len(db.snapshot()) != before+2 {
			t.Fatal("a refused discussion started a turn or wrote history")
		}
	})
}

func awaitIdle(t *testing.T, s *server) room.State {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		st, err := s.load(context.Background())
		busy := len(s.discussions)
		for _, active := range s.active {
			if active {
				busy++
			}
		}
		s.mu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
		if busy == 0 && activeDiscussion(st, "engineering") == "" {
			time.Sleep(5 * time.Millisecond) // let a pending reaction start, then re-check
			s.mu.Lock()
			again := len(s.discussions)
			for _, active := range s.active {
				if active {
					again++
				}
			}
			s.mu.Unlock()
			if again == 0 {
				st, _ = s.load(context.Background())
				return st
			}
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("room did not become idle")
	return room.State{}
}

func TestAgentReactions(t *testing.T) {
	// Synthetic worker configuration; the scripted turn never reaches a provider.
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	reportsTo := func(db *mockDB, report, manager string) {
		db.seed("company.organization.updated", room.OrganizationChart{Revision: 1, Nodes: []room.OrganizationNode{{ActorID: manager}, {ActorID: report, ReportsTo: manager}}})
	}
	const toBlake = `{"action":"send","participantId":"b","text":"Should we ship today?"}`

	t.Run("escalation goes to the agent manager once and stays visible to the human", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, true)
		reportsTo(db, "bot-cb", "bot-ca")
		script.reply = func(name string, _ int) string {
			if name == "Blake" {
				return "[ESCALATE] ship with the known bug?"
			}
			return "No. Fix it first."
		}
		if w := mockRequest(h, "POST", "/api/agents", toBlake); w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		st := awaitIdle(t, s)
		if strings.Join(script.calls, ",") != "Blake,Avery" || !strings.Contains(script.prompts[1], "Blake reports to you and escalated this") || !strings.Contains(script.prompts[1], "ship with the known bug?") {
			t.Fatal("manager did not answer the escalation", script.calls)
		}
		pending := 0
		for _, m := range st.Messages {
			if m.HumanRequest == "explicit" && strings.HasPrefix(m.Text, "[ESCALATE]") {
				pending++
			}
		}
		if pending != 1 || st.Rooms[0].PendingHumanRequests != 1 {
			t.Fatal("escalation not shown to the human", pending, st.Rooms[0].PendingHumanRequests)
		}
	})

	t.Run("a peer or a human manager gets no agent turn", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, true)
		reportsTo(db, "bot-cb", ownerID)
		script.reply = func(string, int) string { return "[ESCALATE] need a decision" }
		mockRequest(h, "POST", "/api/agents", toBlake)
		awaitIdle(t, s)
		if len(script.calls) != 1 {
			t.Fatal("escalation started a turn for a non-manager", script.calls)
		}
	})

	t.Run("agents cannot keep each other running", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, true)
		db.seed("company.organization.updated", room.OrganizationChart{Revision: 1, Nodes: []room.OrganizationNode{{ActorID: "bot-ca", ReportsTo: "bot-cb"}, {ActorID: "bot-cb", ReportsTo: "bot-ca"}}})
		script.reply = func(string, int) string { return "[ESCALATE] over to you" }
		mockRequest(h, "POST", "/api/agents", toBlake)
		st := awaitIdle(t, s)
		if len(script.calls) != 1+maxAgentInitiatedPerHumanMessage || st.Rooms[0].AgentInitiated != maxAgentInitiatedPerHumanMessage {
			t.Fatal("agent-started turns were not bounded", script.calls, st.Rooms[0].AgentInitiated)
		}
		script.reply = func(string, int) string { return "Plain reply" }
		mockRequest(h, "POST", "/api/agents", toBlake)
		if st = awaitIdle(t, s); st.Rooms[0].AgentInitiated != 0 {
			t.Fatal("a human message must reset the allowance")
		}
	})

	t.Run("an agent starts a discussion", func(t *testing.T) {
		s, _, h, script := discussionFixture(t, true)
		script.reply = func(name string, call int) string {
			if call == 1 {
				return "[DISCUSS] which queue library?"
			}
			return "[AGREE] " + name
		}
		mockRequest(h, "POST", "/api/agents", toBlake)
		st := awaitIdle(t, s)
		if strings.Join(script.calls, ",") != "Blake,Avery,Blake" {
			t.Fatal("discussion order after an agent start", script.calls)
		}
		found := false
		for _, m := range st.Messages {
			found = found || (m.ActorType == "system" && m.Text == "Blake started a discussion: which queue library?")
		}
		if !found || !strings.Contains(lastNotice(st), "every agent agreed") {
			t.Fatal("agent-started discussion not announced or not closed", lastNotice(st))
		}
	})

	t.Run("disabled: markers start nothing", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, false)
		reportsTo(db, "bot-cb", "bot-ca")
		script.reply = func(_ string, call int) string {
			if call == 1 {
				return "[ESCALATE] decide please"
			}
			return "[DISCUSS] anything"
		}
		mockRequest(h, "POST", "/api/agents", toBlake)
		st := awaitIdle(t, s)
		mockRequest(h, "POST", "/api/agents", toBlake)
		st = awaitIdle(t, s)
		if len(script.calls) != 2 || st.Rooms[0].PendingHumanRequests != 0 {
			t.Fatal("disabled feature reacted to a marker", script.calls)
		}
	})
}
