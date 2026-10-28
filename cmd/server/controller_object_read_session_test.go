package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"strings"
	"testing"
)

type syntheticObjectReadSessionEvidence struct {
	operation   func(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error
	transition  func(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error
	calls, acks int
}

func (e *syntheticObjectReadSessionEvidence) ValidateObjectReadOperation(ctx context.Context, st room.State, r workerJobRecord, i workerObjectReadIntent, v workerProbeEvent) error {
	e.calls++
	if e.operation != nil {
		return e.operation(ctx, st, r, i, v)
	}
	return ctx.Err()
}
func (e *syntheticObjectReadSessionEvidence) ValidateObjectReadTransition(ctx context.Context, st room.State, r room.ObjectReadReservation, v workerProbeEvent) error {
	e.acks++
	if e.transition != nil {
		return e.transition(ctx, st, r, v)
	}
	return ctx.Err()
}
func objectReadSessionFixture(t *testing.T) (*controllerObjectReadSession, *mockDB, context.Context, *syntheticObjectReadSessionEvidence) {
	t.Helper()
	_, s, db, ctx, i, _ := objectReadLedgerFixture(t)
	st := artifactState(t, s, ctx)
	job := st.ManagedJobs[i.Request.Receipt.Intent.Scope.JobID]
	_, r, _ := retainedObservationFixture(t)
	r.JobID = job.JobID
	r.Version = job.Version
	r.Execution = job.Execution
	r.RoomID = job.RoomID
	r.ParticipantID = job.ParticipantID
	r.ConnectionID = job.ConnectionID
	r.Provider = job.Provider
	r.Deadline = job.Deadline
	r.StartedAt = job.AllocatedAt
	_, _, _, r.Scope, _ = workerMembership(r.RoomID + "/" + r.ParticipantID + "/" + r.ConnectionID + "/worker")
	r.Pair.Repository = job.RepositoryDigest
	r.Pair.Workspace.Identity.Name = "commonroom-tool-work-" + job.JobID
	e := &syntheticObjectReadSessionEvidence{}
	g, err := newControllerObjectReadSession(ctx, s, "synthetic-worker", r, i, e)
	if err != nil {
		t.Fatal(err)
	}
	return g, db, ctx, e
}
func TestControllerObjectReadSessionPinnedLifecycle(t *testing.T) {
	t.Parallel()
	g, db, ctx, e := objectReadSessionFixture(t)
	other := context.WithValue(ctx, actorContextKey{}, "untrusted")
	before := len(db.snapshot())
	local := workerObjectReadRecord{Version: 1, Intent: g.intent}
	for _, event := range []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: strings.Repeat("1", 64)}, {Stage: "finish", ID: strings.Repeat("1", 64), Absent: true}} {
		if err := g.AuthorizeObjectRead(other, g.record, g.intent, event); err != nil {
			t.Fatal("precheck", event, err)
		}
		local.Events = append(local.Events, event)
		if err := g.AcknowledgeObjectRead(other, local, event); err != nil {
			t.Fatal("ack", event, err)
		}
		if err := g.AuthorizeObjectRead(other, g.record, g.intent, event); err != nil {
			t.Fatal("postcheck", event, err)
		}
	}
	if len(db.snapshot()) != before+3 || e.calls != 6 || e.acks != 3 {
		t.Fatal("wrong durable lifecycle", len(db.snapshot()), e.calls, e.acks)
	}
	st := artifactState(t, g.server, ctx)
	saved := st.ObjectReads[controllerObjectReadKey(g.intent)]
	if saved.ActorID != g.actor || !saved.Released {
		t.Fatal("principal or release changed")
	}
	if g.AuthorizeObjectRead(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("old begin reused as operation authority")
	}
	// The exact full acknowledged history remains idempotent, without extra audit.
	if err := g.AcknowledgeObjectRead(ctx, local, local.Events[len(local.Events)-1]); err != nil || len(db.snapshot()) != before+3 {
		t.Fatal("retry", err)
	}
}
func TestControllerObjectReadSessionRejectsHistorySubstitution(t *testing.T) {
	t.Parallel()
	g, db, ctx, e := objectReadSessionFixture(t)
	begin := workerProbeEvent{Stage: "begin"}
	local := workerObjectReadRecord{Version: 1, Intent: g.intent, Events: []workerProbeEvent{begin}}
	if g.AcknowledgeObjectRead(ctx, local, begin) != nil {
		t.Fatal("begin")
	}
	bound := workerProbeEvent{Stage: "bound", ID: strings.Repeat("1", 64)}
	local.Events = append(local.Events, bound)
	if g.AcknowledgeObjectRead(ctx, local, bound) != nil {
		t.Fatal("bound")
	}
	before := len(db.snapshot())
	acks := e.acks
	if g.AcknowledgeObjectRead(ctx, workerObjectReadRecord{Version: 1, Intent: g.intent, Events: []workerProbeEvent{begin}}, begin) == nil {
		t.Fatal("truncated history accepted")
	}
	if len(db.snapshot()) != before || e.acks != acks {
		t.Fatal("substitution reached physical evidence")
	}
}
func TestControllerObjectReadSessionPolicyAndLifetime(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"cancel", "worker", "record", "intent", "evidence", "incident", "stale", "agent"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, e := objectReadSessionFixture(t)
			r := detachedProbeWorkerRecord(g.record)
			i := g.intent
			switch mode {
			case "cancel":
				session, cancel := context.WithCancel(ctx)
				g.lifetime = session
				cancel()
			case "worker":
				g.worker = "other"
			case "record":
				r.Pair.Workspace.Removed = true
			case "intent":
				i.Request.Receipt.Bytes++
			case "evidence":
				e.operation = func(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error {
					return errors.New("synthetic quota denial")
				}
			case "incident":
				db.seed("incident.opened", map[string]string{"reason": "synthetic"})
			case "stale":
				e.operation = func(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error {
					db.seed("agent.background", map[string]string{"stage": "synthetic"})
					return nil
				}
			case "agent":
				g.actor = "dev"
			}
			before := len(db.snapshot())
			err := g.AuthorizeObjectRead(ctx, r, i, workerProbeEvent{Stage: "begin"})
			if err == nil {
				t.Fatal("denied policy granted")
			}
			extra := 0
			if mode == "stale" {
				extra = 1
				if !errors.Is(err, errAuditConflict) {
					t.Fatal(err)
				}
			}
			if len(db.snapshot()) != before+extra {
				t.Fatal("precheck appended grant")
			}
			if mode == "cancel" {
				local := workerObjectReadRecord{Version: 1, Intent: g.intent, Events: []workerProbeEvent{{Stage: "begin"}}}
				if g.AcknowledgeObjectRead(ctx, local, local.Events[0]) == nil {
					t.Fatal("fresh context revived dead session")
				}
			}
		})
	}
}
func TestControllerObjectReadSessionCleanupAndPersistenceFault(t *testing.T) {
	t.Parallel()
	for _, failure := range []bool{false, true} {
		t.Run(map[bool]string{false: "flags-off-cleanup", true: "audit-fault"}[failure], func(t *testing.T) {
			t.Parallel()
			g, db, ctx, _ := objectReadSessionFixture(t)
			local := workerObjectReadRecord{Version: 1, Intent: g.intent}
			for _, v := range []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: strings.Repeat("1", 64)}} {
				local.Events = append(local.Events, v)
				if g.AcknowledgeObjectRead(ctx, local, v) != nil {
					t.Fatal("setup")
				}
			}
			db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
			defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
			v := workerProbeEvent{Stage: "finish", ID: strings.Repeat("1", 64), Absent: true}
			if g.AuthorizeObjectRead(ctx, g.record, g.intent, v) != nil {
				t.Fatal("cleanup blocked")
			}
			local.Events = append(local.Events, v)
			if failure {
				db.failAt = "commit"
			}
			err := g.AcknowledgeObjectRead(ctx, local, v)
			if failure {
				db.failAt = ""
				if err == nil || !g.server.integrityFault {
					t.Fatal("audit failure granted", err)
				}
				g.server.integrityFault = false
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestControllerObjectReadSessionDetachesEvidenceAndStops(t *testing.T) {
	t.Parallel()
	g, _, ctx, e := objectReadSessionFixture(t)
	e.operation = func(_ context.Context, _ room.State, r workerJobRecord, i workerObjectReadIntent, _ workerProbeEvent) error {
		r.Pair.Workspace.Removed = true
		i.Request.Receipt.Digest = strings.Repeat("9", 64)
		return nil
	}
	if g.AuthorizeObjectRead(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) != nil || g.record.Pair.Workspace.Removed || g.intent.Request.Receipt.Digest == strings.Repeat("9", 64) {
		t.Fatal("callback rewrote session")
	}
	g.StopObjectReadAuthority()
	if g.server.integrityFault || g.AuthorizeObjectRead(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("scoped stop failed or caused global integrity incident")
	}
	g.server.integrityFault = false
	if g.AuthorizeObjectRead(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("clearing global fault revived stopped session")
	}
	if _, err := newControllerObjectReadSession(t.Context(), g.server, g.worker, g.record, g.intent, e); err == nil {
		t.Fatal("missing authenticated actor")
	}
	if _, err := newControllerObjectReadSession(ctx, g.server, g.worker, g.record, g.intent, nil); err == nil {
		t.Fatal("missing evidence")
	}
}
