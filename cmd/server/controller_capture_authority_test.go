package main

import (
	"agentroom/internal/room"
	"agentroom/internal/workspacecapture"
	"agentroom/internal/workspacemount"
	"agentroom/internal/workspacesnapshot"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"strings"
	"testing"
	"time"
)

type syntheticCaptureEvidence struct {
	check func(context.Context, room.State, workerJobRecord, workspacecapture.Request, workspacemount.Proof) error
	calls int
}

func (e *syntheticCaptureEvidence) ValidateCapture(ctx context.Context, st room.State, r workerJobRecord, q workspacecapture.Request, p workspacemount.Proof) error {
	e.calls++
	if e.check != nil {
		return e.check(ctx, st, r, q, p)
	}
	return ctx.Err()
}
func controllerCaptureFixture(t *testing.T) (*server, *mockDB, context.Context, workerJobRecord, workspacecapture.Request, workspacemount.Proof) {
	t.Helper()
	s, db, ctx, scope := artifactJournalFixture(t, false)
	st := artifactState(t, s, ctx)
	job := st.ManagedJobs[scope.JobID]
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
	if !retainedObservationRecord(r) {
		t.Fatal("invalid synthetic retained binding", r)
	}
	binding := workspacesnapshot.Binding{Execution: scope.ExecutionScope, BaseCommit: strings.Repeat("7", 40), VolumeName: r.Pair.Workspace.Identity.Name, VolumeNonce: r.Pair.Workspace.Identity.Nonce, VolumeCreatedAt: r.Pair.Workspace.Identity.CreatedAt}
	binding.Execution.RevisionDigest = ""
	q, err := workspacecapture.ContractRequest(st.TaskExecutions[scope.ExecutionID].Definition, binding, strings.Repeat("8", 32), []string{"report"})
	if err != nil {
		t.Fatal(err)
	}
	proof := workspacemount.Proof{Version: 1, Challenge: q.Challenge, MountID: 2, ReadOnly: true, RootReadOnly: true}
	return s, db, ctx, r, q, proof
}
func TestControllerCaptureAuthorityDurableReplayAndPinnedActor(t *testing.T) {
	t.Parallel()
	s, db, ctx, r, q, p := controllerCaptureFixture(t)
	e := &syntheticCaptureEvidence{}
	g, err := newControllerCaptureAuthority(ctx, s, "synthetic-worker", r, q, e)
	if err != nil {
		t.Fatal(err)
	}
	before := len(db.snapshot())
	other := context.WithValue(ctx, actorContextKey{}, "untrusted")
	if err = g.AuthorizeFrozenSource(other, q.Binding, p); err != nil {
		t.Fatal(err)
	}
	if len(db.snapshot()) != before+1 || e.calls != 1 {
		t.Fatal("authorization not durable")
	}
	events := db.snapshot()
	var v controllerCaptureAuthorization
	if json.Unmarshal([]byte(events[len(events)-1].payload), &v) != nil || v.ActorID != ownerID {
		t.Fatal("session actor replaced")
	}
	st := artifactState(t, s, ctx)
	if !validControllerCaptureAuthorization(st, v) || projectControllerCaptureAuthorization(&st, []byte(events[len(events)-1].payload)) != nil {
		t.Fatal("replay failed")
	}
	if !strings.Contains(auditSummary(controllerCaptureAuthorizationKind, nil), "authorization") {
		t.Fatal("unsafe summary")
	}
	// Session metadata is detached from caller slices and retained records.
	q.Members[0].Name = "changed"
	r.Pair.Workspace.Identity.Nonce = strings.Repeat("9", 32)
	if g.AuthorizeFrozenSource(ctx, g.request.Binding, p) != nil {
		t.Fatal("caller mutated pinned authority")
	}
}
func TestControllerCaptureAuthorityDenials(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"evidence", "binding", "challenge", "session canceled", "foreign worker", "missing actor", "stale snapshot", "commit", "feature off", "agent actor"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, ctx, r, q, p := controllerCaptureFixture(t)
			session, cancel := context.WithCancel(ctx)
			defer cancel()
			worker := "synthetic-worker"
			e := &syntheticCaptureEvidence{}
			switch mode {
			case "evidence":
				e.check = func(context.Context, room.State, workerJobRecord, workspacecapture.Request, workspacemount.Proof) error {
					return errors.New("synthetic physical denial")
				}
			case "foreign worker":
				worker = "other"
			case "missing actor":
				session = t.Context()
			case "agent actor":
				session = context.WithValue(ctx, actorContextKey{}, "dev")
			case "stale snapshot":
				e.check = func(context.Context, room.State, workerJobRecord, workspacecapture.Request, workspacemount.Proof) error {
					db.seed("agent.background", map[string]string{"stage": "synthetic"})
					return nil
				}
			}
			g, err := newControllerCaptureAuthority(session, s, worker, r, q, e)
			if mode == "missing actor" {
				if err == nil {
					t.Fatal("unauthenticated session")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			before := len(db.snapshot())
			switch mode {
			case "binding":
				q.Binding.BaseCommit = strings.Repeat("9", 40)
			case "challenge":
				p.Challenge = strings.Repeat("9", 32)
			case "session canceled":
				cancel()
			case "commit":
				db.failAt = "commit"
			case "feature off":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": false})
				before++
			}
			err = g.AuthorizeFrozenSource(ctx, q.Binding, p)
			if mode == "feature off" {
				defer db.seed("feature.flag.updated", map[string]any{"name": "taskDefinitions", "enabled": true})
			}
			if mode == "commit" {
				db.failAt = ""
				if !s.integrityFault {
					t.Fatal("failed persistence did not stop integrity")
				}
				s.integrityFault = false
			}
			if err == nil {
				t.Fatal("denial granted")
			}
			extra := 0
			if mode == "stale snapshot" {
				extra = 1
				if !errors.Is(err, errAuditConflict) {
					t.Fatal("missing CAS conflict", err)
				}
			}
			if len(db.snapshot()) != before+extra {
				t.Fatal("denial recorded grant")
			}
		})
	}
}
func TestControllerCaptureAuthorityCurrentPolicyAndHistoricalReplay(t *testing.T) {
	t.Parallel()
	s, _, ctx, r, q, p := controllerCaptureFixture(t)
	g, err := newControllerCaptureAuthority(ctx, s, "synthetic-worker", r, q, &syntheticCaptureEvidence{})
	if err != nil {
		t.Fatal(err)
	}
	st, v, err := g.current(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"worker", "expiry", "shutdown", "task binding", "record volume", "member", "definition", "feature off", "incident"} {
		t.Run(mode, func(t *testing.T) {
			changed := st
			changed.ManagedJobs = maps.Clone(st.ManagedJobs)
			changed.FeatureFlags = maps.Clone(st.FeatureFlags)
			value := v
			value.Request = copyCaptureRequest(v.Request)
			j := changed.ManagedJobs[q.Binding.Execution.JobID]
			switch mode {
			case "worker":
				j.Worker = "other"
			case "expiry":
				j.Deadline = time.Now().Add(-time.Hour)
				j.AllocatedAt = j.Deadline.Add(-time.Minute)
				j.UpdatedAt = j.AllocatedAt
			case "shutdown":
				j.ShutdownConfirmed = true
				j.Status = "stopped"
			case "task binding":
				j.TaskDispatchID = strings.Repeat("9", 32)
			case "record volume":
				value.Request.Binding.VolumeNonce = "bad"
			case "member":
				value.Request.Members[0].Name = "other.txt"
			case "definition":
				value.Request.Binding.Execution.DefinitionDigest = strings.Repeat("9", 64)
			case "feature off":
				changed.FeatureFlags["taskDefinitions"] = false
			case "incident":
				changed.Incident = true
			}
			changed.ManagedJobs[j.JobID] = j
			historical := validControllerCaptureAuthorization(changed, value)
			if liveControllerCaptureBinding(changed, value, r) {
				t.Fatal("stale/current policy granted live capture")
			}
			if mode == "expiry" || mode == "feature off" || mode == "incident" {
				if !historical {
					t.Fatal("historical replay depends on wall time")
				}
			} else if mode != "shutdown" && historical {
				t.Fatal("historical identity substitution accepted")
			}
		})
	}
}

func TestControllerCaptureAuthorityRejectsPublicInjection(t *testing.T) {
	_, _, handler := mockServer(t)
	if w := mockRequest(handler, "POST", "/api/events", `{"kind":"workspace.capture.authorized","payload":{}}`); w.Code < 400 {
		t.Fatal("public capture authorization injection")
	}
}

func TestControllerCaptureAuthorityMissingInfrastructureRefuses(t *testing.T) {
	var absent *controllerCaptureAuthority
	for _, g := range []*controllerCaptureAuthority{absent, {}} {
		if g.AuthorizeFrozenSource(t.Context(), workspacesnapshot.Binding{}, workspacemount.Proof{}) == nil {
			t.Fatal("missing authority granted observation")
		}
	}
	s, _, ctx, r, q, _ := controllerCaptureFixture(t)
	var evidence *syntheticCaptureEvidence
	if _, err := newControllerCaptureAuthority(ctx, s, "synthetic-worker", r, q, evidence); err == nil {
		t.Fatal("typed-nil physical evidence accepted")
	}
}
