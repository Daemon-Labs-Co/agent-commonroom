package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"reflect"
	"slices"
)

// objectReadSessionWitness MUST independently verify current operation-specific
// export/cleanup permission, creator ownership, approved worker/image/daemon,
// custody/writer exclusion, actual physical quota and account admission. Lifecycle
// checks also attest private persistence. Metadata is never authentication/proof.
// Callbacks must not reenter server/creator/admission locks or the waiting stream.
// There is no permissive default; a source fork can fake witnesses and void isolation.
type objectReadSessionWitness interface {
	ValidateObjectReadOperation(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error
	ValidateObjectReadLifecycle(context.Context, room.State, workerJobRecord, workerObjectReadRecord, workerProbeEvent) error
}
type workspaceObjectReadEvidence struct {
	lifetime    context.Context
	dir, worker string
	record      workerJobRecord
	intent      workerObjectReadIntent
	runner      toolbroker.DockerRunner
	plan        *toolbroker.ObjectReadContainerPlan
	witness     objectReadSessionWitness
}

func newWorkspaceObjectReadEvidence(l *workerWorkspaceObservation, worker string, intent workerObjectReadIntent, witness objectReadSessionWitness) (*workspaceObjectReadEvidence, error) {
	if l == nil || !workerAliasPattern.MatchString(worker) || nilWorkerObservationDependency(witness) {
		return nil, errControllerObjectRead
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || nilWorkerObservationDependency(l.runner) || !objectReadAuthorityBinding(l.record, intent) {
		return nil, errControllerObjectRead
	}
	plan, err := toolbroker.NewObjectReadContainerPlan(intent.Request, intent.Image, l.dockerRoot, intent.Objects)
	if err != nil {
		return nil, err
	}
	return &workspaceObjectReadEvidence{lifetime: l.lifetime, dir: l.dir, worker: worker, record: detachedProbeWorkerRecord(l.record), intent: intent, runner: l.runner, plan: plan, witness: witness}, nil
}
func objectReadControllerEvents(r workerObjectReadRecord) []workerProbeEvent {
	events := slices.Clone(r.Events)
	if len(events) > 0 && (events[len(events)-1].Stage == "confirmed" || events[len(events)-1].Stage == "recovered") {
		events = events[:len(events)-1]
	}
	return events
}
func (e *workspaceObjectReadEvidence) valid(ctx context.Context, st room.State) bool {
	if e == nil || ctx == nil || ctx.Err() != nil || e.lifetime == nil || e.lifetime.Err() != nil || e.plan == nil || nilWorkerObservationDependency(e.witness) || nilWorkerObservationDependency(e.runner) {
		return false
	}
	job, ok := st.ManagedJobs[e.record.JobID]
	b := e.intent.Request.Receipt.Intent.Scope
	return ok && validManagedJob(job) && job.CompanyID == st.Company.ID && job.CompanyID == b.CompanyID && job.RoomID == b.RoomID && job.TaskID == b.TaskID && job.TaskDispatchID == b.ExecutionID && job.TaskDefinitionDigest == b.DefinitionDigest && job.Worker == e.worker && sameWorkerBinding(job, workerJobView{workerJobRecord: e.record})
}

type objectReadLocalSnapshot struct {
	Exists bool
	Record workerObjectReadRecord
}

func (e *workspaceObjectReadEvidence) localOperation(event workerProbeEvent) (objectReadLocalSnapshot, error) {
	if workerProbeStopped(e.dir) {
		return objectReadLocalSnapshot{}, errControllerObjectRead
	}
	records, err := readWorkerObjectReadRecords(e.dir)
	if err != nil {
		return objectReadLocalSnapshot{}, err
	}
	local, exists := records[objectReadRecordName(e.intent)]
	if !exists {
		if event.Stage != "begin" {
			return objectReadLocalSnapshot{}, errControllerObjectRead
		}
		return objectReadLocalSnapshot{}, nil
	}
	if local.Intent != e.intent {
		return objectReadLocalSnapshot{}, errControllerObjectRead
	}
	events := objectReadControllerEvents(local)
	if len(events) == 0 {
		return objectReadLocalSnapshot{}, errControllerObjectRead
	}
	if events[len(events)-1] != event {
		if objectReadRecordSettled(local) {
			return objectReadLocalSnapshot{}, errControllerObjectRead
		}
		events = append(events, event)
		if !validWorkerObjectReadRecord(workerObjectReadRecord{Version: 1, Intent: e.intent, Events: events}) {
			return objectReadLocalSnapshot{}, errControllerObjectRead
		}
	}
	return objectReadLocalSnapshot{true, copyObjectReadRecord(local)}, nil
}
func (e *workspaceObjectReadEvidence) inspect(ctx context.Context, event workerProbeEvent) error {
	switch event.Stage {
	case "begin":
		return e.plan.VerifyNoReferences(ctx, e.runner)
	case "bound":
		return e.plan.VerifyAdmission(ctx, e.runner, event.ID, "created")
	case "finish", "recovery", "recovery-finish":
		if event.ID == "" {
			if event.Stage != "finish" || event.Absent {
				return errControllerObjectRead
			}
			return e.plan.VerifyNoReferences(ctx, e.runner)
		}
		absent, err := e.plan.InspectRecorded(ctx, e.runner, event.ID)
		if err != nil {
			return err
		}
		if event.Absent && !absent {
			return errControllerObjectRead
		}
		return nil
	}
	return errControllerObjectRead
}
func (e *workspaceObjectReadEvidence) ValidateObjectReadOperation(ctx context.Context, st room.State, r workerJobRecord, i workerObjectReadIntent, event workerProbeEvent) error {
	if !e.valid(ctx, st) || !objectReadAuthorityBinding(r, i) || probeWorkerRecordDigest(r) != probeWorkerRecordDigest(e.record) || i != e.intent || !objectReadAuthorityEvent(event) {
		return errControllerObjectRead
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
	if e.witness.ValidateObjectReadOperation(child, st, detachedProbeWorkerRecord(e.record), e.intent, event) != nil || child.Err() != nil {
		return errControllerObjectRead
	}
	if err = e.inspect(child, event); err != nil {
		return err
	}
	if _, err = readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	after, err := e.localOperation(event)
	if err != nil || !reflect.DeepEqual(before, after) {
		return errControllerObjectRead
	}
	if e.witness.ValidateObjectReadOperation(child, st, detachedProbeWorkerRecord(e.record), e.intent, event) != nil || child.Err() != nil || e.lifetime.Err() != nil {
		return errControllerObjectRead
	}
	return nil
}
func (e *workspaceObjectReadEvidence) localTransition(candidate room.ObjectReadReservation, event workerProbeEvent) (workerObjectReadRecord, error) {
	records, err := readWorkerObjectReadRecords(e.dir)
	local, ok := records[objectReadRecordName(e.intent)]
	if err != nil || !ok || local.Intent != e.intent || !slices.Equal(objectReadControllerEvents(local), candidate.Events) || len(candidate.Events) == 0 || candidate.Events[len(candidate.Events)-1] != event {
		return workerObjectReadRecord{}, errControllerObjectRead
	}
	return copyObjectReadRecord(local), nil
}
func (e *workspaceObjectReadEvidence) ValidateObjectReadTransition(ctx context.Context, st room.State, candidate room.ObjectReadReservation, event workerProbeEvent) error {
	if !e.valid(ctx, st) || !validControllerObjectReadRecord(candidate) || candidate.Intent != e.intent || candidate.Worker != e.worker || candidate.CompanyID != st.Company.ID || candidate.RoomID != e.record.RoomID {
		return errControllerObjectRead
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(e.lifetime, cancel)
	defer stop()
	if _, err := readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	before, err := e.localTransition(candidate, event)
	if err != nil {
		return err
	}
	if e.witness.ValidateObjectReadLifecycle(child, st, detachedProbeWorkerRecord(e.record), copyObjectReadRecord(before), event) != nil || child.Err() != nil {
		return errControllerObjectRead
	}
	if err = e.inspect(child, event); err != nil {
		return err
	}
	if _, err = readRetainedWorkspaceObservation(e.dir, e.record); err != nil {
		return err
	}
	after, err := e.localTransition(candidate, event)
	if err != nil || !reflect.DeepEqual(before, after) {
		return errControllerObjectRead
	}
	if e.witness.ValidateObjectReadLifecycle(child, st, detachedProbeWorkerRecord(e.record), copyObjectReadRecord(after), event) != nil || child.Err() != nil || e.lifetime.Err() != nil {
		return errControllerObjectRead
	}
	return nil
}

var _ controllerObjectReadSessionEvidence = (*workspaceObjectReadEvidence)(nil)
