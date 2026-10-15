package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type pairLaunchFake struct {
	t                 *testing.T
	supervisor        *workerPairSupervisor
	calls             []string
	fail              string
	cancelAfterCreate context.CancelFunc
}

func (f *pairLaunchFake) Create(ctx context.Context, identity workerPairIdentity) (string, error) {
	f.calls = append(f.calls, "create:"+identity.Role)
	record, err := readWorkerRecord(f.supervisor.dir, identity.Job)
	if err != nil {
		f.t.Fatal(err)
	}
	role := record.Pair.Tool
	if identity.Role == "engine" {
		role = record.Pair.Engine
	}
	if !role.CreationStarted || role.ID != "" {
		f.t.Fatal("create preceded durable intent")
	}
	if f.cancelAfterCreate != nil {
		f.cancelAfterCreate()
	}
	if f.fail == "create:"+identity.Role {
		return "", errors.New("synthetic failure")
	}
	return strings.Repeat(map[string]string{"tool": "c", "engine": "d"}[identity.Role], 64), nil
}
func (f *pairLaunchFake) Verify(ctx context.Context, identity workerPairIdentity, id string) error {
	f.calls = append(f.calls, "verify:"+identity.Role)
	if f.fail == "verify:"+identity.Role {
		return errors.New("synthetic policy mismatch")
	}
	return nil
}
func (f *pairLaunchFake) Inspect(ctx context.Context, id string) ([]byte, error) {
	f.calls = append(f.calls, "inspect:"+id[:1])
	if f.fail == "inspect:"+id[:1] {
		return nil, errors.New("synthetic inspection failure")
	}
	if f.fail == "resources:"+id[:1] {
		return []byte(fmt.Sprintf(`{"Id":%q,"HostConfig":{"Memory":0}}`, id)), nil
	}
	return []byte(fmt.Sprintf(`{"Id":%q,"HostConfig":{"Memory":536870912,"MemorySwap":536870912,"NanoCpus":500000000,"PidsLimit":64}}`, id)), nil
}
func (f *pairLaunchFake) Start(ctx context.Context, id string) error {
	f.calls = append(f.calls, "start:"+id[:1])
	record, err := readWorkerRecord(f.supervisor.dir, f.supervisor.record.JobID)
	if err != nil || record.Pair.Engine.ID == "" || record.Pair.Tool.ID == "" {
		f.t.Fatal("start preceded both durable identities")
	}
	if f.fail == "start:"+id[:1] {
		return errors.New("synthetic start failure")
	}
	return nil
}
func TestWorkerPairStartOwnsDurableOrdering(t *testing.T) {
	t.Parallel()
	s, err := admitWorkerPair(t.Context(), t.TempDir(), pairRequest(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	fake := &pairLaunchFake{t: t, supervisor: s}
	if err = s.Start(t.Context(), fake); err != nil {
		t.Fatal(err)
	}
	want := []string{"create:tool", "verify:tool", "create:engine", "verify:engine", "inspect:c", "inspect:d", "start:c", "start:d"}
	if !reflect.DeepEqual(fake.calls, want) || s.record.Status != "running" {
		t.Fatalf("calls %v status %s", fake.calls, s.record.Status)
	}
	if err = s.Start(t.Context(), fake); err == nil {
		t.Fatal("replayed startup")
	}
}
func TestWorkerPairStartRetainsUncertainFailure(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"create:tool", "verify:engine", "inspect:c", "inspect:d", "resources:c", "resources:d", "start:c", "start:d"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			s, err := admitWorkerPair(t.Context(), dir, pairRequest(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			fake := &pairLaunchFake{t: t, supervisor: s, fail: failure}
			if err = s.Start(t.Context(), fake); err == nil {
				t.Fatal("failure accepted")
			}
			record, err := readWorkerRecord(dir, s.record.JobID)
			if err != nil || record.ShutdownConfirmed || record.Status == "running" || !record.Pair.Tool.CreationStarted {
				t.Fatal("uncertain creation lost")
			}
			before := len(fake.calls)
			if err = s.Start(t.Context(), fake); err == nil || len(fake.calls) != before {
				t.Fatal("uncertain operation replayed")
			}
		})
	}
}
func TestWorkerPairStartStopsOnCancellation(t *testing.T) {
	t.Parallel()
	for _, latch := range []bool{false, true} {
		t.Run(map[bool]string{false: "context", true: "durable latch"}[latch], func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			dir := t.TempDir()
			s, err := admitWorkerPair(ctx, dir, pairRequest(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			stop := cancel
			if latch {
				stop = func() {
					if err := os.WriteFile(filepath.Join(dir, s.record.JobID+".cancel"), []byte("cancel"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			fake := &pairLaunchFake{t: t, supervisor: s, cancelAfterCreate: stop}
			if err = s.Start(t.Context(), fake); err == nil {
				t.Fatal("cancelled startup accepted")
			}
			for _, call := range fake.calls {
				if strings.HasPrefix(call, "start:") {
					t.Fatal("started after cancellation")
				}
			}
			if s.record.Pair.Engine.CreationStarted || s.record.ShutdownConfirmed {
				t.Fatal("cancelled creation continued or settled")
			}
		})
	}
}

type pairBlockingLaunch struct {
	*pairLaunchFake
	phase   string
	entered chan struct{}
	release chan struct{}
}

func (f *pairBlockingLaunch) Create(ctx context.Context, identity workerPairIdentity) (string, error) {
	if f.phase == "create" {
		close(f.entered)
		<-f.release
	}
	return f.pairLaunchFake.Create(ctx, identity)
}
func (f *pairBlockingLaunch) Start(ctx context.Context, id string) error {
	if f.phase == "start" {
		close(f.entered)
		<-f.release
	}
	return f.pairLaunchFake.Start(ctx, id)
}
func TestWorkerPairCloseRetainsLeaseUntilStartupReturns(t *testing.T) {
	t.Parallel()
	for _, phase := range []string{"create", "start"} {
		t.Run(phase, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			s, err := admitWorkerPair(t.Context(), dir, pairRequest(), nil)
			if err != nil {
				t.Fatal(err)
			}
			fake := &pairBlockingLaunch{pairLaunchFake: &pairLaunchFake{t: t, supervisor: s}, phase: phase, entered: make(chan struct{}), release: make(chan struct{})}
			started := make(chan error, 1)
			go func() { started <- s.Start(t.Context(), fake) }()
			<-fake.entered
			closed1, closed2 := make(chan struct{}), make(chan struct{})
			go func() { s.Close(); close(closed1) }()
			<-s.ctx.Done()
			go func() { s.Close(); close(closed2) }()
			lease, free, err := workerTryLease(dir, ".job-"+s.record.JobID+".lease")
			if lease != nil {
				lease.Close()
			}
			if err != nil || free {
				t.Error("Close released lease while adapter still running")
			}
			select {
			case <-closed1:
				t.Error("Close returned before startup ended")
			default:
			}
			select {
			case <-closed2:
				t.Error("concurrent Close returned before startup ended")
			default:
			}
			close(fake.release)
			if err := <-started; err == nil {
				t.Error("cancelled startup succeeded")
			}
			<-closed1
			<-closed2
			lease, free, err = workerTryLease(dir, ".job-"+s.record.JobID+".lease")
			if lease != nil {
				lease.Close()
			}
			if err != nil || !free {
				t.Error("Close retained lease after startup ended")
			}
		})
	}
}
