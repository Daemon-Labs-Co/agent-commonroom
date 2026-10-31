package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	operations      sync.WaitGroup
	active          bool
	lifetime        context.Context
	dir, dockerRoot string
	record          workerJobRecord
	runner          toolbroker.DockerRunner
	guard           workerWorkspaceObservationGuard
	ownership       *workerLease
}

// Admission and teardown share mu: once active is cleared no Add can race the
// teardown Wait. The completion callback must run only after all cleanup joins.
func (l *workerWorkspaceObservation) beginOperation(ctx context.Context) (context.Context, func(), error) {
	if l == nil || ctx == nil || ctx.Err() != nil {
		return nil, nil, errWorkerProbeJournal
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.active || l.lifetime == nil || l.lifetime.Err() != nil || l.ownership == nil || l.ownership.file == nil {
		return nil, nil, errWorkerProbeJournal
	}
	l.operations.Add(1)
	child, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(l.lifetime, cancel)
	return child, func() { stop(); cancel(); l.operations.Done() }, nil
}

func (l *workerWorkspaceObservation) operationCurrent(ctx context.Context) bool {
	if l == nil || ctx == nil || ctx.Err() != nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.active && l.lifetime != nil && l.lifetime.Err() == nil
}

func (l *workerWorkspaceObservation) Check(ctx context.Context) error {
	return l.checkProbe(ctx, nil, "")
}

// Only the assembled probe may supply an exact stopped read-only helper here.
// Ordinary Check retains the stronger zero-reference contract.
func (l *workerWorkspaceObservation) checkProbe(ctx context.Context, probe *toolbroker.WorkspaceMountProbe, probeID string) error {
	return l.checkWorkspace(ctx, probe, probeID, false)
}
func (l *workerWorkspaceObservation) checkRecovery(ctx context.Context, probe *toolbroker.WorkspaceMountProbe, probeID string) error {
	return l.checkWorkspace(ctx, probe, probeID, true)
}
func (l *workerWorkspaceObservation) checkWorkspace(ctx context.Context, probe *toolbroker.WorkspaceMountProbe, probeID string, recovery bool) error {
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
	current, err := readRetainedWorkspaceObservation(l.dir, l.record)
	if err != nil {
		return err
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
	if err := verifyRetainedWorkspaceReferences(ctx, current, l.dockerRoot, l.runner, probe, probeID, recovery); err != nil {
		return err
	}
	if ctx.Err() != nil || l.lifetime.Err() != nil {
		return errors.New("workspace observation cancelled")
	}
	return nil
}
func readRetainedWorkspaceObservation(dir string, expected workerJobRecord) (workerJobRecord, error) {
	if !retainedObservationRecord(expected) {
		return workerJobRecord{}, errControllerProbe
	}
	current, err := readWorkerRecord(dir, expected.JobID)
	if err != nil || !sameWorkerFenceBinding(current, expected) || !retainedObservationRecord(current) || current.Pair.Workspace.Identity != expected.Pair.Workspace.Identity || current.Pair.Engine != expected.Pair.Engine || current.Pair.Tool != expected.Pair.Tool {
		return workerJobRecord{}, errors.New("workspace observation binding changed")
	}
	scalar, wanted := current, expected
	scalar.Pair = nil
	wanted.Pair = nil
	if scalar != wanted {
		return workerJobRecord{}, errors.New("workspace observation record changed")
	}
	latch, err := openPrivateRegular(filepath.Join(dir, current.JobID+".cancel"), os.O_RDONLY, 8192)
	if err != nil {
		return workerJobRecord{}, errors.New("workspace cancellation unverified")
	}
	raw, readErr := io.ReadAll(io.LimitReader(latch, 7))
	closeErr := latch.Close()
	if readErr != nil || closeErr != nil || string(raw) != "cancel" {
		return workerJobRecord{}, errors.New("workspace cancellation unverified")
	}
	return current, nil
}

func verifyRetainedWorkspaceReferences(ctx context.Context, current workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, probe *toolbroker.WorkspaceMountProbe, probeID string, recovery bool) error {
	for _, filter := range []string{"name=^/" + pairContainerName(current.JobID, "engine") + "$", "name=^/" + pairContainerName(current.JobID, "tool") + "$", "label=commonroom.job=" + current.JobID} {
		remaining, err := pairInventory(ctx, runner, filter)
		expected := ""
		if strings.HasPrefix(filter, "label=") {
			expected = probeID
			if recovery && remaining == "" {
				expected = ""
			}
		}
		if err != nil || remaining != expected {
			return errors.New("workspace writer absence unverified")
		}
	}
	if probeID != "" {
		if probe == nil {
			return errors.New("workspace probe unavailable")
		}
		var verifyErr error
		if recovery {
			_, verifyErr = probe.InspectRecovery(ctx, probeID)
		} else {
			verifyErr = probe.VerifyAdmission(ctx, probeID)
		}
		if verifyErr != nil {
			return errors.New("workspace probe references unverified")
		}
	} else if err := toolbroker.VerifyRetainedWorkspace(ctx, runner, current.JobID, current.Pair.Repository, dockerRoot, current.Pair.Workspace.Identity); err != nil {
		return err
	}
	return ctx.Err()
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
	if use == nil {
		return errors.New("workspace observation callback unavailable")
	}
	return withWorkerWorkspaceOwnership(ctx, dir, expected, dockerRoot, runner, guard, func(l *workerWorkspaceObservation) error {
		if err := l.Check(ctx); err != nil {
			return err
		}
		return use(l)
	})
}

// Internal ownership primitive: use MUST first validate observation or exact-ID
// recovery before acting. Final successful use always requires zero references.
func withWorkerWorkspaceOwnership(ctx context.Context, dir string, expected workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, guard workerWorkspaceObservationGuard, use func(*workerWorkspaceObservation) error) error {
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
	lifetime, expire := context.WithCancel(ctx)
	defer expire()
	l := &workerWorkspaceObservation{ownership: ownership, active: true, lifetime: lifetime, dir: dir, dockerRoot: dockerRoot, record: expected, runner: runner, guard: guard}
	defer func() {
		// Cancel BEFORE acquiring mu: a current journal transition may hold mu
		// while awaiting a dependency that needs this cancellation to return.
		// beginOperation also checks lifetime.Err, so new admission is already
		// refused while teardown waits for the mutex. Join before creator unlock.
		expire()
		l.mu.Lock()
		l.active = false
		l.mu.Unlock()
		l.operations.Wait()
	}()
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
