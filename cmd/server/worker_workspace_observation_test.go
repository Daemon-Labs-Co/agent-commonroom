package main

import (
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type syntheticWorkspaceObservationGuard struct {
	calls  int
	deny   bool
	mutate func(workerJobRecord)
}

func (g *syntheticWorkspaceObservationGuard) AuthorizeWorkspaceObservation(ctx context.Context, r workerJobRecord) error {
	g.calls++
	if g.deny {
		return errors.New("synthetic authorization denied")
	}
	if g.mutate != nil {
		g.mutate(r)
	}
	return ctx.Err()
}
func retainedObservationFixture(t *testing.T) (string, workerJobRecord, *workspaceCleanupDocker) {
	t.Helper()
	dir, s, adapter := workspacePair(t)
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	runtime := newWorkspaceCleanup(t, s)
	for role, id := range map[string]string{"engine": s.record.Pair.Engine.ID, "tool": s.record.Pair.Tool.ID} {
		runtime.containers[id] = pairFakeContainer{name: pairContainerName(s.record.JobID, role), execution: "separated-" + role + "-v1", repository: s.record.Pair.Repository, job: s.record.JobID}
	}
	r, err := s.Shutdown(t.Context(), runtime)
	if err == nil || !retainedObservationRecord(r) || len(runtime.pairFakeDocker.removed) != 2 || len(runtime.containers) != 0 || runtime.removed != 0 {
		t.Fatal("fixture did not reconcile exact real role IDs while retaining volume", err)
	}
	return dir, r, runtime
}

const syntheticObservationDockerRoot = "/home/synthetic/.local/share/docker"

func TestWorkspaceObservationHoldsCreatorLockAndExpires(t *testing.T) {
	t.Parallel()
	dir, r, runtime := retainedObservationFixture(t)
	guard := &syntheticWorkspaceObservationGuard{}
	var escaped *workerWorkspaceObservation
	err := withWorkerWorkspaceObservation(t.Context(), dir, r, syntheticObservationDockerRoot, runtime, guard, func(l *workerWorkspaceObservation) error {
		escaped = l
		competing, free, err := workerTryLease(dir, ".job-"+r.JobID+".lease")
		if competing != nil {
			defer competing.Close()
		}
		if err != nil || free {
			t.Fatal("observation did not hold creator ownership", err)
		}
		if err = l.Check(t.Context()); err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil || guard.calls != 3 {
		t.Fatal("healthy admission/rechecks failed", err, guard.calls)
	}
	if escaped.Check(t.Context()) == nil {
		t.Fatal("escaped observation revived")
	}
	lease, free, err := workerTryLease(dir, ".job-"+r.JobID+".lease")
	if lease != nil {
		defer lease.Close()
	}
	if err != nil || !free {
		t.Fatal("creator lock leaked", err)
	}
	stored, err := readWorkerRecord(dir, r.JobID)
	if err != nil || stored.ShutdownConfirmed || stored.Pair.Workspace.Removed || runtime.removed != 0 {
		t.Fatal("observation claimed disposal", err)
	}
}
func TestWorkspaceObservationRefusesUnsafeAdmission(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"deny", "nil-guard", "typed-nil-runner", "typed-nil", "nil-use", "busy", "no-latch", "bad-latch", "running", "role-unsettled", "changed-volume", "missing-volume", "writer", "foreign-role-name", "daemon", "replacement", "referenced", "extra"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir, r, runtime := retainedObservationFixture(t)
			guard := &syntheticWorkspaceObservationGuard{}
			var authority workerWorkspaceObservationGuard = guard
			var runner toolbroker.DockerRunner = runtime
			called := false
			use := func(*workerWorkspaceObservation) error { called = true; return nil }
			switch mode {
			case "deny":
				guard.deny = true
			case "typed-nil-runner":
				var missing *workspaceCleanupDocker
				runner = missing
			case "nil-guard":
				authority = nil
			case "typed-nil":
				var nilGuard *syntheticWorkspaceObservationGuard
				authority = nilGuard
			case "nil-use":
				use = nil
			case "busy":
				lease, free, err := workerTryLease(dir, ".job-"+r.JobID+".lease")
				if err != nil || !free {
					t.Fatal(err)
				}
				defer lease.Close()
			case "no-latch":
				if err := os.Remove(filepath.Join(dir, r.JobID+".cancel")); err != nil {
					t.Fatal(err)
				}
			case "bad-latch":
				if err := os.WriteFile(filepath.Join(dir, r.JobID+".cancel"), []byte("invalid"), 0600); err != nil {
					t.Fatal(err)
				}
			case "running":
				r.Status = "running"
			case "role-unsettled":
				r.Pair.Tool.Absent = false
			case "changed-volume":
				r.Pair.Workspace.Identity.Nonce = strings.Repeat("f", 32)
			case "missing-volume":
				runtime.volume = false
			case "writer":
				runtime.containers[strings.Repeat("f", 64)] = pairFakeContainer{job: r.JobID, name: "unknown"}
			case "foreign-role-name":
				runtime.containers[strings.Repeat("f", 64)] = pairFakeContainer{job: "foreign", name: pairContainerName(r.JobID, "tool")}
			default:
				runtime.mode = mode
			}
			if err := withWorkerWorkspaceObservation(t.Context(), dir, r, syntheticObservationDockerRoot, runner, authority, use); err == nil || called {
				t.Fatal("unsafe source observation admitted", err, called)
			}
			if runtime.removed != 0 {
				t.Fatal("observation erased data")
			}
		})
	}
}
func TestWorkspaceObservationRechecksAfterUseAndRetainsFailures(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"revoked", "replacement", "attached", "record-change", "record-metadata", "callback-error", "cancelled", "panic"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			dir, r, runtime := retainedObservationFixture(t)
			guard := &syntheticWorkspaceObservationGuard{}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var escaped *workerWorkspaceObservation
			invoke := func() error {
				return withWorkerWorkspaceObservation(ctx, dir, r, syntheticObservationDockerRoot, runtime, guard, func(l *workerWorkspaceObservation) error {
					escaped = l
					switch mode {
					case "revoked":
						guard.deny = true
					case "replacement":
						runtime.mode = "replacement"
					case "attached":
						runtime.mode = "referenced"
					case "record-metadata":
						changed, err := readWorkerRecord(dir, r.JobID)
						if err != nil {
							t.Fatal(err)
						}
						changed.UpdatedAt = changed.UpdatedAt.Add(time.Second)
						if err = writeWorkerRecord(dir, changed, false); err != nil {
							t.Fatal(err)
						}
					case "record-change":
						changed, err := readWorkerRecord(dir, r.JobID)
						if err != nil {
							t.Fatal(err)
						}
						changed.Pair.Workspace.Identity.Nonce = strings.Repeat("f", 32)
						if err = writeWorkerRecord(dir, changed, false); err != nil {
							t.Fatal(err)
						}
					case "callback-error":
						return errors.New("synthetic callback failure")
					case "cancelled":
						cancel()
						if l.Check(context.Background()) == nil {
							t.Fatal("fresh context revived cancelled observation")
						}
					case "panic":
						panic("synthetic callback panic")
					}
					return nil
				})
			}
			if mode == "panic" {
				func() {
					defer func() {
						if recover() == nil {
							t.Fatal("missing callback panic")
						}
					}()
					_ = invoke()
				}()
			} else if err := invoke(); err == nil {
				t.Fatal("post-use uncertainty accepted")
			}
			if escaped == nil || escaped.Check(t.Context()) == nil || runtime.removed != 0 {
				t.Fatal("failed observation retained authority or deleted data")
			}
			lease, free, err := workerTryLease(dir, ".job-"+r.JobID+".lease")
			if lease != nil {
				defer lease.Close()
			}
			if err != nil || !free {
				t.Fatal("failed observation leaked lock", err)
			}
		})
	}
}
func TestWorkspaceObservationGuardCannotMutateLeaseBinding(t *testing.T) {
	t.Parallel()
	dir, r, runtime := retainedObservationFixture(t)
	guard := &syntheticWorkspaceObservationGuard{mutate: func(copy workerJobRecord) {
		copy.Pair.Workspace.Identity.Nonce = strings.Repeat("f", 32)
		copy.Pair.Tool.ID = strings.Repeat("f", 64)
	}}
	if err := withWorkerWorkspaceObservation(t.Context(), dir, r, syntheticObservationDockerRoot, runtime, guard, func(l *workerWorkspaceObservation) error { return l.Check(t.Context()) }); err != nil {
		t.Fatal("guard record aliases immutable binding", err)
	}
}
