package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacecapture"
	"context"
	"errors"
)

// ownedWorkerCapture joins the retained creator lease, durable worker lifecycle
// journal and supervised isolated helper. The runner MUST also own its starter;
// callers cannot splice in a different daemon or provider/host process stream.
// The supplied authorities remain mandatory independent trusted infrastructure,
// not production quota proof by themselves. No public API or fallback exists.
type ownedWorkerCapture struct {
	owner     *workerWorkspaceObservation
	runtime   *toolbroker.CaptureRuntime
	guard     workerCaptureRuntimeGuard
	authority context.Context
}

// Continuous authenticated authority is required even between protocol calls.
// A prior successful exchange or still-live creator lease cannot replace it.
type workerCaptureRuntimeGuard interface {
	workerCaptureJournalGuard
	CaptureAuthorityLifetime() context.Context
}

func newOwnedWorkerCapture(l *workerWorkspaceObservation, q workspacecapture.Request, image string, objects toolbroker.CaptureObjectVolume, guard workerCaptureRuntimeGuard, authority workspacecapture.Authority, custody objectcustody.Journal) (*ownedWorkerCapture, error) {
	if l == nil || nilWorkerObservationDependency(guard) || nilWorkerObservationDependency(authority) || nilWorkerObservationDependency(custody) {
		return nil, errControllerCapture
	}
	lifetime := guard.CaptureAuthorityLifetime()
	if lifetime == nil || lifetime.Err() != nil {
		return nil, errControllerCapture
	}
	l.mu.Lock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil || nilWorkerObservationDependency(l.runner) {
		l.mu.Unlock()
		return nil, errControllerCapture
	}
	runner, root := l.runner, l.dockerRoot
	starter, ok := runner.(toolbroker.CaptureStarter)
	l.mu.Unlock()
	if !ok || nilWorkerObservationDependency(starter) {
		return nil, errControllerCapture
	}
	// Snapshot finite request members before constructing both bindings. Trusted
	// setup must not mutate caller inputs concurrently with this constructor.
	q = copyCaptureRequest(q)
	plan, err := toolbroker.NewCaptureContainerPlan(q, image, root, objects)
	if err != nil {
		return nil, err
	}
	journal, err := newWorkerCaptureJournal(l, q, image, objects, guard)
	if err != nil {
		return nil, err
	}
	runtime, err := toolbroker.NewCaptureRuntime(plan, runner, starter, journal, authority, custody)
	if err != nil {
		return nil, err
	}
	return &ownedWorkerCapture{owner: l, runtime: runtime, guard: guard, authority: lifetime}, nil
}

// Run is one-shot even after failure. Ownership teardown cancels and joins this
// complete operation, including exact-ID cleanup, before releasing its creator
// lock. Dependencies MUST honor cancellation and join their own work; source
// forks supplying dishonest/blocking adapters invalidate that contract.
func (c *ownedWorkerCapture) Run(ctx context.Context) (report workspacecapture.Report, err error) {
	if c == nil || c.owner == nil || c.runtime == nil || nilWorkerObservationDependency(c.guard) || c.authority == nil || ctx == nil || ctx.Err() != nil {
		return workspacecapture.Report{}, errControllerCapture
	}
	l := c.owner
	owned, finish, err := l.beginOperation(ctx)
	if err != nil {
		return workspacecapture.Report{}, errControllerCapture
	}
	defer func() {
		defer finish()
		if err != nil {
			report = workspacecapture.Report{}
			c.guard.StopCaptureAuthority()
		}
	}()
	if c.authority == nil || c.authority.Err() != nil {
		return workspacecapture.Report{}, errControllerCapture
	}
	child, cancel := context.WithCancel(owned)
	defer cancel()
	stop := context.AfterFunc(c.authority, cancel)
	defer stop()
	report, err = c.runtime.Run(child)
	err = errors.Join(err, c.authority.Err())
	if !l.operationCurrent(child) {
		return workspacecapture.Report{}, errors.Join(err, errControllerCapture, objectcustody.ErrUnconfirmed)
	}
	return report, err
}

// recoverRetainedWorkerCapture can reconcile only an already journaled exact ID.
// A fresh trusted guard must independently authorize cleanup and acknowledge it
// durably. There is no starter, frozen-source authority, object journal, report,
// volume deletion or refund capability. Unknown creation outcomes remain incidents.
func recoverRetainedWorkerCapture(ctx context.Context, l *workerWorkspaceObservation, intent workerCaptureIntent, id string, guard workerCaptureRuntimeGuard) (err error) {
	if ctx == nil || ctx.Err() != nil || nilWorkerObservationDependency(guard) {
		return errControllerCapture
	}
	lifetime := guard.CaptureAuthorityLifetime()
	if lifetime == nil || lifetime.Err() != nil {
		return errControllerCapture
	}
	var complete func()
	defer func() {
		if complete != nil {
			defer complete()
		}
		if err != nil {
			guard.StopCaptureAuthority()
		}
	}()
	owned, finish, err := l.beginOperation(ctx)
	if err != nil {
		return err
	}
	complete = finish
	child, cancel := context.WithCancel(owned)
	defer cancel()
	stop := context.AfterFunc(lifetime, cancel)
	defer stop()
	intent = copyCaptureIntent(intent)
	journal, err := newWorkerCaptureJournal(l, intent.Request, intent.Image, intent.Objects, guard)
	if err != nil {
		return err
	}
	l.mu.Lock()
	runner, root := l.runner, l.dockerRoot
	l.mu.Unlock()
	if nilWorkerObservationDependency(runner) {
		return errControllerCapture
	}
	plan, err := toolbroker.NewCaptureContainerPlan(intent.Request, intent.Image, root, intent.Objects)
	if err != nil {
		return err
	}
	recovery, err := toolbroker.NewCaptureRecovery(plan, runner, journal)
	if err != nil {
		return err
	}
	err = errors.Join(recovery.Reconcile(child, id), lifetime.Err())
	if !l.operationCurrent(child) {
		err = errors.Join(err, errControllerCapture, objectcustody.ErrUnconfirmed)
	}
	return err
}
