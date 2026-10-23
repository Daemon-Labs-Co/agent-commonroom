package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func coordinationFixture(t *testing.T) (*server, *mockDB, http.Handler, room.Participant, room.GovernanceAppointment) {
	t.Helper()
	s, db, h := governanceFixture(t, true)
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	p := st.Participants[0]
	appointment := room.GovernanceAppointment{CompanyID: st.Company.ID, RoomID: p.RoomID, ParticipantID: p.ID, Revision: 1}
	db.seed("governance.appointed", map[string]any{"companyId": appointment.CompanyID, "roomId": p.RoomID, "participantId": p.ID, "revision": 1, "performedBy": ownerID})
	return s, db, h, p, appointment
}
func coordinatedContext(p room.GovernanceAppointment) context.Context {
	return context.WithValue(withInitiatingActor(context.Background(), ownerID), coordinationContextKey{}, coordinationBinding{Appointment: p})
}
func replaceCoordinator(t *testing.T, h http.Handler) {
	t.Helper()
	w := governanceRequest(h, "owner", "replace-coordinator", `{"participantId":"b","revision":1}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestCoordinationAdmissionAndSubmissionRevalidateRevision(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"healthy", "queued-transfer", "allocated-transfer", "disabled", "reenabled", "foreign-snapshot", "non-owner"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, h, p, appointment := coordinationFixture(t)
			ctx := coordinatedContext(appointment)
			if mode == "foreign-snapshot" {
				appointment.RoomID = "foreign"
				ctx = coordinatedContext(appointment)
			}
			if mode == "non-owner" {
				ctx = withInitiatingActor(ctx, "other")
			}
			registration := s.managedJobContext(ctx, p).Value(managedJobContextKey{}).(managedJobRegistration)
			if mode == "queued-transfer" {
				replaceCoordinator(t, h)
			}
			if mode == "disabled" || mode == "reenabled" {
				db.seed("feature.flag.updated", map[string]any{"name": "agentCoordination", "enabled": false})
			}
			if mode == "reenabled" {
				db.seed("feature.flag.updated", map[string]any{"name": "agentCoordination", "enabled": true})
			}
			before := len(db.snapshot())
			err := registration.Register("worker", "codex")
			if mode != "healthy" && mode != "allocated-transfer" {
				if err == nil || len(db.snapshot()) != before {
					t.Fatal("stale coordination admitted", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if mode == "allocated-transfer" {
				replaceCoordinator(t, h)
			}
			err = registration.Update("submitted")
			if (err == nil) != (mode == "healthy") {
				t.Fatal("incorrect submission", err)
			}
			st, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			job := st.ManagedJobs[registration.JobID]
			if job.GovernanceRevision != 1 || job.CoordinatorID != p.ID {
				t.Fatal("lost role dependency", job)
			}
			// Revocation cannot prevent cleanup evidence or invent shutdown confirmation.
			if job.ShutdownConfirmed {
				t.Fatal("revocation fabricated shutdown")
			}
			if registration.Update("shutdown_confirmed") != nil {
				t.Fatal("cleanup evidence blocked")
			}
		})
	}
}

func TestCoordinationToolAuditRevalidatesAfterTransferAndRestart(t *testing.T) {
	t.Parallel()
	for _, independent := range []bool{false, true} {
		for _, revoke := range []string{"transfer", "disabled", "reenabled", "removed", "stop"} {
			t.Run(revoke+map[bool]string{true: "/independent", false: "/role"}[independent], func(t *testing.T) {
				t.Parallel()
				s, db, h, p, appointment := coordinationFixture(t)
				c := room.Connection{ID: p.ConnectionID, Protocol: "CLI", Provider: "openai", Endpoint: "ssh://worker", Revision: 1}
				db.seed("connection.updated", c)
				now := time.Now().UTC()
				scope := toolbroker.Scope{Version: 1, JobID: strings.Repeat("a", 32), Room: p.RoomID, Participant: p.ID, Connection: p.ConnectionID, Provider: "codex", RepositoryDigest: strings.Repeat("b", 64), Deadline: now.Add(time.Minute)}
				job := room.ManagedJob{Version: 2, InitiatorID: ownerID, CompanyID: appointment.CompanyID, JobID: scope.JobID, RoomID: p.RoomID, ParticipantID: p.ID, ConnectionID: p.ConnectionID, Worker: "worker", Provider: "codex", Status: "submitted", AllocatedAt: now, UpdatedAt: now, Deadline: scope.Deadline}
				if !independent {
					job.CoordinatorID = p.ID
					job.GovernanceRevision = 1
				}
				db.seed("worker.job.updated", job)
				repository := func() (string, error) { return scope.RepositoryDigest, nil }
				auditor, err := newControllerToolAuditor(t.Context(), s, scope, ownerID, repository)
				if err != nil {
					t.Fatal(err)
				}
				intent := toolbroker.Intent{Scope: scope, CallID: "call", Operation: "read", PayloadDigest: strings.Repeat("c", 64)}
				if auditor.RecordIntent(t.Context(), intent) != nil {
					t.Fatal("healthy intent refused")
				}
				switch revoke {
				case "transfer":
					replaceCoordinator(t, h)
				case "disabled", "reenabled":
					db.seed("feature.flag.updated", map[string]any{"name": "agentCoordination", "enabled": false})
					if revoke == "reenabled" {
						db.seed("feature.flag.updated", map[string]any{"name": "agentCoordination", "enabled": true})
					}
				case "removed":
					db.seed("participant.removed", map[string]string{"id": p.ID})
				case "stop":
					s.stopAll = true
				}
				allowed := independent && (revoke == "transfer" || revoke == "disabled" || revoke == "reenabled")
				before := len(db.snapshot())
				err = auditor.RecordOutcome(t.Context(), intent, toolbroker.Outcome{Status: "succeeded", OutputBytes: 2, OutputDigest: strings.Repeat("d", 64)})
				if (err == nil) != allowed {
					t.Fatal("wrong effect decision", err)
				}
				if !allowed && len(db.snapshot()) != before {
					t.Fatal("revoked outcome persisted")
				}
				restarted := &server{db: db, stopAll: s.stopAll}
				_, err = newControllerToolAuditor(t.Context(), restarted, scope, ownerID, repository)
				if (err == nil) != allowed {
					t.Fatal("restart revived stale authority", err)
				}
				intent.CallID = "next"
				if (auditor.RecordIntent(t.Context(), intent) == nil) != allowed {
					t.Fatal("wrong next-intent decision")
				}
			})
		}
	}
}

func TestCoordinationTransferCancelsOnlyRoleBoundTurnsAndWithholdsStaleResults(t *testing.T) {
	t.Parallel()
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{true: "audit-failure", false: "transfer"}[fail], func(t *testing.T) {
			t.Parallel()
			s, db, h, p, appointment := coordinationFixture(t)
			role, cancel := context.WithCancel(coordinatedContext(appointment))
			defer cancel()
			independent, stop := context.WithCancel(context.Background())
			defer stop()
			s.cancels[p.ID] = cancel
			s.cancels["independent"] = stop
			s.coordinationTurns = map[string]room.GovernanceAppointment{p.ID: appointment}
			if fail {
				db.failAt = "commit"
			}
			w := governanceRequest(h, "owner", "transfer", `{"participantId":"b","revision":1}`)
			if !fail && independent.Err() != nil {
				t.Fatal("transfer cancelled independent work")
			}
			if fail {
				if w.Code != 500 || !s.integrityFault {
					t.Fatal("failed transfer not stopped", w.Code)
				}
				db.failAt = ""
				st, err := s.load(t.Context())
				if err != nil || st.Governance[p.RoomID] != appointment {
					t.Fatal("failed transfer changed appointment", err)
				}
				return
			}
			if w.Code != 200 || role.Err() != context.Canceled {
				t.Fatal("role turn not cancelled", w.Code)
			}
			// Simulate another controller: no local cancellation, but stale result must
			// still be discarded when the current appointment is read at publication.
			s.finishAgentTurn(coordinatedContext(appointment), p, "synthetic stale result", "codex", nil)
			st, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range st.Messages {
				if m.Text == "synthetic stale result" {
					t.Fatal("stale result published")
				}
			}
			if len(st.ManagedJobs) != 0 {
				t.Fatal("transfer manufactured worker proof")
			}
		})
	}
}

func TestCoordinationSendHandlerCapturesDependencyAndRejectsStaleInput(t *testing.T) {
	// Environment config is synthetic and this instance never invokes a provider.
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	for _, mode := range []string{"healthy", "transfer", "disabled", "stale", "other-target", "broadcast", "non-owner"} {
		t.Run(mode, func(t *testing.T) {
			s, db, h, p, appointment := coordinationFixture(t)
			db.seed("connection.updated", room.Connection{ID: p.ConnectionID, Protocol: "CLI", Endpoint: "ssh://worker"})
			body := `{"action":"send","participantId":"a","text":"Synthetic bounded turn","coordinationRevision":1}`
			if mode == "disabled" {
				db.seed("feature.flag.updated", map[string]any{"name": "agentCoordination", "enabled": false})
			}
			if mode == "stale" {
				body = strings.Replace(body, `"coordinationRevision":1`, `"coordinationRevision":2`, 1)
			}
			if mode == "other-target" {
				body = strings.Replace(body, `"participantId":"a"`, `"participantId":"b"`, 1)
			}
			if mode == "broadcast" {
				body = strings.Replace(body, `"action":"send"`, `"action":"broadcast"`, 1)
			}
			if mode == "non-owner" {
				db.seed("actor.created", room.Actor{ID: "human", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{p.RoomID}})
				s.auth.sessions["human"] = ownerSession{ActorID: "human", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
			}
			started := make(chan context.Context, 1)
			release := make(chan struct{})
			defer close(release)
			s.agentTurn = func(ctx context.Context, _ room.Participant, _, _ string) (string, error) {
				started <- ctx
				select {
				case <-release:
				case <-ctx.Done():
				}
				return "synthetic result", nil
			}
			before := len(db.snapshot())
			r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/agents", strings.NewReader(body))
			cookie := "owner"
			if mode == "non-owner" {
				cookie = "human"
			}
			r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: cookie})
			r.Header.Set("X-CSRF-Token", "csrf")
			r.Header.Set("Idempotency-Key", "send-coordination")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if mode != "healthy" && mode != "transfer" {
				want := 409
				if mode == "non-owner" {
					want = 403
				}
				if w.Code != want || len(db.snapshot()) != before {
					t.Fatal("invalid role send admitted", w.Code, w.Body.String())
				}
				return
			}
			if w.Code != 202 {
				t.Fatal(w.Code, w.Body.String())
			}
			var ctx context.Context
			select {
			case ctx = <-started:
			case <-time.After(time.Second):
				t.Fatal("turn did not start")
			}
			if ctx.Value(coordinationContextKey{}) != (coordinationBinding{Appointment: appointment}) || ctx.Value(actorContextKey{}) != ownerID {
				t.Fatal("queue lost dependency or human identity")
			}
			if mode == "transfer" {
				replaceCoordinator(t, h)
				if ctx.Err() != context.Canceled {
					t.Fatal("transfer did not cancel queued turn")
				}
			}
			// Acknowledged replay returns history and cannot dispatch another turn.
			r2 := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/agents", strings.NewReader(body))
			r2.Header = r.Header.Clone()
			r2.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
			replay := httptest.NewRecorder()
			h.ServeHTTP(replay, r2)
			if replay.Code != 202 || replay.Header().Get("Idempotent-Replay") != "true" {
				t.Fatal("receipt replay dispatched or refused", replay.Code)
			}
			s.mu.Lock()
			if cancel := s.cancels[p.ID]; cancel != nil {
				cancel()
			}
			s.mu.Unlock()
			deadline := time.Now().Add(time.Second)
			for {
				s.mu.Lock()
				active := s.active[p.ID]
				s.mu.Unlock()
				if !active {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("turn cleanup incomplete")
				}
				time.Sleep(time.Millisecond)
			}
		})
	}
}

func TestCoordinationDisableCancelsRoleTurnsAndFreshEnableDoesNotReviveOldWork(t *testing.T) {
	t.Parallel()
	s, db, h, p, appointment := coordinationFixture(t)
	role, cancel := context.WithCancel(coordinatedContext(appointment))
	defer cancel()
	independent, stop := context.WithCancel(context.Background())
	defer stop()
	s.coordinationTurns = map[string]room.GovernanceAppointment{p.ID: appointment}
	s.cancels[p.ID] = cancel
	s.cancels["independent"] = stop
	w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"agentCoordination","enabled":false}}`)
	if w.Code != 200 || role.Err() != context.Canceled || independent.Err() != nil {
		t.Fatal("disable cancellation scope", w.Code)
	}
	w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"agentCoordination","enabled":true}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if st.Governance[p.RoomID] != appointment || st.CoordinationEpoch != 1 {
		t.Fatal("lost appointment or revocation generation")
	}
	if coordinationContextAllowed(coordinatedContext(appointment), st, p) {
		t.Fatal("old context revived")
	}
	fresh := context.WithValue(withInitiatingActor(t.Context(), ownerID), coordinationContextKey{}, coordinationBinding{Appointment: appointment, Epoch: st.CoordinationEpoch})
	registration := s.managedJobContext(fresh, p).Value(managedJobContextKey{}).(managedJobRegistration)
	if registration.Register("worker", "claude") != nil || registration.Update("submitted") != nil {
		t.Fatal("new explicit authority refused")
	}
	st, err = (&server{db: db}).load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	job := st.ManagedJobs[registration.JobID]
	if job.CoordinationEpoch != 1 || !coordinationJobAllowed(st, job) {
		t.Fatal("restart lost fresh binding")
	}
	job.GovernanceRevision = 0
	job.CoordinatorID = ""
	job.CoordinationEpoch = 0
	db.seed("worker.job.updated", job)
	if _, err = s.load(t.Context()); err == nil {
		t.Fatal("journal accepted stripping role dependency")
	}
}
