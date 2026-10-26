package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"reflect"
	"slices"
)

// captureLifecycleWitness MUST independently authenticate operation-specific current
// cleanup/source authority, creator ownership, approved worker/image/volumes, account
// capacity/storage accounting and real physical quotas. It must not reenter server/
// admission locks or the same control stream. No permissive/default witness exists.
type captureLifecycleWitness interface {
	ValidateCaptureLifecycle(context.Context, room.State, workerJobRecord, workerCaptureRecord, workerProbeEvent) error
}
type workspaceCaptureLedgerEvidence struct {
	lifetime    context.Context
	dir, worker string
	record      workerJobRecord
	intent      workerCaptureIntent
	runner      toolbroker.DockerRunner
	plan        *toolbroker.CaptureContainerPlan
	witness     captureLifecycleWitness
}

func newWorkspaceCaptureLedgerEvidence(l *workerWorkspaceObservation, worker string, intent workerCaptureIntent, witness captureLifecycleWitness) (*workspaceCaptureLedgerEvidence, error) {
	if l == nil || !workerAliasPattern.MatchString(worker) || nilWorkerObservationDependency(witness) {
		return nil, errControllerCapture
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || !retainedObservationRecord(l.record) || nilWorkerObservationDependency(l.runner) {
		return nil, errControllerCapture
	}
	b := intent.Request.Binding
	v := l.record.Pair.Workspace.Identity
	if b.Execution.JobID != l.record.JobID || b.Execution.RepositoryDigest != l.record.Pair.Repository || b.VolumeName != v.Name || b.VolumeNonce != v.Nonce || b.VolumeCreatedAt != v.CreatedAt {
		return nil, errControllerCapture
	}
	plan, err := toolbroker.NewCaptureContainerPlan(intent.Request, intent.Image, l.dockerRoot, intent.Objects)
	if err != nil {
		return nil, err
	}
	return &workspaceCaptureLedgerEvidence{l.lifetime, l.dir, worker, detachedProbeWorkerRecord(l.record), copyCaptureIntent(intent), l.runner, plan, witness}, nil
}
func captureControllerEvents(local workerCaptureRecord) []workerProbeEvent {
	events := slices.Clone(local.Events)
	if len(events) > 0 {
		last := events[len(events)-1].Stage
		if last == "confirmed" || last == "recovered" {
			events = events[:len(events)-1]
		}
	}
	return events
}
func (e *workspaceCaptureLedgerEvidence) local(candidate room.WorkspaceCaptureReservation, event workerProbeEvent) (workerCaptureRecord, error) {
	records, err := readWorkerCaptureRecords(e.dir)
	local, ok := records[captureRecordName(e.intent)]
	if err != nil || !ok || !captureIntentEqual(local.Intent, e.intent) || !reflect.DeepEqual(captureControllerEvents(local), candidate.Events) || len(candidate.Events) == 0 || candidate.Events[len(candidate.Events)-1] != event {
		return workerCaptureRecord{}, errControllerCapture
	}
	return local, nil
}
func (e *workspaceCaptureLedgerEvidence) ValidateCaptureTransition(ctx context.Context, st room.State, candidate room.WorkspaceCaptureReservation, event workerProbeEvent) error {
	if e == nil || ctx == nil || ctx.Err() != nil || e.lifetime == nil || e.lifetime.Err() != nil || e.plan == nil || nilWorkerObservationDependency(e.witness) || !validControllerCaptureRecord(candidate) || !captureIntentEqual(candidate.Intent, e.intent) || candidate.Worker != e.worker {
		return errControllerCapture
	}
	job, ok := st.ManagedJobs[e.record.JobID]
	b := e.intent.Request.Binding.Execution
	if !ok || !validManagedJob(job) || job.CompanyID != st.Company.ID || candidate.CompanyID != job.CompanyID || candidate.RoomID != job.RoomID || b.CompanyID != job.CompanyID || b.RoomID != job.RoomID || b.TaskID != job.TaskID || b.ExecutionID != job.TaskDispatchID || b.DefinitionDigest != job.TaskDefinitionDigest || job.Worker != e.worker || !sameWorkerBinding(job, workerJobView{workerJobRecord: e.record}) {
		return errControllerCapture
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(e.lifetime, cancel)
	defer stop()
	if _, err := readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	local, err := e.local(candidate, event)
	if err != nil {
		return err
	}
	if e.witness.ValidateCaptureLifecycle(child, st, detachedProbeWorkerRecord(e.record), copyCaptureRecord(local), event) != nil {
		return errControllerCapture
	}
	err = e.inspectTransition(child, event)
	if err != nil {
		return err
	}
	if _, err = readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	after, err := e.local(candidate, event)
	if err != nil || !reflect.DeepEqual(after, local) {
		return errControllerCapture
	}
	if e.witness.ValidateCaptureLifecycle(child, st, detachedProbeWorkerRecord(e.record), copyCaptureRecord(after), event) != nil || child.Err() != nil || e.lifetime.Err() != nil {
		return errControllerCapture
	}
	return nil
}

var _ controllerCaptureLedgerEvidence = (*workspaceCaptureLedgerEvidence)(nil)

// inspectTransition observes exact known identities only; it never performs lifecycle effects.
func (e *workspaceCaptureLedgerEvidence) inspectTransition(ctx context.Context, event workerProbeEvent) error {
	var err error
	switch event.Stage {
	case "begin":
		err = e.plan.VerifyNoReferences(ctx, e.runner)
	case "bound":
		err = e.plan.VerifyAdmission(ctx, e.runner, event.ID, "created")
	case "finish", "recovery", "recovery-finish":
		if event.ID == "" {
			if event.Absent || event.Stage != "finish" {
				return errControllerCapture
			}
			err = e.plan.VerifyNoReferences(ctx, e.runner)
		} else {
			var absent bool
			absent, err = e.plan.InspectRecorded(ctx, e.runner, event.ID)
			if err == nil && event.Absent && !absent {
				err = errControllerCapture
			}
		}
	default:
		return errControllerCapture
	}
	return err
}
