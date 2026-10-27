package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/room"
	"context"
)

// artifactReleaseEvidence independently authenticates the approved isolated custody
// reader, actual sealed object/root identity, current export/retention policy and
// physical quota/account state. It must not trust peer metadata or reenter server
// locks/the waiting read stream. No permissive/default implementation exists.
type artifactReleaseEvidence interface {
	AuthorizeRelease(context.Context, room.State, objectcustody.Receipt) error
}
type controllerArtifactReadSession struct {
	server        *server
	lifetime      context.Context
	cancel        context.CancelFunc
	actor, worker string
	receipt       objectcustody.Receipt
	evidence      artifactReleaseEvidence
	journal       *controllerArtifactJournal
}

// Trusted authenticated private setup supplies identity and exact receipt. This
// adapter creates no public endpoint or network authentication, never selects paths,
// and grants no reserve/seal/host execution authority. Canceled sessions cannot revive.
func newControllerArtifactReadSession(ctx context.Context, s *server, worker string, r objectcustody.Receipt, e artifactReleaseEvidence) (*controllerArtifactReadSession, error) {
	if ctx == nil || ctx.Err() != nil || s == nil || !workerAliasPattern.MatchString(worker) || !objectcustody.ValidReceipt(r) || nilWorkerObservationDependency(e) {
		return nil, errArtifactCustody
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	if !membershipPartPattern.MatchString(actor) {
		return nil, errArtifactCustody
	}
	lifetime, cancel := context.WithCancel(ctx)
	g := &controllerArtifactReadSession{server: s, lifetime: lifetime, cancel: cancel, actor: actor, worker: worker, receipt: r, evidence: e}
	journal, err := newControllerArtifactJournal(s, g)
	if err != nil {
		cancel()
		return nil, err
	}
	g.journal = journal
	return g, nil
}
func (g *controllerArtifactReadSession) callContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if g == nil || ctx == nil || ctx.Err() != nil || g.lifetime == nil || g.lifetime.Err() != nil {
		return nil, nil, errArtifactCustody
	}
	child, cancel := context.WithCancel(context.WithValue(ctx, actorContextKey{}, g.actor))
	stop := context.AfterFunc(g.lifetime, cancel)
	return child, func() { stop(); cancel() }, nil
}
func (g *controllerArtifactReadSession) AuthorizeRead(ctx context.Context, scope objectcustody.Scope, r objectcustody.Receipt) error {
	if g == nil || r != g.receipt || scope != g.receipt.Intent.Scope {
		return errArtifactCustody
	}
	child, cancel, err := g.callContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return g.journal.AuthorizeRead(child, scope, r)
}
func (g *controllerArtifactReadSession) AcknowledgeRead(ctx context.Context, r objectcustody.Receipt) error {
	if g == nil || r != g.receipt {
		return errArtifactCustody
	}
	child, cancel, err := g.callContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return g.journal.AcknowledgeRead(child, r)
}
func (g *controllerArtifactReadSession) pinned(ctx context.Context, st room.State, r objectcustody.Receipt) bool {
	if ctx == nil || ctx.Err() != nil || g.lifetime.Err() != nil || r != g.receipt {
		return false
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	job, ok := st.ManagedJobs[r.Intent.Scope.JobID]
	return actor == g.actor && ok && validManagedJob(job) && job.Worker == g.worker && artifactRecordedSource(st, r.Intent)
}

// Called only inside the journal's locked policy snapshot. The post-evidence load
// checks current access and sealed/reservation linkage before either metadata approval
// or durable acknowledgment. A changed audit revision is a conflict, not a grant.
func (g *controllerArtifactReadSession) AuthorizeRelease(ctx context.Context, st room.State, r objectcustody.Receipt) error {
	if !g.pinned(ctx, st, r) || nilWorkerObservationDependency(g.evidence) {
		return errArtifactCustody
	}
	if g.evidence.AuthorizeRelease(ctx, st, r) != nil || ctx.Err() != nil || g.lifetime.Err() != nil {
		return errArtifactCustody
	}
	fresh, err := g.journal.current(ctx, r.Intent, false)
	if err != nil || !g.pinned(ctx, fresh, r) {
		return errArtifactCustody
	}
	event := artifactCustodyEvent{Version: 1, ActorID: g.actor, Intent: r.Intent, Receipt: &r}
	if validateArtifactTransition(fresh, "artifact.custody.read", event) != nil {
		return errArtifactCustody
	}
	if fresh.EventCount != st.EventCount {
		return errAuditConflict
	}
	return nil
}

// A read session cannot become a writer by being passed as a full custody guard.
func (*controllerArtifactReadSession) ValidateSource(context.Context, room.State, objectcustody.Intent) error {
	return errArtifactCustody
}
func (*controllerArtifactReadSession) EstablishQuota(context.Context, room.StorageReservation) error {
	return errArtifactCustody
}
func (*controllerArtifactReadSession) ValidateSeal(context.Context, room.State, objectcustody.Receipt) error {
	return errArtifactCustody
}

// Scoped revocation is deliberately nonblocking. Persistence failures still trip
// the existing global journal integrity latch. Transfer failures never refund data.
func (g *controllerArtifactReadSession) StopReadAuthority() {
	if g != nil && g.cancel != nil {
		g.cancel()
	}
}

var _ objectcustody.ReadReleaseGuard = (*controllerArtifactReadSession)(nil)
