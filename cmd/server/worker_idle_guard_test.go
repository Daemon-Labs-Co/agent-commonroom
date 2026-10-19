package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkerIdleDeploymentPolicy(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		flag, minutes string
		enabled       bool
		want          int
		invalid       bool
	}{
		{"", "", true, 5, false}, {"true", "1", true, 1, false}, {"false", "60", false, 60, false}, {"false", "0", false, 0, true}, {"yes", "5", false, 0, true}, {"true", "-1", false, 0, true}, {"true", "61", false, 0, true}, {"true", "1.5", false, 0, true},
	} {
		p, err := readWorkerIdlePolicy(func(k string) string {
			if k == "COMMONROOM_AGENT_IDLE_SHUTDOWN" {
				return tc.flag
			}
			return tc.minutes
		})
		if (err != nil) != tc.invalid || err == nil && (p.Enabled != tc.enabled || p.Minutes != tc.want) {
			t.Fatal(tc, p, err)
		}
	}
}
func idleWorkspace(t *testing.T, enabled bool) (string, *workerPairSupervisor, *workspaceCleanupDocker, time.Time) {
	t.Helper()
	dir, s, adapter := workspacePair(t)
	// The cleanup runner sees the completed immutable startup record below.
	docker := newWorkspaceCleanup(t, s)
	if err := s.configureIdle(workerIdlePolicy{enabled, 5}, ownedShutdownRunner{t, docker}); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	docker.record = s.record
	for role, id := range map[string]string{"engine": s.record.Pair.Engine.ID, "tool": s.record.Pair.Tool.ID} {
		docker.containers[id] = pairFakeContainer{name: pairContainerName(s.record.JobID, role), execution: "separated-" + role + "-v1", repository: s.record.Pair.Repository, job: s.record.JobID}
	}
	s.mu.Lock()
	since := time.Now()
	s.idle.since = since
	s.mu.Unlock()
	t.Cleanup(s.Close)
	return dir, s, docker, since
}
func TestWorkerIdleGuardExactBoundaryAndRetention(t *testing.T) {
	t.Parallel()
	dir, s, docker, since := idleWorkspace(t, true)
	if expired, err := s.expireIdle(since.Add(5*time.Minute - time.Nanosecond)); expired || err != nil {
		t.Fatal("early idle shutdown", err)
	}
	expired, err := s.expireIdle(since.Add(5 * time.Minute))
	// Present source intentionally prevents full-job shutdown confirmation.
	if !expired || err == nil || !s.closed || s.ctx.Err() == nil || !docker.volume || docker.removed != 0 {
		t.Fatal("idle shutdown violated retention", expired, err)
	}
	record, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if len(docker.pairFakeDocker.removed) != 2 {
		t.Fatal("idle stop did not remove both containers")
	}
	if record.IdleShutdownRequestedAt.IsZero() || record.IdleTimeoutMinutes != 5 || record.ShutdownConfirmed || !pairRoleSettled(record.Pair.Engine) || !pairRoleSettled(record.Pair.Tool) {
		t.Fatal("idle intent/absence evidence missing")
	}
	if _, err := os.Stat(filepath.Join(dir, record.JobID+".cancel")); err != nil {
		t.Fatal("idle stop lacks durable tombstone")
	}
	if expired, _ := s.expireIdle(since.Add(6 * time.Minute)); expired {
		t.Fatal("idle cleanup repeated")
	}
	if err := s.BeginCreate(context.Background(), "engine"); err == nil {
		t.Fatal("idle job revived")
	}
}
func TestWorkerIdleGuardDisabledAndBusy(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"disabled", "turn", "queued", "startup"} {
		t.Run(mode, func(t *testing.T) {
			_, s, docker, since := idleWorkspace(t, mode != "disabled")
			switch mode {
			case "turn":
				s.turn.Lock()
				defer s.turn.Unlock()
			case "startup":
				s.startup.Lock()
				defer s.startup.Unlock()
			case "queued":
				if err := s.queueTurn(); err != nil {
					t.Fatal(err)
				}
				defer s.finishQueuedTurn()
			}
			if expired, err := s.expireIdle(since.Add(6 * time.Minute)); expired || err != nil || s.closed || !docker.volume || docker.removed != 0 {
				t.Fatal("idle guard interrupted busy/disabled supervisor", mode, err)
			}
		})
	}
}
func TestWorkerIdlePersistenceFailureHasNoRuntimeEffects(t *testing.T) {
	t.Parallel()
	dir, s, docker, since := idleWorkspace(t, true)
	// Deterministic private-journal failure without changing global permissions.
	if err := os.Remove(filepath.Join(dir, s.record.JobID+".json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, s.record.JobID+".json"), 0700); err != nil {
		t.Fatal(err)
	}
	expired, err := s.expireIdle(since.Add(5 * time.Minute))
	if !expired || err == nil || !s.closed || docker.removed != 0 || len(docker.pairFakeDocker.removed) != 0 {
		t.Fatal("unrecorded idle action reached runtime")
	}
}
func TestWorkerIdlePolicyCannotBeRebound(t *testing.T) {
	t.Parallel()
	_, s, docker, _ := idleWorkspace(t, true)
	if s.configureIdle(workerIdlePolicy{false, 1}, docker) == nil {
		t.Fatal("running policy overwritten")
	}
	if s.configureIdle(workerIdlePolicy{true, 0}, docker) == nil {
		t.Fatal("invalid policy accepted")
	}
}

func TestWorkerIdleQueuedArrivalResetsFullInterval(t *testing.T) {
	t.Parallel()
	_, s, _, since := idleWorkspace(t, true)
	if err := s.queueTurn(); err != nil {
		t.Fatal(err)
	}
	if expired, err := s.expireIdle(since.Add(5 * time.Minute)); expired || err != nil {
		t.Fatal("queued work interrupted")
	}
	s.finishQueuedTurn()
	s.mu.Lock()
	finished := s.idle.since
	s.mu.Unlock()
	if expired, err := s.expireIdle(finished.Add(5*time.Minute - time.Nanosecond)); expired || err != nil {
		t.Fatal("full idle interval not reset")
	}
	if expired, _ := s.expireIdle(finished.Add(5 * time.Minute)); !expired {
		t.Fatal("reset idle interval never expires")
	}
}
func TestRunSnapshotsWorkerIdleConfiguration(t *testing.T) {
	for _, mode := range []string{"true", "false", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("COMMONROOM_AGENT_IDLE_SHUTDOWN", mode)
			t.Setenv("COMMONROOM_AGENT_IDLE_MINUTES", "2")
			_, s, adapter := workspacePair(t)
			docker := newWorkspaceCleanup(t, s)
			ran := false
			_, err := s.Run(t.Context(), adapter, ownedShutdownRunner{t, docker}, func(context.Context) error {
				ran = true
				s.mu.Lock()
				defer s.mu.Unlock()
				if s.idle == nil || s.idle.policy.Minutes != 2 || s.idle.policy.Enabled != (mode == "true") || s.idle.armed != (mode == "true") {
					t.Fatal("deployment policy not enforced")
				}
				docker.record = s.record
				return nil
			})
			if err == nil {
				t.Fatal("retained workspace must not certify full shutdown")
			}
			if mode == "invalid" && (ran || len(adapter.calls) != 0) {
				t.Fatal("invalid policy launched work")
			}
			if mode != "invalid" && !ran {
				t.Fatal("valid policy blocked work", err)
			}
		})
	}
}

func TestWorkerIdleCleanupUncertaintyCannotRevivePair(t *testing.T) {
	t.Parallel()
	dir, s, docker, since := idleWorkspace(t, true)
	docker.failRemove = s.record.Pair.Engine.ID
	expired, err := s.expireIdle(since.Add(5 * time.Minute))
	if !expired || err == nil {
		t.Fatal("cleanup uncertainty treated as success")
	}
	stored, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.ShutdownConfirmed || stored.Pair.Engine.Absent || !stored.Pair.Tool.Absent || !docker.volume {
		t.Fatal("uncertain cleanup proof/source lost")
	}
	if s.queueTurn() == nil {
		t.Fatal("replacement turn admitted after uncertainty")
	}
	if expired, _ := s.expireIdle(since.Add(6 * time.Minute)); expired {
		t.Fatal("uncertain cleanup automatically retried")
	}
}

func TestWorkerIdleEvidenceValidation(t *testing.T) {
	t.Parallel()
	_, s, _, since := idleWorkspace(t, true)
	s.mu.Lock()
	original := s.record
	s.mu.Unlock()
	for _, mode := range []string{"missing-time", "missing-interval", "out-of-range", "before-admission"} {
		record := original
		record.DiagnosticCode = "idle_shutdown_requested"
		record.IdleShutdownRequestedAt = since
		record.IdleTimeoutMinutes = 5
		switch mode {
		case "missing-time":
			record.IdleShutdownRequestedAt = time.Time{}
		case "missing-interval":
			record.IdleTimeoutMinutes = 0
		case "out-of-range":
			record.IdleTimeoutMinutes = 61
		case "before-admission":
			record.IdleShutdownRequestedAt = record.StartedAt.Add(-time.Second)
		}
		if validWorkerRecord(record) {
			t.Fatal("invalid idle evidence accepted", mode)
		}
	}
}
