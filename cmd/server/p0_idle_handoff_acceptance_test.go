package main

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func driveP0IdleArrivingWork(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	dir, worker, docker, since := idleWorkspace(t, true)
	seconds, err := strconv.Atoi(scenario.Given["idle_seconds"])
	if err != nil {
		t.Fatal(err)
	}
	before, err := readWorkerRecord(dir, worker.record.JobID)
	if err != nil {
		t.Fatal(err)
	}
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if err := worker.queueTurn(); err != nil {
		t.Fatal(err)
	}
	// Authorized work arrives just before expiry; the next tick reaches the old
	// boundary. Without the pending-turn check this would start shutdown.
	expired, tickErr := worker.expireIdle(since.Add(time.Duration(seconds+1) * time.Second))
	completedAfter := time.Now()
	worker.finishQueuedTurn()
	completedBefore := time.Now()
	worker.mu.Lock()
	reset, pending := worker.idle.since, worker.pendingTurns
	worker.mu.Unlock()
	if reset.After(since) && !reset.Before(completedAfter) && !reset.After(completedBefore) && pending == 0 {
		out.Effects = append(out.Effects, "idle_timer_reset")
	}
	early, earlyErr := worker.expireIdle(reset.Add(5*time.Minute - time.Nanosecond))
	if expired || tickErr != nil || early || earlyErr != nil || worker.closed || len(docker.pairFakeDocker.removed) != 0 {
		out.Effects = append(out.Effects, "idle_shutdown")
	}
	after, err := readWorkerRecord(dir, worker.record.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) || !docker.volume || docker.removed != 0 || len(docker.containers) != 2 {
		out.Effects = append(out.Effects, "retained_state_changed")
	}
	return out
}

func driveP0IdleOwnerInterval(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	seconds, err := strconv.Atoi(scenario.Given["owner_config_seconds"])
	if err != nil || seconds%60 != 0 {
		t.Fatal("invalid configured interval", err)
	}
	// Exercise the actual authenticated owner setting before binding its durable
	// snapshot to a separately admitted synthetic worker within the hard deadline.
	s, _, handler := idleSettingsHTTP(t)
	body, err := json.Marshal(map[string]any{"kind": "agents.idle_policy.updated", "payload": map[string]int{"minutes": seconds / 60, "revision": 0}})
	if err != nil {
		t.Fatal(err)
	}
	if response := mockRequest(handler, http.MethodPost, "/api/events", string(body)); response.Code != http.StatusOK {
		t.Fatal("owner idle setting refused", response.Code)
	}
	state, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	policy, err := idlePolicyForState(state)
	if err != nil || !policy.Enabled || state.AgentIdleSettings.Revision != 1 {
		t.Fatal("owner setting not durable", err)
	}
	f := newIdleNoticeIntervalFixture(t, policy.Minutes)
	before := f.db.snapshot()
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	announce := f.announce
	f.announce = func(ctx context.Context, record workerJobRecord) error {
		if err := announce(ctx, record); err != nil {
			return err
		}
		st, err := f.s.load(ctx)
		if err != nil {
			return err
		}
		if notice, exists := st.IdleStopNotices[record.JobID]; exists && notice.Job.IdleMinutes == policy.Minutes {
			out.Effects = append(out.Effects, "idle_announced")
		}
		return nil
	}
	idleSeconds, err := strconv.Atoi(scenario.Given["idle_seconds"])
	if err != nil {
		t.Fatal(err)
	}
	tick := f.now.Add(time.Duration(idleSeconds-seconds) * time.Second)
	if expired, err := f.worker.expireIdle(tick.Add(-time.Nanosecond)); expired || err != nil {
		t.Fatal("configured interval expired early", err)
	}
	expired, stopErr := f.worker.expireIdle(tick)
	if len(f.docker.pairFakeDocker.removed) == 2 && len(f.docker.containers) == 0 {
		out.Effects = append(out.Effects, "fenced_pair")
	}
	if !expired || stopErr == nil || !f.worker.closed || f.worker.ctx.Err() == nil {
		t.Fatal("expiry invented full shutdown proof or lost fencing")
	}
	if !f.docker.volume || f.docker.removed != 0 {
		out.Effects = append(out.Effects, "workspace_discarded")
	}
	after := f.db.snapshot()
	if len(after) < len(before) || !reflect.DeepEqual(before, after[:len(before)]) {
		out.Effects = append(out.Effects, "history_deleted")
	}
	return out
}
