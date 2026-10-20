package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// Combine actual worker cleanup, controller reconciliation and quota admission.
// The evidence-return adapter substitutes for SSH only; its result is derived
// from the worker's durable record, never from the catalog's expected decision.
func driveP0IdleCleanupUncertain(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	seconds, err := strconv.Atoi(scenario.Given["idle_seconds"])
	if err != nil || seconds < 60 || seconds >= 150 || seconds%60 != 0 {
		t.Fatal("uncertainty fixture must fit the hard job lifetime", err)
	}
	configured, err := strconv.Atoi(scenario.Given["owner_config_seconds"])
	if err != nil || configured != seconds {
		t.Fatal("uncertainty case must tick at its configured interval", err)
	}
	f := newIdleNoticeIntervalFixture(t, configured/60)
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	announce := f.announce
	f.announce = func(ctx context.Context, record workerJobRecord) error {
		if err := announce(ctx, record); err != nil {
			return err
		}
		state, err := f.s.load(ctx)
		if err != nil {
			return err
		}
		if _, exists := state.IdleStopNotices[record.JobID]; exists {
			out.Effects = append(out.Effects, "idle_announced")
		}
		return nil
	}
	f.docker.failRemove = f.worker.record.Pair.Engine.ID
	expired, cleanupErr := f.worker.expireIdle(f.now)
	if !expired || cleanupErr == nil {
		t.Fatal("injected cleanup uncertainty was not observed")
	}
	stored, err := readWorkerRecord(f.worker.dir, f.job.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if !stored.Pair.Engine.Absent && stored.Pair.Tool.Absent && len(f.docker.pairFakeDocker.removed) == 1 {
		out.Effects = append(out.Effects, "partial_pair_cleanup")
	}
	if stored.ShutdownConfirmed || !f.docker.volume || f.docker.removed != 0 {
		t.Fatal("uncertain cleanup lost source or invented absence proof")
	}
	f.s.jobFence = func(_ context.Context, alias string, request workerFenceRequest) error {
		if alias != f.job.Worker || request != managedJobFence(f.job) {
			t.Fatal("controller requested a different worker/job")
		}
		if !workerJobSettled(stored) {
			return errors.New("synthetic transport reports observed unresolved worker")
		}
		return nil
	}
	shutdownErr, auditErr := f.s.reconcileManagedJob(t.Context(), f.job)
	if shutdownErr == nil || auditErr != nil {
		t.Fatal("controller failed to retain uncertain outcome", shutdownErr, auditErr)
	}
	// Rebuild from the journal rather than relying on the original stop latch.
	restarted := &server{db: f.db}
	state, err := restarted.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	job := state.ManagedJobs[f.job.JobID]
	if job.Status == "shutdown_unconfirmed" && !job.ShutdownConfirmed {
		out.Effects = append(out.Effects, "shutdown_unconfirmed")
	}
	if restarted.agentsBlockedLocked(state, f.job.RoomID) {
		out.Effects = append(out.Effects, "execution_stopped")
	}
	before := f.db.snapshot()
	p := room.Participant{ID: f.job.ParticipantID, RoomID: f.job.RoomID, ConnectionID: f.job.ConnectionID}
	registration := restarted.managedJobContext(t.Context(), p).Value(managedJobContextKey{}).(managedJobRegistration)
	if registration.Register(f.job.Worker, f.job.Provider) != nil {
		out.State = "block"
	} else {
		out.Effects = append(out.Effects, "replacement_started")
	}
	if !reflect.DeepEqual(before, f.db.snapshot()) {
		out.Effects = append(out.Effects, "replacement_journal_changed")
	}
	// Fill the other two physical slots and test a new, independent allocation.
	// An erroneous release of the uncertain pair would let this one fit.
	other := quotaJob(2, true)
	other.CompanyID = state.Company.ID
	f.db.seed("worker.job.updated", other)
	state, err = restarted.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	next := quotaJob(20, false)
	next.CompanyID = state.Company.ID
	if admitManagedContainerQuota(state, next) == nil {
		out.Effects = append(out.Effects, "container_slots_released")
	}
	if f.worker.queueTurn() == nil {
		out.Effects = append(out.Effects, "worker_revived")
	}
	removed := len(f.docker.pairFakeDocker.removed)
	if expired, _ := f.worker.expireIdle(f.now.Add(time.Minute)); expired || removed != len(f.docker.pairFakeDocker.removed) {
		out.Effects = append(out.Effects, "cleanup_retried")
	}
	return out
}
