package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

// workerWorkspaceObservationGuard is trusted controller infrastructure. Each call
// must enforce current human/task/export authority and durably acknowledge bounded
// source observation before success. It is not a model callback or quota waiver.
// No production controller transport/guard is assembled here.
type workerWorkspaceObservationGuard interface {
	AuthorizeWorkspaceObservation(context.Context, workerJobRecord) error
}

// workerWorkspaceObservation holds the actual creator lock during a trusted
// internal callback. It proves retained-volume admission, NOT a frozen read-only
// root. A future exporter must additionally attest an isolated read-only mount.
// Never use this alone as workspacesnapshot.FrozenLease. No public endpoint or
// host root exists. A fork bypassing the lock or mount contract invalidates claims.
type workerWorkspaceObservation struct {
	mu              sync.Mutex
	active          bool
	lifetime        context.Context
	dir, dockerRoot string
	record          workerJobRecord
	runner          toolbroker.DockerRunner
	guard           workerWorkspaceObservationGuard
}

func (l *workerWorkspaceObservation) Check(ctx context.Context) error {
	if l == nil {
		return errors.New("workspace observation unavailable")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || ctx == nil || ctx.Err() != nil || l.lifetime.Err() != nil {
		return errors.New("workspace observation expired")
	}
	child, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(l.lifetime, cancel)
	defer stop()
	defer cancel()
	ctx = child
	current, err := readWorkerRecord(l.dir, l.record.JobID)
	if err != nil || !sameWorkerFenceBinding(current, l.record) || !retainedObservationRecord(current) || current.Pair.Workspace.Identity != l.record.Pair.Workspace.Identity || current.Pair.Engine != l.record.Pair.Engine || current.Pair.Tool != l.record.Pair.Tool {
		return errors.New("workspace observation binding changed")
	}
	scalar, wanted := current, l.record
	scalar.Pair = nil
	wanted.Pair = nil
	if scalar != wanted {
		return errors.New("workspace observation record changed")
	}
	latch, err := openPrivateRegular(filepath.Join(l.dir, current.JobID+".cancel"), os.O_RDONLY, 8192)
	if err != nil {
		return errors.New("workspace cancellation unverified")
	}
	raw, readErr := io.ReadAll(io.LimitReader(latch, 7))
	closeErr := latch.Close()
	if readErr != nil || closeErr != nil || string(raw) != "cancel" {
		return errors.New("workspace cancellation unverified")
	}
	// Give the guard a detached record: callbacks must not be able to mutate the
	// immutable lease binding through Pair/Workspace pointers.
	copy := current
	pair := *current.Pair
	workspace := *pair.Workspace
	pair.Workspace = &workspace
	copy.Pair = &pair
	if l.guard.AuthorizeWorkspaceObservation(ctx, copy) != nil || ctx.Err() != nil || l.lifetime.Err() != nil {
		return errors.New("workspace observation unauthorized")
	}
	for _, filter := range []string{"name=^/" + pairContainerName(current.JobID, "engine") + "$", "name=^/" + pairContainerName(current.JobID, "tool") + "$", "label=commonroom.job=" + current.JobID} {
		remaining, err := pairInventory(ctx, l.runner, filter)
		if err != nil || remaining != "" {
			return errors.New("workspace writer absence unverified")
		}
	}
	if err := toolbroker.VerifyRetainedWorkspace(ctx, l.runner, current.JobID, current.Pair.Repository, l.dockerRoot, current.Pair.Workspace.Identity); err != nil {
		return err
	}
	if ctx.Err() != nil || l.lifetime.Err() != nil {
		return errors.New("workspace observation cancelled")
	}
	return nil
}
func retainedObservationRecord(r workerJobRecord) bool {
	return validWorkerRecord(r) && r.Version == workerWorkspaceLifecycleVersion && r.Status == "shutdown_unconfirmed" && !r.ShutdownConfirmed && r.Pair != nil && r.Pair.Workspace != nil && !r.Pair.Workspace.Removed && toolbroker.ValidWorkspaceVolume(r.JobID, r.Pair.Workspace.Identity, true) && r.Pair.Engine.CreationStarted && r.Pair.Tool.CreationStarted && validPairContainerID(r.Pair.Engine.ID) && validPairContainerID(r.Pair.Tool.ID) && r.Pair.Engine.Absent && r.Pair.Tool.Absent
}

// withWorkerWorkspaceObservation never stops a live creator or acquires a host
// mount. It admits only an already-reconciled retained workspace. The persistent
// cancellation latch prevents managed recreation; the exclusive creator lease
// excludes concurrent cleanup/disposal until the callback and final checks end.
// Callback failures/cancellation retain all objects and accounting; no cleanup
// success, export receipt or release authority is inferred from this return value.
func withWorkerWorkspaceObservation(ctx context.Context, dir string, expected workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, guard workerWorkspaceObservationGuard, use func(*workerWorkspaceObservation) error) error {
	if ctx == nil || ctx.Err() != nil || !retainedObservationRecord(expected) || nilWorkerObservationDependency(runner) || nilWorkerObservationDependency(guard) || use == nil {
		return errors.New("workspace observation dependencies unavailable")
	}
	// Detach caller-owned pointers before waiting; caller must not mutate inputs
	// concurrently during invocation.
	pair := *expected.Pair
	workspace := *pair.Workspace
	pair.Workspace = &workspace
	expected.Pair = &pair
	ownership, free, err := workerTryLease(dir, ".job-"+expected.JobID+".lease")
	if err != nil || !free {
		return errors.New("workspace creator busy")
	}
	defer ownership.Close()
	l := &workerWorkspaceObservation{active: true, lifetime: ctx, dir: dir, dockerRoot: dockerRoot, record: expected, runner: runner, guard: guard}
	defer func() { l.mu.Lock(); l.active = false; l.mu.Unlock() }()
	if err = l.Check(ctx); err != nil {
		return err
	}
	if err = use(l); err != nil {
		return err
	}
	return l.Check(ctx)
}

func nilWorkerObservationDependency(v any) bool {
	if v == nil {
		return true
	}
	r := reflect.ValueOf(v)
	switch r.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return r.IsNil()
	}
	return false
}
