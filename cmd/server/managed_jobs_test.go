package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func syntheticManagedJob(t *testing.T) (*server, *mockDB, http.Handler, managedJobRegistration) {
	t.Helper()
	s, db, _ := mockServer(t)
	p := room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection", Name: "Synthetic"}
	db.seed("participant.created", p)
	ctx := s.managedJobContext(context.Background(), p)
	registration := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	if err := registration.Register("worker", "codex"); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/job-diagnostics", s.jobDiagnostics)
	mux.HandleFunc("/api/job-recovery", s.jobRecovery)
	return s, db, s.auth.protect(s.actorGuard(mux)), registration
}
func TestJobJournalIsMandatoryWithDiagnosticsOffAndFailureStopsExecution(t *testing.T) {
	t.Parallel()
	s, db, _, registration := syntheticManagedJob(t)
	db.seed("feature.flag.updated", map[string]any{"name": "agentBackgroundWork", "enabled": false})
	db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
	if err := registration.Update("submitted"); err != nil {
		t.Fatal(err)
	}
	st, err := s.load(context.Background())
	if err != nil || st.ManagedJobs[registration.JobID].Status != "submitted" {
		t.Fatal("flag disabled required journal")
	}
	other, stop := context.WithCancel(context.Background())
	defer stop()
	s.cancels["other"] = stop
	if err = registration.Update("shutdown_unconfirmed"); err != nil || !s.stopAll || other.Err() == nil {
		t.Fatal("unconfirmed termination did not stop application", err)
	}
	if err = registration.Update("shutdown_confirmed"); err != nil {
		t.Fatal(err)
	}
	if err = registration.Update("shutdown_unconfirmed"); err != nil {
		t.Fatal(err)
	}
	st, _ = s.load(context.Background())
	if !st.ManagedJobs[registration.JobID].ShutdownConfirmed {
		t.Fatal("late uncertainty undid verified proof")
	}
	fresh, db2, _, reg := syntheticManagedJob(t)
	db2.failAt = "insert"
	if reg.Update("submitted") == nil || !fresh.integrityFault {
		t.Fatal("failed journal update admitted work")
	}
}
func TestJobAllocationCannotRepeatOrOverrideUnresolvedMembership(t *testing.T) {
	t.Parallel()
	s, _, _, registration := syntheticManagedJob(t)
	if registration.Register("worker", "codex") == nil {
		t.Fatal("same turn registered twice")
	}
	p := room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection"}
	next := s.managedJobContext(context.Background(), p).Value(managedJobContextKey{}).(managedJobRegistration)
	if next.Register("worker", "codex") == nil {
		t.Fatal("unresolved durable job replaced")
	}
	if registration.Update("shutdown_confirmed") != nil {
		t.Fatal("confirmation failed")
	}
	if next.Register("worker", "codex") != nil {
		t.Fatal("verified job prevented explicit next turn")
	}
}
func TestJobRecoveryPermissionsFlagsProofAndActiveTurnIdentity(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	for _, mode := range []string{"verified", "unknown-worker", "held-supervisor", "failed-stop", "audit-failure", "other-active-turn", "disabled-diagnostics", "non-owner"} {
		t.Run(mode, func(t *testing.T) {
			s, db, h, registration := syntheticManagedJob(t)
			st, _ := s.load(context.Background())
			job := st.ManagedJobs[registration.JobID]
			request := managedJobFence(job)
			record, _ := request.record()
			view := workerJobView{workerJobRecord: record, LeaseState: "free"}
			s.jobSnapshot = func(context.Context, string) (workerJobsSnapshot, error) {
				if mode == "unknown-worker" {
					return workerJobsSnapshot{}, errors.New("synthetic transport outage")
				}
				if mode == "held-supervisor" {
					view.LeaseState = "held"
				}
				return workerJobsSnapshot{Version: 2, Jobs: []workerJobView{view}}, nil
			}
			calls := 0
			s.jobFence = func(context.Context, string, workerFenceRequest) error {
				calls++
				if mode == "failed-stop" {
					return errors.New("synthetic shutdown outage")
				}
				return nil
			}
			if mode == "audit-failure" {
				db.failAt = "insert"
			}
			if mode == "other-active-turn" {
				s.active[job.ParticipantID] = true
				s.activeJobs[job.ParticipantID] = strings.Repeat("e", 32)
			}
			if mode == "disabled-diagnostics" {
				db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
				if mockRequest(h, "GET", "/api/job-diagnostics", "").Code != 403 {
					t.Fatal("disabled diagnostics exposed")
				}
			}
			if mode == "non-owner" {
				db.seed("actor.revised", room.Actor{ID: "qa", Kind: "human", Role: "QA", Enabled: true})
				s.auth.sessions["owner"] = ownerSession{ActorID: "qa", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
			}
			w := mockRequest(h, "POST", "/api/job-recovery", `{"action":"reconcile","jobId":"`+job.JobID+`","reason":"Reviewed synthetic job"}`)
			expected := 200
			switch mode {
			case "unknown-worker", "failed-stop", "audit-failure":
				expected = 503
			case "held-supervisor", "other-active-turn":
				expected = 409
			case "non-owner":
				expected = 403
			}
			if w.Code != expected {
				t.Fatal(mode, w.Code, w.Body.String())
			}
			if (mode == "audit-failure" || mode == "other-active-turn" || mode == "non-owner" || mode == "held-supervisor" || mode == "unknown-worker") && calls != 0 {
				t.Fatal("worker touched before authorization and proof")
			}
			if mode == "failed-stop" && !s.stopAll {
				t.Fatal("failed reconciliation left dispatch enabled")
			}
		})
	}
}
func TestJobDiagnosticsNeverExposeRawWorkerFieldsAndRespectRoom(t *testing.T) {
	t.Parallel()
	s, _, _, registration := syntheticManagedJob(t)
	st, _ := s.load(context.Background())
	job := st.ManagedJobs[registration.JobID]
	record, _ := managedJobFence(job).record()
	record.DiagnosticCode = "synthetic-private-output"
	record.Status = "synthetic-secret-status"
	views := boundedJobDiagnostics(st, workerJobsSnapshot{Version: 2, Jobs: []workerJobView{{workerJobRecord: record, LeaseState: "free"}}}, nil, "engineering", job.ParticipantID)
	if len(views) != 1 || views[0].WorkerStatus != "unverified" || views[0].DiagnosticCode == record.DiagnosticCode {
		t.Fatal("raw worker data exposed")
	}
	if len(boundedJobDiagnostics(st, workerJobsSnapshot{}, nil, "another-room", "")) != 0 {
		t.Fatal("job crossed room scope")
	}
	if boundedJobDiagnostics(st, workerJobsSnapshot{}, errors.New("private-error"), "engineering", "")[0].CanReconcile {
		t.Fatal("unavailable worker became stopped")
	}
}

func TestStartupAndResumeRefuseUnconfirmedControllerJournalEvenIfWorkerIdle(t *testing.T) {
	t.Parallel()
	s, _, _, _ := syntheticManagedJob(t)
	s.recoveryIdle = func(context.Context, string) (bool, error) { return true, nil }
	s.verifyWorkerAtStartup(context.Background(), "ssh://worker")
	if !s.stopAll {
		t.Fatal("missing remote job made unresolved controller intent safe")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/agent-safety", s.agentSafety)
	h := s.auth.protect(s.actorGuard(mux))
	w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"resume","scope":"application"}`)
	if w.Code != 409 || !s.stopAll {
		t.Fatal("resume bypassed unresolved intent", w.Code)
	}
}

func TestJobRecoveryRejectsMissingCSRFReasonAndWrongRoomBeforeWorker(t *testing.T) {
	t.Parallel()
	s, _, h, registration := syntheticManagedJob(t)
	s.jobFence = func(context.Context, string, workerFenceRequest) error {
		t.Fatal("invalid request contacted worker")
		return nil
	}
	for _, body := range []string{`{"action":"reconcile","jobId":"` + registration.JobID + `","reason":" "}`, `{"action":"unsafe","jobId":"` + registration.JobID + `","reason":"Synthetic"}`} {
		if mockRequest(h, "POST", "/api/job-recovery", body).Code != 400 {
			t.Fatal("invalid recovery body accepted")
		}
	}
	if mockRequest(h, "POST", "/api/job-recovery?room=another-room", `{"action":"cancel","jobId":"`+registration.JobID+`","reason":"Synthetic"}`).Code != 404 {
		t.Fatal("room scope bypassed")
	}
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/job-recovery", strings.NewReader(`{"action":"cancel","jobId":"`+registration.JobID+`","reason":"Synthetic"}`))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("CSRF bypassed", w.Code)
	}
}
