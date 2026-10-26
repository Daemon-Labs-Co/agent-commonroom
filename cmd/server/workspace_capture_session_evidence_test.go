package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type syntheticCaptureOperationWitness struct {
	syntheticLifecycleWitness
	operations int
	mode       string
	change     func()
}

func (w *syntheticCaptureOperationWitness) ValidateCaptureOperation(ctx context.Context, _ room.State, r workerJobRecord, i workerCaptureIntent, _ workerProbeEvent) error {
	w.operations++
	if w.mode == "denied" || w.mode == "revoked" && w.operations > 1 {
		return errControllerCapture
	}
	if w.change != nil && w.operations == 1 {
		w.change()
	}
	r.Pair.Workspace.Removed = true
	i.Request.Members[0].Name = "mutated"
	return ctx.Err()
}
func TestWorkspaceCaptureOperationPhysicalHistory(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"begin-pre", "begin-post", "bound-pre", "bound-post", "finish-pre", "finish-post", "recovery-pre", "recovery-post", "denied", "revoked", "extra reference", "false absence", "corrupt history", "changed history", "record changed", "foreign record", "foreign worker", "expired", "stopped", "old event", "unknown adoption"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, g, l := captureJournalFixtureScope(t)
			id := strings.Repeat("1", 64)
			event := workerProbeEvent{Stage: "bound", ID: id}
			physical := "created"
			if mode != "begin-pre" {
				beginCaptureJournal(t, j)
			}
			if mode == "begin-pre" || mode == "begin-post" {
				event = workerProbeEvent{Stage: "begin"}
				physical = "vacant"
			}
			if mode == "bound-post" || strings.HasPrefix(mode, "finish-") || strings.HasPrefix(mode, "recovery-") || mode == "false absence" || mode == "old event" {
				if j.BindCapture(t.Context(), j.intent.Request, id) != nil {
					t.Fatal("bind")
				}
			}
			if strings.HasPrefix(mode, "finish-") || mode == "false absence" {
				event = workerProbeEvent{Stage: "finish", ID: id, Absent: true}
				physical = "absent"
			}
			if mode == "finish-post" || mode == "old event" {
				if j.FinishCapture(t.Context(), j.intent.Request, id, true) != nil {
					t.Fatal("finish")
				}
			}
			if strings.HasPrefix(mode, "recovery-") {
				if j.FinishCapture(t.Context(), j.intent.Request, id, false) != nil {
					t.Fatal("uncertain")
				}
				event = workerProbeEvent{Stage: "recovery", ID: id}
				physical = ""
				if mode == "recovery-post" && j.BeginCaptureRecovery(t.Context(), j.intent.Request, j.intent.Objects, j.intent.Image, id) != nil {
					t.Fatal("recovery")
				}
			}
			if mode == "unknown adoption" {
				if j.FinishCapture(t.Context(), j.intent.Request, "", false) != nil {
					t.Fatal("unknown finish")
				}
				event = workerProbeEvent{Stage: "recovery", ID: id}
			}
			f := &physicalCaptureFixture{probe: g.f, request: j.intent.Request, objects: j.intent.Objects, mode: physical}
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
			w := &syntheticCaptureOperationWitness{mode: mode}
			e, err := newWorkspaceCaptureSessionEvidence(l, "worker", j.intent, w)
			if err != nil {
				t.Fatal(err)
			}
			r := detachedProbeWorkerRecord(j.record)
			switch mode {
			case "extra reference":
				f.mode = mode
			case "record changed":
				f.mode = mode
			case "false absence":
				f.mode = "created"
			case "foreign record":
				r.Pair.Workspace.Removed = true
			case "foreign worker":
				e.worker = "other"
			case "expired":
				dead, cancel := context.WithCancel(ctx)
				cancel()
				e.lifetime = dead
			case "stopped":
				if err = os.WriteFile(filepath.Join(j.dir, workerProbeStop), []byte("stop"), 0600); err != nil {
					t.Fatal(err)
				}
			case "corrupt history":
				if err = os.WriteFile(filepath.Join(j.dir, captureRecordName(j.intent)), []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
			case "changed history":
				w.change = func() {
					records, err := readWorkerCaptureRecords(j.dir)
					if err != nil {
						t.Fatal(err)
					}
					local := records[captureRecordName(j.intent)]
					local.Events = append(local.Events, event)
					raw, _ := json.Marshal(local)
					if err = os.WriteFile(filepath.Join(j.dir, captureRecordName(j.intent)), raw, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			err = e.ValidateCaptureOperation(ctx, st, r, j.intent, event)
			healthy := strings.HasSuffix(mode, "-pre") || strings.HasSuffix(mode, "-post")
			if (err == nil) != healthy {
				t.Fatal(mode, err)
			}
			if healthy && w.operations != 2 {
				t.Fatal("witness not rechecked", w.operations)
			}
			if mode == "denied" && f.inspections != 0 {
				t.Fatal("inspected before independent permission")
			}
			if g.f.creates != 0 || g.f.started || g.f.removed {
				t.Fatal("operation observer performed lifecycle effects")
			}
			if j.intent.Request.Members[0].Name == "mutated" || j.record.Pair.Workspace.Removed {
				t.Fatal("callback changed pinned metadata")
			}
		})
	}
}
func TestWorkspaceCaptureSessionEvidenceRequiresScope(t *testing.T) {
	var e *workspaceCaptureSessionEvidence
	if e.ValidateCaptureOperation(t.Context(), room.State{}, workerJobRecord{}, workerCaptureIntent{}, workerProbeEvent{}) == nil {
		t.Fatal("missing scope")
	}
	j, _, l := captureJournalFixtureScope(t)
	if _, err := newWorkspaceCaptureSessionEvidence(l, "worker", j.intent, nil); err == nil {
		t.Fatal("missing witness")
	}
	if _, err := newWorkspaceCaptureSessionEvidence(nil, "worker", j.intent, &syntheticCaptureOperationWitness{}); err == nil {
		t.Fatal("missing creator")
	}
}
