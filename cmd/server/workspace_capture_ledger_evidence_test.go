package main

import (
	"agentroom/internal/room"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type syntheticLifecycleWitness struct {
	calls  int
	deny   bool
	revoke bool
}

func (w *syntheticLifecycleWitness) ValidateCaptureLifecycle(ctx context.Context, _ room.State, _ workerJobRecord, _ workerCaptureRecord, _ workerProbeEvent) error {
	w.calls++
	if w.deny || w.revoke && w.calls > 1 {
		return errControllerCapture
	}
	return ctx.Err()
}
func TestWorkspaceCaptureLedgerPhysicalStages(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"begin", "bound", "finish", "recovery", "false absence", "local mismatch", "quota denied", "revoked", "extra reference", "corrupt local", "expired"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, g, l := captureJournalFixtureScope(t)
			beginCaptureJournal(t, j)
			id := strings.Repeat("1", 64)
			stage := "begin"
			dockerMode := "vacant"
			if mode != "begin" && mode != "quota denied" && mode != "expired" {
				if j.BindCapture(t.Context(), j.intent.Request, id) != nil {
					t.Fatal("bind")
				}
				stage = "bound"
				dockerMode = "created"
			}
			if mode == "finish" || mode == "false absence" {
				if j.FinishCapture(t.Context(), j.intent.Request, id, true) != nil {
					t.Fatal("finish")
				}
				stage = "finish"
				if mode == "finish" {
					dockerMode = "absent"
				}
			}
			if mode == "recovery" {
				if j.FinishCapture(t.Context(), j.intent.Request, id, false) != nil || j.BeginCaptureRecovery(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image, id) != nil {
					t.Fatal("recovery")
				}
				stage = "recovery"
				dockerMode = ""
			}
			f := &physicalCaptureFixture{probe: g.f, request: j.intent.Request, objects: j.intent.Objects, mode: dockerMode}
			f.probe.image = j.intent.Image
			f.probe.name = "commonroom-capture-" + j.record.JobID + "-" + j.intent.Request.Challenge
			l.runner = f
			s, _, ctx := physicalControllerFixture(t, g.f)
			st := probeControllerState(t, s, ctx)
			job := st.ManagedJobs[j.record.JobID]
			b := j.intent.Request.Binding.Execution
			job.TaskID = b.TaskID
			job.TaskDispatchID = b.ExecutionID
			job.TaskDefinitionDigest = b.DefinitionDigest
			st.ManagedJobs[job.JobID] = job
			witness := &syntheticLifecycleWitness{deny: mode == "quota denied", revoke: mode == "revoked"}
			e, err := newWorkspaceCaptureLedgerEvidence(l, "worker", j.intent, witness)
			if err != nil {
				t.Fatal(err)
			}
			records, err := readWorkerCaptureRecords(j.dir)
			if err != nil {
				t.Fatal(err)
			}
			local := records[captureRecordName(j.intent)]
			events := captureControllerEvents(local)
			event := events[len(events)-1]
			if event.Stage != stage {
				t.Fatal("wrong synthetic stage")
			}
			candidate := room.WorkspaceCaptureReservation{Version: 1, CompanyID: st.Company.ID, RoomID: job.RoomID, Worker: "worker", ActorID: ownerID, Intent: copyCaptureIntent(j.intent), Events: events, Released: releasedProbeEvents(events)}
			if mode == "local mismatch" {
				candidate.Events[len(candidate.Events)-1].ID = strings.Repeat("9", 64)
				event = candidate.Events[len(candidate.Events)-1]
			}
			if mode == "extra reference" {
				f.mode = "extra reference"
			}
			if mode == "corrupt local" {
				if err = os.WriteFile(filepath.Join(j.dir, captureRecordName(j.intent)), []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "expired" {
				e.lifetime = func() context.Context { c, cancel := context.WithCancel(ctx); cancel(); return c }()
			}
			err = e.ValidateCaptureTransition(ctx, st, candidate, event)
			healthy := mode == "begin" || mode == "bound" || mode == "finish" || mode == "recovery"
			if (err == nil) != healthy {
				t.Fatal("unexpected physical lifecycle result", mode, err)
			}
			if healthy && witness.calls != 2 {
				t.Fatal("witness not rechecked")
			}
			if mode == "quota denied" && f.inspections != 0 {
				t.Fatal("inspection before quota permission")
			}
			if g.f.creates != 0 || g.f.removed || g.f.started {
				t.Fatal("evidence performed lifecycle effect")
			}
		})
	}
}
func TestWorkspaceCaptureLedgerMissingScope(t *testing.T) {
	var missing *workspaceCaptureLedgerEvidence
	if missing.ValidateCaptureTransition(t.Context(), room.State{}, room.WorkspaceCaptureReservation{}, workerProbeEvent{}) == nil {
		t.Fatal("nil evidence admitted")
	}
	if _, err := newWorkspaceCaptureLedgerEvidence(nil, "worker", workerCaptureIntent{}, nil); err == nil {
		t.Fatal("missing scope admitted")
	}
}
