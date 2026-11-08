package main

// Room membership log. Every agent added to or removed from a room leaves a
// durable `room.membership` event naming the human who did it, the room and the
// agent. Like messages, these records live in the append-only events table; no
// operation edits or deletes them, and there is no feature flag that disables
// recording (audit integrity is mandatory).

import (
	"agentroom/internal/room"
	"encoding/json"
)

type membershipRecord struct {
	Action        string `json:"action"` // "added" or "removed"
	ParticipantID string `json:"participantId"`
	Name          string `json:"name"`
	RoomID        string `json:"roomId"`
	ActorID       string `json:"actorId"`
	Reason        string `json:"reason,omitempty"`
}

func membershipChange(action string, p room.Participant, actorID, reason string) auditChange {
	return auditChange{Kind: "room.membership", Payload: membershipRecord{Action: action, ParticipantID: p.ID, Name: p.Name, RoomID: p.RoomID, ActorID: actorID, Reason: reason}}
}

// connectionRemovalChanges records the implicit removal of every agent that used
// a deleted connection, which the projection otherwise drops silently.
func connectionRemovalChanges(st room.State, connectionID, actorID string) []auditChange {
	var changes []auditChange
	for _, p := range st.Participants {
		if p.ConnectionID == connectionID {
			changes = append(changes, membershipChange("removed", p, actorID, "connection removed"))
		}
	}
	return changes
}

func connectionIDOf(req eventRequest) string {
	var v struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(req.Payload, &v)
	return v.ID
}
