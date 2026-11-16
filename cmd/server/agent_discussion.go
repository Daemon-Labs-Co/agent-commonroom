package main

// Bounded agent discussions. After a human asks for one, the room's agents take
// turns replying to each other without a human prompt for every reply. Each turn
// is an ordinary managed turn: isolated worker, daily turn budget, context sharing
// policy and emergency stop all apply. Discussion text is never an approval.

import (
	"agentroom/internal/room"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// discussionRoundCeiling is the highest round limit an owner can use. Owners may
// lower it per room, never raise it. discussionRoundMaximum is the fixed safety
// bound that no configuration may exceed.
const (
	discussionRoundCeiling = 3
	discussionRoundMaximum = 10
)

// agentDiscussionEnabled defaults off; while off, agents reply only when a human,
// workflow or task dispatch starts their turn.
func agentDiscussionEnabled(st room.State) bool {
	return st.FeatureFlags["agentDiscussion"]
}

type discussionStarted struct {
	ID        string `json:"id"`
	RoomID    string `json:"roomId"`
	MessageID string `json:"messageId"`
	ActorID   string `json:"actorId"`
	Rounds    int    `json:"rounds"`
}

type discussionEnded struct {
	ID     string `json:"id"`
	RoomID string `json:"roomId"`
	Reason string `json:"reason"`
	Rounds int    `json:"rounds"`
	Turns  int    `json:"turns"`
}

type discussionLimit struct {
	RoomID string `json:"roomId"`
	Rounds int    `json:"rounds"`
}

// roomDiscussionRounds is the room's effective round limit.
func roomDiscussionRounds(st room.State, roomID string) int {
	for _, r := range st.Rooms {
		if r.ID == roomID && r.DiscussionRounds >= 1 && r.DiscussionRounds < discussionRoundCeiling {
			return r.DiscussionRounds
		}
	}
	return discussionRoundCeiling
}

func activeDiscussion(st room.State, roomID string) string {
	for _, r := range st.Rooms {
		if r.ID == roomID {
			return r.DiscussionID
		}
	}
	return ""
}

func discussionInstruction(round, rounds int, topic string) string {
	return fmt.Sprintf("Discussion round %d of %d between the agents in this room. Topic from the human: %s\nRead the other agents' latest replies above and respond to them directly and briefly: say what you agree with, challenge what you think is wrong and give reasons. Begin your first line with [AGREE] if you accept the current direction, or [ESCALATE] followed by the question if the disagreement needs a decision from a human. Otherwise begin with no marker and continue the discussion. No marker grants approval or permission.", round, rounds, topic)
}

func discussionNotice(roomID, text string) room.Message {
	return room.Message{ID: id(), RoomID: roomID, ActorID: "commonroom-system", ActorType: "system", Author: "CommonRoom", Text: text, Created: time.Now().UTC()}
}

// endDiscussionLocked records the end once and tells the room why. Caller holds s.mu.
func (s *server) endDiscussionLocked(d discussionStarted, reason string, rounds, turns int) {
	delete(s.discussions, d.RoomID)
	st, err := s.load(context.Background())
	if err != nil || activeDiscussion(st, d.RoomID) != d.ID {
		return
	}
	notice := discussionNotice(d.RoomID, fmt.Sprintf("Discussion ended after %d turn(s): %s.", turns, reason))
	_ = s.appendBatch(context.Background(), []auditChange{
		{Kind: "room.discussion.ended", Payload: discussionEnded{ID: d.ID, RoomID: d.RoomID, Reason: reason, Rounds: rounds, Turns: turns}},
		{Kind: "message.created", Payload: notice},
	})
}

// runDiscussion drives the turns sequentially so each agent sees earlier replies.
func (s *server) runDiscussion(ctx context.Context, d discussionStarted, topic string, order []string) {
	turns, round := 0, 0
	reason := "the round limit was reached"
	defer func() {
		s.mu.Lock()
		s.endDiscussionLocked(d, reason, round, turns)
		s.mu.Unlock()
		if reason == "every agent agreed" {
			s.suggestAfterAgreement(ctx, d.RoomID)
		}
	}()
	for round = 1; round <= d.Rounds; round++ {
		agreed := 0
		for _, pid := range order {
			stance, stop := s.discussionTurn(ctx, d, topic, pid, round)
			if stop != "" {
				reason = stop
				return
			}
			turns++
			if stance == "escalate" {
				reason = "an agent escalated to a human"
				return
			}
			if stance == "agree" {
				agreed++
			}
		}
		if agreed == len(order) {
			reason = "every agent agreed"
			return
		}
	}
	round = d.Rounds
}

// discussionTurn runs one agent's turn and reports its stance, or why the
// discussion must stop.
func (s *server) discussionTurn(ctx context.Context, d discussionStarted, topic, pid string, round int) (stance, stop string) {
	text, stop := s.agentInitiatedTurn(ctx, d.RoomID, pid, discussionInstruction(round, d.Rounds, topic), "Taking part in a room discussion", func(all, st room.State) string {
		switch {
		case activeDiscussion(all, d.RoomID) != d.ID:
			return "it was replaced or closed"
		case !agentDiscussionEnabled(all):
			return "agent discussions were turned off"
		case humanPostedAfter(st, d.MessageID):
			return "a human posted in the room"
		}
		return ""
	})
	if stop != "" {
		return "", stop
	}
	stance, _, _ = stanceMarker(text)
	return stance, ""
}

// agentInitiatedTurn runs one managed turn that no human message directly asked
// for. It applies every check a human-started turn gets (emergency stop, admission
// and daily limits, isolation, context size) plus the caller's own precheck, and
// returns the agent's reply or the reason no turn ran.
func (s *server) agentInitiatedTurn(ctx context.Context, roomID, pid, instruction, activity string, precheck func(all, st room.State) string) (text, stop string) {
	s.mu.Lock()
	if ctx.Err() != nil {
		s.mu.Unlock()
		return "", "a human stopped it"
	}
	all, err := s.load(context.Background())
	if err != nil {
		s.mu.Unlock()
		return "", "room state was unavailable"
	}
	st := scoped(all, roomID)
	if all.Incident || s.agentsBlockedLocked(all, roomID) {
		stop = "agents were stopped"
	} else if precheck != nil {
		stop = precheck(all, st)
	}
	var p room.Participant
	endpoint := ""
	if stop == "" {
		for _, candidate := range st.Participants {
			if candidate.ID == pid && candidate.RoomID == roomID {
				p = candidate
			}
		}
		for _, c := range st.Connections {
			if c.ID == p.ConnectionID && room.SupportsAgent(c) {
				endpoint = c.Endpoint
			}
		}
		switch {
		case p.ID == "" || endpoint == "":
			stop = "an agent left the room or lost its connection"
		case s.active[p.ID]:
			stop = p.Name + " was already working"
		default:
			if e := s.agentAdmission(all, roomID, 1); e != nil {
				stop = e.Error()
			} else if e := isolationAvailable(endpoint); e != nil {
				stop = e.Error()
			}
		}
	}
	prompt := ""
	if stop == "" {
		if prompt, err = chatPrompt(st, p, instruction); err != nil {
			stop = err.Error()
		}
	}
	if stop != "" {
		s.mu.Unlock()
		return "", stop
	}
	before := len(st.Messages)
	p.Status = "Working"
	if err = s.appendBatch(context.Background(), []auditChange{
		{Kind: "agent.activity", Payload: map[string]any{"participantId": p.ID, "activity": activity, "startedAt": time.Now().UTC()}},
		{Kind: "participant.session", Payload: p},
	}); err != nil {
		s.mu.Unlock()
		return "", "the turn could not be recorded"
	}
	s.active[p.ID] = true
	turnCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	s.cancels[p.ID] = cancel
	s.mu.Unlock()

	s.runAgent(turnCtx, cancel, p, prompt, endpoint)

	s.mu.Lock()
	defer s.mu.Unlock()
	if ctx.Err() != nil {
		return "", "a human stopped it"
	}
	after, err := s.load(context.Background())
	if err != nil {
		return "", "room state was unavailable"
	}
	after = scoped(after, roomID)
	for _, m := range after.Messages[min(before, len(after.Messages)):] {
		if m.ActorType == "agent" && m.Author == p.Name {
			return m.Text, ""
		}
	}
	return "", p.Name + " did not reply"
}

func humanPostedAfter(st room.State, messageID string) bool {
	seen := false
	for _, m := range st.Messages {
		if m.ID == messageID {
			seen = true
			continue
		}
		if seen && m.ActorType == "human" {
			return true
		}
	}
	return false
}

// discussionTopic trims and bounds the human's topic.
func discussionTopic(text string) string {
	return strings.TrimSpace(text)
}

// startDiscussionLocked validates and records a human's request, then starts the
// turn loop. Caller holds s.mu and has passed the integrity and stop checks.
func (s *server) startDiscussionLocked(w http.ResponseWriter, r *http.Request, all, st room.State, roomID, text string, rounds int) {
	if !agentDiscussionEnabled(all) {
		http.Error(w, "Agent discussions are disabled", 403)
		return
	}
	human, ok := member(all, requestActor(r))
	if !ok || !human.Enabled || human.Kind != "human" {
		http.Error(w, "Only a human in this room can start a discussion", 403)
		return
	}
	topic := discussionTopic(text)
	if topic == "" {
		http.Error(w, "A topic is required", 400)
		return
	}
	limit := roomDiscussionRounds(all, roomID)
	if rounds == 0 {
		rounds = limit
	}
	if rounds < 1 || rounds > limit {
		http.Error(w, fmt.Sprintf("Rounds must be between 1 and %d for this room", limit), 400)
		return
	}
	changes := []auditChange{}
	if stale := activeDiscussion(all, roomID); stale != "" {
		if s.discussions[roomID] != nil {
			http.Error(w, "A discussion is already running in this room", 409)
			return
		}
		changes = append(changes, auditChange{Kind: "room.discussion.ended", Payload: discussionEnded{ID: stale, RoomID: roomID, Reason: "interrupted by a restart"}})
	}
	order := []string{}
	var first room.Participant
	for _, p := range st.Participants {
		if p.RoomID != roomID || s.active[p.ID] {
			continue
		}
		for _, c := range st.Connections {
			if c.ID == p.ConnectionID && room.SupportsAgent(c) && isolationAvailable(c.Endpoint) == nil {
				if len(order) == 0 {
					first = p
				}
				order = append(order, p.ID)
			}
		}
	}
	if len(order) < 2 {
		http.Error(w, "A discussion needs at least two available agents in this room", 409)
		return
	}
	if e := s.agentAdmission(all, roomID, 1); e != nil {
		http.Error(w, e.Error(), http.StatusTooManyRequests)
		return
	}
	if _, e := chatPrompt(st, first, discussionInstruction(1, rounds, topic)); e != nil {
		http.Error(w, e.Error(), 413)
		return
	}
	m := room.Message{RoomID: roomID, ID: id(), ActorID: human.ID, ActorType: "human", Author: human.Name, Text: "@room (discussion) " + topic, Created: time.Now().UTC()}
	d := discussionStarted{ID: id(), RoomID: roomID, MessageID: m.ID, ActorID: human.ID, Rounds: rounds}
	changes = append(changes, auditChange{Kind: "message.created", Payload: m}, auditChange{Kind: "room.discussion.started", Payload: d})
	response := map[string]any{"discussionId": d.ID, "rounds": rounds, "agents": len(order)}
	if err := s.appendBatch(withOperationResponse(r.Context(), 202, response), changes); err != nil {
		operationError(w, err)
		return
	}
	ctx, cancel := context.WithCancel(detachedInitiatorContext(r.Context()))
	if s.discussions == nil {
		s.discussions = map[string]context.CancelFunc{}
	}
	s.discussions[roomID] = cancel
	s.inFlight.Add(1)
	go func() {
		defer s.inFlight.Add(-1)
		defer cancel()
		s.runDiscussion(ctx, d, topic, order)
	}()
	w.WriteHeader(http.StatusAccepted)
	reply(w, response)
}

// stopDiscussionLocked lets any human in the room end the discussion. It works
// during an incident or emergency stop. Caller holds s.mu.
func (s *server) stopDiscussionLocked(w http.ResponseWriter, r *http.Request, st room.State, roomID string) {
	human, ok := member(st, requestActor(r))
	if !ok || !human.Enabled || human.Kind != "human" {
		http.Error(w, "Only a human in this room can stop a discussion", 403)
		return
	}
	current := activeDiscussion(st, roomID)
	if current == "" {
		http.Error(w, "No discussion is running in this room", 409)
		return
	}
	if cancel := s.discussions[roomID]; cancel != nil {
		cancel()
		for _, p := range st.Participants {
			if p.RoomID == roomID && s.active[p.ID] && s.cancels[p.ID] != nil {
				s.cancels[p.ID]()
			}
		}
	} else if err := s.appendBatch(r.Context(), []auditChange{
		{Kind: "room.discussion.ended", Payload: discussionEnded{ID: current, RoomID: roomID, Reason: "stopped after a restart"}},
		{Kind: "message.created", Payload: discussionNotice(roomID, "Discussion ended: stopped by a human after a restart.")},
	}); err != nil {
		operationError(w, err)
		return
	}
	reply(w, map[string]bool{"stopping": true})
}
