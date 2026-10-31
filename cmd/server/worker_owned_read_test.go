package main

import (
	"agentroom/internal/objectcustody"
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type blockingOwnedReadGuard struct {
	lifetime                                      context.Context
	entered, canceled, release, stopping, stopped chan struct{}
}

func (g *blockingOwnedReadGuard) ObjectReadAuthorityLifetime() context.Context { return g.lifetime }
func (g *blockingOwnedReadGuard) AuthorizeObjectRead(ctx context.Context, _ workerJobRecord, _ workerObjectReadIntent, _ workerProbeEvent) error {
	close(g.entered)
	<-ctx.Done()
	close(g.canceled)
	<-g.release
	return ctx.Err()
}
func (g *blockingOwnedReadGuard) AcknowledgeObjectRead(context.Context, workerObjectReadRecord, workerProbeEvent) error {
	return errArtifactCustody
}
func (g *blockingOwnedReadGuard) StopObjectReadAuthority() { close(g.stopping); <-g.stopped }

type ownedReadRelease struct{ stopped bool }

func (g *ownedReadRelease) AuthorizeRead(context.Context, objectcustody.Scope, objectcustody.Receipt) error {
	return errArtifactCustody
}
func (g *ownedReadRelease) AcknowledgeRead(context.Context, objectcustody.Receipt) error {
	return errArtifactCustody
}
func (g *ownedReadRelease) StopReadAuthority() { g.stopped = true }

func TestOwnedReadJoinsRevocationBeforeCreatorUnlock(t *testing.T) {
	for _, recovery := range []bool{false, true} {
		name := "reader"
		if recovery {
			name = "recovery"
		}
		t.Run(name, func(t *testing.T) { t.Parallel(); checkOwnedReadJoin(t, recovery) })
	}
}
func checkOwnedReadJoin(t *testing.T, recovery bool) {
	t.Helper()
	j, _, l, capture := objectReadJournalFixture(t)
	id := strings.Repeat("1", 64)
	if recovery {
		if err := j.BeginObjectRead(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image); err != nil {
			t.Fatal(err)
		}
		if err := j.BindObjectRead(t.Context(), j.intent.Request, id); err != nil {
			t.Fatal(err)
		}
		if err := j.FinishObjectRead(t.Context(), j.intent.Request, id, false); err != nil {
			t.Fatal(err)
		}
	}
	// No daemon operation may occur: the first current-authority call blocks.
	runner := &assembledObjectReadFixture{physicalObjectReadFixture: &physicalObjectReadFixture{probe: &assembledProbeFixture{t: t, dir: l.dir, record: l.record}, intent: j.intent}, t: t}
	l.ownership.Close()
	witness := &assembledProbeFixture{t: t, dir: l.dir, record: l.record}
	guard := &blockingOwnedReadGuard{lifetime: t.Context(), entered: make(chan struct{}), canceled: make(chan struct{}), release: make(chan struct{}), stopping: make(chan struct{}), stopped: make(chan struct{})}
	release := &ownedReadRelease{}
	ownerDone, runDone := make(chan error, 1), make(chan error, 1)
	go func() {
		ownerDone <- withWorkerWorkspaceOwnership(t.Context(), l.dir, l.record, l.dockerRoot, runner, witness, func(owner *workerWorkspaceObservation) error {
			if recovery {
				go func() { runDone <- recoverRetainedWorkerObjectRead(t.Context(), owner, j.intent, id, guard) }()
			} else {
				runtime, err := newWorkerObjectReadRuntime(owner, j.intent, guard, release)
				if err != nil {
					return err
				}
				go func() { _, err := runtime.Run(t.Context()); runDone <- err }()
			}
			select {
			case <-guard.entered:
				return errors.New("synthetic owner callback exit")
			case <-time.After(2 * time.Second):
				return errors.New("reader did not begin")
			}
		})
	}()
	select {
	case <-guard.canceled:
	case err := <-ownerDone:
		t.Fatal("owner exited before reader cancellation", err)
	case <-time.After(2 * time.Second):
		t.Fatal("reader not canceled")
	}
	assertHeld := func() {
		t.Helper()
		lease, free, err := workerTryLease(l.dir, ".job-"+l.record.JobID+".lease")
		if lease != nil {
			lease.Close()
		}
		if err != nil || free {
			t.Fatal("creator unlocked before reader completion", err)
		}
		select {
		case <-ownerDone:
			t.Fatal("owner finished before reader revocation")
		default:
		}
	}
	assertHeld()
	close(guard.release)
	select {
	case <-guard.stopping:
	case <-time.After(2 * time.Second):
		t.Fatal("operation authority not revoked")
	}
	assertHeld()
	close(guard.stopped)
	select {
	case err := <-ownerDone:
		if err == nil {
			t.Fatal("owner failure lost")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("owner did not join reader")
	}
	select {
	case err := <-runDone:
		if err == nil {
			t.Fatal("reader succeeded after owner exit")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reader not finished")
	}
	if !recovery && !release.stopped {
		t.Fatal("byte-release authority survived failure")
	}
	want := 0
	if recovery {
		want = 1
	}
	if slots, err := workerObjectReadSlots(j.dir); err != nil || slots != want {
		t.Fatal("denied begin persisted a reader", slots, err)
	}
	// Assert no retained capture history was modified by the denied reader.
	records, err := readWorkerCaptureRecords(capture.dir)
	if err != nil || len(records) != 0 {
		t.Fatal("reader touched capture history", err)
	}
}

// Compile-time checks ensure the independent release boundary cannot be replaced
// by the lifecycle guard. Neither test double is a production authority.
var _ objectcustody.ReadReleaseGuard = (*ownedReadRelease)(nil)
var _ workerObjectReadRuntimeGuard = (*blockingOwnedReadGuard)(nil)
var _ toolbroker.ObjectReadStarter = (*assembledObjectReadFixture)(nil)
