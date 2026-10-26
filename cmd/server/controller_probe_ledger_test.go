package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

type syntheticControllerProbeEvidence struct {
	validate func(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error
}

func (g *syntheticControllerProbeEvidence) ValidateProbe(ctx context.Context, st room.State, r room.WorkspaceProbeReservation, e workerProbeEvent) error {
	if g.validate != nil {
		return g.validate(ctx, st, r, e)
	}
	return ctx.Err()
}
func controllerProbeFixture(t *testing.T) (*server, *mockDB, context.Context, toolbroker.WorkspaceProbeIntent) {
	t.Helper()
	s, db, _ := mockServer(t)
	job := quotaJob(1, true)
	job.InitiatorID = ownerID
	job.RoomID = "engineering"
	job.Status = "shutdown_unconfirmed"
	db.seed("worker.job.updated", job)
	i := toolbroker.WorkspaceProbeIntent{JobID: job.JobID, Repository: job.RepositoryDigest, Image: "sha256:" + strings.Repeat("d", 64), Challenge: strings.Repeat("e", 32), Volume: toolbroker.WorkspaceVolumeIdentity{Version: 1, Name: "commonroom-tool-work-" + job.JobID, Nonce: strings.Repeat("c", 32), CreatedAt: "2026-10-25T00:00:00Z"}}
	return s, db, context.WithValue(t.Context(), actorContextKey{}, ownerID), i
}
func probeControllerState(t *testing.T, s *server, ctx context.Context) room.State {
	t.Helper()
	st, e := s.load(ctx)
	if e != nil {
		t.Fatal(e)
	}
	return st
}
func TestControllerProbeLedgerRestartDedupAndQuota(t *testing.T) {
	t.Parallel()
	s, db, ctx, i := controllerProbeFixture(t)
	calls := 0
	g := &syntheticControllerProbeEvidence{validate: func(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error {
		calls++
		return nil
	}}
	j, e := newControllerProbeLedger(s, g)
	if e != nil {
		t.Fatal(e)
	}
	begin := workerProbeEvent{Stage: "begin"}
	if e = j.Acknowledge(ctx, i, begin); e != nil {
		t.Fatal(e)
	}
	st := probeControllerState(t, s, ctx)
	if n, e := controllerProbeSlots(st); e != nil || n != 1 {
		t.Fatal("probe reservation missing", e)
	}
	if admitManagedContainerQuota(st, quotaJob(9, true)) == nil || admitManagedContainerQuota(st, quotaJob(9, false)) != nil {
		t.Fatal("normal job quota omitted probe")
	}
	restarted := &server{db: db}
	other, _ := newControllerProbeLedger(restarted, g)
	before := len(db.snapshot())
	if e = other.Acknowledge(ctx, i, begin); e != nil || len(db.snapshot()) != before || calls != 2 {
		t.Fatal("duplicate acknowledgment not revalidated/deduplicated", e)
	}
	pending := i
	pending.Challenge = strings.Repeat("f", 32)
	if other.Acknowledge(ctx, pending, begin) == nil {
		t.Fatal("unresolved creator admitted twice")
	}
	id := strings.Repeat("1", 64)
	if e = other.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}); e != nil {
		t.Fatal(e)
	}
	if e = other.Acknowledge(ctx, i, workerProbeEvent{Stage: "finish", ID: id, Absent: true}); e != nil {
		t.Fatal(e)
	}
	st = probeControllerState(t, restarted, ctx)
	if n, e := controllerProbeSlots(st); e != nil || n != 0 || admitManagedContainerQuota(st, quotaJob(9, true)) != nil {
		t.Fatal("acknowledged absence did not release probe slot", e)
	}
	if e = other.Acknowledge(ctx, pending, begin); e != nil {
		t.Fatal("fresh challenge refused after cleanup", e)
	}
}
func TestControllerProbeLedgerRetainsUnknownAndRecovery(t *testing.T) {
	t.Parallel()
	for _, known := range []bool{false, true} {
		t.Run(fmt.Sprint(known), func(t *testing.T) {
			t.Parallel()
			s, _, ctx, i := controllerProbeFixture(t)
			j, _ := newControllerProbeLedger(s, &syntheticControllerProbeEvidence{})
			id := ""
			if known {
				id = strings.Repeat("1", 64)
			}
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil || j.Acknowledge(ctx, i, workerProbeEvent{Stage: "finish", ID: id}) != nil {
				t.Fatal("uncertain sequence")
			}
			st := probeControllerState(t, s, ctx)
			if n, e := controllerProbeSlots(st); e != nil || n != 1 {
				t.Fatal("uncertainty freed slot", e)
			}
			recovery := workerProbeEvent{Stage: "recovery", ID: id}
			if !known {
				if j.Acknowledge(ctx, i, recovery) == nil {
					t.Fatal("unknown ID adopted")
				}
				return
			}
			if j.Acknowledge(ctx, i, recovery) != nil || j.Acknowledge(ctx, i, recovery) != nil {
				t.Fatal("recovery intent/retry")
			}
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "recovery-finish", ID: id, Absent: true}) != nil {
				t.Fatal("cleanup outcome")
			}
			st = probeControllerState(t, s, ctx)
			if n, e := controllerProbeSlots(st); e != nil || n != 0 {
				t.Fatal("recovered slot not released", e)
			}
		})
	}
}
func TestControllerProbeLedgerAuthorityEvidenceAndFailure(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"missing actor", "agent", "disabled", "wrong room", "evidence denied", "revoked during evidence", "audit failure", "wrong repository", "wrong worker"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, i := controllerProbeFixture(t)
			if mode == "missing actor" {
				ctx = t.Context()
			}
			if mode == "agent" || mode == "disabled" || mode == "wrong room" {
				a := room.Actor{ID: "probe-human", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}}
				if mode == "agent" {
					a.Kind = "agent"
				}
				if mode == "disabled" {
					a.Enabled = false
				}
				if mode == "wrong room" {
					a.RoomIDs = []string{"elsewhere"}
				}
				db.seed("actor.revised", a)
				ctx = context.WithValue(t.Context(), actorContextKey{}, a.ID)
			}
			g := &syntheticControllerProbeEvidence{validate: func(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error {
				if mode == "evidence denied" {
					return errors.New("synthetic denied")
				}
				if mode == "revoked during evidence" {
					db.seed("actor.revised", room.Actor{ID: ownerID, Kind: "human", Enabled: false})
				}
				if mode == "audit failure" {
					db.failAt = "commit"
				}
				return nil
			}}
			if mode == "wrong repository" {
				i.Repository = strings.Repeat("f", 64)
			}
			if mode == "wrong worker" {
				job := quotaJob(1, true)
				job.InitiatorID = ownerID
				job.RoomID = "engineering"
				job.Status = "shutdown_unconfirmed"
				job.Worker = "different"
				db.seed("worker.job.updated", job)
			}
			j, _ := newControllerProbeLedger(s, g)
			before := len(db.snapshot())
			err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"})
			if mode == "wrong worker" {
				if err == nil {
					t.Fatal("changed worker binding replay accepted")
				}
				return
			}
			if err == nil {
				t.Fatal("unsafe acknowledgment")
			}
			if mode != "revoked during evidence" && len(db.snapshot()) != before {
				t.Fatal("failed admission partially charged")
			}
			if mode == "audit failure" && !s.integrityFault {
				t.Fatal("audit failure did not stop authority")
			}
		})
	}
}
func TestControllerProbeLedgerConcurrentControllersLastSlot(t *testing.T) {
	t.Parallel()
	s, db, ctx, i := controllerProbeFixture(t)
	db.seed("worker.job.updated", quotaJob(2, false))
	other := &server{db: db}
	a, _ := newControllerProbeLedger(s, &syntheticControllerProbeEvidence{})
	b, _ := newControllerProbeLedger(other, &syntheticControllerProbeEvidence{})
	second := i
	second.Challenge = strings.Repeat("f", 32)
	var wg sync.WaitGroup
	result := make(chan error, 2)
	for _, v := range []struct {
		j *controllerProbeLedger
		i toolbroker.WorkspaceProbeIntent
	}{{a, i}, {b, second}} {
		wg.Add(1)
		go func(j *controllerProbeLedger, i toolbroker.WorkspaceProbeIntent) {
			defer wg.Done()
			result <- j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"})
		}(v.j, v.i)
	}
	wg.Wait()
	close(result)
	success := 0
	for err := range result {
		if err == nil {
			success++
		} else if !errors.Is(err, errAuditConflict) && !errors.Is(err, errControllerProbe) {
			t.Fatal(err)
		}
	}
	st := probeControllerState(t, s, ctx)
	if success != 1 || len(st.WorkspaceProbes) != 1 {
		t.Fatal("last slot overallocated", success)
	}
	if s.integrityFault || other.integrityFault {
		t.Fatal("ordinary conflict tripped integrity")
	}
}
func TestControllerProbeLedgerPrivateStateAndInjection(t *testing.T) {
	t.Parallel()
	s, _, ctx, i := controllerProbeFixture(t)
	j, _ := newControllerProbeLedger(s, &syntheticControllerProbeEvidence{})
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil {
		t.Fatal("begin")
	}
	raw, e := json.Marshal(probeControllerState(t, s, ctx))
	if e != nil || strings.Contains(string(raw), i.Volume.Nonce) || strings.Contains(string(raw), i.Challenge) {
		t.Fatal("private probe state exposed")
	}
	_, _, h := mockServer(t)
	w := mockRequest(h, "POST", "/api/events", `{"kind":"workspace.probe.acknowledged","payload":{}}`)
	if w.Code < 400 {
		t.Fatal("public event injection accepted")
	}
	var missing *syntheticControllerProbeEvidence
	if _, e := newControllerProbeLedger(s, missing); e == nil {
		t.Fatal("typed nil evidence accepted")
	}
}

func TestControllerProbeLedgerCleanupWhileLaunchesStopped(t *testing.T) {
	t.Parallel()
	s, db, ctx, i := controllerProbeFixture(t)
	j, _ := newControllerProbeLedger(s, &syntheticControllerProbeEvidence{})
	id := strings.Repeat("1", 64)
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil || j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}) != nil {
		t.Fatal("begin/bound")
	}
	db.seed("agents.stop", map[string]string{"scope": "application"})
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}) == nil {
		t.Fatal("halted launch authorization replayed")
	}
	if e := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "finish", ID: id, Absent: true}); e != nil {
		t.Fatal("cleanup blocked by launch stop", e)
	}
}
func TestControllerProbeLedgerEvidenceSnapshotConflict(t *testing.T) {
	t.Parallel()
	s, db, ctx, i := controllerProbeFixture(t)
	g := &syntheticControllerProbeEvidence{validate: func(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error {
		db.seed("worker.job.updated", quotaJob(2, false))
		return nil
	}}
	j, _ := newControllerProbeLedger(s, g)
	if err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}); !errors.Is(err, errAuditConflict) || s.integrityFault {
		t.Fatal("stale evidence admitted or false incident", err)
	}
	if len(probeControllerState(t, s, ctx).WorkspaceProbes) != 0 {
		t.Fatal("stale callback partially charged")
	}
}
func TestControllerProbeLedgerRejectsForgedHistory(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"unknown absence", "wrong ID", "wrong image", "worker", "room", "company", "duplicate", "terminal", "wrong reservation key"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, _, ctx, i := controllerProbeFixture(t)
			j, _ := newControllerProbeLedger(s, &syntheticControllerProbeEvidence{})
			id := strings.Repeat("1", 64)
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil || j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}) != nil {
				t.Fatal("setup")
			}
			st := probeControllerState(t, s, ctx)
			v := controllerProbeEvent{1, ownerID, st.Company.ID, "engineering", st.ManagedJobs[i.JobID].Worker, i, workerProbeEvent{Stage: "finish", ID: id, Absent: true}}
			switch mode {
			case "unknown absence":
				v.Event.ID = ""
			case "wrong ID":
				v.Event.ID = strings.Repeat("2", 64)
			case "wrong image":
				v.Intent.Image = "sha256:" + strings.Repeat("f", 64)
			case "worker":
				v.Worker = "foreign"
			case "room":
				v.RoomID = "elsewhere"
			case "company":
				v.CompanyID = "foreign"
			case "duplicate":
				v.Event = workerProbeEvent{Stage: "bound", ID: id}
			case "terminal":
				v.Event.Stage = "recovered"
			case "wrong reservation key":
				r := st.WorkspaceProbes[controllerProbeKey(i)]
				delete(st.WorkspaceProbes, controllerProbeKey(i))
				st.WorkspaceProbes["foreign"] = r
				if _, e := controllerProbeSlots(st); e == nil {
					t.Fatal("malformed quota record accepted")
				}
				return
			}
			raw, _ := json.Marshal(v)
			if projectControllerProbe(&st, raw) == nil {
				t.Fatal("forged controller event replayed")
			}
		})
	}
}

func TestControllerProbeLedgerHumanRoomAccessRevocation(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	actor := room.Actor{ID: "probe-human", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}}
	db.seed("actor.revised", actor)
	job := quotaJob(1, true)
	job.InitiatorID = actor.ID
	job.RoomID = "engineering"
	job.Status = "shutdown_unconfirmed"
	db.seed("worker.job.updated", job)
	i := toolbroker.WorkspaceProbeIntent{JobID: job.JobID, Repository: job.RepositoryDigest, Image: "sha256:" + strings.Repeat("d", 64), Challenge: strings.Repeat("e", 32), Volume: toolbroker.WorkspaceVolumeIdentity{Version: 1, Name: "commonroom-tool-work-" + job.JobID, Nonce: strings.Repeat("c", 32), CreatedAt: "2026-10-25T00:00:00Z"}}
	ctx := context.WithValue(t.Context(), actorContextKey{}, actor.ID)
	j, _ := newControllerProbeLedger(s, &syntheticControllerProbeEvidence{})
	if err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}); err != nil {
		t.Fatal("authorized human refused", err)
	}
	actor.RoomIDs = []string{"elsewhere"}
	db.seed("actor.revised", actor)
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: strings.Repeat("1", 64)}) == nil {
		t.Fatal("revoked room access accepted")
	}
	actor.RoomIDs = []string{"engineering"}
	actor.Kind = "agent"
	db.seed("actor.revised", actor)
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: strings.Repeat("1", 64)}) == nil {
		t.Fatal("agent usurped human authority")
	}
}
