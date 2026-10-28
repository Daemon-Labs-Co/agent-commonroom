package main

import (
	"agentroom/internal/objectread"
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacecapture"
	"agentroom/internal/workspacesnapshot"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
)

type syntheticObjectReadEvidence struct {
	check func(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error
	calls int
}

func (e *syntheticObjectReadEvidence) ValidateObjectReadTransition(ctx context.Context, st room.State, r room.ObjectReadReservation, v workerProbeEvent) error {
	e.calls++
	if e.check != nil {
		return e.check(ctx, st, r, v)
	}
	return ctx.Err()
}
func objectReadLedgerFixture(t *testing.T) (*controllerObjectReadLedger, *server, *mockDB, context.Context, workerObjectReadIntent, *syntheticObjectReadEvidence) {
	t.Helper()
	s, db, ctx, scope := artifactJournalFixture(t, true)
	custody, err := newControllerArtifactJournal(s, &syntheticArtifactGuard{})
	if err != nil {
		t.Fatal(err)
	}
	store, _ := artifactStore(t, custody)
	receipt, err := store.Put(ctx, scope, 128, strings.NewReader("synthetic reader bytes"))
	if err != nil {
		t.Fatal(err)
	}
	e := &syntheticObjectReadEvidence{}
	j, err := newControllerObjectReadLedger(s, e)
	if err != nil {
		t.Fatal(err)
	}
	i := workerObjectReadIntent{Request: objectread.Request{Receipt: receipt, Challenge: strings.Repeat("7", 32)}, Image: "sha256:" + strings.Repeat("d", 64), Objects: toolbroker.CaptureObjectVolume{Name: "commonroom-custody-" + scope.JobID, Nonce: strings.Repeat("f", 32), CreatedAt: "2026-10-20T00:00:00Z"}}
	if !toolbroker.ValidObjectReadMetadata(i.Request, i.Image, i.Objects) {
		t.Fatal("invalid synthetic metadata")
	}
	return j, s, db, ctx, i, e
}
func TestControllerObjectReadLedgerLifecycleAndCapacity(t *testing.T) {
	t.Parallel()
	j, s, db, ctx, i, e := objectReadLedgerFixture(t)
	before := len(db.snapshot())
	begin := workerProbeEvent{Stage: "begin"}
	if err := j.Acknowledge(ctx, i, begin); err != nil {
		t.Fatal(err)
	}
	st := artifactState(t, s, ctx)
	if n, err := controllerHelperSlots(st); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	next := st.ManagedJobs[i.Request.Receipt.Intent.Scope.JobID]
	next.JobID = strings.Repeat("9", 32)
	if admitManagedContainerQuota(st, next) == nil {
		t.Fatal("reader omitted from container quota")
	}
	if err := j.Acknowledge(ctx, i, begin); err != nil || len(db.snapshot()) != before+1 || e.calls != 2 {
		t.Fatal("retry did not revalidate without append", err)
	}
	id := strings.Repeat("1", 64)
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}) != nil || j.Acknowledge(ctx, i, workerProbeEvent{Stage: "finish", ID: id, Absent: true}) != nil {
		t.Fatal("finish")
	}
	st = artifactState(t, s, ctx)
	if n, err := controllerObjectReadSlots(st); err != nil || n != 0 {
		t.Fatal(n, err)
	}
	if st.ArtifactCustody[i.Request.Receipt.Intent.ObjectID].ReadAcknowledgments != 0 {
		t.Fatal("lifecycle manufactured byte release")
	}
	if _, ok := st.ManagedStorage.Releases[i.Request.Receipt.Intent.ObjectID]; ok {
		t.Fatal("helper cleanup refunded object storage")
	}
	if j.Acknowledge(ctx, i, begin) == nil {
		t.Fatal("settled reader reopened")
	}
	events := db.snapshot()
	if len(events) != before+3 || projectControllerObjectRead(&st, []byte(events[len(events)-1].payload)) == nil {
		t.Fatal("duplicate audit replay accepted")
	}
	if !strings.Contains(auditSummary(controllerObjectReadEventKind, nil), "delivery not confirmed") {
		t.Fatal("summary claims delivery")
	}
}
func TestControllerObjectReadLedgerRecovery(t *testing.T) {
	t.Parallel()
	for _, known := range []bool{false, true} {
		t.Run(map[bool]string{true: "known", false: "unknown"}[known], func(t *testing.T) {
			t.Parallel()
			j, s, db, ctx, i, e := objectReadLedgerFixture(t)
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil {
				t.Fatal("begin")
			}
			id := ""
			if known {
				id = strings.Repeat("1", 64)
				if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "bound", ID: id}) != nil {
					t.Fatal("bound")
				}
			}
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "finish", ID: id}) != nil {
				t.Fatal("uncertain finish")
			}
			if n, err := controllerObjectReadSlots(artifactState(t, s, ctx)); err != nil || n != 1 {
				t.Fatal("uncertainty not charged", n, err)
			}
			if !known {
				if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "recovery", ID: strings.Repeat("2", 64)}) == nil {
					t.Fatal("unknown ID adopted")
				}
				return
			}
			recovery := workerProbeEvent{Stage: "recovery", ID: id}
			if j.Acknowledge(ctx, i, recovery) != nil {
				t.Fatal("recovery")
			}
			before, calls := len(db.snapshot()), e.calls
			for range 8 {
				if j.Acknowledge(ctx, i, recovery) != nil {
					t.Fatal("recovery retry")
				}
			}
			if len(db.snapshot()) != before || e.calls != calls+8 {
				t.Fatal("unbounded history or skipped evidence")
			}
			e.check = func(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error {
				return errors.New("fresh cleanup denied")
			}
			if j.Acknowledge(ctx, i, recovery) == nil {
				t.Fatal("cached cleanup permission")
			}
			e.check = nil
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "recovery-finish", ID: strings.Repeat("2", 64), Absent: true}) == nil {
				t.Fatal("foreign ID released")
			}
			db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
			defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
			if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "recovery-finish", ID: id, Absent: true}) != nil {
				t.Fatal("editing off blocked cleanup")
			}
			if n, err := controllerObjectReadSlots(artifactState(t, s, ctx)); err != nil || n != 0 {
				t.Fatal(n, err)
			}
			if j.Acknowledge(ctx, i, recovery) == nil {
				t.Fatal("settled recovery reopened")
			}
		})
	}
}
func TestControllerObjectReadLedgerRefusals(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"physical", "commit", "CAS", "agent", "missing actor", "receipt", "source", "company", "canceled", "integrity", "revoked"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			j, s, db, ctx, i, e := objectReadLedgerFixture(t)
			switch mode {
			case "physical":
				e.check = func(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error {
					return errors.New("synthetic quota denial")
				}
			case "commit":
				db.failAt = "commit"
			case "CAS":
				e.check = func(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error {
					db.seed("agent.background", map[string]string{"stage": "synthetic"})
					return nil
				}
			case "agent":
				ctx = context.WithValue(ctx, actorContextKey{}, "bot-openai")
			case "missing actor":
				ctx = t.Context()
			case "receipt":
				i.Request.Receipt.Digest = strings.Repeat("9", 64)
			case "source":
				i.Request.Receipt.Intent.Scope.MemberName = "foreign.txt"
			case "company":
				i.Request.Receipt.Intent.Scope.CompanyID = "foreign"
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "integrity":
				s.integrityFault = true
			case "revoked":
				e.check = func(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error {
					db.seed("incident.opened", map[string]string{"reason": "synthetic"})
					return nil
				}
			}
			before := len(db.snapshot())
			err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"})
			if mode == "commit" {
				db.failAt = ""
				if !s.integrityFault {
					t.Fatal("failed persistence did not halt")
				}
				s.integrityFault = false
			}
			if mode == "integrity" {
				s.integrityFault = false
			}
			if err == nil {
				t.Fatal("denied operation admitted")
			}
			extra := 0
			if mode == "CAS" || mode == "revoked" {
				extra = 1
			}
			if mode == "CAS" && (!errors.Is(err, errAuditConflict) || s.integrityFault) {
				t.Fatal("wrong conflict policy", err)
			}
			if len(db.snapshot()) != before+extra {
				t.Fatal("refusal appended reader grant")
			}
		})
	}
}
func TestControllerObjectReadLedgerEditingOffRetainedRead(t *testing.T) {
	t.Parallel()
	j, s, db, ctx, i, _ := objectReadLedgerFixture(t)
	db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
	defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
	if err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}); err != nil {
		t.Fatal("editing flag erased retained read", err)
	}
	if n, err := controllerObjectReadSlots(artifactState(t, s, ctx)); err != nil || n != 1 {
		t.Fatal(n, err)
	}
}
func TestControllerObjectReadLedgerMalformedAndPrivateHistory(t *testing.T) {
	t.Parallel()
	j, s, _, ctx, i, _ := objectReadLedgerFixture(t)
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil {
		t.Fatal("setup")
	}
	st := artifactState(t, s, ctx)
	r := st.ObjectReads[controllerObjectReadKey(i)]
	r.Released = true
	st.ObjectReads[controllerObjectReadKey(i)] = r
	if _, err := controllerHelperSlots(st); err == nil {
		t.Fatal("forged release accepted")
	}
	raw, _ := json.Marshal(controllerObjectReadEvent{Version: 1, ActorID: ownerID, CompanyID: st.Company.ID, RoomID: i.Request.Receipt.Intent.Scope.RoomID, Worker: "synthetic-worker", Intent: i, Event: workerProbeEvent{Stage: "confirmed"}})
	if projectControllerObjectRead(&st, raw) == nil {
		t.Fatal("worker confirmation injected")
	}
	raw, _ = json.Marshal(st)
	if strings.Contains(string(raw), "objectReads") || strings.Contains(string(raw), i.Request.Challenge) {
		t.Fatal("private ledger leaked in state")
	}
	_, _, h := mockServer(t)
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"artifact.reader.acknowledged","payload":{}}`); w.Code < 400 {
		t.Fatal("public reader injection")
	}
	var missing *syntheticObjectReadEvidence
	if _, err := newControllerObjectReadLedger(s, missing); err == nil {
		t.Fatal("typed nil evidence accepted")
	}
}
func TestControllerObjectReadLedgerDetachedEvidence(t *testing.T) {
	t.Parallel()
	j, s, _, ctx, i, e := objectReadLedgerFixture(t)
	e.check = func(_ context.Context, _ room.State, r room.ObjectReadReservation, _ workerProbeEvent) error {
		r.Events[0].Stage = "forged"
		return nil
	}
	if err := j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}); err != nil {
		t.Fatal(err)
	}
	if artifactState(t, s, ctx).ObjectReads[controllerObjectReadKey(i)].Events[0].Stage != "begin" {
		t.Fatal("callback mutated durable state")
	}
}

func TestControllerObjectReadLedgerConcurrentControllers(t *testing.T) {
	t.Parallel()
	a, s, db, ctx, i, _ := objectReadLedgerFixture(t)
	other := &server{db: db}
	b, err := newControllerObjectReadLedger(other, &syntheticObjectReadEvidence{})
	if err != nil {
		t.Fatal(err)
	}
	second := i
	second.Request.Challenge = strings.Repeat("6", 32)
	var wg sync.WaitGroup
	result := make(chan error, 2)
	for _, v := range []struct {
		j *controllerObjectReadLedger
		i workerObjectReadIntent
	}{{a, i}, {b, second}} {
		wg.Add(1)
		go func(j *controllerObjectReadLedger, i workerObjectReadIntent) {
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
		} else if !errors.Is(err, errAuditConflict) && !errors.Is(err, errControllerObjectRead) {
			t.Fatal(err)
		}
	}
	st := artifactState(t, s, ctx)
	if success != 1 || len(st.ObjectReads) != 1 || s.integrityFault || other.integrityFault {
		t.Fatal("concurrent reader admission not fenced", success)
	}
}
func TestControllerObjectReadLedgerCaptureExclusionAndCombinedQuota(t *testing.T) {
	t.Parallel()
	j, s, _, ctx, i, _ := objectReadLedgerFixture(t)
	st := artifactState(t, s, ctx)
	scope := i.Request.Receipt.Intent.Scope
	binding := workspacesnapshot.Binding{Execution: scope.ExecutionScope, BaseCommit: strings.Repeat("8", 40), VolumeName: "commonroom-tool-work-" + scope.JobID, VolumeNonce: strings.Repeat("c", 32), VolumeCreatedAt: i.Objects.CreatedAt}
	binding.Execution.RevisionDigest = ""
	q, err := workspacecapture.ContractRequest(st.TaskExecutions[scope.ExecutionID].Definition, binding, strings.Repeat("8", 32), []string{"report"})
	if err != nil {
		t.Fatal(err)
	}
	capture := workerCaptureIntent{Request: q, Image: i.Image, Objects: i.Objects}
	v := controllerCaptureEvent{Version: 1, ActorID: ownerID, CompanyID: st.Company.ID, RoomID: scope.RoomID, Worker: "synthetic-worker", Intent: capture, Event: workerProbeEvent{Stage: "begin"}}
	if j.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil {
		t.Fatal("reader begin")
	}
	st = artifactState(t, s, ctx)
	if _, _, err := validateControllerCaptureEvent(st, v); err == nil {
		t.Fatal("capture admitted alongside reader")
	}
	// Both directions are checked against the same exact task/job, independently
	// of the account container limit.
	delete(st.ObjectReads, controllerObjectReadKey(i))
	r, _, err := validateControllerCaptureEvent(st, v)
	if err != nil {
		t.Fatal(err)
	}
	st.WorkspaceCaptures = map[string]room.WorkspaceCaptureReservation{controllerCaptureKey(capture): r}
	read := controllerObjectReadEvent{Version: 1, ActorID: ownerID, CompanyID: st.Company.ID, RoomID: scope.RoomID, Worker: "synthetic-worker", Intent: i, Event: workerProbeEvent{Stage: "begin"}}
	if _, _, err := validateControllerObjectReadEvent(st, read); err == nil {
		t.Fatal("reader admitted alongside capture")
	}
	st = artifactState(t, s, ctx)
	st.WorkspaceCaptures = map[string]room.WorkspaceCaptureReservation{controllerCaptureKey(capture): r}
	probe := toolbroker.WorkspaceProbeIntent{JobID: scope.JobID, Repository: scope.RepositoryDigest, Image: i.Image, Challenge: strings.Repeat("5", 32), Volume: toolbroker.WorkspaceVolumeIdentity{Version: 1, Name: binding.VolumeName, Nonce: binding.VolumeNonce, CreatedAt: binding.VolumeCreatedAt}}
	st.WorkspaceProbes = map[string]room.WorkspaceProbeReservation{controllerProbeKey(probe): {Version: 1, ActorID: ownerID, CompanyID: st.Company.ID, RoomID: scope.RoomID, Worker: "synthetic-worker", Intent: probe, Events: []workerProbeEvent{{Stage: "begin"}}}}
	if n, err := controllerHelperSlots(st); err != nil || n != 3 {
		t.Fatal("combined accounting", n, err)
	}
	if admitControllerProbeQuota(st) == nil {
		t.Fatal("probe ignored reader/capture capacity")
	}
}
