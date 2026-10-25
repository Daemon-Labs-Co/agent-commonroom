package main

import (
	"agentroom/internal/room"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

// Hash stable configuration only. Session threads, status/telemetry, consent
// projection and derived company/actor fields legitimately change during a turn.
func taskParticipantConfigurationDigest(p room.Participant) string {
	p.Status = ""
	p.ThreadID = ""
	p.ExecutionMode = ""
	p.Background = nil
	p.StatusDetails = nil
	p.DisclosureVersion = 0
	p.CompanyID = ""
	p.ActorID = ""
	raw, _ := json.Marshal(p)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func taskConnectionConfigurationDigest(c room.Connection) string {
	c.CompanyID = ""
	raw, _ := json.Marshal(c)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func bindTaskExecutionConfiguration(st room.State, p room.Participant, e *room.TaskExecution) error {
	var participant room.Participant
	var connection room.Connection
	participants, connections := 0, 0
	for _, v := range st.Participants {
		if v.ID == p.ID && v.RoomID == p.RoomID && v.ConnectionID == p.ConnectionID {
			participant = v
			participants++
		}
	}
	for _, v := range st.Connections {
		if v.ID == p.ConnectionID {
			connection = v
			connections++
		}
	}
	if participant.Revision < 0 || connection.Revision < 0 || participants != 1 || connections != 1 || p.RoomID != e.RoomID || p.ID != e.ParticipantID || connection.Protocol != "CLI" {
		return errors.New("task execution configuration unavailable")
	}
	e.ContextPolicyEpoch = st.TaskContextPolicyEpochs[e.RoomID]
	e.ConfigurationVersion = 1
	e.ParticipantRevision = participant.Revision
	e.ParticipantConfigurationDigest = taskParticipantConfigurationDigest(participant)
	e.ConnectionID = connection.ID
	e.ConnectionRevision = connection.Revision
	e.ConnectionConfigurationDigest = taskConnectionConfigurationDigest(connection)
	return nil
}
func taskExecutionConfigurationCurrent(st room.State, e room.TaskExecution) bool {
	if e.ConfigurationVersion != 1 {
		return false
	}
	candidate := e
	p := room.Participant{ID: e.ParticipantID, RoomID: e.RoomID, ConnectionID: e.ConnectionID}
	if bindTaskExecutionConfiguration(st, p, &candidate) != nil {
		return false
	}
	return candidate.ContextPolicyEpoch == e.ContextPolicyEpoch && candidate.ParticipantRevision == e.ParticipantRevision && candidate.ParticipantConfigurationDigest == e.ParticipantConfigurationDigest && candidate.ConnectionID == e.ConnectionID && candidate.ConnectionRevision == e.ConnectionRevision && candidate.ConnectionConfigurationDigest == e.ConnectionConfigurationDigest
}
func validTaskExecutionConfiguration(st room.State, e room.TaskExecution) bool {
	if e.ConfigurationVersion == 0 { // Retained pre-configuration history, never pending authority.
		return e.ContextPolicyEpoch == 0 && e.ParticipantRevision == 0 && e.ParticipantConfigurationDigest == "" && e.ConnectionID == "" && e.ConnectionRevision == 0 && e.ConnectionConfigurationDigest == ""
	}
	return e.ParticipantRevision >= 0 && e.ConnectionRevision >= 0 && taskExecutionConfigurationCurrent(st, e)
}

// Policy edits revoke pending authority even after an allow/deny/allow cycle.
func taskExecutionExportCurrent(st room.State, e room.TaskExecution) bool {
	if e.ContextPolicyEpoch != st.TaskContextPolicyEpochs[e.RoomID] {
		return false
	}
	for _, p := range st.Participants {
		if p.ID == e.ParticipantID && p.RoomID == e.RoomID && p.ConnectionID == e.ConnectionID {
			_, _, err := contextDestination(st, p)
			return err == nil
		}
	}
	return false
}
