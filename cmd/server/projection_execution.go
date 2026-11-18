package main

// Event reducers for managed jobs, tool calls, artifacts, storage and workspace custody. Each reducer applies one kind of recorded event to
// the projection and returns an error only for an event that must never have been recorded.

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"encoding/json"
	"errors"
	"time"
)

// reduceWorkerIdleStopAnnounced reduces worker.idle_stop.announced.
func reduceWorkerIdleStopAnnounced(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectIdleNotice(st, data); err != nil {
		return err
	}
	return nil
}

// reduceWorkspaceDisposalProposed reduces workspace.disposal.proposed, workspace.disposal.claimed, workspace.disposal.revoked, workspace.disposal.succeeded, workspace.disposal.unconfirmed.
func reduceWorkspaceDisposalProposed(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectWorkspaceDisposal(st, kind, data); err != nil {
		return err
	}
	return nil
}

// reduceArtifactReaderAcknowledged reduces artifact.reader.acknowledged.
func reduceArtifactReaderAcknowledged(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectControllerObjectRead(st, data); err != nil {
		return err
	}
	return nil
}

// reduceWorkspaceCaptureAcknowledged reduces workspace.capture.acknowledged.
func reduceWorkspaceCaptureAcknowledged(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectControllerCapture(st, data); err != nil {
		return err
	}
	return nil
}

// reduceWorkspaceCaptureAuthorized reduces workspace.capture.authorized.
func reduceWorkspaceCaptureAuthorized(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectControllerCaptureAuthorization(st, data); err != nil {
		return err
	}
	return nil
}

// reduceWorkspaceProbeAuthorized reduces workspace.probe.authorized.
func reduceWorkspaceProbeAuthorized(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectControllerProbeAuthorization(st, data); err != nil {
		return err
	}
	return nil
}

// reduceWorkspaceProbeAcknowledged reduces workspace.probe.acknowledged.
func reduceWorkspaceProbeAcknowledged(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectControllerProbe(st, data); err != nil {
		return err
	}
	return nil
}

// reduceArtifactManifestRecorded reduces artifact.manifest.recorded.
func reduceArtifactManifestRecorded(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectArtifactManifest(st, data); err != nil {
		return err
	}
	return nil
}

// reduceArtifactCustodyReserved reduces artifact.custody.reserved, artifact.custody.ready, artifact.custody.sealed, artifact.custody.read.
func reduceArtifactCustodyReserved(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectArtifactCustody(st, kind, data); err != nil {
		return err
	}
	return nil
}

// reduceStorageInventoryRecorded reduces storage.inventory.recorded, storage.reserved, storage.released.
func reduceStorageInventoryRecorded(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectManagedStorage(st, kind, data); err != nil {
		return err
	}
	return nil
}

// reduceWorkerToolIntent reduces worker.tool.intent, worker.tool.outcome.
func reduceWorkerToolIntent(proj *projection, kind string, data []byte) error {
	st := &proj.st
	if err := projectToolAudit(st, kind, data); err != nil {
		return err
	}
	return nil
}

// reduceOperationReceipt reduces operation.receipt.
func reduceOperationReceipt(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var receipt room.OperationReceipt
	if json.Unmarshal(data, &receipt) != nil || len(receipt.Key) != 64 || len(receipt.Fingerprint) != 64 || receipt.Status < 200 || receipt.Status >= 300 || !json.Valid(receipt.Response) {
		return errors.New("Invalid operation receipt")
	}
	if st.OperationReceipts == nil {
		st.OperationReceipts = map[string]room.OperationReceipt{}
	}
	st.OperationReceipts[receipt.Key] = receipt
	return nil
}

// reduceToolCallIntent reduces tool.call.intent.
func reduceToolCallIntent(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var intent toolbroker.Intent
	// ValidAuditIntent's deadline check is admission-time-relative; replay
	// happens long after that deadline passed, so validate shape as of the
	// intent's own deadline instant rather than wall-clock now (see its doc
	// comment: "outcome persistence may instead use the original intent's
	// admission instant" — the same reasoning applies to replaying history).
	if json.Unmarshal(data, &intent) != nil || !toolbroker.ValidAuditIntent(intent, intent.Scope.Deadline.Add(-time.Nanosecond)) {
		return errors.New("invalid tool call intent event")
	}
	if st.ToolCalls == nil {
		st.ToolCalls = map[string]room.ToolCallAudit{}
	}
	st.ToolCalls[intent.CallID] = room.ToolCallAudit{
		CallID: intent.CallID, JobID: intent.Scope.JobID, RoomID: intent.Scope.Room,
		ParticipantID: intent.Scope.Participant, ConnectionID: intent.Scope.Connection,
		Provider: intent.Scope.Provider, Operation: intent.Operation,
		PayloadDigest: intent.PayloadDigest, Status: "pending",
	}
	return nil
}

// reduceToolCallOutcome reduces tool.call.outcome.
func reduceToolCallOutcome(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var record toolCallOutcomeRecord
	if json.Unmarshal(data, &record) != nil || !toolbroker.ValidAuditOutcome(record.Outcome) {
		return errors.New("invalid tool call outcome event")
	}
	existing, ok := st.ToolCalls[record.Intent.CallID]
	if !ok {
		return errors.New("tool call outcome has no recorded intent")
	}
	existing.Status, existing.OutputBytes, existing.OutputDigest = record.Outcome.Status, record.Outcome.OutputBytes, record.Outcome.OutputDigest
	st.ToolCalls[record.Intent.CallID] = existing
	return nil
}

// reduceWorkerJobUpdated reduces worker.job.updated.
func reduceWorkerJobUpdated(proj *projection, kind string, data []byte) error {
	st := &proj.st
	var job room.ManagedJob
	if json.Unmarshal(data, &job) != nil || !validManagedJob(job) {
		return errors.New("invalid managed job event")
	}
	if st.ManagedJobs == nil {
		st.ManagedJobs = map[string]room.ManagedJob{}
	}
	if _, exists := st.ManagedJobs[job.JobID]; !exists && job.GovernanceRevision > 0 && !coordinationJobAllowed(*st, job) {
		return errors.New("invalid initial coordination dependency")
	}
	if prior, exists := st.ManagedJobs[job.JobID]; exists {
		binding := job
		binding.Status, binding.DiagnosticCode, binding.UpdatedAt, binding.ShutdownConfirmed = prior.Status, prior.DiagnosticCode, prior.UpdatedAt, prior.ShutdownConfirmed
		if binding != prior || prior.ShutdownConfirmed && !job.ShutdownConfirmed {
			return errors.New("managed job binding changed")
		}
	}
	st.ManagedJobs[job.JobID] = job
	return nil
}
