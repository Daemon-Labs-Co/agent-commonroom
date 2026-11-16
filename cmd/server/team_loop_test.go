package main

import (
	"agentroom/internal/room"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// TestTeamLoopEndToEnd walks the whole loop with scripted agents: two agents
// discuss and agree, the room asks for task suggestions, the owner captures,
// approves, assigns and sends the task, the assignee does it, and the manager
// then assigns follow-up work to its report.
func TestTeamLoopEndToEnd(t *testing.T) {
	// Synthetic worker configuration; scripted turns never reach a provider.
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	s, db, _, script := discussionFixture(t, true)
	for _, flag := range []string{"governingSuggestions", "taskProposals", "agentObjections", "managerAssignedTasks"} {
		db.seed("feature.flag.updated", map[string]any{"name": flag, "enabled": true})
	}
	db.seed("company.organization.updated", room.OrganizationChart{Revision: 1, Nodes: []room.OrganizationNode{{ActorID: "bot-ca"}, {ActorID: "bot-cb", ReportsTo: "bot-ca"}}})
	taskTurns := []string{}
	s.taskReady = func(string) error { return nil }
	s.taskTurn = func(_ context.Context, p room.Participant, prompt, _ string) (string, error) {
		taskTurns = append(taskTurns, p.Name)
		return "Login built and tested.", nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/agents", s.agents)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/task-proposals", s.taskProposals)
	mux.HandleFunc("/api/task-dispatch", s.taskDispatch)
	h := s.auth.protect(s.actorGuard(mux))

	// 1. Discussion: disagreement in round one, agreement in round two, then
	// the first agent is asked for suggestions automatically.
	script.reply = func(name string, call int) string {
		switch call {
		case 1:
			return "Use sessions with server-side storage."
		case 2:
			return "I would rather use signed tokens; sessions need shared storage."
		case 3, 4:
			return "[AGREE] sessions, with storage we already run."
		default:
			return "[PROPOSE TASK] Build login :: Users can sign in and out :: Blake"
		}
	}
	if w := mockRequest(h, "POST", "/api/agents", `{"action":"discuss","text":"How should login work?"}`); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	st := awaitIdle(t, s)
	if strings.Join(script.calls, ",") != "Avery,Blake,Avery,Blake,Avery" {
		t.Fatal("discussion and suggestion turns", script.calls)
	}
	if !strings.Contains(script.prompts[1], "Use sessions with server-side storage.") || !strings.Contains(script.prompts[2], "signed tokens") {
		t.Fatal("agents did not see each other's arguments")
	}
	if len(st.TaskSuggestions) != 1 || st.TaskSuggestions[0].Assignee != "Blake" || len(st.Tasks) != 0 {
		t.Fatalf("expected one suggestion and no task yet: %+v", st.TaskSuggestions)
	}

	// 2. The owner captures, approves, assigns and sends: the same four requests
	// the interface's "Approve and send" button makes.
	sug := st.TaskSuggestions[0]
	if w := proposalRequest(h, "owner", "loop-capture-key", fmt.Sprintf(`{"action":"capture","sourceMessageId":%q,"title":%q,"criteria":%q}`, sug.MessageID, sug.Title, sug.Criteria)); w.Code >= 300 {
		t.Fatal("capture", w.Code, w.Body.String())
	}
	st, _ = s.load(t.Context())
	var proposal room.TaskProposal
	for _, p := range st.TaskProposals {
		proposal = p
	}
	if w := proposalRequest(h, "owner", "loop-approve-key", fmt.Sprintf(`{"action":"approve","proposalId":%q,"revision":%d,"title":%q,"criteria":%q}`, proposal.ID, proposal.Revision, sug.Title, sug.Criteria)); w.Code >= 300 {
		t.Fatal("approve", w.Code, w.Body.String())
	}
	st, _ = s.load(t.Context())
	if len(st.Tasks) != 1 || st.Tasks[0].Status != "Draft" || st.TaskSuggestions[0].Status != "captured" {
		t.Fatal("approval did not create exactly one draft task")
	}
	task := st.Tasks[0]
	assign := fmt.Sprintf(`{"kind":"task.revised","payload":{"id":%q,"proposalId":%q,"title":%q,"criteria":%q,"assignee":"b","reviewer":"a","status":"Assigned","revision":%d}}`, task.ID, task.ProposalID, task.Title, task.Criteria, task.Revision)
	if w := mockRequest(h, "POST", "/api/events", assign); w.Code != 200 {
		t.Fatal("assign", w.Code, w.Body.String())
	}
	st, _ = s.load(t.Context())
	if w := dispatchRequest(h, "loop-dispatch-key", fmt.Sprintf(`{"taskId":%q,"revision":%d,"mode":"work"}`, task.ID, st.Tasks[0].Revision)); w.Code != 202 {
		t.Fatal("dispatch", w.Code, w.Body.String())
	}
	st = awaitIdle(t, s)
	if strings.Join(taskTurns, ",") != "Blake" || st.Tasks[0].DispatchStatus != "completed" || st.Tasks[0].Status == "Accepted" {
		t.Fatal("the assignee did not do the task, or it was accepted without a human", taskTurns, st.Tasks[0].DispatchStatus, st.Tasks[0].Status)
	}

	// 3. The manager hands follow-up work to its report.
	script.reply = func(name string, _ int) string {
		if name == "Avery" {
			return "[ASSIGN Blake] Add password reset\nReuse the login session code."
		}
		return "Password reset added."
	}
	before := len(script.calls)
	if w := mockRequest(h, "POST", "/api/agents", `{"action":"send","participantId":"a","text":"What is next?"}`); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	st = awaitIdle(t, s)
	if strings.Join(script.calls[before:], ",") != "Avery,Blake" || !hasNotice(st, "Avery assigned Blake: Add password reset") {
		t.Fatal("manager assignment did not reach the report", script.calls[before:])
	}
	for _, m := range st.Messages {
		if m.ActorType == "agent" && m.Author == "" {
			t.Fatal("an agent message lost its author")
		}
	}
	if st.Rooms[0].DiscussionID != "" || len(st.Objections) != 0 {
		t.Fatal("loop left a discussion or objection open")
	}
}
