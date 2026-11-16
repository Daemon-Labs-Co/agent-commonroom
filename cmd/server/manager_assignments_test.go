package main

import (
	"agentroom/internal/room"
	"strings"
	"testing"
)

func hasNotice(st room.State, fragment string) bool {
	for _, m := range st.Messages {
		if m.ActorType == "system" && strings.Contains(m.Text, fragment) {
			return true
		}
	}
	return false
}

func TestManagerAssignments(t *testing.T) {
	// Synthetic worker configuration; the scripted turn never reaches a provider.
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	const toAvery = `{"action":"send","participantId":"a","text":"Plan the parser work"}`
	const toBlake = `{"action":"send","participantId":"b","text":"Plan the parser work"}`
	hierarchy := room.OrganizationChart{Revision: 1, Nodes: []room.OrganizationNode{{ActorID: "bot-ca"}, {ActorID: "bot-cb", ReportsTo: "bot-ca"}}}
	enable := func(db *mockDB) {
		db.seed("feature.flag.updated", map[string]any{"name": "managerAssignedTasks", "enabled": true})
		db.seed("company.organization.updated", hierarchy)
	}

	t.Run("a manager assigns work to its report", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, false)
		enable(db)
		script.reply = func(name string, _ int) string {
			if name == "Avery" {
				return "[ASSIGN blake] Write the parser\nUse the existing tokenizer."
			}
			return "Parser written."
		}
		if w := mockRequest(h, "POST", "/api/agents", toAvery); w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		st := awaitIdle(t, s)
		if strings.Join(script.calls, ",") != "Avery,Blake" {
			t.Fatal("report did not get a turn", script.calls)
		}
		if p := script.prompts[1]; !strings.Contains(p, "Avery, who you report to, assigned you this work: Write the parser") || !strings.Contains(p, "Use the existing tokenizer.") || !strings.Contains(p, "cannot approve, accept or publish") {
			t.Fatal("assignment instruction incomplete")
		}
		if !hasNotice(st, "Avery assigned Blake: Write the parser") || st.Rooms[0].AssignmentsToday != 1 || len(st.Tasks) != 0 {
			t.Fatal("assignment not announced, not counted, or it created a task record")
		}
	})

	t.Run("only the reporting line may assign", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, false)
		enable(db)
		script.reply = func(string, int) string { return "[ASSIGN Avery] Do my work" }
		mockRequest(h, "POST", "/api/agents", toBlake)
		st := awaitIdle(t, s)
		if len(script.calls) != 1 || !hasNotice(st, "Avery does not report to Blake") || st.Rooms[0].AssignmentsToday != 0 {
			t.Fatal("a report assigned work to its manager", script.calls)
		}
		script.reply = func(string, int) string { return "[ASSIGN Nobody] Do it" }
		mockRequest(h, "POST", "/api/agents", toAvery)
		if st = awaitIdle(t, s); !hasNotice(st, "no single agent named Nobody") {
			t.Fatal("unknown assignee not refused visibly")
		}
	})

	t.Run("owner's daily limit applies immediately", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, false)
		enable(db)
		if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.assignment_limit.updated","payload":{"limit":51}}`); w.Code != 400 {
			t.Fatal("limit above the maximum accepted", w.Code)
		}
		if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.assignment_limit.updated","payload":{}}`); w.Code != 400 {
			t.Fatal("missing limit accepted", w.Code)
		}
		if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.assignment_limit.updated","payload":{"limit":1}}`); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		script.reply = func(name string, _ int) string {
			if name == "Avery" {
				return "[ASSIGN Blake] One more thing"
			}
			return "Done."
		}
		mockRequest(h, "POST", "/api/agents", toAvery)
		awaitIdle(t, s)
		mockRequest(h, "POST", "/api/agents", toAvery)
		st := awaitIdle(t, s)
		if strings.Join(script.calls, ",") != "Avery,Blake,Avery" || !hasNotice(st, "daily limit of 1") {
			t.Fatal("daily limit not enforced", script.calls)
		}
		if w := mockRequest(h, "POST", "/api/events", `{"kind":"room.assignment_limit.updated","payload":{"limit":0}}`); w.Code != 200 {
			t.Fatal("zero must be allowed", w.Code)
		}
		if st, _ = s.load(t.Context()); roomAssignmentLimit(st, "engineering") != 0 {
			t.Fatal("zero limit not stored")
		}
	})

	t.Run("disabled: the marker is plain text", func(t *testing.T) {
		s, db, h, script := discussionFixture(t, false)
		db.seed("company.organization.updated", hierarchy)
		script.reply = func(string, int) string { return "[ASSIGN Blake] Write the parser" }
		mockRequest(h, "POST", "/api/agents", toAvery)
		st := awaitIdle(t, s)
		if len(script.calls) != 1 || hasNotice(st, "assign") || st.Rooms[0].AssignmentsToday != 0 {
			t.Fatal("disabled feature acted on an assignment", script.calls)
		}
		if roomAssignmentLimit(st, "engineering") != defaultManagerAssignmentLimit {
			t.Fatal("default limit")
		}
	})
}

func TestTeamLoopGuidanceMatchesEnabledFeaturesAndReportingLine(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "ca", Provider: "openai", Protocol: "CLI", Endpoint: "ssh://worker", Name: "Codex"})
	db.seed("connection.created", room.Connection{ID: "cb", Provider: "anthropic", Protocol: "CLI", Endpoint: "ssh://worker/claude", Name: "Claude"})
	db.seed("participant.created", room.Participant{ID: "a", RoomID: "engineering", ConnectionID: "ca", Name: "Avery", Rank: 1})
	db.seed("participant.created", room.Participant{ID: "b", RoomID: "engineering", ConnectionID: "cb", Name: "Blake", Rank: 1})
	db.seed("company.organization.updated", room.OrganizationChart{Revision: 1, Nodes: []room.OrganizationNode{{ActorID: "bot-ca"}, {ActorID: "bot-cb", ReportsTo: "bot-ca"}}})
	prompts := func() (string, string) {
		st, err := s.load(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		a, errA := chatPrompt(st, st.Participants[0], "hello")
		b, errB := chatPrompt(st, st.Participants[1], "hello")
		if errA != nil || errB != nil {
			t.Fatal(errA, errB)
		}
		return a, b
	}
	if a, _ := prompts(); strings.Contains(a, "[DISCUSS]") || strings.Contains(a, "[PROPOSE TASK]") || strings.Contains(a, "[ASSIGN") {
		t.Fatal("disabled features were advertised to an agent")
	}
	for _, f := range []string{"agentDiscussion", "governingSuggestions", "managerAssignedTasks"} {
		db.seed("feature.flag.updated", map[string]any{"name": f, "enabled": true})
	}
	manager, report := prompts()
	if !strings.Contains(manager, "[DISCUSS]") || !strings.Contains(manager, "[PROPOSE TASK] short title") || !strings.Contains(manager, "These agents report to you in this room: Blake") {
		t.Fatal("manager not told what it may use")
	}
	if strings.Contains(report, "[ASSIGN") || !strings.Contains(report, "[DISCUSS]") {
		t.Fatal("a report was told it may assign work, or lost the shared guidance")
	}
}
