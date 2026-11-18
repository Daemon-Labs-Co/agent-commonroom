package main

// Event reducers for rooms, messages, memory and objections. Each reducer applies one kind of recorded event to
// the projection and returns an error only for an event that must never have been recorded.

import (
	"agentroom/internal/room"
	"encoding/json"
	"errors"
)

// reduceRoomOutcomeReviewed reduces room.outcome.reviewed.
func reduceRoomOutcomeReviewed(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectRoomOutcomeReview(st, data); err != nil {
		return err
	}
	return nil
}

// reduceRoomCreated reduces room.created, room.updated.
func reduceRoomCreated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if kind == "room.updated" {
		projectActors(st)
	}
	var v room.Room
	json.Unmarshal(data, &v)
	if kind == "room.created" {
		st.Rooms = append(st.Rooms, v)
	} else {
		for i, old := range st.Rooms {
			if old.ID == v.ID {
				if old.PersonalActorID != "" {
					// A changed saved name is an explicit override.
					// Saving other settings keeps automatic labels derived.
					v.CustomPersonalName = old.CustomPersonalName || v.Name != old.Name
					v.PersonalActorID = old.PersonalActorID
				}
				st.Rooms[i] = v
			}
		}
	}
	return nil
}

// reduceEvidenceCreated reduces evidence.created.
func reduceEvidenceCreated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Evidence
	json.Unmarshal(data, &v)
	st.Evidence = append(st.Evidence, v)
	return nil
}

// reduceRoomDiscussionStarted reduces room.discussion.started.
func reduceRoomDiscussionStarted(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v discussionStarted
	json.Unmarshal(data, &v)
	for i := range st.Rooms {
		if st.Rooms[i].ID == v.RoomID {
			st.Rooms[i].DiscussionID = v.ID
		}
	}
	return nil
}

// reduceRoomDiscussionEnded reduces room.discussion.ended.
func reduceRoomDiscussionEnded(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v discussionEnded
	json.Unmarshal(data, &v)
	for i := range st.Rooms {
		if st.Rooms[i].ID == v.RoomID && st.Rooms[i].DiscussionID == v.ID {
			st.Rooms[i].DiscussionID = ""
		}
	}
	return nil
}

// reduceRoomDiscussionLimitUpdated reduces room.discussion_limit.updated.
func reduceRoomDiscussionLimitUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v discussionLimit
	json.Unmarshal(data, &v)
	for i := range st.Rooms {
		if st.Rooms[i].ID == v.RoomID {
			st.Rooms[i].DiscussionRounds = v.Rounds
		}
	}
	return nil
}

// reduceRoomContextCleared reduces room.context.cleared.
func reduceRoomContextCleared(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v contextClearPayload
	json.Unmarshal(data, &v)
	for i := range st.Rooms {
		if st.Rooms[i].ID == v.RoomID {
			st.Rooms[i].ContextBoundaryID = v.MessageID
		}
	}
	return nil
}

// reduceRoomContextExportUpdated reduces room.context_export.updated.
func reduceRoomContextExportUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		RoomID string                   `json:"roomId"`
		Policy room.ContextExportPolicy `json:"policy"`
	}
	json.Unmarshal(data, &v)
	if !hasRoom(*st, v.RoomID) {
		projectActors(st)
	}
	for i := range st.Rooms {
		if st.Rooms[i].ID == v.RoomID {
			if st.TaskContextPolicyEpochs == nil {
				st.TaskContextPolicyEpochs = map[string]int{}
			}
			if st.TaskContextPolicyEpochs[v.RoomID] == int(^uint(0)>>1) {
				return errors.New("context policy generation exhausted")
			}
			st.TaskContextPolicyEpochs[v.RoomID]++
			st.Rooms[i].ContextExport = &v.Policy
			st.Rooms[i].Revision++
		}
	}
	return nil
}

// reduceMemoryRevised reduces memory.revised.
func reduceMemoryRevised(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Memory
	json.Unmarshal(data, &v)
	if v.RoomID == "" {
		v.RoomID = "engineering"
	}
	found := false
	for i, m := range st.Memories {
		if m.ID == v.ID {
			st.Memories[i] = v
			found = true
		}
	}
	if !found {
		st.Memories = append(st.Memories, v)
	}
	return nil
}

// reduceObjectionResolved reduces objection.resolved.
func reduceObjectionResolved(proj *projection, kind string, data []byte) error {
	var v objectionResolution
	json.Unmarshal(data, &v)
	proj.objectionResolutions[v.MessageID] = v
	return nil
}

// reduceMessageAttentionAcknowledged reduces message.attention.acknowledged.
func reduceMessageAttentionAcknowledged(proj *projection, kind string, data []byte) error {
	var ack struct {
		MessageID string `json:"messageId"`
		ActorID   string `json:"actorId"`
	}
	json.Unmarshal(data, &ack)
	proj.acknowledged[ack.MessageID] = ack.ActorID
	return nil
}

// reduceRoomAssignmentLimitUpdated reduces room.assignment_limit.updated.
func reduceRoomAssignmentLimitUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v assignmentLimit
	json.Unmarshal(data, &v)
	for i := range st.Rooms {
		if st.Rooms[i].ID == v.RoomID {
			limit := v.Limit
			st.Rooms[i].AssignmentLimit = &limit
		}
	}
	return nil
}

// reduceMessageCreated reduces message.created.
func reduceMessageCreated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Message
	json.Unmarshal(data, &v)
	if v.RoomID == "" {
		v.RoomID = "engineering"
	}
	if v.ActorType == "human" {
		for i := range st.Rooms {
			if st.Rooms[i].ID == v.RoomID {
				st.Rooms[i].AgentInitiated = 0
			}
		}
	}
	if v.Provider == "" && v.Author != "You" {
		for _, p := range st.Participants {
			if p.Name == v.Author && p.RoomID == v.RoomID {
				for _, c := range st.Connections {
					if c.ID == p.ConnectionID {
						v.Provider = c.Provider
						if v.Provider == "" && room.SupportsAgent(c) {
							v.Provider = providerForEndpoint(c.Endpoint)
						}
					}
				}
			}
		}
	}
	st.Messages = append(st.Messages, v)
	return nil
}
