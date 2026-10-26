package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type syntheticCaptureLedgerEvidence struct {
	check func(context.Context, room.State, room.WorkspaceCaptureReservation, workerProbeEvent) error
	calls int
}

func (e *syntheticCaptureLedgerEvidence) ValidateCaptureTransition(ctx context.Context, st room.State, r room.WorkspaceCaptureReservation, v workerProbeEvent) error {
	e.calls++
	if e.check != nil {
		return e.check(ctx, st, r, v)
	}
	return ctx.Err()
}
func captureLedgerFixture(t *testing.T) (*controllerCaptureLedger, *server, *mockDB, context.Context, workerCaptureIntent, *syntheticCaptureLedgerEvidence) {
	t.Helper()
	s, db, ctx, r, q, _ := controllerCaptureFixture(t)
	e := &syntheticCaptureLedgerEvidence{}
	ledger, err := newControllerCaptureLedger(s, e)
	if err != nil {
		t.Fatal(err)
	}
	i := workerCaptureIntent{Request: q, Image: "sha256:" + strings.Repeat("d", 64), Objects: toolbroker.CaptureObjectVolume{Name: "commonroom-custody-" + r.JobID, Nonce: strings.Repeat("f", 32), CreatedAt: r.Pair.Workspace.Identity.CreatedAt}}
	return ledger, s, db, ctx, i, e
}
func TestControllerCaptureLedgerDurableCapacityReplay(t *testing.T) {
	t.Parallel()
	j, s, db, ctx, i, e := captureLedgerFixture(t)
	before := len(db.snapshot())
	begin := workerProbeEvent{Stage: "begin"}
	if j.Acknowledge(ctx, i, begin) != nil {
		t.Fatal("begin")
	}
	st := artifactState(t, s, ctx)
	if n, err := controllerCaptureSlots(st); err != nil || n != 1 {
		t.Fatal("capture not charged", n, err)
	}
	next := st.ManagedJobs[i.Request.Binding.Execution.JobID]
	next.JobID = strings.Repeat("9", 32)
	if admitManagedContainerQuota(st, next) == nil {
		t.Fatal("new role pair ignored capture charge")
	}
	if j.Acknowledge(ctx, i, begin) != nil || len(db.snapshot()) != before+1 {
		t.Fatal("idempotent retry appended twice")
	}
	id := strings.Repeat("1", 64)
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}) != nil || j.Acknowledge(ctx, i, workerProbeEvent{Stage: "finish", ID: id, Absent: true}) != nil {
		t.Fatal("finish")
	}
	st = artifactState(t, s, ctx)
	if n, err := controllerCaptureSlots(st); err != nil || n != 0 {
		t.Fatal("observed helper absence not released", n, err)
	}
	if len(db.snapshot()) != before+3 || e.calls != 4 {
		t.Fatal("wrong acknowledgment history")
	}
	events := db.snapshot()
	if projectControllerCapture(&st, []byte(events[len(events)-1].payload)) == nil {
		t.Fatal("duplicate audit replay accepted")
	}
	if !strings.Contains(auditSummary(controllerCaptureEventKind, nil), "acknowledged") {
		t.Fatal("unsafe summary")
	}
}
func TestControllerCaptureLedgerUnknownAndKnownRecovery(t *testing.T) {
	t.Parallel()
	for _, known := range []bool{false, true} {
		t.Run(map[bool]string{true: "known", false: "unknown"}[known], func(t *testing.T) {
			t.Parallel()
			j, s, _, ctx, i, _ := captureLedgerFixture(t)
			j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"})
			id := ""
			if known {
				id = strings.Repeat("1", 64)
				if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}) != nil {
					t.Fatal("bind")
				}
			}
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "finish", ID: id, Absent: false}) != nil {
				t.Fatal("uncertain finish")
			}
			st := artifactState(t, s, ctx)
			if n, err := controllerCaptureSlots(st); err != nil || n != 1 {
				t.Fatal("uncertainty lost charge", n, err)
			}
			if !known {
				id = strings.Repeat("9", 64)
			}
			err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "recovery", ID: id})
			if !known {
				if err == nil {
					t.Fatal("unknown ID adopted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "recovery-finish", ID: id, Absent: true}) != nil {
				t.Fatal("recovery finish")
			}
			st = artifactState(t, s, ctx)
			if n, err := controllerCaptureSlots(st); err != nil || n != 0 {
				t.Fatal("recovery charge", n, err)
			}
		})
	}
}
func TestControllerCaptureLedgerDenialsAndFencing(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"physical", "commit", "CAS", "agent", "missing actor", "wrong definition", "wrong member", "flag off"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, s, db, ctx, i, e := captureLedgerFixture(t)
			switch mode {
			case "physical":
				e.check = func(context.Context, room.State, room.WorkspaceCaptureReservation, workerProbeEvent) error {
					return errors.New("synthetic physical denial")
				}
			case "commit":
				db.failAt = "commit"
			case "CAS":
				e.check = func(context.Context, room.State, room.WorkspaceCaptureReservation, workerProbeEvent) error {
					db.seed("agent.background", map[string]string{"stage": "synthetic"})
					return nil
				}
			case "agent":
				ctx = context.WithValue(ctx, actorContextKey{}, "dev")
			case "missing actor":
				ctx = t.Context()
			case "wrong definition":
				i.Request.Binding.Execution.DefinitionDigest = strings.Repeat("9", 64)
			case "wrong member":
				i.Request.Members[0].Name = "other.txt"
			case "flag off":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
				defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
			}
			before := len(db.snapshot())
			err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"})
			if mode == "commit" {
				db.failAt = ""
				if !s.integrityFault {
					t.Fatal("persistence did not halt")
				}
				s.integrityFault = false
			}
			if err == nil {
				t.Fatal("denial granted")
			}
			extra := 0
			if mode == "CAS" {
				extra = 1
				if !errors.Is(err, errAuditConflict) || s.integrityFault {
					t.Fatal("wrong conflict behavior", err)
				}
			}
			if len(db.snapshot()) != before+extra {
				t.Fatal("denial appended grant")
			}
		})
	}
}
func TestControllerCaptureLedgerCleanupWithFlagsOffAndCombinedQuota(t *testing.T) {
	t.Parallel()
	j, s, db, ctx, i, _ := captureLedgerFixture(t)
	id := strings.Repeat("1", 64)
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil || j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}) != nil {
		t.Fatal("setup")
	}
	st := artifactState(t, s, ctx)
	probe := toolbroker.WorkspaceProbeIntent{JobID: i.Request.Binding.Execution.JobID, Repository: i.Request.Binding.Execution.RepositoryDigest, Image: i.Image, Challenge: strings.Repeat("7", 32), Volume: toolbroker.WorkspaceVolumeIdentity{Version: 1, Name: i.Request.Binding.VolumeName, Nonce: i.Request.Binding.VolumeNonce, CreatedAt: i.Request.Binding.VolumeCreatedAt}}
	st.WorkspaceProbes = map[string]room.WorkspaceProbeReservation{controllerProbeKey(probe): {Version: 1, CompanyID: st.Company.ID, RoomID: i.Request.Binding.Execution.RoomID, Worker: "synthetic-worker", ActorID: ownerID, Intent: probe, Events: []workerProbeEvent{{Stage: "begin"}}}}
	if n, err := controllerHelperSlots(st); err != nil || n != 2 {
		t.Fatal("combined helpers", n, err)
	}
	if admitControllerProbeQuota(st) == nil {
		t.Fatal("probe ignored capture charge")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
	defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "finish", ID: id, Absent: true}) != nil {
		t.Fatal("flag off blocked authorized cleanup")
	}
}
func TestControllerCaptureLedgerRejectsPublicAndMalformedHistory(t *testing.T) {
	_, _, handler := mockServer(t)
	if w := mockRequest(handler, "POST", "/api/events", `{"kind":"workspace.capture.acknowledged","payload":{}}`); w.Code < 400 {
		t.Fatal("public ledger injection")
	}
	j, s, _, ctx, i, _ := captureLedgerFixture(t)
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil {
		t.Fatal("setup")
	}
	st := artifactState(t, s, ctx)
	r := st.WorkspaceCaptures[controllerCaptureKey(i)]
	r.Released = true
	st.WorkspaceCaptures[controllerCaptureKey(i)] = r
	if _, err := controllerHelperSlots(st); err == nil {
		t.Fatal("forged release accepted")
	}
	raw, _ := json.Marshal(controllerCaptureEvent{Version: 1, ActorID: ownerID, CompanyID: st.Company.ID, RoomID: i.Request.Binding.Execution.RoomID, Worker: "synthetic-worker", Intent: i, Event: workerProbeEvent{Stage: "confirmed"}})
	if projectControllerCapture(&st, raw) == nil {
		t.Fatal("worker-only confirmation injected")
	}
}

func TestControllerCaptureLedgerEvidenceCannotMutatePinnedIntent(t *testing.T) {
	t.Parallel()
	j, s, _, ctx, i, e := captureLedgerFixture(t)
	e.check = func(_ context.Context, _ room.State, r room.WorkspaceCaptureReservation, _ workerProbeEvent) error {
		r.Intent.Request.Members[0].Name = "changed.txt"
		r.Events[0].Stage = "forged"
		return nil
	}
	if err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}); err != nil {
		t.Fatal(err)
	}
	st := artifactState(t, s, ctx)
	saved := st.WorkspaceCaptures[controllerCaptureKey(i)]
	if saved.Intent.Request.Members[0].Name != i.Request.Members[0].Name || saved.Events[0].Stage != "begin" {
		t.Fatal("callback mutated durable intent")
	}
}
