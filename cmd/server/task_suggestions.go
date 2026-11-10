package main

// Task suggestions from agents. A suggestion is derived from an agent's own
// message and is only a prompt for a human: capturing it goes through the existing
// owner-only proposal endpoint, so the rule that only an authenticated human
// creates proposal and task facts is unchanged.

import (
	"agentroom/internal/room"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const maxSuggestionsPerMessage = 5

// taskSuggestionsEnabled defaults off; while off no suggestions are derived and
// no agent is asked for any.
func taskSuggestionsEnabled(st room.State) bool {
	return st.FeatureFlags["governingSuggestions"]
}

type suggestionDismissed struct {
	ID      string `json:"id"`
	RoomID  string `json:"roomId"`
	ActorID string `json:"actorId"`
}

// parseTaskSuggestions reads up to five "[PROPOSE TASK] title :: criteria ::
// assignee" lines from one agent message. Criteria and assignee are optional.
func parseTaskSuggestions(m room.Message) []room.TaskSuggestion {
	out := []room.TaskSuggestion{}
	for _, line := range strings.Split(m.Text, "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 14 || !strings.EqualFold(line[:14], "[PROPOSE TASK]") {
			continue
		}
		parts := strings.SplitN(line[14:], "::", 3)
		title := strings.TrimSpace(parts[0])
		if title == "" || len(title) > 200 {
			continue
		}
		s := room.TaskSuggestion{ID: fmt.Sprintf("%s#%d", m.ID, len(out)+1), RoomID: m.RoomID, MessageID: m.ID, ActorID: m.ActorID, Author: m.Author, Title: title, Status: "open"}
		if len(parts) > 1 {
			if s.Criteria = strings.TrimSpace(parts[1]); len(s.Criteria) > 4000 {
				s.Criteria = s.Criteria[:4000]
			}
		}
		if len(parts) > 2 {
			s.Assignee = strings.TrimSpace(parts[2])
		}
		out = append(out, s)
		if len(out) == maxSuggestionsPerMessage {
			break
		}
	}
	return out
}

// projectTaskSuggestions derives suggestions and their status from stored messages.
func projectTaskSuggestions(st *room.State, dismissed map[string]bool) {
	st.TaskSuggestions = []room.TaskSuggestion{}
	if !taskSuggestionsEnabled(*st) {
		return
	}
	for _, m := range st.Messages {
		if m.ActorType != "agent" && m.ActorType != "extension" {
			continue
		}
		for _, s := range parseTaskSuggestions(m) {
			for _, p := range st.TaskProposals {
				if p.SourceMessageID == m.ID && strings.EqualFold(strings.TrimSpace(p.Title), s.Title) {
					s.Status = "captured"
				}
			}
			if s.Status == "open" && dismissed[s.ID] {
				s.Status = "dismissed"
			}
			st.TaskSuggestions = append(st.TaskSuggestions, s)
		}
	}
}

const suggestionInstruction = "Review the recent room conversation and the room's goal. Reply with at most five concrete tasks that the discussion has settled on and that are not already tasks. Put each on its own line in exactly this form:\n[PROPOSE TASK] short title :: acceptance criteria that can be checked :: name of the agent best placed to do it\nAdd one short sentence before the list if useful. Suggest nothing that was disputed or left open; if there is nothing ready, say so and propose no task. These are suggestions for a human to review; they do not create or approve work."

// suggestionAgent picks who is asked: the room's governing agent when one is
// appointed, otherwise the first available agent in the room.
func suggestionAgent(s *server, all room.State, roomID string) (room.Participant, bool) {
	if appointment := all.Governance[roomID]; currentCoordination(all, appointment) {
		for _, p := range all.Participants {
			if p.ID == appointment.ParticipantID && p.RoomID == roomID {
				return p, true
			}
		}
	}
	for _, p := range all.Participants {
		if p.RoomID != roomID {
			continue
		}
		for _, c := range all.Connections {
			if c.ID == p.ConnectionID && room.SupportsAgent(c) && isolationAvailable(c.Endpoint) == nil {
				return p, true
			}
		}
	}
	return room.Participant{}, false
}

// suggestTasksLocked handles a human's request for suggestions. Caller holds s.mu.
func (s *server) suggestTasksLocked(w http.ResponseWriter, r *http.Request, all room.State, roomID string) {
	if !taskSuggestionsEnabled(all) {
		http.Error(w, "Task suggestions are disabled", 403)
		return
	}
	human, ok := member(all, requestActor(r))
	if !ok || !human.Enabled || human.Kind != "human" {
		http.Error(w, "Only a human in this room can ask for task suggestions", 403)
		return
	}
	p, ok := suggestionAgent(s, all, roomID)
	if !ok {
		http.Error(w, "No agent in this room is available to suggest tasks", 409)
		return
	}
	if s.active[p.ID] {
		http.Error(w, p.Name+" is already working", 409)
		return
	}
	if e := s.agentAdmission(all, roomID, 1); e != nil {
		http.Error(w, e.Error(), http.StatusTooManyRequests)
		return
	}
	m := room.Message{RoomID: roomID, ID: id(), ActorID: human.ID, ActorType: "human", Author: human.Name, Text: "@" + p.Name + " Suggest tasks from this discussion.", Created: time.Now().UTC()}
	response := map[string]any{"participantId": p.ID, "name": p.Name}
	if err := s.appendBatch(withOperationResponse(r.Context(), 202, response), []auditChange{{Kind: "message.created", Payload: m}}); err != nil {
		operationError(w, err)
		return
	}
	ctx := detachedInitiatorContext(r.Context())
	go s.agentInitiatedTurn(ctx, roomID, p.ID, suggestionInstruction, "Suggesting tasks from the discussion", nil)
	w.WriteHeader(http.StatusAccepted)
	reply(w, response)
}

// suggestAfterAgreement asks for suggestions once when a discussion ends with
// every agent agreeing. It counts as agent-started activity for the loop guard.
func (s *server) suggestAfterAgreement(ctx context.Context, roomID string) {
	s.mu.Lock()
	all, err := s.load(context.Background())
	if err != nil || !taskSuggestionsEnabled(all) || all.Incident || s.agentsBlockedLocked(all, roomID) || roomAgentInitiated(all, roomID) >= maxAgentInitiatedPerHumanMessage {
		s.mu.Unlock()
		return
	}
	p, ok := suggestionAgent(s, all, roomID)
	if !ok || s.appendBatch(context.Background(), []auditChange{{Kind: "agent.initiated", Payload: agentInitiated{RoomID: roomID, Kind: "task suggestions", ParticipantID: p.ID}}}) != nil {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	s.agentInitiatedTurn(detachedInitiatorContext(ctx), roomID, p.ID, suggestionInstruction, "Suggesting tasks from the discussion", nil)
}
