package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacemount"
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func physicalControllerFixture(t *testing.T, f *assembledProbeFixture) (*server, *mockDB, context.Context) {
	t.Helper()
	s, db, _ := mockServer(t)
	r := f.record
	job := room.ManagedJob{Version: r.Version, Execution: r.Execution, RepositoryDigest: r.Pair.Repository, InitiatorID: ownerID, CompanyID: "company", JobID: r.JobID, RoomID: r.RoomID, ParticipantID: r.ParticipantID, ConnectionID: r.ConnectionID, Worker: "worker", Provider: r.Provider, Status: "shutdown_unconfirmed", AllocatedAt: r.StartedAt, UpdatedAt: r.UpdatedAt, Deadline: r.Deadline}
	db.seed("worker.job.updated", job)
	f.durable = func(stage string) bool {
		for _, event := range db.snapshot() {
			if event.kind == controllerProbeEventKind {
				var v controllerProbeEvent
				if json.Unmarshal([]byte(event.payload), &v) == nil && v.Event.Stage == stage {
					return true
				}
			}
		}
		return false
	}
	return s, db, context.WithValue(t.Context(), actorContextKey{}, ownerID)
}
func TestWorkspaceProbePhysicalControllerLifecycle(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"healthy", "extra reference", "role returns", "late reference", "writable", "replacement", "forged proof", "cleanup failure", "lost create"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			f := newAssembledProbeFixture(t, mode)
			s, _, ctx := physicalControllerFixture(t, f)
			var escaped *workspaceProbePhysicalEvidence
			var stopped atomic.Int32
			done := make(chan error, 1)
			image := "sha256:" + strings.Repeat("d", 64)
			factory := func(lease *workerWorkspaceObservation, i toolbroker.WorkspaceProbeIntent) (workerWorkspaceProbeGuard, func(), error) {
				e, err := newWorkspaceProbePhysicalEvidence(lease, "worker", image)
				if err != nil {
					return nil, nil, err
				}
				escaped = e
				guard, err := newControllerWorkspaceProbeAuthority(ctx, s, "worker", f.record, i, e)
				if err != nil {
					return nil, nil, err
				}
				left, right := net.Pipe()
				client, err := newProbeAuthorityClient(left, f.record, i, func() { stopped.Add(1) })
				if err != nil {
					left.Close()
					right.Close()
					return nil, nil, err
				}
				go func() { done <- serveProbeAuthority(ctx, right, f.record, i, guard) }()
				return client, func() { left.Close(); <-done }, nil
			}
			proof, err := probeRetainedWorkerWorkspaceAuthority(ctx, f.dir, f.record, syntheticObservationDockerRoot, f, image, factory)
			if mode == "healthy" {
				if err != nil || !workspacemount.ValidProof(proof, f.challenge) || !f.removed || stopped.Load() != 0 {
					t.Fatal("physical lifecycle failed", err, proof, stopped.Load())
				}
				st := probeControllerState(t, s, ctx)
				if slots, err := controllerProbeSlots(st); err != nil || slots != 0 {
					t.Fatal("account slot retained after verified absence", slots, err)
				}
				records, err := readWorkerProbeRecords(f.dir)
				if err != nil || len(records) != 1 {
					t.Fatal("local journal missing", err)
				}
				for _, r := range records {
					if !workerProbeSettled(r) {
						t.Fatal("local slot not confirmed")
					}
				}
				if escaped == nil || escaped.lifetime.Err() == nil {
					t.Fatal("physical evidence outlived creator lock")
				}
				if escaped.ValidateWorkspaceProbe(ctx, st, f.record, toolbroker.WorkspaceProbeIntent{JobID: f.record.JobID, Repository: f.record.Pair.Repository, Volume: f.record.Pair.Workspace.Identity, Image: image, Challenge: f.challenge}, "observe") == nil {
					t.Fatal("escaped physical authority revived")
				}
			} else {
				if err == nil || proof != (workspacemount.Proof{}) {
					t.Fatal("unsafe physical proof released", err, proof)
				}
				if (mode == "extra reference" || mode == "role returns") && f.started {
					t.Fatal("started with foreign reference")
				}
				if mode == "lost create" && (!f.exists || f.removed) {
					t.Fatal("lost create adopted")
				}
			}
			if f.base.removed != 0 {
				t.Fatal("retained workspace deleted")
			}
		})
	}
}
func TestWorkspaceProbePhysicalEvidenceNeedsActualOwnership(t *testing.T) {
	t.Parallel()
	f := newAssembledProbeFixture(t, "healthy")
	image := "sha256:" + strings.Repeat("d", 64)
	if _, err := newWorkspaceProbePhysicalEvidence(nil, "worker", image); err == nil {
		t.Fatal("missing ownership accepted")
	}
	fabricated := &workerWorkspaceObservation{active: true, lifetime: t.Context(), dir: f.dir, dockerRoot: syntheticObservationDockerRoot, record: f.record, runner: f}
	if _, err := newWorkspaceProbePhysicalEvidence(fabricated, "worker", image); err == nil {
		t.Fatal("unowned metadata accepted")
	}
	var e *workspaceProbePhysicalEvidence
	err := withWorkerWorkspaceObservation(t.Context(), f.dir, f.record, syntheticObservationDockerRoot, f, &syntheticWorkspaceObservationGuard{}, func(lease *workerWorkspaceObservation) error {
		var err error
		e, err = newWorkspaceProbePhysicalEvidence(lease, "worker", image)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.lifetime.Err() == nil {
		t.Fatal("evidence lifetime not canceled at callback exit")
	}
}

func TestWorkspaceProbePhysicalRejectsSubstitutionAndFalseAbsence(t *testing.T) {
	t.Parallel()
	f := newAssembledProbeFixture(t, "healthy")
	s, _, ctx := physicalControllerFixture(t, f)
	image := "sha256:" + strings.Repeat("d", 64)
	err := withWorkerWorkspaceObservation(ctx, f.dir, f.record, syntheticObservationDockerRoot, f, &syntheticWorkspaceObservationGuard{}, func(l *workerWorkspaceObservation) error {
		e, err := newWorkspaceProbePhysicalEvidence(l, "worker", image)
		if err != nil {
			return err
		}
		st := probeControllerState(t, s, ctx)
		i := toolbroker.WorkspaceProbeIntent{JobID: f.record.JobID, Repository: f.record.Pair.Repository, Volume: f.record.Pair.Workspace.Identity, Image: image, Challenge: strings.Repeat("e", 32)}
		if err := e.ValidateWorkspaceProbe(ctx, st, f.record, i, "observe"); err != nil {
			t.Fatal("healthy physical observation refused", err)
		}
		foreign := i
		foreign.Image = "sha256:" + strings.Repeat("f", 64)
		if e.ValidateWorkspaceProbe(ctx, st, f.record, foreign, "observe") == nil {
			t.Fatal("unreviewed image accepted")
		}
		foreign = i
		foreign.Volume.Nonce = strings.Repeat("f", 32)
		if e.ValidateWorkspaceProbe(ctx, st, f.record, foreign, "observe") == nil {
			t.Fatal("foreign volume accepted")
		}
		job := st.ManagedJobs[i.JobID]
		job.Worker = "other"
		st.ManagedJobs[i.JobID] = job
		if e.ValidateWorkspaceProbe(ctx, st, f.record, i, "observe") == nil {
			t.Fatal("wrong authenticated worker accepted")
		}
		job.Worker = "worker"
		st.ManagedJobs[i.JobID] = job
		f.name = "commonroom-export-" + i.JobID + "-" + i.Challenge
		f.challenge = i.Challenge
		f.image = image
		f.exists = true
		events := []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: f.id}, {Stage: "finish", ID: f.id, Absent: true}}
		local := workerProbeRecord{Version: 1, Intent: i, Events: events}
		raw, _ := json.Marshal(local)
		if err := replacePrivateDurably(filepath.Join(f.dir, probeRecordName(i)), raw, syncParentDirectory); err != nil {
			return err
		}
		reservation := room.WorkspaceProbeReservation{Version: 1, CompanyID: st.Company.ID, RoomID: f.record.RoomID, Worker: "worker", ActorID: ownerID, Intent: i, Events: events, Released: true}
		if e.ValidateProbe(ctx, st, reservation, events[2]) == nil {
			t.Fatal("live exact ID accepted as absent")
		}
		f.exists = false
		if e.ValidateProbe(ctx, st, reservation, events[2]) != nil {
			t.Fatal("actual exact absence refused")
		}
		// A controller frame cannot replace the worker's durable last stage.
		local.Events = local.Events[:2]
		raw, _ = json.Marshal(local)
		if err := replacePrivateDurably(filepath.Join(f.dir, probeRecordName(i)), raw, syncParentDirectory); err != nil {
			return err
		}
		if e.ValidateProbe(ctx, st, reservation, events[2]) == nil {
			t.Fatal("unrecorded lifecycle event accepted")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
