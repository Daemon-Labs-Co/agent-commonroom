package main

import (
	"agentroom/internal/room"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestAdmissionBoundsIncludeOtherRoomsAndRemovedActiveJobs(t *testing.T) {
	t.Parallel()
	s, _, _ := mockServer(t)
	st := room.State{}
	for i := 0; i < maxRoomAgentTurns; i++ {
		pid := fmt.Sprint(i)
		s.active[pid] = true
		st.Participants = append(st.Participants, room.Participant{ID: pid, RoomID: "one"})
	}
	if s.agentAdmission(st, "one", 1) == nil || s.agentAdmission(st, "two", 1) != nil {
		t.Fatal("room budget not isolated")
	}
	for i := maxRoomAgentTurns; i < maxActiveAgentTurns; i++ {
		s.active[fmt.Sprint(i)] = true
	}
	if s.agentAdmission(st, "two", 1) == nil {
		t.Fatal("missing participant bypassed global bound")
	}
	delete(s.active, "7")
	if s.agentAdmission(st, "two", 1) != nil || s.agentAdmission(st, "two", 2) == nil {
		t.Fatal("batch admission not atomic")
	}
}

func TestChatQuotaRejectsBeforeProviderOrHistoryWrite(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "c", Protocol: "CLI", Endpoint: "codex"})
	db.seed("participant.created", room.Participant{ID: "p", RoomID: "engineering", ConnectionID: "c"})
	for i := 0; i < maxActiveAgentTurns; i++ {
		s.active[fmt.Sprint(i)] = true
	}
	before := len(db.snapshot())
	w := mockRequest(http.HandlerFunc(s.agents), "POST", "/api/agents", `{"action":"broadcast","text":"Synthetic request"}`)
	if w.Code != 429 || len(db.snapshot()) != before || s.active["p"] {
		t.Fatal("overload mutated or dispatched", w.Code, w.Body.String())
	}
}

func TestWorkflowQuotaPreventsProviderDispatch(t *testing.T) {
	t.Parallel()
	s, _, h := executionServer(t)
	for i := 0; i < maxActiveAgentTurns; i++ {
		s.active[fmt.Sprint(i)] = true
	}
	s.workflowTurn = func(context.Context, room.Participant, string, string) (string, error) {
		t.Error("overloaded workflow launched provider")
		return "", nil
	}
	startExecution(t, h)
	run := awaitExecution(t, s, "failed")
	if !strings.Contains(run.Detail, "concurrency limit") {
		t.Fatal(run.Detail)
	}
}
