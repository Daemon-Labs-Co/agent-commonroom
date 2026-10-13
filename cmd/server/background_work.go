package main

import (
	"agentroom/internal/room"
	"context"
	"time"
)

type backgroundCallbackKey struct{}
type backgroundCallback func(string, string)

var backgroundLabels = map[string]string{
	"preparing": "Preparing isolated execution", "starting": "Worker is starting the job",
	"running": "Worker reports the job running", "finished": "Worker reports the job finished",
	"failed": "Worker reports failure", "cancelled": "Worker reports cancellation",
	"unreachable": "Worker status could not be confirmed", "shutdown_confirmed": "Worker confirmed container shutdown",
	"shutdown_unconfirmed": "Worker shutdown remains unconfirmed", "command_started": "Agent started a command",
	"command_finished": "Agent finished a command", "tool_started": "Agent started a tool operation",
	"tool_finished": "Agent finished a tool operation", "completed": "Controller received a completed response",
	"turn_failed": "Controller received a failed response", "timed_out": "Turn reached its hard time limit",
	"stop_requested": "Turn cancellation was requested",
}

func backgroundEnabled(st room.State) bool {
	enabled, exists := st.FeatureFlags["agentBackgroundWork"]
	return !exists || enabled
}

func emitBackground(ctx context.Context, stage, job string) {
	if cb, ok := ctx.Value(backgroundCallbackKey{}).(backgroundCallback); ok {
		cb(stage, job)
	}
}

func codexBackgroundStage(eventType, itemType string) string {
	phase := ""
	switch eventType {
	case "item.started":
		phase = "started"
	case "item.completed":
		phase = "finished"
	default:
		return ""
	}
	switch itemType {
	case "command_execution":
		return "command_" + phase
	case "mcp_tool_call", "web_search":
		return "tool_" + phase
	}
	return ""
}

// backgroundContext accepts only fixed operational labels. Model text, commands,
// outputs and credentials must never become telemetry. This is observability,
// not proof of useful progress or authority to extend execution deadlines.
func (s *server) backgroundContext(ctx context.Context, p room.Participant) (context.Context, func(error)) {
	s.mu.Lock()
	st, err := s.load(ctx)
	enabled := err == nil && backgroundEnabled(st)
	s.mu.Unlock()
	if !enabled {
		return ctx, func(error) {}
	}
	turn := id()
	count := 0
	closed := false
	stoppedJobs := map[string]bool{}
	callback := backgroundCallback(func(stage, job string) {
		label, ok := backgroundLabels[stage]
		if !ok || (job != "" && !validWorkerJobID(job)) {
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		final := stage == "completed" || stage == "turn_failed" || stage == "timed_out" || stage == "stop_requested"
		if closed || count >= 64 || (count >= 60 && !final && stage != "shutdown_confirmed" && stage != "shutdown_unconfirmed") {
			return
		}
		if job != "" && stoppedJobs[job] && stage != "shutdown_confirmed" && stage != "shutdown_unconfirmed" {
			return
		}
		if stage == "shutdown_confirmed" {
			stoppedJobs[job] = true
		}
		if final {
			closed = true
		}
		latest, e := s.load(context.Background())
		if e != nil || !backgroundEnabled(latest) {
			return
		}
		found := false
		for _, current := range latest.Participants {
			if current.ID == p.ID && current.RoomID == p.RoomID && current.ConnectionID == p.ConnectionID {
				found = true
			}
		}
		if !found {
			return
		}
		count++
		_ = s.append(context.Background(), "agent.background", room.BackgroundActivity{ParticipantID: p.ID, RoomID: p.RoomID, TurnID: turn, JobID: job, Stage: stage, Label: label, ObservedAt: time.Now().UTC()})
	})
	ctx = context.WithValue(ctx, backgroundCallbackKey{}, callback)
	emitBackground(ctx, "preparing", "")
	return ctx, func(err error) {
		stage := "completed"
		if ctx.Err() == context.DeadlineExceeded {
			stage = "timed_out"
		} else if ctx.Err() == context.Canceled {
			stage = "stop_requested"
		} else if err != nil {
			stage = "turn_failed"
		}
		emitBackground(ctx, stage, "")
	}
}

// projectBackground bounds delivery; complete retained audit events remain in
// History. Turning this feature off omits timelines without changing execution.
func projectBackground(st *room.State, events map[string][]room.BackgroundActivity) {
	for i := range st.Participants {
		p := &st.Participants[i]
		p.Background = nil
		if !backgroundEnabled(*st) {
			continue
		}
		for _, event := range events[p.ID] {
			if event.RoomID == p.RoomID {
				p.Background = append(p.Background, event)
			}
		}
	}
}
