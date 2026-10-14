package main

import (
	"agentroom/internal/room"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// safeAgentFailure uses allowlisted explanations rather than displaying raw
// stderr, which can contain credentials, private paths or prompt material.
func safeAgentFailure(raw string) (reason, next string) {
	text := strings.ToLower(raw)
	switch {
	case strings.Contains(text, "credential-separated"):
		return "Development is blocked because native tools share provider credentials.", "The administrator must install and verify the separate provider/tool bridge. Signing in again or disabling isolation will not fix this."
	case strings.Contains(text, "deadline exceeded"), strings.Contains(text, "timed out"):
		return "The agent did not finish before its time limit.", "Inspect partial work, then retry or split the task into smaller steps."
	case strings.Contains(text, "signal: killed"):
		return "The agent process was killed; the recorded error does not identify the cause.", "Check the turn duration and worker resource limits. Older records may omit a timeout cause; inspect partial work before retrying."
	case strings.Contains(text, "concurrency limit"):
		return "Execution capacity is currently full.", "Wait for active turns to finish, then explicitly retry the message or resume the workflow."
	case strings.Contains(text, "bwrap"), strings.Contains(text, "new namespace"):
		return "The execution sandbox could not create its required namespace.", "Ask the administrator to verify the isolated development worker; keep isolation enabled."
	case strings.Contains(text, "read-only"), strings.Contains(text, "read only"):
		return "The requested operation needs a writable workspace.", "Verify that this room has an approved writable development workspace."
	case strings.Contains(text, "proxy refused"), strings.Contains(text, "connection refused"), strings.Contains(text, "gateway"):
		return "The provider connection or isolated network gateway was unavailable.", "Test the connection and check worker/network readiness."
	case strings.Contains(text, "sign in"), strings.Contains(text, "login"), strings.Contains(text, "unauthorized"), strings.Contains(text, "credential"):
		return "Provider authentication or credentials were unavailable.", "The connection owner should check sign-in and test the connection."
	case strings.Contains(text, "docker"), strings.Contains(text, "isolated worker"), strings.Contains(text, "development workspace"), strings.Contains(text, "repository"):
		return "The isolated worker or approved development workspace was not ready.", "Ask the administrator to check the room's worker and repository configuration."
	case strings.Contains(text, "explicit passed/changes"):
		return "The reviewer did not return the required explicit verdict.", "Inspect the workflow review output and resume only after correcting its instructions."
	case strings.Contains(text, "iteration limit"):
		return "The workflow reached its iteration limit.", "Inspect the remaining issues; cancel and revise the workflow."
	case strings.Contains(text, "output exceeds"), strings.Contains(text, "context exceeds"), strings.Contains(text, "output exceeded"):
		return "The step exceeded its output or context limit.", "Request a shorter report or split the workflow."
	case strings.Contains(text, "cancelled"), strings.Contains(text, "emergency stop"), strings.Contains(text, "feature disabled"):
		return "Execution was stopped or disabled.", "The owner must resolve the stop before explicitly resuming work."
	case strings.Contains(text, "integrity"):
		return "Audit integrity could not be verified.", "The owner must investigate history integrity before allowing further execution."
	default:
		return "The last agent turn failed; a specific cause was not recognized.", "Ask the owner to inspect the corresponding agent error in History and test the connection."
	}
}

// projectAgentElapsed reports controller dispatch age, not process health or
// useful progress. No heartbeat may extend the hard execution budget.
func projectAgentElapsed(st *room.State, starts map[string]time.Time, now time.Time) {
	for i := range st.Participants {
		p := &st.Participants[i]
		start := starts[p.ID]
		if p.Status != "Working" || p.StatusDetails == nil || start.IsZero() || start.After(now) {
			continue
		}
		elapsed := int(now.Sub(start).Seconds())
		p.StatusDetails.ElapsedSeconds = elapsed
		if elapsed >= 120 && !st.Incident && !st.AgentsStopped && !containsRoom(st.AgentStopRooms, p.RoomID) {
			p.StatusDetails.LongRunning = true
			p.StatusDetails.Reason = fmt.Sprintf("This turn has been running for %d seconds since dispatch; process liveness and progress are not confirmed.", elapsed)
			p.StatusDetails.Next = "Check partial work and the worker. The hard time limit still applies; use emergency stop if necessary."
		}
	}
}

// statusActivity bounds operator-controlled workflow titles for safe display.
func statusActivity(text string) string {
	text = strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, text))
	chars := []rune(text)
	if len(chars) > 180 {
		text = string(chars[:180]) + "…"
	}
	return text
}

// projectAgentStatusDetails derives current explanations from retained events.
// Disabled delivery omits details server-side; historical audit records remain.
func projectAgentStatusDetails(st *room.State, activities, failures map[string]string) {
	enabled, exists := st.FeatureFlags["agentStatusDetails"]
	for i := range st.Participants {
		p := &st.Participants[i]
		p.StatusDetails = nil
		if exists && !enabled {
			continue
		}
		d := &room.AgentStatusDetails{Activity: statusActivity(activities[p.ID])}
		if d.Activity == "" {
			d.Activity = "Last agent turn (activity was not recorded)"
		}
		switch {
		case st.Incident:
			d.Reason = "Agent execution is blocked by an integrity incident."
			d.Next = "The owner must investigate integrity and authorize recovery."
		case st.AgentsStopped || containsRoom(st.AgentStopRooms, p.RoomID):
			d.Reason = "Emergency stop is active."
			d.Next = "The owner must explicitly resume execution after investigating."
		case p.Status == "Blocked":
			if failures[p.ID] == "" {
				d.Reason = "This agent is blocked, but no detailed failure was recorded."
				d.Next = "Ask the owner to inspect History and test the connection."
			} else {
				d.Reason, d.Next = safeAgentFailure(failures[p.ID])
			}
		case p.Status == "Working":
			d.Reason = "The agent is currently working."
			d.Next = "Wait for its result, or use emergency stop if necessary."
		case p.Status == "Stopped":
			d.Reason = "The last turn was stopped."
			d.Next = "Inspect partial work before sending another request."
		default:
			d.Reason = "No active agent turn is recorded."
			d.Next = "Send a message or assign a workflow step."
		}
		p.StatusDetails = d
	}
}
func containsRoom(rooms []string, id string) bool {
	for _, room := range rooms {
		if room == id {
			return true
		}
	}
	return false
}
