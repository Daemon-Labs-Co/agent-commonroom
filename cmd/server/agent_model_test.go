package main

import (
	"agentroom/internal/room"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func modelFixture(t *testing.T, flagOff bool) (*server, *mockDB, func(kind, payload string) (int, string)) {
	t.Helper()
	s, db, h := governanceFixture(t, false)
	db.seed("connection.created", room.Connection{ID: "w", Provider: "anthropic", Protocol: "CLI", Endpoint: "ssh://commonroom-server/claude", Name: "Claude"})
	if flagOff {
		db.seed("feature.flag.updated", map[string]any{"name": "agentModelSelection", "enabled": false})
	}
	return s, db, func(kind, payload string) (int, string) {
		w := mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":%q,"payload":%s}`, kind, payload))
		return w.Code, w.Body.String()
	}
}

func participantByName(t *testing.T, s *server, name string) room.Participant {
	t.Helper()
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range st.Participants {
		if p.Name == name {
			return p
		}
	}
	t.Fatal("participant not found", name)
	return room.Participant{}
}

func TestAgentModelChosenAtInvitationAndChangedLater(t *testing.T) {
	t.Parallel()
	s, db, post := modelFixture(t, false)
	if code, body := post("participant.created", `{"connectionId":"w","name":"Builder","rank":1,"model":" claude-example-1 "}`); code != 200 {
		t.Fatal(code, body)
	}
	p := participantByName(t, s, "Builder")
	if p.Model != "claude-example-1" {
		t.Fatalf("model not stored trimmed: %q", p.Model)
	}
	p.ThreadID = "session-1"
	db.seed("participant.session", p)
	p = participantByName(t, s, "Builder")
	p.Model = "claude-example-2"
	payload := fmt.Sprintf(`{"id":%q,"connectionId":"w","name":"Builder","rank":1,"revision":%d,"model":"claude-example-2"}`, p.ID, p.Revision)
	if code, body := post("participant.updated", payload); code != 200 {
		t.Fatal(code, body)
	}
	got := participantByName(t, s, "Builder")
	if got.Model != "claude-example-2" || got.ThreadID != "" {
		t.Fatalf("a model change must start a fresh session: model=%q thread=%q", got.Model, got.ThreadID)
	}
}

func TestAgentModelRejectsUnsafeNames(t *testing.T) {
	t.Parallel()
	_, _, post := modelFixture(t, false)
	for _, bad := range []string{"--dangerously-skip-permissions x", "m;rm -rf", "-leading", "a b", strings.Repeat("a", 65), "$(id)"} {
		if code, _ := post("participant.created", fmt.Sprintf(`{"connectionId":"w","name":"X","rank":1,"model":%q}`, bad)); code != 400 {
			t.Fatalf("model %q accepted with %d", bad, code)
		}
	}
}

func TestAgentModelFlagOffBlocksChoicesButKeepsExisting(t *testing.T) {
	t.Parallel()
	s, _, post := modelFixture(t, true)
	if code, _ := post("participant.created", `{"connectionId":"w","name":"Blocked","rank":1,"model":"claude-example-1"}`); code != 400 {
		t.Fatal("a model choice was accepted while disabled", code)
	}
	if code, body := post("participant.created", `{"connectionId":"w","name":"Plain","rank":1}`); code != 200 {
		t.Fatal("agents without a model must still be addable", code, body)
	}
	if participantByName(t, s, "Plain").Model != "" {
		t.Fatal("unexpected model")
	}
}

func TestModelArgsAndFailureMessage(t *testing.T) {
	t.Parallel()
	if got := modelArgs("codex", "m-1"); !reflect.DeepEqual(got, []string{"-m", "m-1"}) {
		t.Fatal(got)
	}
	if got := modelArgs("claude", "m-1"); !reflect.DeepEqual(got, []string{"--model", "m-1"}) {
		t.Fatal(got)
	}
	if modelArgs("claude", "") != nil || modelArgs("claude", "bad name") != nil {
		t.Fatal("empty or unsafe model must add no arguments")
	}
	if err := explainModelFailure("m-1", fmt.Errorf("unknown model")); err == nil || !strings.Contains(err.Error(), "profile") {
		t.Fatal("model failure not explained", err)
	}
	if err := explainModelFailure("", fmt.Errorf("unknown model")); strings.Contains(err.Error(), "profile") {
		t.Fatal("default-model failure must not blame a model")
	}
}
