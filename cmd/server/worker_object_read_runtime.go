package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/toolbroker"
	"context"
	"errors"
)

// workerObjectReadRuntime assembles the mandatory local journal with the fixed
// isolated reader runtime. A real retained-workspace creator lease, authenticated
// operation guard, same-daemon starter and independently authorized byte-release
// guard are all required. None is inferred from model metadata or substituted with
// host/SSH execution. There is no optional isolation switch or default authority.
// A source fork can replace these dependencies and invalidate the boundary.
// The runtime requires continuous authority, not just successful past exchanges.
type workerObjectReadRuntimeGuard interface {
	workerObjectReadJournalGuard
	ObjectReadAuthorityLifetime() context.Context
}
type workerObjectReadRuntime struct {
	runtime *toolbroker.ObjectReadRuntime
	guard   workerObjectReadRuntimeGuard
	creator context.Context
	owner   *workerWorkspaceObservation
	release objectcustody.ReadReleaseGuard
}

func newWorkerObjectReadRuntime(l *workerWorkspaceObservation, i workerObjectReadIntent, g workerObjectReadRuntimeGuard, release objectcustody.ReadReleaseGuard) (*workerObjectReadRuntime, error) {
	if nilWorkerObservationDependency(release) || nilWorkerObservationDependency(g) || g.ObjectReadAuthorityLifetime() == nil || g.ObjectReadAuthorityLifetime().Err() != nil {
		return nil, errWorkerProbeJournal
	}
	journal, err := newWorkerObjectReadJournal(l, i.Request, i.Image, i.Objects, g)
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || nilWorkerObservationDependency(l.runner) {
		return nil, errWorkerProbeJournal
	}
	// One trusted adapter owns both inspection and starting. A caller cannot
	// splice in a separately selected daemon or provider/host stream.
	starter, ok := l.runner.(toolbroker.ObjectReadStarter)
	if !ok || nilWorkerObservationDependency(starter) {
		return nil, errWorkerProbeJournal
	}
	plan, err := toolbroker.NewObjectReadContainerPlan(i.Request, i.Image, l.dockerRoot, i.Objects)
	if err != nil {
		return nil, err
	}
	runtime, err := toolbroker.NewObjectReadRuntime(plan, l.runner, starter, journal, release)
	if err != nil {
		return nil, err
	}
	return &workerObjectReadRuntime{runtime: runtime, guard: g, creator: l.lifetime, owner: l, release: release}, nil
}

// The caller owns and must join/close its private control transport on all paths.
// Runtime errors revoke operation authority as well as the runtime's byte-release
// authority. Success still requires exit, exact-ID cleanup, both durable journals
// and current release permission; it never means user delivery or task acceptance.
func (r *workerObjectReadRuntime) Run(ctx context.Context) (raw []byte, err error) {
	if r == nil || r.runtime == nil || nilWorkerObservationDependency(r.guard) || ctx == nil || r.guard.ObjectReadAuthorityLifetime() == nil || r.creator == nil {
		return nil, errWorkerProbeJournal
	}
	var complete func()
	defer func() {
		if complete != nil {
			defer complete()
		}
		if err != nil {
			clear(raw)
			raw = nil
			r.guard.StopObjectReadAuthority()
			if !nilWorkerObservationDependency(r.release) {
				r.release.StopReadAuthority()
			}
		}
	}()
	owned, finish, err := r.owner.beginOperation(ctx)
	if err != nil {
		return nil, err
	}
	complete = finish
	child, cancel := context.WithCancel(owned)
	defer cancel()
	lifetime := r.guard.ObjectReadAuthorityLifetime()
	if lifetime.Err() != nil || r.creator.Err() != nil {
		return nil, errWorkerProbeJournal
	}
	stop := context.AfterFunc(lifetime, cancel)
	defer stop()
	raw, err = r.runtime.Run(child)
	err = errors.Join(err, ctx.Err(), lifetime.Err(), r.creator.Err())
	if !r.owner.operationCurrent(child) {
		err = errors.Join(err, errWorkerProbeJournal, objectcustody.ErrUnconfirmed)
	}
	return raw, err
}

// recoverRetainedWorkerObjectRead assembles cleanup only. It requires fresh scoped
// authority and a local durable known-ID history; it has no starter, byte-release
// guard, reader, object deletion, storage refund or unknown-ID adoption capability.
func recoverRetainedWorkerObjectRead(ctx context.Context, l *workerWorkspaceObservation, i workerObjectReadIntent, id string, g workerObjectReadRuntimeGuard) (err error) {
	if ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(g) || g.ObjectReadAuthorityLifetime() == nil || g.ObjectReadAuthorityLifetime().Err() != nil {
		return errWorkerProbeJournal
	}
	var complete func()
	defer func() {
		if complete != nil {
			defer complete()
		}
		if err != nil {
			g.StopObjectReadAuthority()
		}
	}()
	owned, finish, err := l.beginOperation(ctx)
	if err != nil {
		return err
	}
	complete = finish
	ctx = owned
	journal, err := newWorkerObjectReadJournal(l, i.Request, i.Image, i.Objects, g)
	if err != nil {
		return err
	}
	l.mu.Lock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || nilWorkerObservationDependency(l.runner) {
		l.mu.Unlock()
		return errWorkerProbeJournal
	}
	plan, err := toolbroker.NewObjectReadContainerPlan(i.Request, i.Image, l.dockerRoot, i.Objects)
	runner := l.runner
	creator := l.lifetime
	l.mu.Unlock()
	if err != nil {
		return err
	}
	recovery, err := toolbroker.NewObjectReadRecovery(plan, runner, journal)
	if err != nil {
		return err
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	lifetime := g.ObjectReadAuthorityLifetime()
	stop := context.AfterFunc(lifetime, cancel)
	defer stop()
	stopCreator := context.AfterFunc(creator, cancel)
	defer stopCreator()
	if creator.Err() != nil {
		return errWorkerProbeJournal
	}
	return errors.Join(recovery.Reconcile(child, id), ctx.Err(), lifetime.Err(), creator.Err())
}
