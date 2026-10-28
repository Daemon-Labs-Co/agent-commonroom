package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
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

func TestControllerObjectReadSessionCanceledFinishCannotRelease(t *testing.T) {
	t.Parallel()
	g, db, ctx, _ := objectReadSessionFixture(t)
	local := workerObjectReadRecord{Version: 1, Intent: g.intent}
	id := strings.Repeat("1", 64)
	for _, event := range []workerProbeEvent{{Stage: "begin"}, {Stage: "bound", ID: id}} {
		local.Events = append(local.Events, event)
		if err := g.AcknowledgeObjectRead(ctx, local, event); err != nil {
			t.Fatal(err)
		}
	}
	g.StopObjectReadAuthority()
	before := len(db.snapshot())
	absent := workerProbeEvent{Stage: "finish", ID: id, Absent: true}
	denied := copyObjectReadRecord(local)
	denied.Events = append(denied.Events, absent)
	if g.AcknowledgeObjectRead(ctx, denied, absent) == nil || len(db.snapshot()) != before {
		t.Fatal("canceled read released capacity")
	}
	uncertain := workerProbeEvent{Stage: "finish", ID: id}
	local.Events = append(local.Events, uncertain)
	if err := g.AcknowledgeObjectRead(ctx, local, uncertain); err != nil {
		t.Fatal("canceled finish lost uncertainty", err)
	}
	if n, err := controllerObjectReadSlots(artifactState(t, g.server, ctx)); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	if g.AuthorizeObjectRead(ctx, g.record, g.intent, workerProbeEvent{Stage: "recovery", ID: id}) == nil {
		t.Fatal("canceled session gained fresh cleanup authority")
	}
}

func TestControllerObjectReadSessionThroughWorkerJournal(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"normal", "lost-ack", "audit-failure"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			g, db, ctx, e := objectReadSessionFixture(t)
			dir := t.TempDir()
			if err := os.Chmod(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := writeWorkerRecord(dir, g.record, true); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, g.record.JobID+".cancel"), []byte("cancel"), 0600); err != nil {
				t.Fatal(err)
			}
			lease, free, err := workerTryLease(dir, ".job-"+g.record.JobID+".lease")
			if err != nil || !free {
				t.Fatal(err)
			}
			defer lease.Close()
			l := &workerWorkspaceObservation{ownership: lease, active: true, lifetime: ctx, dir: dir, dockerRoot: syntheticObservationDockerRoot, record: detachedProbeWorkerRecord(g.record)}
			var stopped atomic.Int32
			left, right := net.Pipe()
			c, err := newObjectReadAuthorityClient(left, g.record, g.intent, func() { stopped.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			defer c.StopObjectReadAuthority()
			defer right.Close()
			j, err := newWorkerObjectReadJournal(l, g.intent.Request, g.intent.Image, g.intent.Objects, c)
			if err != nil {
				t.Fatal(err)
			}
			// Independent synthetic evidence verifies actual local persistence before the
			// controller event. It does not call the waiting stream or reacquire its locks.
			e.transition = func(_ context.Context, _ room.State, r room.ObjectReadReservation, event workerProbeEvent) error {
				records, err := readWorkerObjectReadRecords(dir)
				local, ok := records[objectReadRecordName(g.intent)]
				if err != nil || !ok || !slices.Equal(local.Events, r.Events) || local.Events[len(local.Events)-1] != event {
					t.Error("controller ack preceded local persistence")
					return errControllerObjectRead
				}
				return nil
			}
			done := make(chan error, 1)
			if mode == "lost-ack" {
				// First authorization succeeds; drop the response only after begin has been
				// committed, retaining both local and controller capacity on link loss.
				serverStream := &objectReadNthWriteDrop{Conn: right, dropAt: 3}
				go func() { done <- serveObjectReadAuthority(ctx, serverStream, g.record, g.intent, g) }()
			} else {
				go func() { done <- serveObjectReadAuthority(ctx, right, g.record, g.intent, g) }()
			}
			if mode == "audit-failure" {
				db.failAt = "commit"
			}
			err = j.BeginObjectRead(ctx, j.intent.Request, j.intent.Objects, j.intent.Image)
			if mode != "normal" {
				c.StopObjectReadAuthority()
				<-done
				db.failAt = ""
				if mode == "audit-failure" {
					if !g.server.integrityFault {
						t.Fatal("audit failure not latched")
					}
					g.server.integrityFault = false
				}
				if err == nil {
					t.Fatal("uncertain acknowledgment granted")
				}
				if n, e := workerObjectReadSlots(dir); e != nil || n != 1 {
					t.Fatal("local charge lost", n, e)
				}
				st := artifactState(t, g.server, ctx)
				want := 0
				if mode == "lost-ack" {
					want = 1
				}
				if n, e := controllerObjectReadSlots(st); e != nil || n != want {
					t.Fatal("controller charge", n, e)
				}
				if !workerProbeStopped(dir) || stopped.Load() != 1 {
					t.Fatal("uncertainty failed to stop worker")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			id := strings.Repeat("1", 64)
			if j.BindObjectRead(ctx, j.intent.Request, id) != nil || j.FinishObjectRead(ctx, j.intent.Request, id, true) != nil {
				t.Fatal("normal journal lifecycle")
			}
			if n, e := workerObjectReadSlots(dir); e != nil || n != 0 {
				t.Fatal("worker confirmation", n, e)
			}
			if n, e := controllerObjectReadSlots(artifactState(t, g.server, ctx)); e != nil || n != 0 {
				t.Fatal("controller release", n, e)
			}
			c.StopObjectReadAuthority()
			<-done
			if stopped.Load() != 1 || g.lifetime.Err() == nil || g.server.integrityFault {
				t.Fatal("normal close did not revoke only session")
			}
		})
	}
}

type objectReadNthWriteDrop struct {
	net.Conn
	writes int
	dropAt int
}

func (s *objectReadNthWriteDrop) Write(p []byte) (int, error) {
	s.writes++
	if s.writes == s.dropAt {
		_ = s.Conn.Close()
		return 0, errors.New("synthetic lost ack")
	}
	return s.Conn.Write(p)
}
