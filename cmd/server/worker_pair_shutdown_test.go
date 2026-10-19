package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type ownedShutdownRunner struct {
	t       *testing.T
	runtime toolbroker.DockerRunner
}

func (r ownedShutdownRunner) Run(ctx context.Context, args []string, input []byte, out io.Writer) error {
	if owned, _ := ctx.Value(workerCleanupOwnershipKey{}).(bool); !owned {
		r.t.Fatal("cleanup ran before creator lease acquisition")
	}
	if discard, _ := ctx.Value(workerWorkspaceDiscardKey{}).(*toolbroker.WorkspaceVolumeIdentity); discard != nil {
		r.t.Fatal("ordinary shutdown granted workspace discard")
	}
	return r.runtime.Run(ctx, args, input, out)
}
func TestSupervisorShutdownFencesBothRolesAfterReleasingCreator(t *testing.T) {
	t.Parallel()
	dir, s, runtime := pairFixture(t)
	record, err := s.Shutdown(t.Context(), ownedShutdownRunner{t, runtime})
	if err != nil || !record.ShutdownConfirmed || len(runtime.removed) != 2 {
		t.Fatal("pair shutdown not proven", err, record.ShutdownConfirmed, runtime.removed)
	}
	if _, err = os.Stat(filepath.Join(dir, record.JobID+".cancel")); err != nil {
		t.Fatal("missing cancellation tombstone", err)
	}
	if s.ctx.Err() == nil || !s.closed {
		t.Fatal("supervisor not cancelled")
	}
	if err = s.BeginCreate(context.Background(), "engine"); err == nil {
		t.Fatal("closed supervisor revived")
	}
	again, err := s.Shutdown(t.Context(), ownedShutdownRunner{t, runtime})
	if err != nil || !again.ShutdownConfirmed || len(runtime.removed) != 2 {
		t.Fatal("idempotent stop reexecuted removal", err)
	}
}
func TestSupervisorShutdownFailureRetainsUnresolvedEvidence(t *testing.T) {
	t.Parallel()
	_, s, runtime := pairFixture(t)
	runtime.failRemove = s.record.Pair.Tool.ID
	record, err := s.Shutdown(t.Context(), ownedShutdownRunner{t, runtime})
	if err == nil || record.ShutdownConfirmed || record.Status != "shutdown_unconfirmed" {
		t.Fatal("failed removal fabricated shutdown", err)
	}
	if !record.Pair.Engine.Absent || record.Pair.Tool.Absent {
		t.Fatalf("independent role cleanup skipped or fabricated: engine=%+v tool=%+v error=%v", record.Pair.Engine, record.Pair.Tool, err)
	}
	if len(runtime.containers) != 1 {
		t.Fatal("unexpected surviving containers", len(runtime.containers))
	}
}
func TestSupervisorShutdownAlwaysRetainsPresentWorkspace(t *testing.T) {
	t.Parallel()
	_, s, fake := workspacePair(t)
	if err := s.Start(t.Context(), fake); err != nil {
		t.Fatal(err)
	}
	runtime := newWorkspaceCleanup(t, s)
	for role, id := range map[string]string{"engine": s.record.Pair.Engine.ID, "tool": s.record.Pair.Tool.ID} {
		runtime.containers[id] = pairFakeContainer{name: pairContainerName(s.record.JobID, role), execution: "separated-" + role + "-v1", repository: s.record.Pair.Repository, job: s.record.JobID}
	}
	record, err := s.Shutdown(t.Context(), ownedShutdownRunner{t, runtime})
	if err == nil || record.ShutdownConfirmed || !runtime.volume || runtime.removed != 0 {
		t.Fatal("unapproved workspace removed or shutdown claimed", err)
	}
	if !record.Pair.Engine.Absent || !record.Pair.Tool.Absent || len(runtime.containers) != 0 {
		t.Fatal("containers not removed before retention")
	}
	if record.Pair.Workspace.Removed {
		t.Fatal("workspace evidence lost")
	}
}
func TestSupervisorShutdownCancelledCleanupDoesNotClaimAbsence(t *testing.T) {
	t.Parallel()
	_, s, runtime := pairFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	record, err := s.Shutdown(ctx, ownedShutdownRunner{t, runtime})
	if err == nil || record.ShutdownConfirmed || len(runtime.removed) != 0 {
		t.Fatal("cancelled observation became proof")
	}
	if s.ctx.Err() == nil {
		t.Fatal("cancelled cleanup left supervisor running")
	}
}

func TestSupervisorRunAlwaysReconcilesTurnResults(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"success", "turn failure", "panic", "startup failure", "audit loss"} {
		t.Run(mode, func(t *testing.T) {
			_, s, adapter := workspacePair(t)
			runtime := newWorkspaceCleanup(t, s)
			runtime.volume = false // Synthetic successful inventory proves no retained volume.
			calls := 0
			if mode == "startup failure" {
				adapter.fail = "create:tool"
			}
			turn := func(ctx context.Context) error {
				calls++
				runtime.record = s.record
				for role, id := range map[string]string{"engine": s.record.Pair.Engine.ID, "tool": s.record.Pair.Tool.ID} {
					runtime.containers[id] = pairFakeContainer{name: pairContainerName(s.record.JobID, role), execution: "separated-" + role + "-v1", repository: s.record.Pair.Repository, job: s.record.JobID}
				}
				switch mode {
				case "turn failure":
					return errors.New("synthetic turn failure")
				case "panic":
					panic("synthetic driver panic")
				case "audit loss":
					s.cancel()
				}
				return nil
			}
			if mode == "panic" {
				func() {
					defer func() {
						if recover() == nil {
							t.Error("driver panic swallowed")
						}
					}()
					s.Run(t.Context(), adapter, ownedShutdownRunner{t, runtime}, turn)
				}()
				record, err := readWorkerRecord(s.dir, s.record.JobID)
				if err != nil || !record.ShutdownConfirmed || len(runtime.containers) != 0 {
					t.Fatal("panic skipped cleanup", err)
				}
			} else {
				record, err := s.Run(t.Context(), adapter, ownedShutdownRunner{t, runtime}, turn)
				if !record.ShutdownConfirmed || (mode == "success" && err != nil) || (mode != "success" && err == nil) {
					t.Fatal("lifecycle result incorrect", mode, err, record.ShutdownConfirmed)
				}
				if len(runtime.containers) != 0 {
					t.Fatal("turn return skipped container removal")
				}
			}
			if (mode == "startup failure" && calls != 0) || (mode != "startup failure" && calls != 1) {
				t.Fatal("wrong number of model turn calls", calls)
			}
		})
	}
}

func TestShutdownWaitsForActiveTurnBeforeReleasingCreator(t *testing.T) {
	t.Parallel()
	dir, s, adapter := workspacePair(t)
	runtime := newWorkspaceCleanup(t, s)
	runtime.volume = false
	entered, cancelled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	runDone := make(chan error, 1)
	go func() {
		_, err := s.Run(t.Context(), adapter, ownedShutdownRunner{t, runtime}, func(ctx context.Context) error {
			runtime.record = s.record
			close(entered)
			<-ctx.Done()
			close(cancelled)
			<-release
			return ctx.Err()
		})
		runDone <- err
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("turn not reached")
	}
	stopDone := make(chan error, 1)
	go func() { _, err := s.Shutdown(t.Context(), ownedShutdownRunner{t, runtime}); stopDone <- err }()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("turn not cancelled")
	}
	lease, free, err := workerTryLease(dir, ".job-"+s.record.JobID+".lease")
	if lease != nil {
		lease.Close()
	}
	if err != nil || free {
		t.Fatal("creator ownership released during active turn", err)
	}
	close(release)
	awaitAuditResult(t, runDone)
	awaitAuditResult(t, stopDone)
}
