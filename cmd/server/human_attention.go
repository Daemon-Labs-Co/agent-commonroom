package main

import (
	"agentroom/internal/room"
	"strings"
)

// humanAttentionEnabled controls detection, projection and acknowledgment, never
// authorization or workflow acceptance. Existing requests remain in history.
func humanAttentionEnabled(st room.State) bool {
	enabled, exists := st.FeatureFlags["humanAttention"]
	return !exists || enabled
}

// humanRequestKind honors an explicit first-line protocol; conservative phrases
// are suggestions for older free-text responses, not authoritative decisions.
func humanRequestKind(text string) string {
	first := strings.SplitN(strings.TrimSpace(text), "\n", 2)[0]
	if strings.HasPrefix(strings.ToUpper(first), "[HUMAN ACTION REQUIRED]") {
		return "explicit"
	}
	lower := strings.ToLower(text)
	for _, phrase := range []string{"please approve", "please confirm", "please provide", "please choose", "please decide", "please clarify", "i need your approval", "i need you to", "awaiting your approval", "waiting for your decision", "human approval is required"} {
		if strings.Contains(lower, phrase) {
			return "suggested"
		}
	}
	if strings.Contains(lower, "to proceed,") || strings.Contains(lower, "to unblock implementation,") {
		for _, phrase := range []string{"provide", "identify", "confirm", "approve", "assign"} {
			if strings.Contains(lower, phrase) {
				return "suggested"
			}
		}
	}
	return ""
}

// projectHumanAttention adds display metadata without rewriting message history.
func projectHumanAttention(st *room.State, acknowledged map[string]string) {
	for i := range st.Rooms {
		st.Rooms[i].PendingHumanRequests = 0
	}
	for i := range st.Messages {
		m := &st.Messages[i]
		m.HumanRequest = ""
		m.AttentionAcknowledgedBy = ""
		if !humanAttentionEnabled(*st) || m.ActorType != "agent" {
			continue
		}
		m.HumanRequest = humanRequestKind(m.Text)
		if m.HumanRequest != "" {
			m.AttentionAcknowledgedBy = acknowledged[m.ID]
			if m.AttentionAcknowledgedBy == "" {
				for j := range st.Rooms {
					if st.Rooms[j].ID == m.RoomID {
						st.Rooms[j].PendingHumanRequests++
					}
				}
			}
		}
	}
}

const humanAttentionGuidance = "\nIf your response requires a human decision, approval, clarification or missing information, begin its first line with [HUMAN ACTION REQUIRED] and state the requested action clearly. Do not use this marker for routine updates or questions directed only to another agent. The marker does not authorize work or approve a task.\n"
