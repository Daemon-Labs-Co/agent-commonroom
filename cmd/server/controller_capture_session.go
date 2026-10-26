package main

import (
	"agentroom/internal/room"
	"context"
	"reflect"
)

// controllerCaptureSessionEvidence supplies independent operation-specific creator,
// physical quota/account and volume/image checks before a local journal transition,
// plus persisted local/physical evidence for controller acknowledgments. It must not
// reenter server/admission locks or the waiting control stream. No default exists.
type controllerCaptureSessionEvidence interface {
	controllerCaptureLedgerEvidence
	ValidateCaptureOperation(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error
}
type controllerCaptureSession struct {
	server        *server
	lifetime      context.Context
	cancel        context.CancelFunc
	actor, worker string
	record        workerJobRecord
	intent        workerCaptureIntent
	evidence      controllerCaptureSessionEvidence
	ledger        *controllerCaptureLedger
}

// Session identity comes only from authenticated private setup, never peer fields.
// This adapter checks policy; it does not establish transport authentication or
// implement physical quotas. There is no host launch or optional isolation bypass.
func newControllerCaptureSession(ctx context.Context, s *server, worker string, r workerJobRecord, i workerCaptureIntent, e controllerCaptureSessionEvidence) (*controllerCaptureSession, error) {
	if ctx == nil || ctx.Err() != nil || s == nil || !workerAliasPattern.MatchString(worker) || !captureAuthorityBinding(r, i) || nilWorkerObservationDependency(e) {
		return nil, errControllerCapture
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	if !membershipPartPattern.MatchString(actor) {
		return nil, errControllerCapture
	}
	lifetime, cancel := context.WithCancel(ctx)
	g := &controllerCaptureSession{server: s, lifetime: lifetime, cancel: cancel, actor: actor, worker: worker, record: detachedProbeWorkerRecord(r), intent: copyCaptureIntent(i), evidence: e}
	ledger, err := newControllerCaptureLedger(s, g)
	if err != nil {
		cancel()
		return nil, err
	}
	g.ledger = ledger
	return g, nil
}
func (g *controllerCaptureSession) callContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if g == nil || ctx == nil || ctx.Err() != nil || g.lifetime == nil || g.lifetime.Err() != nil {
		return nil, nil, errControllerCapture
	}
	child, cancel := context.WithCancel(context.WithValue(ctx, actorContextKey{}, g.actor))
	stop := context.AfterFunc(g.lifetime, cancel)
	return child, func() { stop(); cancel() }, nil
}
func (g *controllerCaptureSession) pinned(ctx context.Context, st room.State) bool {
	job, ok := st.ManagedJobs[g.record.JobID]
	return ok && ctx.Err() == nil && g.lifetime.Err() == nil && job.Worker == g.worker && sameWorkerBinding(job, workerJobView{workerJobRecord: g.record})
}
func (g *controllerCaptureSession) candidate(ctx context.Context, event workerProbeEvent) (room.State, room.WorkspaceCaptureReservation, error) {
	st, v, err := g.ledger.current(ctx, g.intent, event)
	if err != nil || !g.pinned(ctx, st) {
		return st, room.WorkspaceCaptureReservation{}, errControllerCapture
	}
	candidate, _, err := validateControllerCaptureEvent(st, v)
	// An old idempotent ledger event is not fresh operation authority. Only the
	// current last transition can be rechecked after its acknowledgment.
	if err != nil || len(candidate.Events) == 0 || candidate.Events[len(candidate.Events)-1] != event {
		return st, room.WorkspaceCaptureReservation{}, errControllerCapture
	}
	return st, candidate, nil
}
func (g *controllerCaptureSession) AuthorizeCapture(ctx context.Context, r workerJobRecord, i workerCaptureIntent, event workerProbeEvent) error {
	if g == nil || !retainedObservationRecord(r) || probeWorkerRecordDigest(r) != probeWorkerRecordDigest(g.record) || !captureIntentEqual(i, g.intent) || !captureAuthorityEvent(event) {
		return errControllerCapture
	}
	child, cancel, err := g.callContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	child, storageCancel := g.server.storageContext(child)
	defer storageCancel()
	s := g.server
	s.mu.Lock()
	defer s.mu.Unlock()
	st, _, err := g.candidate(child, event)
	if err != nil {
		return err
	}
	if g.evidence.ValidateCaptureOperation(child, st, detachedProbeWorkerRecord(g.record), copyCaptureIntent(g.intent), event) != nil || child.Err() != nil {
		return errControllerCapture
	}
	fresh, _, err := g.candidate(child, event)
	if err != nil {
		return err
	}
	if fresh.EventCount != st.EventCount {
		return errAuditConflict
	}
	return nil // No capacity is reserved until the separate durable acknowledgment.
}
func (g *controllerCaptureSession) AcknowledgeCapture(ctx context.Context, local workerCaptureRecord, event workerProbeEvent) error {
	if g == nil || !captureAuthorityRecord(local, g.intent, event) {
		return errControllerCapture
	}
	child, cancel, err := g.callContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	// Pin the exact received history inside the ledger's locked, fenced snapshot.
	// A concurrent call may not replace this request's history on a shared session.
	expected := copyCaptureRecord(local)
	ledger, err := newControllerCaptureLedger(g.server, &captureSessionAcknowledgment{session: g, local: expected})
	if err != nil {
		return err
	}
	return ledger.Acknowledge(child, g.intent, event)
}

type captureSessionAcknowledgment struct {
	session *controllerCaptureSession
	local   workerCaptureRecord
}

func (a *captureSessionAcknowledgment) ValidateCaptureTransition(ctx context.Context, st room.State, r room.WorkspaceCaptureReservation, e workerProbeEvent) error {
	if !reflect.DeepEqual(r.Events, a.local.Events) || !captureIntentEqual(r.Intent, a.local.Intent) {
		return errControllerCapture
	}
	return a.session.ValidateCaptureTransition(ctx, st, r, e)
}
func (g *controllerCaptureSession) ValidateCaptureTransition(ctx context.Context, st room.State, r room.WorkspaceCaptureReservation, event workerProbeEvent) error {
	if !g.pinned(ctx, st) || r.Worker != g.worker || !captureIntentEqual(r.Intent, g.intent) {
		return errControllerCapture
	}
	if g.evidence.ValidateCaptureTransition(ctx, st, copyControllerCaptureReservation(r), event) != nil || ctx.Err() != nil || g.lifetime.Err() != nil {
		return errControllerCapture
	}
	return nil
}
func (g *controllerCaptureSession) StopCaptureAuthority() {
	if g == nil || g.server == nil {
		return
	}
	if g.cancel != nil {
		g.cancel()
	}
	g.server.mu.Lock()
	defer g.server.mu.Unlock()
	g.server.tripIntegrityLocked()
}

var _ workerCaptureJournalGuard = (*controllerCaptureSession)(nil)
