package main

import (
	"agentroom/internal/room"
	"strings"
	"testing"
	"time"
)

// Negative cases drive the real owner HTTP boundary and observed journal/ledger
// state. Positive disposal stays pending until exact storage credit is supported.
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
