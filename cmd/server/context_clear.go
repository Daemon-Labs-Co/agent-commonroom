package main

// Room context clearing: a non-destructive boundary. Application history stays
// visible to humans and in the audit log; agents stop receiving earlier messages
// and every room participant starts a fresh native provider session.

import (
	"agentroom/internal/room"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const maxContextClearReasonRunes = 500

// contextClearEnabled is a UX/product toggle. Disabling it only rejects new
// clears; boundaries already recorded keep being enforced in prompts.
func contextClearEnabled(st room.State) bool {
	return st.FeatureFlags["contextClear"] || !flagDefined(st, "contextClear")
}

// contextClearPayload is the audited room event. MessageID is the visible notice
// that becomes the new first message agents may see, so the stated reason reaches
// every agent exactly as if the human had said it to the whole room.
type contextClearPayload struct {
	RoomID    string `json:"roomId"`
	MessageID string `json:"messageId"`
	ActorID   string `json:"actorId"`
	Reason    string `json:"reason"`
}

// prepareContextClear validates a clear request. Any enabled human with access
// to the room may clear it (room access is enforced earlier by the actor guard);
// agent actors cannot, and a reason is mandatory.
func (s *server) prepareContextClear(w http.ResponseWriter, r *http.Request, st room.State, roomID string, rawReason string) (contextClearPayload, bool) {
	if !contextClearEnabled(st) {
		http.Error(w, "Clearing room context is disabled", 403)
		return contextClearPayload{}, false
	}
	actor, ok := member(st, requestActor(r))
	if !ok || !actor.Enabled || actor.Kind != "human" {
		http.Error(w, "Only a human in this room can clear its context", 403)
		return contextClearPayload{}, false
	}
	reason := strings.TrimSpace(rawReason)
	if reason == "" {
		http.Error(w, "A reason is required to clear context", 400)
		return contextClearPayload{}, false
	}
	if utf8.RuneCountInString(reason) > maxContextClearReasonRunes {
		http.Error(w, "Reason is too long (500 characters maximum)", 400)
		return contextClearPayload{}, false
	}
	for _, p := range st.Participants {
		if p.RoomID == roomID && s.active[p.ID] {
			http.Error(w, "Stop room agents before clearing context", 409)
			return contextClearPayload{}, false
		}
	}
	return contextClearPayload{RoomID: roomID, MessageID: id(), ActorID: actor.ID, Reason: reason}, true
}

// contextClearNotice is the human-visible message recording who cleared context
// and why.
func contextClearNotice(st room.State, c contextClearPayload) room.Message {
	actor, _ := member(st, c.ActorID)
	return room.Message{
		ID:        c.MessageID,
		RoomID:    c.RoomID,
		ActorID:   c.ActorID,
		ActorType: "human",
		Author:    actor.Name,
		Text:      "Cleared agent context. Reason: " + c.Reason,
		Created:   time.Now().UTC(),
	}
}

// contextClearBatch is the atomic change set: notice, boundary, session resets.
func contextClearBatch(st room.State, c contextClearPayload) []auditChange {
	changes := []auditChange{
		{Kind: "message.created", Payload: contextClearNotice(st, c)},
		{Kind: "room.context.cleared", Payload: c},
	}
	for _, p := range st.Participants {
		if p.RoomID == c.RoomID {
			p.ThreadID = ""
			changes = append(changes, auditChange{Kind: "participant.session", Payload: p})
		}
	}
	return changes
}

// messagesAfterBoundary returns the index of the first message agents may see.
// The boundary notice itself is included so the reason is visible.
func messagesAfterBoundary(msgs []room.Message, roomID, boundaryID string) int {
	if boundaryID == "" {
		return 0
	}
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].ID == boundaryID && msgs[i].RoomID == roomID {
			return i
		}
	}
	return 0
}

func roomContextBoundary(st room.State, roomID string) string {
	for _, r := range st.Rooms {
		if r.ID == roomID {
			return r.ContextBoundaryID
		}
	}
	return ""
}
