package main

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

// Drive real idle admission with explicit clock ticks and mocked Docker. These
// refusal cases do not establish default-expiry audit or production admission.
func driveP0IdleAuditLifecycle(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	policy, err := readWorkerIdlePolicy(func(key string) string {
		if key == "COMMONROOM_AGENT_IDLE_SHUTDOWN" && scenario.Given["idle_feature"] == "off" {
			return "false"
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	dir, worker, docker, since := idleWorkspace(t, policy.Enabled)
	before, err := readWorkerRecord(dir, worker.record.JobID)
	if err != nil {
		t.Fatal(err)
	}
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if policy.Minutes != worker.idle.policy.Minutes {
		t.Fatal("fixture differs from deployed default interval")
	}
	secondsText := scenario.Given["idle_seconds"]
	if scenario.Given["work"] == "active" {
		secondsText = scenario.Given["elapsed_seconds"]
		worker.turn.Lock()
		defer worker.turn.Unlock()
	}
	seconds, err := strconv.Atoi(secondsText)
	if err != nil {
		t.Fatal(err)
	}
	expired, stopErr := worker.expireIdle(since.Add(time.Duration(seconds) * time.Second))
	if len(docker.pairFakeDocker.removed) != 0 {
		out.Effects = append(out.Effects, "shutdown_started")
	}
	if !docker.volume || docker.removed != 0 {
		out.Effects = append(out.Effects, "workspace_discarded")
	}
	after, err := readWorkerRecord(dir, worker.record.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		out.Effects = append(out.Effects, "worker_journal_changed")
	}
	if expired || stopErr != nil {
		out.Effects = append(out.Effects, "idle_shutdown")
	}
	if scenario.ID != "idle-feature-off" {
		out.Values["running"] = strconv.FormatBool(!worker.closed && worker.ctx.Err() == nil && len(docker.containers) == 2)
	}
	return out
}
