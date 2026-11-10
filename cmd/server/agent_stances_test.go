package main

import (
	"agentroom/internal/room"
	"strings"
	"testing"
)

func agentMessage(id, actor, author, text string) room.Message {
	return room.Message{ID: id, RoomID: "engineering", ActorID: actor, ActorType: "agent", Author: author, Text: text}
}

func TestStanceMarkerParsesOnlyTheFirstLine(t *testing.T) {
	t.Parallel()
	cases := []struct{ text, stance, target, detail string }{
		{"[AGREE] looks right", "agree", "", "looks right"},
		{"  [object: Synthetic feature] it skips validation\nmore", "object", "Synthetic feature", "it skips validation"},
		{"[OBJECT] too risky", "object", "", "too risky"},
		{"[ESCALATE] which database?", "escalate", "", "which database?"},
		{"[PROPOSE TASK] add login", "propose", "", "add login"},
		{"[WITHDRAW: Synthetic feature]", "withdraw", "Synthetic feature", ""},
		{"Plain reply\n[OBJECT] quoted later", "", "", ""},
		{"I think [OBJECT] is wrong here", "", "", ""},
		{"[APPROVE] ship it", "", "", ""},
	}
	for _, c := range cases {
		stance, target, detail := stanceMarker(c.text)
		if stance != c.stance || target != c.target || detail != c.detail {
			t.Fatalf("%q => %q %q %q", c.text, stance, target, detail)
		}
	}
}

func TestObjectionBlocksOnlyItsTaskUntilAHumanResolvesIt(t *testing.T) {
	t.Parallel()
	s, db, h, release := dispatchFixture(t)
	defer close(release)
	db.seed("feature.flag.updated", map[string]any{"name": "agentObjections", "enabled": true})
	db.seed("task.revised", room.Task{ID: "other", RoomID: "engineering", Title: "Unrelated work", Criteria: "c", Assignee: "dev", Reviewer: "dev", Status: "Assigned", Revision: 1})
	db.seed("message.created", agentMessage("m1", "bot-a", "Avery", "[OBJECT: synthetic FEATURE] it skips validation"))
	db.seed("message.created", agentMessage("m2", "bot-a", "Avery", "[OBJECT: No such task] vague worry"))
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Objections) != 2 || st.Objections[0].TargetKind != "task" || st.Objections[0].TargetID != "task" || st.Objections[1].TargetKind != "" || st.Objections[1].Status != "open" {
		t.Fatalf("objections not derived: %+v", st.Objections)
	}
	if w := dispatchRequest(h, "objection-key-1", dispatchBody); w.Code != 409 || !strings.Contains(w.Body.String(), "Avery") {
		t.Fatal("objected task was dispatched", w.Code, w.Body.String())
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"objection.resolved","payload":{"messageId":"m1","decision":"maybe"}}`); w.Code != 400 {
		t.Fatal("invalid decision accepted", w.Code)
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"objection.resolved","payload":{"messageId":"missing","decision":"upheld"}}`); w.Code != 404 {
		t.Fatal("unknown objection resolved", w.Code)
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"objection.resolved","payload":{"messageId":"m1","decision":"overruled","note":"validated elsewhere"}}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ = s.load(t.Context())
	if o := st.Objections[0]; o.Status != "resolved" || o.Decision != "overruled" || o.ResolvedBy != ownerID || o.Note != "validated elsewhere" {
		t.Fatalf("resolution not recorded: %+v", o)
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"objection.resolved","payload":{"messageId":"m1","decision":"upheld"}}`); w.Code != 404 {
		t.Fatal("objection resolved twice", w.Code)
	}
	if w := dispatchRequest(h, "objection-key-2", dispatchBody); w.Code != 202 {
		t.Fatal("resolved objection still blocks", w.Code, w.Body.String())
	}
}

func TestAgentWithdrawsOnlyItsOwnObjection(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("feature.flag.updated", map[string]any{"name": "agentObjections", "enabled": true})
	db.seed("task.revised", room.Task{ID: "task", RoomID: "engineering", Title: "Synthetic feature", Revision: 1})
	db.seed("message.created", agentMessage("m1", "bot-a", "Avery", "[OBJECT: Synthetic feature] unsafe"))
	db.seed("message.created", agentMessage("m2", "bot-b", "Blake", "[WITHDRAW: Synthetic feature]"))
	st, _ := s.load(t.Context())
	if st.Objections[0].Status != "open" {
		t.Fatal("another agent withdrew the objection")
	}
	db.seed("message.created", agentMessage("m3", "bot-a", "Avery", "[WITHDRAW: Synthetic feature] convinced"))
	st, _ = s.load(t.Context())
	if o := st.Objections[0]; o.Status != "resolved" || o.Decision != "withdrawn" {
		t.Fatalf("own withdrawal ignored: %+v", o)
	}
	if _, blocked := openObjection(st, "task", "task"); blocked {
		t.Fatal("withdrawn objection still blocks")
	}
}

func TestObjectionsOffLeavesMarkersAsTextAndBlocksNothing(t *testing.T) {
	t.Parallel()
	s, db, h, release := dispatchFixture(t)
	defer close(release)
	db.seed("message.created", agentMessage("m1", "bot-a", "Avery", "[OBJECT: Synthetic feature] unsafe"))
	db.seed("message.created", room.Message{ID: "h1", RoomID: "engineering", ActorID: ownerID, ActorType: "human", Author: "Owner", Text: "[OBJECT: Synthetic feature] human text is not a marker"})
	st, _ := s.load(t.Context())
	if len(st.Objections) != 0 || st.Messages[len(st.Messages)-2].Stance != "" {
		t.Fatal("disabled feature derived objections")
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"objection.resolved","payload":{"messageId":"m1","decision":"upheld"}}`); w.Code != 403 {
		t.Fatal("resolution accepted while disabled", w.Code)
	}
	if w := dispatchRequest(h, "objection-key-1", dispatchBody); w.Code != 202 {
		t.Fatal("disabled objection blocked dispatch", w.Code, w.Body.String())
	}
	db.seed("feature.flag.updated", map[string]any{"name": "agentObjections", "enabled": true})
	st, _ = s.load(t.Context())
	if len(st.Objections) != 1 {
		t.Fatalf("human text must never become an objection: %+v", st.Objections)
	}
	prompt, err := chatPrompt(st, st.Participants[0], "hello")
	if err != nil || !strings.Contains(prompt, "[OBJECT: exact task or proposal title]") {
		t.Fatal("agents are not told about markers when enabled", err)
	}
}
