package main

import (
	"agentroom/internal/room"
	"context"
	"reflect"
)

// captureSessionWitness must independently verify operation-specific current
// creator/worker/account admission and physically enforced quotas. Lifecycle witness
// checks additionally verify persisted transitions. Neither can trust stream claims,
// reacquire server/admission locks or call the waiting stream. No default exists.
type captureSessionWitness interface {
	captureLifecycleWitness
	ValidateCaptureOperation(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error
}
type workspaceCaptureSessionEvidence struct {
	*workspaceCaptureLedgerEvidence
	operation captureSessionWitness
}

func newWorkspaceCaptureSessionEvidence(l *workerWorkspaceObservation, worker string, intent workerCaptureIntent, witness captureSessionWitness) (*workspaceCaptureSessionEvidence, error) {
	if nilWorkerObservationDependency(witness) {
		return nil, errControllerCapture
	}
	ledger, err := newWorkspaceCaptureLedgerEvidence(l, worker, intent, witness)
	if err != nil {
		return nil, err
	}
	return &workspaceCaptureSessionEvidence{workspaceCaptureLedgerEvidence: ledger, operation: witness}, nil
}

type captureLocalOperationState struct {
	exists bool
	record workerCaptureRecord
}

func (e *workspaceCaptureSessionEvidence) localOperation(event workerProbeEvent) (captureLocalOperationState, error) {
	var zero captureLocalOperationState
	if workerProbeStopped(e.dir) {
		return zero, errControllerCapture
	}
	records, err := readWorkerCaptureRecords(e.dir)
	if err != nil {
		return zero, err
	}
	local, exists := records[captureRecordName(e.intent)]
	if !exists {
		if event.Stage != "begin" {
			return zero, errControllerCapture
		}
		return zero, nil
	}
	if !captureIntentEqual(local.Intent, e.intent) {
		return zero, errControllerCapture
	}
	events := captureControllerEvents(local)
	if len(events) == 0 {
		return zero, errControllerCapture
	}
	// Before persistence, the proposed event must extend valid retained history.
	// After persistence, only the exact latest event may be rechecked. Terminal local
	// markers remain in the compared snapshot and cannot authorize a new lifecycle.
	if events[len(events)-1] != event {
		if captureRecordSettled(local) {
			return zero, errControllerCapture
		}
		events = append(events, event)
		if !validWorkerCaptureRecord(workerCaptureRecord{Version: 1, Intent: e.intent, Events: events}) {
			return zero, errControllerCapture
		}
	}
	return captureLocalOperationState{exists: true, record: copyCaptureRecord(local)}, nil
}
func (e *workspaceCaptureSessionEvidence) ValidateCaptureOperation(ctx context.Context, st room.State, r workerJobRecord, intent workerCaptureIntent, event workerProbeEvent) error {
	if e == nil || e.workspaceCaptureLedgerEvidence == nil || ctx == nil || ctx.Err() != nil || e.lifetime == nil || e.lifetime.Err() != nil || e.plan == nil || nilWorkerObservationDependency(e.operation) || !retainedObservationRecord(r) || probeWorkerRecordDigest(r) != probeWorkerRecordDigest(e.record) || !captureIntentEqual(intent, e.intent) || !captureAuthorityEvent(event) {
		return errControllerCapture
	}
	job, ok := st.ManagedJobs[e.record.JobID]
	b := e.intent.Request.Binding.Execution
	if !ok || !validManagedJob(job) || job.CompanyID != st.Company.ID || b.CompanyID != job.CompanyID || b.RoomID != job.RoomID || b.TaskID != job.TaskID || b.ExecutionID != job.TaskDispatchID || b.DefinitionDigest != job.TaskDefinitionDigest || job.Worker != e.worker || !sameWorkerBinding(job, workerJobView{workerJobRecord: e.record}) {
		return errControllerCapture
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(e.lifetime, cancel)
	defer stop()
	if _, err := readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	before, err := e.localOperation(event)
	if err != nil {
		return err
	}
	if e.operation.ValidateCaptureOperation(child, st, detachedProbeWorkerRecord(e.record), copyCaptureIntent(e.intent), event) != nil || child.Err() != nil {
		return errControllerCapture
	}
	if err = e.inspectTransition(child, event); err != nil {
		return err
	}
	if _, err = readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	after, err := e.localOperation(event)
	if err != nil || !reflect.DeepEqual(before, after) {
		return errControllerCapture
	}
	if e.operation.ValidateCaptureOperation(child, st, detachedProbeWorkerRecord(e.record), copyCaptureIntent(e.intent), event) != nil || child.Err() != nil || e.lifetime.Err() != nil {
		return errControllerCapture
	}
	return nil
}

var _ controllerCaptureSessionEvidence = (*workspaceCaptureSessionEvidence)(nil)
