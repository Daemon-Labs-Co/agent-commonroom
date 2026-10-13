package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestBackgroundWorkerScopeAndUntrustedStatus(t *testing.T) {
	t.Parallel()
	job := strings.Repeat("a", 32)
	for _, tc := range []struct {
		record workerJobRecord
		want   string
	}{
		{workerJobRecord{JobID: job, Provider: "codex", Status: "running"}, "running"},
		{workerJobRecord{JobID: "foreign", Provider: "codex", Status: "running"}, "unreachable"},
		{workerJobRecord{JobID: job, Provider: "claude", Status: "running"}, "unreachable"},
		{workerJobRecord{JobID: job, Provider: "codex", Status: "secret-command"}, "unreachable"},
	} {
		raw, _ := json.Marshal(tc.record)
		if got := backgroundWorkerStage(raw, job, "codex"); got != tc.want {
			t.Fatal(got)
		}
	}
	if backgroundWorkerStage([]byte(strings.Repeat("x", 9000)), job, "codex") != "unreachable" {
		t.Fatal("unbounded worker response")
	}
}

func TestCodexTelemetryOnlyMapsOperationalEnums(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ event, item, want string }{
		{"item.started", "command_execution", "command_started"},
		{"item.completed", "command_execution", "command_finished"},
		{"item.started", "web_search", "tool_started"},
		{"item.completed", "reasoning", ""},
		{"item.completed", "agent_message", ""},
		{"secret command", "command_execution", ""},
	} {
		if got := codexBackgroundStage(tc.event, tc.item); got != tc.want {
			t.Fatal(tc, got)
		}
	}
}

func TestBackgroundLifecycleBoundedAndFlagEnforced(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	p := room.Participant{ID: "p", RoomID: "engineering", ConnectionID: "c"}
	db.seed("participant.created", p)
	ctx, finish := s.backgroundContext(context.Background(), p)
	emitBackground(ctx, "secret command token=synthetic", "")
	job := strings.Repeat("a", 32)
	emitBackground(ctx, "running", job)
	emitBackground(ctx, "shutdown_confirmed", job)
	emitBackground(ctx, "running", job) // Late worker response must not undo shutdown.
	finish(nil)
	emitBackground(ctx, "running", job)
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Participants[0].Background) != 4 {
		t.Fatal(st.Participants[0].Background)
	}
	db.seed("feature.flag.updated", map[string]any{"name": "agentBackgroundWork", "enabled": false})
	st, _ = s.load(context.Background())
	if len(st.Participants[0].Background) != 0 {
		t.Fatal("flag off exposed telemetry")
	}
	before := len(db.snapshot())
	ctx2, finish2 := s.backgroundContext(context.Background(), p)
	emitBackground(ctx2, "running", job)
	finish2(nil)
	if len(db.snapshot()) != before {
		t.Fatal("flag off collected telemetry")
	}
}

func TestPublicCallerCannotForgeBackgroundTelemetry(t *testing.T) {
	t.Parallel()
	_, db, h := mockServer(t)
	before := len(db.snapshot())
	w := mockRequest(h, "POST", "/api/events", `{"kind":"agent.background","payload":{"participantId":"p","roomId":"engineering","stage":"running","label":"fake"}}`)
	if w.Code < 400 || len(db.snapshot()) != before {
		t.Fatal("public telemetry accepted", w.Code)
	}
}

func TestBackgroundProjectionIsNotDuplicatedInProfileAudit(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	p := room.Participant{ID: "p", RoomID: "engineering", Background: []room.BackgroundActivity{{Stage: "running"}}}
	if err := s.append(context.Background(), "participant.session", p); err != nil {
		t.Fatal(err)
	}
	for _, event := range db.snapshot() {
		if strings.Contains(string(event.payload), "background") {
			t.Fatal("telemetry duplicated in profile")
		}
	}
}

func TestBackgroundProgressCannotFloodOrExposeOtherRooms(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	p := room.Participant{ID: "p", RoomID: "engineering"}
	db.seed("participant.created", p)
	ctx, finish := s.backgroundContext(context.Background(), p)
	for i := 0; i < 1000; i++ {
		emitBackground(ctx, "command_started", "")
	}
	finish(nil)
	st, _ := s.load(context.Background())
	events := st.Participants[0].Background
	if len(events) > 64 || events[len(events)-1].Stage != "completed" {
		t.Fatal("unbounded progress or terminal lost")
	}
	db.seed("agent.background", room.BackgroundActivity{ParticipantID: "p", RoomID: "other", Stage: "running", Label: "private raw command"})
	st, _ = s.load(context.Background())
	for _, event := range st.Participants[0].Background {
		if event.RoomID != "engineering" || strings.Contains(event.Label, "private") {
			t.Fatal("foreign telemetry delivered")
		}
	}
}
