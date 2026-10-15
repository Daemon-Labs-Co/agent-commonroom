package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"encoding/json"
	"errors"
)

func toolAuditKey(job, call string) string { return job + "." + call }

func toolAuditIntent(call room.ManagedToolCall) toolbroker.Intent {
	return toolbroker.Intent{Scope: toolbroker.Scope{Version: 1, JobID: call.JobID, Room: call.RoomID, Participant: call.ParticipantID, Connection: call.ConnectionID, Provider: call.Provider, RepositoryDigest: call.RepositoryDigest, Deadline: call.Deadline}, CallID: call.CallID, Operation: call.Operation, PayloadDigest: call.PayloadDigest}
}

// projectToolAudit checks historical transitions, not current access policy.
// Permission changes must not invalidate previously valid audit evidence.
func projectToolAudit(st *room.State, kind string, raw []byte) error {
	var call room.ManagedToolCall
	if json.Unmarshal(raw, &call) != nil || !toolbroker.ValidAuditIntent(toolAuditIntent(call), call.RecordedAt) || call.ActorID == "" || call.CompanyID == "" || call.RecordedAt.IsZero() || call.UpdatedAt.Before(call.RecordedAt) {
		return errors.New("invalid tool audit event")
	}
	if st.ToolAudits == nil {
		st.ToolAudits = map[string]room.ManagedToolCall{}
	}
	key := toolAuditKey(call.JobID, call.CallID)
	prior, exists := st.ToolAudits[key]
	if kind == "worker.tool.intent" {
		if exists || call.Status != "pending" || call.OutputBytes != 0 || call.OutputDigest != "" || !call.RecordedAt.Equal(call.UpdatedAt) {
			return errors.New("invalid tool audit intent")
		}
	} else {
		outcome := toolbroker.Outcome{Status: call.Status, OutputBytes: call.OutputBytes, OutputDigest: call.OutputDigest}
		if !exists || prior.Status != "pending" || !toolbroker.ValidAuditOutcome(outcome) {
			return errors.New("invalid tool audit outcome")
		}
		candidate := call
		candidate.Status, candidate.OutputBytes, candidate.OutputDigest, candidate.UpdatedAt = prior.Status, prior.OutputBytes, prior.OutputDigest, prior.UpdatedAt
		if candidate != prior {
			return errors.New("tool audit scope changed")
		}
	}
	st.ToolAudits[key] = call
	return nil
}
