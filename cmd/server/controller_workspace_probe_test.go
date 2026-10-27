package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"agentroom/internal/workspacemount"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
)

type syntheticWorkspaceProbeEvidence struct {
	observe func(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error
	stage   func(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error
}

func (e *syntheticWorkspaceProbeEvidence) ValidateWorkspaceProbe(ctx context.Context, st room.State, r workerJobRecord, i toolbroker.WorkspaceProbeIntent, op string) error {
	if e.observe != nil {
		return e.observe(ctx, st, r, i, op)
	}
	return ctx.Err()
}
func (e *syntheticWorkspaceProbeEvidence) ValidateProbe(ctx context.Context, st room.State, r room.WorkspaceProbeReservation, event workerProbeEvent) error {
	if e.stage != nil {
		return e.stage(ctx, st, r, event)
	}
	return ctx.Err()
}
func controllerWorkspaceAuthorityFixture(t *testing.T) (*server, *mockDB, context.Context, workerJobRecord, toolbroker.WorkspaceProbeIntent) {
	t.Helper()
	_, r, _ := retainedObservationFixture(t)
	s, db, _ := mockServer(t)
	job := room.ManagedJob{Version: r.Version, Execution: r.Execution, RepositoryDigest: r.Pair.Repository, InitiatorID: ownerID, CompanyID: "company", JobID: r.JobID, RoomID: r.RoomID, ParticipantID: r.ParticipantID, ConnectionID: r.ConnectionID, Worker: "worker", Provider: r.Provider, Status: "shutdown_unconfirmed", AllocatedAt: r.StartedAt, UpdatedAt: r.UpdatedAt, Deadline: r.Deadline}
	if !validManagedJob(job) {
		t.Fatal("invalid synthetic job")
	}
	db.seed("worker.job.updated", job)
	i := toolbroker.WorkspaceProbeIntent{JobID: r.JobID, Repository: r.Pair.Repository, Volume: r.Pair.Workspace.Identity, Image: "sha256:" + strings.Repeat("d", 64), Challenge: strings.Repeat("e", 32)}
	return s, db, context.WithValue(t.Context(), actorContextKey{}, ownerID), r, i
}
func TestControllerWorkspaceProbeAuthorityDurableAndReplay(t *testing.T) {
	t.Parallel()
	s, db, ctx, r, i := controllerWorkspaceAuthorityFixture(t)
	e := &syntheticWorkspaceProbeEvidence{}
	g, err := newControllerWorkspaceProbeAuthority(ctx, s, "worker", r, i, e)
	if err != nil {
		t.Fatal(err)
	}
	before := len(db.snapshot())
	if g.Authorize(ctx, i) != nil || g.AuthorizeWorkspaceObservation(ctx, r) != nil || g.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) != nil {
		t.Fatal("healthy guard refused")
	}
	if len(db.snapshot()) != before+3 {
		t.Fatal("checks not durably acknowledged")
	}
	st := probeControllerState(t, &server{db: db}, ctx)
	if n, err := controllerProbeSlots(st); err != nil || n != 1 {
		t.Fatal("account slot not replayed", n, err)
	}
	if !strings.Contains(auditSummary(controllerProbeAuthorizationKind, []byte(`{}`)), "authorization") {
		t.Fatal("missing safe audit summary")
	}
	// The supplied operation context cannot replace the authenticated session actor.
	other := context.WithValue(ctx, actorContextKey{}, "untrusted")
	if g.Authorize(other, i) != nil {
		t.Fatal("caller context replaced session principal")
	}
	events := db.snapshot()
	var v controllerProbeAuthorization
	if json.Unmarshal([]byte(events[len(events)-1].payload), &v) != nil || v.ActorID != ownerID {
		t.Fatal("actor substituted", v)
	}
}
func TestControllerWorkspaceProbeAuthorityRefusalAndFencing(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"evidence", "persistence", "revoked session", "foreign worker", "missing principal", "stale snapshot", "current job changed", "record changed"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, r, i := controllerWorkspaceAuthorityFixture(t)
			e := &syntheticWorkspaceProbeEvidence{}
			session, cancel := context.WithCancel(ctx)
			defer cancel()
			worker := "worker"
			switch mode {
			case "evidence":
				e.observe = func(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error {
					return errors.New("synthetic denied")
				}
			case "persistence":
				db.failAt = "commit"
			case "foreign worker":
				worker = "other"
			case "missing principal":
				session = t.Context()
			case "stale snapshot":
				e.observe = func(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error {
					db.seed("agent.background", map[string]string{"stage": "synthetic"})
					return nil
				}
			}
			g, err := newControllerWorkspaceProbeAuthority(session, s, worker, r, i, e)
			if mode == "missing principal" {
				if err == nil {
					t.Fatal("anonymous session admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if mode == "revoked session" {
				cancel()
			}
			if mode == "current job changed" {
				job := probeControllerState(t, s, ctx).ManagedJobs[i.JobID]
				job.Worker = "other"
				db.seed("worker.job.updated", job)
			}
			before := len(db.snapshot())
			if mode == "record changed" {
				r.DiagnosticCode = "other"
				err = g.AuthorizeWorkspaceObservation(ctx, r)
			} else {
				err = g.Authorize(ctx, i)
			}
			if err == nil {
				t.Fatal("unsafe authority acknowledged")
			}
			if mode != "stale snapshot" && len(db.snapshot()) != before {
				t.Fatal("refusal appended grant")
			}
			if mode == "persistence" && !s.integrityFault {
				t.Fatal("audit failure did not halt")
			}
			if mode == "stale snapshot" && s.integrityFault {
				t.Fatal("ordinary conflict tripped integrity")
			}
		})
	}
}
func TestControllerWorkspaceProbeStagePinsWorkerInsideLedger(t *testing.T) {
	t.Parallel()
	s, db, ctx, r, i := controllerWorkspaceAuthorityFixture(t)
	calls := 0
	e := &syntheticWorkspaceProbeEvidence{stage: func(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error {
		calls++
		return nil
	}}
	g, err := newControllerWorkspaceProbeAuthority(ctx, s, "worker", r, i, e)
	if err != nil {
		t.Fatal(err)
	}
	job := probeControllerState(t, s, ctx).ManagedJobs[i.JobID]
	job.Worker = "other"
	db.seed("worker.job.updated", job)
	if g.Acknowledge(ctx, i, workerProbeEvent{Stage: "begin"}) == nil || calls != 0 {
		t.Fatal("changed worker reached physical evidence")
	}
}
func TestControllerWorkspaceProbeChannelIntegrated(t *testing.T) {
	t.Parallel()
	f := newAssembledProbeFixture(t, "healthy")
	s, db, _ := mockServer(t)
	r := f.record
	job := room.ManagedJob{Version: r.Version, Execution: r.Execution, RepositoryDigest: r.Pair.Repository, InitiatorID: ownerID, CompanyID: "company", JobID: r.JobID, RoomID: r.RoomID, ParticipantID: r.ParticipantID, ConnectionID: r.ConnectionID, Worker: "worker", Provider: r.Provider, Status: "shutdown_unconfirmed", AllocatedAt: r.StartedAt, UpdatedAt: r.UpdatedAt, Deadline: r.Deadline}
	db.seed("worker.job.updated", job)
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	e := &syntheticWorkspaceProbeEvidence{observe: func(ctx context.Context, st room.State, r workerJobRecord, i toolbroker.WorkspaceProbeIntent, op string) error {
		f.checkLock()
		return nil
	}, stage: func(ctx context.Context, st room.State, r room.WorkspaceProbeReservation, event workerProbeEvent) error {
		return f.Acknowledge(ctx, r.Intent, event)
	}}
	done := make(chan error, 1)
	connect := func(call context.Context, r workerJobRecord, i toolbroker.WorkspaceProbeIntent) (io.ReadWriteCloser, error) {
		g, err := newControllerWorkspaceProbeAuthority(ctx, s, "worker", r, i, e)
		if err != nil {
			return nil, err
		}
		left, right := net.Pipe()
		go func() { done <- serveProbeAuthority(call, right, r, i, g) }()
		return left, nil
	}
	proof, err := probeRetainedWorkerWorkspaceChannel(ctx, f.dir, r, syntheticObservationDockerRoot, f, "sha256:"+strings.Repeat("d", 64), connect, func() { t.Error("healthy authority stopped") })
	<-done
	if err != nil || !workspacemount.ValidProof(proof, f.challenge) || !f.removed {
		t.Fatal("assembled authority failed", err, proof)
	}
	st := probeControllerState(t, s, ctx)
	if slots, err := controllerProbeSlots(st); err != nil || slots != 0 {
		t.Fatal("confirmed physical absence did not release slot", slots, err)
	}
	if len(st.WorkspaceProbes) != 1 {
		t.Fatal("lifecycle not durably retained")
	}
}

func TestControllerWorkspaceProbeHumanPermissionsAndRevocation(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"agent", "disabled", "wrong room", "unrelated human", "revoked during evidence", "session canceled during evidence", "stale task"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, r, i := controllerWorkspaceAuthorityFixture(t)
			e := &syntheticWorkspaceProbeEvidence{}
			if mode == "agent" || mode == "disabled" || mode == "wrong room" || mode == "unrelated human" {
				actor := room.Actor{ID: "synthetic-human", Kind: "human", Enabled: true, RoomIDs: []string{r.RoomID}}
				if mode == "agent" {
					actor.Kind = "agent"
				}
				if mode == "disabled" {
					actor.Enabled = false
				}
				if mode == "wrong room" {
					actor.RoomIDs = []string{"elsewhere"}
				}
				db.seed("actor.revised", actor)
				ctx = context.WithValue(t.Context(), actorContextKey{}, actor.ID)
			}
			session, cancel := context.WithCancel(ctx)
			defer cancel()
			if mode == "revoked during evidence" {
				e.observe = func(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error {
					db.seed("actor.revised", room.Actor{ID: ownerID, Kind: "human", Enabled: false})
					return nil
				}
			}
			if mode == "session canceled during evidence" {
				e.observe = func(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error {
					cancel()
					return nil
				}
			}
			if mode == "stale task" {
				job := probeControllerState(t, s, ctx).ManagedJobs[i.JobID]
				job.TaskID = "task"
				job.TaskDispatchID = "dispatch"
				db.seed("worker.job.updated", job)
			}
			g, err := newControllerWorkspaceProbeAuthority(session, s, "worker", r, i, e)
			if err != nil {
				t.Fatal(err)
			}
			before := len(db.snapshot())
			if g.Authorize(t.Context(), i) == nil {
				t.Fatal("unauthorized grant")
			}
			for _, event := range db.snapshot()[before:] {
				if event.kind == controllerProbeAuthorizationKind {
					t.Fatal("revoked grant persisted")
				}
			}
		})
	}
}
func TestControllerWorkspaceProbeInternalOnlyAndFeatureIndependence(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			t.Parallel()
			s, db, ctx, r, i := controllerWorkspaceAuthorityFixture(t)
			for _, name := range []string{"taskDispatch", "taskDefinitions", "jobDiagnostics"} {
				db.seed("feature.flag.updated", map[string]any{"name": name, "enabled": enabled})
			}
			st := probeControllerState(t, s, ctx)
			for _, name := range []string{"taskDispatch", "taskDefinitions", "jobDiagnostics"} {
				if st.FeatureFlags[name] != enabled {
					t.Fatal("flag fixture not applied", name)
				}
			}
			g, err := newControllerWorkspaceProbeAuthority(ctx, s, "worker", r, i, &syntheticWorkspaceProbeEvidence{})
			if err != nil {
				t.Fatal(err)
			}
			if g.Authorize(ctx, i) != nil {
				t.Fatal("legacy metadata authority depended on optional flags")
			}
			// Optional flags cannot bypass evidence or human authorization.
			db.seed("actor.revised", room.Actor{ID: ownerID, Kind: "human", Enabled: false})
			if g.Authorize(ctx, i) == nil {
				t.Fatal("flag bypassed revoked human")
			}
		})
	}
	_, _, handler := mockServer(t)
	if w := mockRequest(handler, "POST", "/api/events", `{"kind":"workspace.probe.authorized","payload":{}}`); w.Code < 400 {
		t.Fatal("public authorization injection")
	}
}

func TestControllerWorkspaceProbeChannelAuditFailureWithholdsProof(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"begin", "bound", "finish"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			f := newAssembledProbeFixture(t, "healthy")
			s, db, _ := mockServer(t)
			r := f.record
			job := room.ManagedJob{Version: r.Version, Execution: r.Execution, RepositoryDigest: r.Pair.Repository, InitiatorID: ownerID, CompanyID: "company", JobID: r.JobID, RoomID: r.RoomID, ParticipantID: r.ParticipantID, ConnectionID: r.ConnectionID, Worker: "worker", Provider: r.Provider, Status: "shutdown_unconfirmed", AllocatedAt: r.StartedAt, UpdatedAt: r.UpdatedAt, Deadline: r.Deadline}
			db.seed("worker.job.updated", job)
			ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
			e := &syntheticWorkspaceProbeEvidence{observe: func(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error {
				f.checkLock()
				return nil
			}, stage: func(ctx context.Context, st room.State, res room.WorkspaceProbeReservation, event workerProbeEvent) error {
				if err := f.Acknowledge(ctx, res.Intent, event); err != nil {
					return err
				}
				if event.Stage == stage {
					db.failAt = "commit"
				}
				return nil
			}}
			done := make(chan error, 1)
			connect := func(call context.Context, r workerJobRecord, i toolbroker.WorkspaceProbeIntent) (io.ReadWriteCloser, error) {
				g, err := newControllerWorkspaceProbeAuthority(ctx, s, "worker", r, i, e)
				if err != nil {
					return nil, err
				}
				left, right := net.Pipe()
				go func() { done <- serveProbeAuthority(call, right, r, i, g) }()
				return left, nil
			}
			stopped := false
			proof, err := probeRetainedWorkerWorkspaceChannel(ctx, f.dir, r, syntheticObservationDockerRoot, f, "sha256:"+strings.Repeat("d", 64), connect, func() { stopped = true })
			<-done
			if err == nil || proof != (workspacemount.Proof{}) || !stopped || !s.integrityFault || !workerProbeStopped(f.dir) {
				t.Fatal("failed durable audit released proof/authority", err, proof)
			}
			if stage == "begin" && f.creates != 0 {
				t.Fatal("creation preceded durable begin")
			}
			if stage == "bound" && f.started {
				t.Fatal("start preceded durable bound")
			}
			// Failed absence acknowledgments never release the prior controller charge.
			if stage == "finish" {
				st := probeControllerState(t, s, ctx)
				if n, err := controllerProbeSlots(st); err != nil || n != 1 {
					t.Fatal("failed cleanup audit released account slot", n, err)
				}
			}
		})
	}
}
