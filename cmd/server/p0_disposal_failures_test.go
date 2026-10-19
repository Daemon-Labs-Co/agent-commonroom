package main

import (
	"context"
	"io"
	"reflect"
	"testing"
)

// A successful remove followed by unavailable inventory must remain uncertain.
// This adapter only wraps synthetic Docker state; it cannot launch a process.
type disposalLostReplyDocker struct{ *workspaceCleanupDocker }

func (d disposalLostReplyDocker) Run(ctx context.Context, args []string, in []byte, out io.Writer) error {
	err := d.workspaceCleanupDocker.Run(ctx, args, in, out)
	if err == nil && len(args) > 1 && args[0] == "volume" && args[1] == "rm" {
		d.mode = "daemon"
	}
	return err
}

func driveP0DisposalFailure(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true, true)
	dir := t.TempDir()
	record := f.view.workerJobRecord
	if err := writeWorkerRecord(dir, record, true); err != nil {
		t.Fatal(err)
	}
	docker := newWorkspaceCleanup(t, &workerPairSupervisor{record: record})
	if scenario.ID == "discard-remove-fails" {
		docker.mode = "remove"
	} else if scenario.ID != "discard-absence-unknown" {
		t.Fatal("unsupported disposal failure")
	}
	p := f.propose(t)
	f.now.Store(p.NotBefore.UnixNano())
	original := f.s.jobFence
	f.s.jobFence = func(ctx context.Context, alias string, r workerFenceRequest) error {
		if err := original(ctx, alias, r); err != nil {
			return err
		}
		_, err := fenceWorkerJobContainers(ctx, dir, r, func(ctx context.Context, v workerJobRecord) (workerJobRecord, error) {
			if scenario.ID == "discard-absence-unknown" {
				return stopWorkerPair(ctx, v, disposalLostReplyDocker{docker})
			}
			return stopWorkerPair(ctx, v, docker)
		})
		return err
	}
	before := f.db.snapshot()
	response := f.approve(t, p)
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if response.Code >= 500 {
		out.State = "block"
	}
	st, err := (&server{db: f.db}).load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if st.WorkspaceDisposals[p.ID].Status == "unconfirmed" && !st.ManagedJobs[p.Job.JobID].ShutdownConfirmed {
		out.Effects = append(out.Effects, "cleanup_unconfirmed")
	}
	total, err := managedStorageTotal(st)
	if err != nil {
		t.Fatal(err)
	}
	if total != 42 || len(st.ManagedStorage.Releases) != 0 {
		out.Effects = append(out.Effects, "storage_reservation_released")
	}
	after := f.db.snapshot()
	if len(after) < len(before) || !reflect.DeepEqual(after[:len(before)], before) {
		out.Effects = append(out.Effects, "history_mutated")
	}
	for _, e := range after[min(len(before), len(after)):] {
		if e.kind != "workspace.disposal.claimed" && e.kind != "workspace.disposal.unconfirmed" {
			out.Effects = append(out.Effects, "unexpected:"+e.kind)
		}
	}
	stored, err := readWorkerRecord(dir, record.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.ShutdownConfirmed || stored.Pair.Workspace.Removed {
		t.Fatal("worker certified uncertain removal")
	}
	if scenario.ID == "discard-remove-fails" && (!docker.volume || docker.removed != 0) {
		t.Fatal("failed removal lost source")
	}
	if scenario.ID == "discard-absence-unknown" && (docker.volume || docker.removed != 1) {
		t.Fatal("lost confirmation did not follow actual synthetic removal")
	}
	if retry := f.approve(t, p); retry.Code != 409 || f.calls.Load() != 1 || !reflect.DeepEqual(after, f.db.snapshot()) {
		t.Fatal("uncertain disposal was retried")
	}
	return out
}

func driveP0CancelRetainsWorkspace(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	dir, s, adapter := workspacePair(t)
	if err := s.Start(t.Context(), adapter); err != nil {
		t.Fatal(err)
	}
	docker := newWorkspaceCleanup(t, s)
	s.Close()
	request := workspaceFence(s)
	request.DiscardWorkspace = false
	request.ExpectedWorkspace = nil
	result, _ := fenceWorkerJobContainers(t.Context(), dir, request, func(ctx context.Context, r workerJobRecord) (workerJobRecord, error) {
		return stopWorkerPair(ctx, r, docker)
	})
	out := pendingP0Observation{State: "block", Values: map[string]string{}}
	stored, err := readWorkerRecord(dir, s.record.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if pairRoleSettled(result.Pair.Engine) && pairRoleSettled(result.Pair.Tool) && pairRoleSettled(stored.Pair.Engine) && pairRoleSettled(stored.Pair.Tool) && len(docker.containers) == 0 {
		out.State = "allow"
		out.Effects = append(out.Effects, "fenced_pair")
	}
	if docker.volume && docker.removed == 0 && !stored.Pair.Workspace.Removed && !stored.ShutdownConfirmed {
		out.Effects = append(out.Effects, "workspace_retained")
	}
	if docker.removed != 0 || stored.Pair.Workspace.Removed {
		out.Effects = append(out.Effects, "workspace_removed")
	}
	return out
}

func driveP0DisposalHistoryProtection(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true, true)
	p := f.propose(t)
	f.now.Store(p.NotBefore.UnixNano())
	if response := f.approve(t, p); response.Code != 200 {
		t.Fatal(response.Code)
	}
	f.db.seed("feature.flag.updated", map[string]any{"name": "workspaceDisposal", "enabled": false})
	before := f.db.snapshot()
	response := mockRequest(f.h, "POST", "/api/events", `{"kind":"audit.history.purged","payload":{"all":true}}`)
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if response.Code == 400 || response.Code == 403 {
		out.State = "deny"
	}
	after := f.db.snapshot()
	if len(after) < len(before) {
		out.Effects = append(out.Effects, "history_deleted")
	} else if !reflect.DeepEqual(before, after[:len(before)]) {
		out.Effects = append(out.Effects, "history_mutated")
	}
	if len(after) > len(before) {
		for _, event := range after[min(len(before), len(after)):] {
			out.Effects = append(out.Effects, event.kind)
		}
	}
	return out
}
