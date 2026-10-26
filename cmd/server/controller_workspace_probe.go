package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

const controllerProbeAuthorizationKind = "workspace.probe.authorized"

// controllerWorkspaceProbeEvidence is independently authenticated physical worker
// infrastructure, never agent output. ValidateWorkspaceProbe must verify the exact
// approved worker/creator lease, retained volume, reviewed image and complete
// reference inventory at the requested observation. It must not reenter server
// locks. ValidateProbe additionally proves exact lifecycle IDs/absence before the
// ledger can release a slot. No permissive physical implementation is supplied.
type controllerWorkspaceProbeEvidence interface {
	controllerProbeEvidence
	ValidateWorkspaceProbe(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error
}

type controllerProbeAuthorization struct {
	Version      int                             `json:"version"`
	ActorID      string                          `json:"actorId"`
	CompanyID    string                          `json:"companyId"`
	Worker       string                          `json:"worker"`
	Operation    string                          `json:"operation"`
	RecordDigest string                          `json:"recordDigest"`
	Intent       toolbroker.WorkspaceProbeIntent `json:"intent"`
}

type controllerWorkspaceProbeAuthority struct {
	server        *server
	lifetime      context.Context
	actor, worker string
	record        workerJobRecord
	intent        toolbroker.WorkspaceProbeIntent
	evidence      controllerWorkspaceProbeEvidence
	ledger        *controllerProbeLedger
}

func probeWorkerRecordDigest(r workerJobRecord) string {
	raw, _ := json.Marshal(r)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func detachedProbeWorkerRecord(r workerJobRecord) workerJobRecord {
	pair := *r.Pair
	volume := *pair.Workspace
	pair.Workspace = &volume
	r.Pair = &pair
	return r
}

// The context's human principal and worker identity MUST come from authenticated
// private session setup, not decoded peer frames. The immutable intent/record are
// pinned by that setup. Calls reauthorize their principal; a canceled session may
// not be revived by passing a fresh cleanup context. New cleanup needs a separately
// authenticated session. This authorizes metadata probes only, not file export.
func newControllerWorkspaceProbeAuthority(ctx context.Context, s *server, worker string, r workerJobRecord, i toolbroker.WorkspaceProbeIntent, e controllerWorkspaceProbeEvidence) (*controllerWorkspaceProbeAuthority, error) {
	if ctx == nil || ctx.Err() != nil || s == nil || !workerAliasPattern.MatchString(worker) || nilWorkerObservationDependency(e) || !retainedObservationRecord(r) || !validWorkerProbeIntent(i) || r.JobID != i.JobID || r.Pair.Repository != i.Repository || r.Pair.Workspace.Identity != i.Volume {
		return nil, errControllerProbe
	}
	actor, _ := ctx.Value(actorContextKey{}).(string)
	if !membershipPartPattern.MatchString(actor) {
		return nil, errControllerProbe
	}
	g := &controllerWorkspaceProbeAuthority{server: s, lifetime: ctx, actor: actor, worker: worker, record: detachedProbeWorkerRecord(r), intent: i, evidence: e}
	ledger, err := newControllerProbeLedger(s, g)
	if err != nil {
		return nil, err
	}
	g.ledger = ledger
	return g, nil
}
func validControllerProbeAuthorization(st room.State, v controllerProbeAuthorization) bool {
	job, ok := st.ManagedJobs[v.Intent.JobID]
	return ok && v.Version == 1 && (v.Operation == "observe" || v.Operation == "authorize") && probeHex(v.RecordDigest, 32) && validWorkerProbeIntent(v.Intent) && validManagedJob(job) && job.Version == workerWorkspaceLifecycleVersion && job.Status == "shutdown_unconfirmed" && !job.ShutdownConfirmed && job.CompanyID == st.Company.ID && v.CompanyID == job.CompanyID && v.Worker == job.Worker && v.Intent.Repository == job.RepositoryDigest && artifactRecordedActor(st, v.ActorID, job.RoomID) && (v.ActorID == ownerID || v.ActorID == job.InitiatorID)
}
func projectControllerProbeAuthorization(st *room.State, raw []byte) error {
	var v controllerProbeAuthorization
	if json.Unmarshal(raw, &v) != nil || !validControllerProbeAuthorization(*st, v) {
		return errControllerProbe
	}
	return nil // Audit-only evidence, never a reusable capability or capacity release.
}
func (g *controllerWorkspaceProbeAuthority) callContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if g == nil || ctx == nil || ctx.Err() != nil || g.lifetime == nil || g.lifetime.Err() != nil {
		return nil, nil, errControllerProbe
	}
	child, cancel := context.WithCancel(context.WithValue(ctx, actorContextKey{}, g.actor))
	stop := context.AfterFunc(g.lifetime, cancel)
	return child, func() { stop(); cancel() }, nil
}
func (g *controllerWorkspaceProbeAuthority) current(ctx context.Context, operation string) (room.State, controllerProbeAuthorization, error) {
	st, err := g.server.load(ctx)
	v := controllerProbeAuthorization{1, g.actor, st.Company.ID, g.worker, operation, probeWorkerRecordDigest(g.record), g.intent}
	job := st.ManagedJobs[g.intent.JobID]
	if err != nil || ctx.Err() != nil || g.lifetime.Err() != nil || g.server.integrityFault || st.Incident || !artifactActor(st, ctx, job.RoomID) || g.server.agentsBlockedLocked(st, job.RoomID) || !validControllerProbeAuthorization(st, v) || !sameWorkerBinding(job, workerJobView{workerJobRecord: g.record}) || !managedTaskCurrent(st, job) || !coordinationJobAllowed(st, job) || !delegatedJobAllowed(st, job) {
		return st, v, errControllerProbe
	}
	return st, v, nil
}
func (g *controllerWorkspaceProbeAuthority) authorize(ctx context.Context, operation string) error {
	ctx, cancel, err := g.callContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	s := g.server
	ctx, storageCancel := s.storageContext(ctx)
	defer storageCancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	st, _, err := g.current(ctx, operation)
	if err != nil {
		return err
	}
	if g.evidence.ValidateWorkspaceProbe(ctx, st, detachedProbeWorkerRecord(g.record), g.intent, operation) != nil || ctx.Err() != nil {
		return errControllerProbe
	}
	fresh, again, err := g.current(ctx, operation)
	if err != nil {
		return err
	}
	if fresh.EventCount != st.EventCount {
		return errAuditConflict
	}
	// All grants are durable before the private stream acknowledges them. A local
	// lock is insufficient across controllers, so append fences the event revision.
	err = s.append(context.WithValue(ctx, expectedEventCountKey{}, fresh.EventCount), controllerProbeAuthorizationKind, again)
	if err != nil && !errors.Is(err, errAuditConflict) {
		s.tripIntegrityLocked()
	}
	return err
}
func (g *controllerWorkspaceProbeAuthority) Authorize(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error {
	if g == nil || i != g.intent {
		return errControllerProbe
	}
	return g.authorize(ctx, "authorize")
}
func (g *controllerWorkspaceProbeAuthority) AuthorizeWorkspaceObservation(ctx context.Context, r workerJobRecord) error {
	if g == nil || !retainedObservationRecord(r) || probeWorkerRecordDigest(r) != probeWorkerRecordDigest(g.record) {
		return errControllerProbe
	}
	return g.authorize(ctx, "observe")
}
func (g *controllerWorkspaceProbeAuthority) Acknowledge(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error {
	if g == nil || i != g.intent {
		return errControllerProbe
	}
	ctx, cancel, err := g.callContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return g.ledger.Acknowledge(ctx, i, e)
}

// Invoked inside the ledger's fenced authorization snapshot, so a worker change
// between calls cannot turn a session for one worker into authority over another.
func (g *controllerWorkspaceProbeAuthority) ValidateProbe(ctx context.Context, st room.State, r room.WorkspaceProbeReservation, e workerProbeEvent) error {
	job := st.ManagedJobs[g.intent.JobID]
	if ctx.Err() != nil || g.lifetime.Err() != nil || r.Intent != g.intent || r.Worker != g.worker || job.Worker != g.worker || !sameWorkerBinding(job, workerJobView{workerJobRecord: g.record}) {
		return errControllerProbe
	}
	if err := g.evidence.ValidateProbe(ctx, st, r, e); err != nil {
		return err
	}
	if ctx.Err() != nil || g.lifetime.Err() != nil {
		return errControllerProbe
	}
	return nil
}

func (g *controllerWorkspaceProbeAuthority) StopAuthority() {
	if g == nil || g.server == nil {
		return
	}
	g.server.mu.Lock()
	defer g.server.mu.Unlock()
	g.server.tripIntegrityLocked()
}

var _ workerWorkspaceProbeGuard = (*controllerWorkspaceProbeAuthority)(nil)
