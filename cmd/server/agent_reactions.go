package main

// Reactions to an agent's own reply: escalating to the agent it reports to, and
// starting a discussion. These are the only turns an agent's text can start, so
// they are tightly bounded: each needs the agentDiscussion flag, passes every
// check a human-started turn gets, and a room allows only a small number of them
// until a human posts again. Nothing here approves, dispatches or accepts work.

import (
	"agentroom/internal/room"
	"context"
	"fmt"
)

// maxAgentInitiatedPerHumanMessage bounds agent-started activity between human
// messages, so agents cannot keep each other running.
const maxAgentInitiatedPerHumanMessage = 2

type agentInitiated struct {
	RoomID        string `json:"roomId"`
	Kind          string `json:"kind"`
	ParticipantID string `json:"participantId"`
	TargetID      string `json:"targetId,omitempty"`
}

func participantActor(p room.Participant) string {
	if p.ActorID != "" {
		return p.ActorID
	}
	return "bot-" + p.ConnectionID
}

// managerOf returns the actor an actor reports to directly, or "".
func managerOf(chart room.OrganizationChart, actor string) string {
	for _, n := range chart.Nodes {
		if n.ActorID == actor {
			return n.ReportsTo
		}
	}
	return ""
}

func roomAgentInitiated(st room.State, roomID string) int {
	for _, r := range st.Rooms {
		if r.ID == roomID {
			return r.AgentInitiated
		}
	}
	return 0
}

func escalationInstruction(from, question string) string {
	return fmt.Sprintf("%s reports to you and escalated this to you for a decision: %s\nAnswer once, briefly, with your decision and reason. You cannot approve, accept or publish work. If it needs a human decision, begin your first line with [HUMAN ACTION REQUIRED] and say what you need.", from, question)
}

// reactToAgentReply runs after an agent's chat or discussion reply is committed.
func (s *server) reactToAgentReply(ctx context.Context, p room.Participant, text string) {
	stance, _, detail := stanceMarker(text)
	if stance != "escalate" && stance != "discuss" {
		return
	}
	base := detachedInitiatorContext(ctx)
	s.mu.Lock()
	all, err := s.load(context.Background())
	if err != nil || !agentDiscussionEnabled(all) || all.Incident || s.agentsBlockedLocked(all, p.RoomID) || roomAgentInitiated(all, p.RoomID) >= maxAgentInitiatedPerHumanMessage {
		s.mu.Unlock()
		return
	}
	st := scoped(all, p.RoomID)
	committed := false
	for i := len(st.Messages) - 1; i >= 0 && !committed; i-- {
		if m := st.Messages[i]; m.ActorType == "agent" && m.Author == p.Name {
			committed = m.Text == text
			break
		}
	}
	if !committed {
		s.mu.Unlock()
		return
	}
	if stance == "escalate" {
		manager := managerOf(all.Organization, participantActor(p))
		target := ""
		for _, candidate := range st.Participants {
			if candidate.RoomID == p.RoomID && candidate.ID != p.ID && manager != "" && participantActor(candidate) == manager {
				target = candidate.ID
			}
		}
		if target == "" {
			s.mu.Unlock()
			return // A human manager, or none: the human-attention item covers it.
		}
		if s.appendBatch(context.Background(), []auditChange{{Kind: "agent.initiated", Payload: agentInitiated{RoomID: p.RoomID, Kind: "escalation", ParticipantID: p.ID, TargetID: target}}}) != nil {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
		s.agentInitiatedTurn(base, p.RoomID, target, escalationInstruction(p.Name, detail), "Answering an escalation from "+p.Name, nil)
		return
	}
	// [DISCUSS]: ignored inside a running discussion and without a topic.
	if detail == "" || activeDiscussion(all, p.RoomID) != "" || s.discussions[p.RoomID] != nil {
		s.mu.Unlock()
		return
	}
	order := []string{}
	for _, candidate := range st.Participants {
		if candidate.RoomID != p.RoomID || candidate.ID == p.ID || s.active[candidate.ID] {
			continue
		}
		for _, c := range st.Connections {
			if c.ID == candidate.ConnectionID && room.SupportsAgent(c) && isolationAvailable(c.Endpoint) == nil {
				order = append(order, candidate.ID)
			}
		}
	}
	if len(order) == 0 {
		s.mu.Unlock()
		return
	}
	order = append(order, p.ID)
	notice := discussionNotice(p.RoomID, p.Name+" started a discussion: "+detail)
	d := discussionStarted{ID: id(), RoomID: p.RoomID, MessageID: notice.ID, ActorID: participantActor(p), Rounds: roomDiscussionRounds(all, p.RoomID)}
	if s.appendBatch(context.Background(), []auditChange{
		{Kind: "agent.initiated", Payload: agentInitiated{RoomID: p.RoomID, Kind: "discussion", ParticipantID: p.ID}},
		{Kind: "message.created", Payload: notice},
		{Kind: "room.discussion.started", Payload: d},
	}) != nil {
		s.mu.Unlock()
		return
	}
	dctx, cancel := context.WithCancel(base)
	if s.discussions == nil {
		s.discussions = map[string]context.CancelFunc{}
	}
	s.discussions[p.RoomID] = cancel
	s.mu.Unlock()
	defer cancel()
	s.runDiscussion(dctx, d, detail, order)
}

// projectEscalations makes every [ESCALATE] reply a human-attention item while
// stances or discussions are enabled, so a human always sees an escalation even
// when an agent manager answers it first. Acknowledgment uses the existing flow.
func projectEscalations(st *room.State, acknowledged map[string]string) {
	if !humanAttentionEnabled(*st) || (!agentObjectionsEnabled(*st) && !agentDiscussionEnabled(*st)) {
		return
	}
	for i := range st.Messages {
		m := &st.Messages[i]
		if m.ActorType != "agent" || m.HumanRequest != "" {
			continue
		}
		if stance, _, _ := stanceMarker(m.Text); stance != "escalate" {
			continue
		}
		m.HumanRequest = "explicit"
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
