package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"errors"
	"strings"
	"testing"
)

type syntheticCaptureSessionEvidence struct {
	operation   func(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error
	transition  func(context.Context, room.State, room.WorkspaceCaptureReservation, workerProbeEvent) error
	calls, acks int
}

func (e *syntheticCaptureSessionEvidence) ValidateCaptureOperation(ctx context.Context, st room.State, r workerJobRecord, i workerCaptureIntent, v workerProbeEvent) error {
	e.calls++
	if e.operation != nil {
		return e.operation(ctx, st, r, i, v)
	}
	return ctx.Err()
}
func (e *syntheticCaptureSessionEvidence) ValidateCaptureTransition(ctx context.Context, st room.State, r room.WorkspaceCaptureReservation, v workerProbeEvent) error {
	e.acks++
	if e.transition != nil {
		return e.transition(ctx, st, r, v)
	}
	return ctx.Err()
}
func captureSessionFixture(t *testing.T) (*controllerCaptureSession, *mockDB, context.Context, *syntheticCaptureSessionEvidence) {
	t.Helper()
	s, db, ctx, r, q, _ := controllerCaptureFixture(t)
	i := workerCaptureIntent{Request: q, Image: "sha256:" + strings.Repeat("d", 64), Objects: toolbroker.CaptureObjectVolume{Name: "commonroom-custody-" + r.JobID, Nonce: strings.Repeat("f", 32), CreatedAt: r.Pair.Workspace.Identity.CreatedAt}}
	e := &syntheticCaptureSessionEvidence{}
	g, err := newControllerCaptureSession(ctx, s, "synthetic-worker", r, i, e)
	if err != nil {
		t.Fatal(err)
	}
	return g, db, ctx, e
}
func TestControllerCaptureSessionPinnedLifecycle(t *testing.T) {
	t.Parallel()
	g, db, ctx, e := captureSessionFixture(t)
	other := context.WithValue(ctx, actorContextKey{}, "untrusted")
	before := len(db.snapshot())
	local := workerCaptureRecord{Version: 1, Intent: copyCaptureIntent(g.intent)}
	for _, event := range []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: strings.Repeat("1", 64)}, {Stage: "finish", ID: strings.Repeat("1", 64), Absent: true}} {
		if err := g.AuthorizeCapture(other, g.record, g.intent, event); err != nil {
			t.Fatal("precheck", event, err)
		}
		local.Events = append(local.Events, event)
		if err := g.AcknowledgeCapture(other, local, event); err != nil {
			t.Fatal("ack", event, err)
		}
		if err := g.AuthorizeCapture(other, g.record, g.intent, event); err != nil {
			t.Fatal("postcheck", event, err)
		}
	}
	if len(db.snapshot()) != before+3 || e.calls != 6 || e.acks != 3 {
		t.Fatal("wrong durable lifecycle", len(db.snapshot()), e.calls, e.acks)
	}
	st := artifactState(t, g.server, ctx)
	saved := st.WorkspaceCaptures[controllerCaptureKey(g.intent)]
	if saved.ActorID != g.actor || !saved.Released {
		t.Fatal("principal or release changed")
	}
	if g.AuthorizeCapture(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("old begin reused as operation authority")
	}
	// The exact full acknowledged history remains idempotent, without extra audit.
	if err := g.AcknowledgeCapture(ctx, local, local.Events[len(local.Events)-1]); err != nil || len(db.snapshot()) != before+3 {
		t.Fatal("retry", err)
	}
}
func TestControllerCaptureSessionRejectsHistorySubstitution(t *testing.T) {
	t.Parallel()
	g, db, ctx, e := captureSessionFixture(t)
	begin := workerProbeEvent{Stage: "begin"}
	local := workerCaptureRecord{Version: 1, Intent: g.intent, Events: []workerProbeEvent{begin}}
	if g.AcknowledgeCapture(ctx, local, begin) != nil {
		t.Fatal("begin")
	}
	bound := workerProbeEvent{Stage: "bound", ID: strings.Repeat("1", 64)}
	local.Events = append(local.Events, bound)
	if g.AcknowledgeCapture(ctx, local, bound) != nil {
		t.Fatal("bound")
	}
	before := len(db.snapshot())
	acks := e.acks
	if g.AcknowledgeCapture(ctx, workerCaptureRecord{Version: 1, Intent: g.intent, Events: []workerProbeEvent{begin}}, begin) == nil {
		t.Fatal("truncated history accepted")
	}
	if len(db.snapshot()) != before || e.acks != acks {
		t.Fatal("substitution reached physical evidence")
	}
}
func TestControllerCaptureSessionPolicyAndLifetime(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"cancel", "worker", "record", "intent", "evidence", "flag", "stale", "agent"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, e := captureSessionFixture(t)
			r := detachedProbeWorkerRecord(g.record)
			i := copyCaptureIntent(g.intent)
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
				i.Request.Members[0].MaxBytes++
			case "evidence":
				e.operation = func(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error {
					return errors.New("synthetic quota denial")
				}
			case "flag":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
				defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
			case "stale":
				e.operation = func(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error {
					db.seed("agent.background", map[string]string{"stage": "synthetic"})
					return nil
				}
			case "agent":
				g.actor = "dev"
			}
			before := len(db.snapshot())
			err := g.AuthorizeCapture(ctx, r, i, workerProbeEvent{Stage: "begin"})
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
				local := workerCaptureRecord{Version: 1, Intent: g.intent, Events: []workerProbeEvent{{Stage: "begin"}}}
				if g.AcknowledgeCapture(ctx, local, local.Events[0]) == nil {
					t.Fatal("fresh context revived dead session")
				}
			}
		})
	}
}
func TestControllerCaptureSessionCleanupAndPersistenceFault(t *testing.T) {
	t.Parallel()
	for _, failure := range []bool{false, true} {
		t.Run(map[bool]string{false: "flags-off-cleanup", true: "audit-fault"}[failure], func(t *testing.T) {
			t.Parallel()
			g, db, ctx, _ := captureSessionFixture(t)
			local := workerCaptureRecord{Version: 1, Intent: g.intent}
			for _, v := range []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: strings.Repeat("1", 64)}} {
				local.Events = append(local.Events, v)
				if g.AcknowledgeCapture(ctx, local, v) != nil {
					t.Fatal("setup")
				}
			}
			db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
			defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
			v := workerProbeEvent{Stage: "finish", ID: strings.Repeat("1", 64), Absent: true}
			if g.AuthorizeCapture(ctx, g.record, g.intent, v) != nil {
				t.Fatal("cleanup blocked")
			}
			local.Events = append(local.Events, v)
			if failure {
				db.failAt = "commit"
			}
			err := g.AcknowledgeCapture(ctx, local, v)
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
func TestControllerCaptureSessionDetachesEvidenceAndStops(t *testing.T) {
	t.Parallel()
	g, _, ctx, e := captureSessionFixture(t)
	e.operation = func(_ context.Context, _ room.State, r workerJobRecord, i workerCaptureIntent, _ workerProbeEvent) error {
		r.Pair.Workspace.Removed = true
		i.Request.Members[0].Name = "changed"
		return nil
	}
	if g.AuthorizeCapture(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) != nil || g.record.Pair.Workspace.Removed || g.intent.Request.Members[0].Name == "changed" {
		t.Fatal("callback rewrote session")
	}
	g.StopCaptureAuthority()
	if !g.server.integrityFault || g.AuthorizeCapture(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("stop did not revoke session")
	}
	g.server.integrityFault = false
	if g.AuthorizeCapture(ctx, g.record, g.intent, workerProbeEvent{Stage: "begin"}) == nil {
		t.Fatal("clearing global fault revived stopped session")
	}
	if _, err := newControllerCaptureSession(t.Context(), g.server, g.worker, g.record, g.intent, e); err == nil {
		t.Fatal("missing authenticated actor")
	}
	if _, err := newControllerCaptureSession(ctx, g.server, g.worker, g.record, g.intent, nil); err == nil {
		t.Fatal("missing evidence")
	}
}
