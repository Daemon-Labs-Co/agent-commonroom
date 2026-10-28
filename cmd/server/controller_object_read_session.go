package main

import (
	"agentroom/internal/room"
	"context"
	"slices"
)

// Evidence independently authenticates operation-specific export/cleanup authority,
// creator ownership, actual physical quotas and private persisted local history.
// It must not reenter controller/admission locks or the waiting authority stream.
// No permissive default or network authentication mechanism is supplied here.
type controllerObjectReadSessionEvidence interface {
	controllerObjectReadEvidence
	ValidateObjectReadOperation(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error
}
type controllerObjectReadSession struct {
	server        *server
	lifetime      context.Context
	cancel        context.CancelFunc
	actor, worker string
	record        workerJobRecord
	intent        workerObjectReadIntent
	evidence      controllerObjectReadSessionEvidence
	ledger        *controllerObjectReadLedger
}

// Only authenticated private setup may supply principal, worker and creator.
// This pins existing authentication; decoding a record never authenticates a peer.
// Managed reads remain isolated, with no host/SSH fallback or security toggle.
func newControllerObjectReadSession(ctx context.Context, s *server, worker string, r workerJobRecord, i workerObjectReadIntent, e controllerObjectReadSessionEvidence) (*controllerObjectReadSession, error) {
	if ctx == nil || ctx.Err() != nil || s == nil || !workerAliasPattern.MatchString(worker) || !objectReadAuthorityBinding(r, i) || nilWorkerObservationDependency(e) {
		return nil, errControllerObjectRead
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	if !membershipPartPattern.MatchString(actor) {
		return nil, errControllerObjectRead
	}
	lifetime, cancel := context.WithCancel(ctx)
	g := &controllerObjectReadSession{server: s, lifetime: lifetime, cancel: cancel, actor: actor, worker: worker, record: detachedProbeWorkerRecord(r), intent: i, evidence: e}
	ledger, err := newControllerObjectReadLedger(s, g)
	if err != nil {
		cancel()
		return nil, err
	}
	g.ledger = ledger
	return g, nil
}

// Cancellation may still record finish uncertainty under current human authority.
// It never permits confirmed absence/release, a new operation or fresh recovery.
func objectReadUncertainFinish(e workerProbeEvent) bool { return e.Stage == "finish" && !e.Absent }
func (g *controllerObjectReadSession) callContext(ctx context.Context, e workerProbeEvent) (context.Context, context.CancelFunc, error) {
	if g == nil || ctx == nil || ctx.Err() != nil || g.lifetime == nil || (g.lifetime.Err() != nil && !objectReadUncertainFinish(e)) {
		return nil, nil, errControllerObjectRead
	}
	child, cancel := context.WithCancel(context.WithValue(ctx, actorContextKey{}, g.actor))
	if objectReadUncertainFinish(e) {
		return child, cancel, nil
	}
	stop := context.AfterFunc(g.lifetime, cancel)
	return child, func() { stop(); cancel() }, nil
}
func (g *controllerObjectReadSession) pinned(ctx context.Context, st room.State, e workerProbeEvent) bool {
	job, ok := st.ManagedJobs[g.record.JobID]
	return ok && ctx.Err() == nil && (g.lifetime.Err() == nil || objectReadUncertainFinish(e)) && job.Worker == g.worker && sameWorkerBinding(job, workerJobView{workerJobRecord: g.record})
}
func (g *controllerObjectReadSession) candidate(ctx context.Context, event workerProbeEvent) (room.State, room.ObjectReadReservation, error) {
	st, v, err := g.ledger.current(ctx, g.intent, event)
	if err != nil || !g.pinned(ctx, st, event) {
		return st, room.ObjectReadReservation{}, errControllerObjectRead
	}
	candidate, _, err := validateControllerObjectReadEvent(st, v)
	if err != nil || len(candidate.Events) == 0 || candidate.Events[len(candidate.Events)-1] != event {
		return st, room.ObjectReadReservation{}, errControllerObjectRead
	}
	return st, candidate, nil
}
func (g *controllerObjectReadSession) AuthorizeObjectRead(ctx context.Context, r workerJobRecord, i workerObjectReadIntent, event workerProbeEvent) error {
	if g == nil || !objectReadAuthorityBinding(r, i) || probeWorkerRecordDigest(r) != probeWorkerRecordDigest(g.record) || i != g.intent || !captureAuthorityEvent(event) {
		return errControllerObjectRead
	}
	child, cancel, err := g.callContext(ctx, event)
	if err != nil {
		return err
	}
	defer cancel()
	child, storageCancel := g.server.storageContext(child)
	defer storageCancel()
	g.server.mu.Lock()
	defer g.server.mu.Unlock()
	st, _, err := g.candidate(child, event)
	if err != nil {
		return err
	}
	if g.evidence.ValidateObjectReadOperation(child, st, detachedProbeWorkerRecord(g.record), g.intent, event) != nil || child.Err() != nil {
		return errControllerObjectRead
	}
	fresh, _, err := g.candidate(child, event)
	if err != nil {
		return err
	}
	if fresh.EventCount != st.EventCount {
		return errAuditConflict
	}
	return nil
}
func objectReadAuthorityRecord(r workerObjectReadRecord, i workerObjectReadIntent, e workerProbeEvent) bool {
	return validWorkerObjectReadRecord(r) && r.Intent == i && len(r.Events) > 0 && r.Events[len(r.Events)-1] == e && captureAuthorityEvent(e)
}
func (g *controllerObjectReadSession) AcknowledgeObjectRead(ctx context.Context, local workerObjectReadRecord, event workerProbeEvent) error {
	if g == nil || !objectReadAuthorityRecord(local, g.intent, event) {
		return errControllerObjectRead
	}
	child, cancel, err := g.callContext(ctx, event)
	if err != nil {
		return err
	}
	defer cancel()
	ledger, err := newControllerObjectReadLedger(g.server, &objectReadSessionAcknowledgment{g, copyObjectReadRecord(local)})
	if err != nil {
		return err
	}
	return ledger.Acknowledge(child, g.intent, event)
}

type objectReadSessionAcknowledgment struct {
	session *controllerObjectReadSession
	local   workerObjectReadRecord
}

func (a *objectReadSessionAcknowledgment) ValidateObjectReadTransition(ctx context.Context, st room.State, r room.ObjectReadReservation, e workerProbeEvent) error {
	if r.Intent != a.local.Intent || !slices.Equal(r.Events, a.local.Events) {
		return errControllerObjectRead
	}
	return a.session.ValidateObjectReadTransition(ctx, st, r, e)
}
func (g *controllerObjectReadSession) ValidateObjectReadTransition(ctx context.Context, st room.State, r room.ObjectReadReservation, event workerProbeEvent) error {
	if !g.pinned(ctx, st, event) || r.Worker != g.worker || r.Intent != g.intent {
		return errControllerObjectRead
	}
	r.Events = slices.Clone(r.Events)
	if g.evidence.ValidateObjectReadTransition(ctx, st, r, event) != nil || ctx.Err() != nil || (!objectReadUncertainFinish(event) && g.lifetime.Err() != nil) {
		return errControllerObjectRead
	}
	return nil
}

// Scoped revocation is nonblocking: it cannot be mistaken for cleanup evidence.
// Actual controller persistence faults independently trip the global integrity latch.
func (g *controllerObjectReadSession) StopObjectReadAuthority() {
	if g != nil && g.cancel != nil {
		g.cancel()
	}
}

var _ workerObjectReadJournalGuard = (*controllerObjectReadSession)(nil)
