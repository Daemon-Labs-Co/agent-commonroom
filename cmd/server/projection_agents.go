package main

// Event reducers for connections, participants and agent activity. Each reducer applies one kind of recorded event to
// the projection and returns an error only for an event that must never have been recorded.

import (
	"agentroom/internal/room"
	"encoding/json"
	"time"
)

// reduceConnectionCreated reduces connection.created.
func reduceConnectionCreated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Connection
	json.Unmarshal(data, &v)
	st.Connections = append(st.Connections, v)
	st.Actors = append(st.Actors, room.Actor{ID: "bot-" + v.ID, Name: v.Name, Kind: "agent", Role: "Agent", Enabled: true, Revision: 1})
	return nil
}

// reduceConnectionDeleted reduces connection.deleted.
func reduceConnectionDeleted(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		ID string `json:"id"`
	}
	json.Unmarshal(data, &v)
	connections := st.Connections[:0]
	for _, c := range st.Connections {
		if c.ID != v.ID {
			connections = append(connections, c)
		}
	}
	st.Connections = connections
	participants := st.Participants[:0]
	for _, member := range st.Participants {
		if member.ConnectionID != v.ID {
			participants = append(participants, member)
		}
	}
	st.Participants = participants
	return nil
}

// reduceConnectionUpdated reduces connection.updated.
func reduceConnectionUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Connection
	json.Unmarshal(data, &v)
	for i, old := range st.Connections {
		if old.ID != v.ID {
			continue
		}
		st.Connections[i] = v
		if old.Endpoint != v.Endpoint || old.Protocol != v.Protocol {
			for j, member := range st.Participants {
				if member.ConnectionID == v.ID {
					st.Participants[j].ThreadID = ""
					st.Participants[j].ExecutionMode = ""
					st.Participants[j].Status = ""
					if room.SupportsAgent(v) {
						st.Participants[j].Status = "Idle"
					}
				}
			}
		}
	}
	return nil
}

// reduceParticipantCreated reduces participant.created.
func reduceParticipantCreated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Participant
	json.Unmarshal(data, &v)
	if v.RoomID == "" {
		v.RoomID = "engineering"
	}
	st.Participants = append(st.Participants, v)
	return nil
}

// reduceAgentActivity reduces agent.activity.
func reduceAgentActivity(proj *projection, kind string, data []byte) error {
	var activity struct {
		ParticipantID string    `json:"participantId"`
		Activity      string    `json:"activity"`
		StartedAt     time.Time `json:"startedAt"`
	}
	json.Unmarshal(data, &activity)
	proj.activities[activity.ParticipantID] = activity.Activity
	proj.activityStarts[activity.ParticipantID] = activity.StartedAt
	return nil
}

// reduceAgentBackground reduces agent.background.
func reduceAgentBackground(proj *projection, kind string, data []byte) error {
	var event room.BackgroundActivity
	if json.Unmarshal(data, &event) == nil && backgroundLabels[event.Stage] != "" {
		event.Label = backgroundLabels[event.Stage]
		proj.background[event.ParticipantID] = append(proj.background[event.ParticipantID], event)
		if len(proj.background[event.ParticipantID]) > 64 {
			proj.background[event.ParticipantID] = proj.background[event.ParticipantID][1:]
		}
	}
	return nil
}

// reduceAgentError reduces agent.error.
func reduceAgentError(proj *projection, kind string, data []byte) error {
	var failure struct {
		ParticipantID string `json:"participantId"`
		Error         string `json:"error"`
	}
	json.Unmarshal(data, &failure)
	proj.failures[failure.ParticipantID] = failure.Error
	return nil
}

// reduceAgentRecovered reduces agent.recovered.
func reduceAgentRecovered(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var recovered struct {
		Participant room.Participant `json:"participant"`
	}
	if json.Unmarshal(data, &recovered) == nil {
		for i, p := range st.Participants {
			if p.ID == recovered.Participant.ID && p.RoomID == recovered.Participant.RoomID {
				st.Participants[i] = recovered.Participant
				delete(proj.failures, p.ID)
			}
		}
	}
	return nil
}

// reduceParticipantUpdated reduces participant.updated, participant.session.
func reduceParticipantUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v room.Participant
	json.Unmarshal(data, &v)
	if v.RoomID == "" {
		v.RoomID = "engineering"
	}
	for i, p := range st.Participants {
		if p.ID == v.ID {
			st.Participants[i] = v
			if kind == "participant.session" && v.Status != "Blocked" {
				delete(proj.failures, v.ID)
			}
			if v.ConnectionID != p.ConnectionID {
				delete(proj.failures, v.ID)
				delete(proj.activities, v.ID)
				delete(proj.activityStarts, v.ID)
			}
		}
	}
	return nil
}

// reduceParticipantRemoved reduces participant.removed.
func reduceParticipantRemoved(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v struct {
		ID string `json:"id"`
	}
	json.Unmarshal(data, &v)
	kept := st.Participants[:0]
	for _, p := range st.Participants {
		if p.ID != v.ID {
			kept = append(kept, p)
		}
	}
	st.Participants = kept
	return nil
}

// reduceAgentAssignment reduces agent.assignment.
func reduceAgentAssignment(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v managerAssignment
	json.Unmarshal(data, &v)
	day := v.Assigned.UTC().Format("2006-01-02")
	for i := range st.Rooms {
		if st.Rooms[i].ID == v.RoomID {
			if st.Rooms[i].AssignmentDay != day {
				st.Rooms[i].AssignmentDay, st.Rooms[i].AssignmentsToday = day, 0
			}
			st.Rooms[i].AssignmentsToday++
		}
	}
	return nil
}

// reduceAgentInitiated reduces agent.initiated.
func reduceAgentInitiated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var v agentInitiated
	json.Unmarshal(data, &v)
	for i := range st.Rooms {
		if st.Rooms[i].ID == v.RoomID {
			st.Rooms[i].AgentInitiated++
		}
	}
	return nil
}
