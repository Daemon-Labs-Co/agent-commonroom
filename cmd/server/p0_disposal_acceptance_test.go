package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Negative cases drive the real owner HTTP boundary and observed journal/ledger
// state. Positive disposal separately exercises worker receipt and credit integration.
func driveP0DisposalRefusal(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true)
	proposalID := strings.Repeat("f", 32)
	if scenario.ID != "discard-missing" {
		p := f.propose(t)
		proposalID = p.ID
		f.now.Store(p.NotBefore.UnixNano())
	}
	switch scenario.ID {
	case "discard-wrong-actor":
		f.db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true})
		f.s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	case "discard-wrong-volume":
		f.view.Pair.Workspace.Identity.Nonce = strings.Repeat("f", 32)
	case "discard-stale":
		if response := f.post(t, map[string]string{"action": "revoke", "proposalId": proposalID}); response.Code != 200 {
			t.Fatal(response.Code)
		}
	case "discard-missing":
	default:
		t.Fatal("unimplemented disposal refusal driver")
	}
	before := len(f.db.snapshot())
	response := f.post(t, map[string]string{"action": "approve", "proposalId": proposalID, "confirm": f.job.JobID})
	out := pendingP0Observation{State: "allow", Values: map[string]string{}}
	if response.Code == 400 || response.Code == 403 || response.Code == 404 || response.Code == 409 {
		out.State = "deny"
	} else if response.Code >= 500 {
		out.State = "block"
	}
	for _, event := range f.db.snapshot()[before:] {
		out.Effects = append(out.Effects, event.kind)
	}
	if f.calls.Load() != 0 {
		out.Effects = append(out.Effects, "worker_disposal_sent")
	}
	st, err := f.s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	total, err := managedStorageTotal(st)
	if err != nil {
		t.Fatal(err)
	}
	if total != 42 {
		out.Effects = append(out.Effects, "storage_reservation_released")
	}
	return out
}

// Exercises the owner HTTP flow, durable worker journal, creator ownership,
// exact receipt validation and controller refund with mocked Docker and database.
// This is hermetic component integration, not live filesystem quota evidence.
func driveP0ApprovedDisposal(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true, true)
	dir := t.TempDir()
	record := f.view.workerJobRecord
	if err := writeWorkerRecord(dir, record, true); err != nil {
		t.Fatal(err)
	}
	supervisor := &workerPairSupervisor{record: record}
	docker := newWorkspaceCleanup(t, supervisor)
	p := f.propose(t)
	f.now.Store(p.NotBefore.UnixNano())
	out := pendingP0Observation{State: "block", Values: map[string]string{}}
	before := f.db.snapshot()
	original := f.s.jobFence
	f.s.jobFence = func(ctx context.Context, alias string, request workerFenceRequest) error {
		if err := original(ctx, alias, request); err != nil {
			return err
		}
		st, err := f.s.load(ctx)
		if err != nil {
			return err
		}
		approved := st.WorkspaceDisposals[p.ID]
		if approved.Status == "claimed" && approved.ActorID == ownerID && approved.ClaimedAt.Before(approved.ExpiresAt) && !approved.ClaimedAt.Before(approved.ReviewObservedAt.Add(disposalReviewDelay)) {
			out.Effects = append(out.Effects, "discard_approval_checked")
		}
		events := f.db.snapshot()
		if len(events) > len(before) && events[len(events)-1].kind == "workspace.disposal.claimed" {
			out.Effects = append(out.Effects, "discard_intent_durable")
		}
		result, err := fenceWorkerJobContainers(ctx, dir, request, func(ctx context.Context, r workerJobRecord) (workerJobRecord, error) {
			return stopWorkerPair(ctx, r, docker)
		})
		if docker.removed == 1 {
			out.Effects = append(out.Effects, "workspace_removed")
		}
		if err != nil {
			return err
		}
		expected, err := request.record()
		if err != nil {
			return err
		}
		if !validWorkerFenceAck(request, expected, workerFenceAcknowledgement(request, result)) {
			return errors.New("invalid disposal receipt")
		}
		stored, err := readWorkerRecord(dir, record.JobID)
		if err != nil {
			return err
		}
		if !docker.volume && stored.ShutdownConfirmed && stored.Pair.Workspace.Removed {
			out.Effects = append(out.Effects, "absence_proven")
		}
		return nil
	}
	response := f.approve(t, p)
	if response.Code == 200 {
		out.State = "allow"
	}
	st, err := (&server{db: f.db}).load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	total, err := managedStorageTotal(st)
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 && len(st.ManagedStorage.Releases) == 1 {
		out.Effects = append(out.Effects, "storage_reservation_released")
	}
	after := f.db.snapshot()
	for _, event := range after[min(len(before), len(after)):] {
		switch event.kind {
		case "workspace.disposal.claimed", "workspace.disposal.succeeded", "worker.job.updated", "storage.released":
		default:
			out.Effects = append(out.Effects, "unexpected:"+event.kind)
		}
	}
	if len(after) < len(before) {
		out.Effects = append(out.Effects, "history_deleted")
	} else {
		for i, event := range before {
			if event.kind != after[i].kind || string(event.payload) != string(after[i].payload) {
				out.Effects = append(out.Effects, "history_deleted")
				break
			}
		}
	}
	return out
}
