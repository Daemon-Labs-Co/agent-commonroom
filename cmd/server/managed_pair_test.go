package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func managedPairFixture(t *testing.T) (*server, *mockDB, http.Handler, managedJobRegistration) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("connection.created", room.Connection{ID: "pair-connection", Name: "Synthetic provider", Protocol: "CLI", Provider: "openai", Endpoint: "ssh://worker", Revision: 1})
	p := room.Participant{ID: "pair-agent", RoomID: "engineering", ConnectionID: "pair-connection", Revision: 1}
	db.seed("participant.created", p)
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	repository := strings.Repeat("b", 64)
	ctx = s.managedExecutionJobContext(ctx, p, repository, func() (string, error) { return repository, nil })
	registration := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	if err := registration.Register("worker", "codex"); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/job-recovery", s.jobRecovery)
	mux.HandleFunc("/api/job-diagnostics", s.jobDiagnostics)
	return s, db, s.auth.protect(s.actorGuard(mux)), registration
}
func TestManagedPairJobPreservesVersionRepositoryAndPrincipal(t *testing.T) {
	t.Parallel()
	s, _, _, registration := managedPairFixture(t)
	if registration.Version != workerPairLifecycleVersion || registration.Repository == "" {
		t.Fatal("pair registration lost protocol")
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	job := st.ManagedJobs[registration.JobID]
	if job.Version != workerPairLifecycleVersion || job.Execution != workerPairExecution || job.RepositoryDigest != registration.Repository || job.InitiatorID != ownerID || job.CompanyID != st.Company.ID {
		t.Fatal("durable pair binding incomplete", job)
	}
	request := managedJobFence(job)
	if request.Version != 3 || request.Repository != job.RepositoryDigest {
		t.Fatal("pair recovery downgraded protocol")
	}
	record, err := request.record()
	if err != nil || record.Version != workerPairLifecycleVersion || record.Pair.Repository != job.RepositoryDigest {
		t.Fatal("pair fence lost binding", err)
	}
	if err := registration.Update("submitted"); err != nil {
		t.Fatal(err)
	}
}
func TestManagedPairRegistrationRechecksPolicyAndRequiresHuman(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"missing human", "disabled policy", "changed repository", "policy error", "expired context"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, _ := mockServer(t)
			p := room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection"}
			db.seed("participant.created", p)
			actor := ownerID
			if mode == "missing human" {
				actor = ""
			}
			ctx := context.WithValue(t.Context(), actorContextKey{}, actor)
			if mode == "expired context" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			repository := strings.Repeat("b", 64)
			current := repository
			var policyErr error
			ctx = s.managedExecutionJobContext(ctx, p, repository, func() (string, error) { return current, policyErr })
			switch mode {
			case "disabled policy":
				current = ""
			case "changed repository":
				current = strings.Repeat("c", 64)
			case "policy error":
				policyErr = errors.New("synthetic unavailable policy")
			}
			reg := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
			if err := reg.Register("worker", "codex"); err == nil {
				t.Fatal("unapproved paired job persisted")
			}
			st, err := s.load(t.Context())
			if err != nil || len(st.ManagedJobs) != 0 {
				t.Fatal("refused job gained durable authority", err)
			}
		})
	}
}
func TestManagedPairRecoveryUsesMatchingWorkerVersionAndDigest(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	for _, mode := range []string{"success", "diagnostics off", "wrong repository", "wrong deadline", "legacy worker", "held supervisor", "failed shutdown"} {
		t.Run(mode, func(t *testing.T) {
			s, db, h, registration := managedPairFixture(t)
			st, _ := s.load(t.Context())
			job := st.ManagedJobs[registration.JobID]
			view, _ := managedJobFence(job).record()
			view.Deadline = job.Deadline
			lease := "free"
			switch mode {
			case "diagnostics off":
				db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
			case "wrong repository":
				view.Pair.Repository = strings.Repeat("c", 64)
			case "wrong deadline":
				view.Deadline = view.Deadline.Add(time.Second)
			case "legacy worker":
				view.Version = 2
				view.Execution = ""
				view.Pair = nil
			case "held supervisor":
				lease = "held"
			}
			s.jobSnapshot = func(context.Context, string) (workerJobsSnapshot, error) {
				return workerJobsSnapshot{Version: 2, Jobs: []workerJobView{{workerJobRecord: view, LeaseState: lease}}}, nil
			}
			calls := 0
			s.jobFence = func(ctx context.Context, alias string, request workerFenceRequest) error {
				calls++
				if alias != "worker" || request.Version != 3 || request.Repository != job.RepositoryDigest || request.JobID != job.JobID {
					t.Fatal("recovery changed pair authority")
				}
				if mode == "failed shutdown" {
					return errors.New("synthetic tool survivor")
				}
				return nil
			}
			response := mockRequest(h, "POST", "/api/job-recovery", `{"action":"reconcile","jobId":"`+job.JobID+`","reason":"Reviewed synthetic pair"}`)
			want := 200
			switch mode {
			case "wrong repository", "wrong deadline", "legacy worker", "held supervisor":
				want = 409
			case "failed shutdown":
				want = 503
			}
			if response.Code != want {
				t.Fatal(mode, response.Code, response.Body.String())
			}
			if want == 409 && calls != 0 {
				t.Fatal("unverified pair contacted for recovery")
			}
			latest, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			if latest.ManagedJobs[job.JobID].ShutdownConfirmed != (want == 200) {
				t.Fatal("incorrect durable pair settlement")
			}
		})
	}
}
func TestManagedPairProjectionRejectsVersionAndRepositoryChanges(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"repository", "downgrade"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, _, registration := managedPairFixture(t)
			st, _ := s.load(t.Context())
			job := st.ManagedJobs[registration.JobID]
			if mode == "repository" {
				job.RepositoryDigest = strings.Repeat("c", 64)
			} else {
				job.Version = 2
				job.RepositoryDigest = ""
				job.Execution = ""
			}
			db.seed("worker.job.updated", job)
			if _, err := s.load(t.Context()); err == nil {
				t.Fatal("pair identity mutation accepted")
			}
		})
	}
}
func TestManagedPairAuditCannotFollowNewRepositoryPolicy(t *testing.T) {
	t.Parallel()
	s, _, _, registration := managedPairFixture(t)
	if err := registration.Update("submitted"); err != nil {
		t.Fatal(err)
	}
	st, _ := s.load(t.Context())
	job := st.ManagedJobs[registration.JobID]
	scope := toolbroker.Scope{Version: 1, JobID: job.JobID, Room: job.RoomID, Participant: job.ParticipantID, Connection: job.ConnectionID, Provider: job.Provider, RepositoryDigest: strings.Repeat("c", 64), Deadline: job.Deadline}
	if _, err := newControllerToolAuditor(t.Context(), s, scope, ownerID, func() (string, error) { return scope.RepositoryDigest, nil }); err != toolbroker.ErrAudit {
		t.Fatal("new policy silently rebound original job", err)
	}
	scope.RepositoryDigest = job.RepositoryDigest
	if _, err := newControllerToolAuditor(t.Context(), s, scope, ownerID, func() (string, error) { return scope.RepositoryDigest, nil }); err != nil {
		t.Fatal("original pair audit unavailable", err)
	}
}
func TestPairedRegistrationCannotUseLegacyRemoteLaunch(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	_, _, _, registration := managedPairFixture(t)
	registration.Register = func(string, string) error { t.Fatal("paired registration reached legacy dispatch"); return nil }
	registration.Update = func(string) error { t.Fatal("legacy dispatch changed paired lifecycle"); return nil }
	ctx := context.WithValue(t.Context(), managedJobContextKey{}, registration)
	// Configured transport must reject the protocol mismatch before registration
	// or any subprocess preparation.
	command, cleanup, err := remoteAgentCommand(ctx, "ssh://worker", "engineering/pair-agent/pair-connection/worker", append(providerTurnArgs("codex", false, false), "-"))
	if err == nil || !strings.Contains(err.Error(), "Durable managed job registration") || command != nil || cleanup != nil {
		t.Fatal("paired identity reached legacy launch")
	}
}

func TestManagedPairLifecycleUpdatesRetainInitiatingHuman(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	p := room.Participant{ID: "pair-agent", RoomID: "engineering", ConnectionID: "pair-connection"}
	db.seed("participant.created", p)
	db.seed("actor.revised", room.Actor{ID: "human", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}, Revision: 1})
	repository := strings.Repeat("b", 64)
	ctx := context.WithValue(t.Context(), actorContextKey{}, "human")
	ctx = s.managedExecutionJobContext(ctx, p, repository, func() (string, error) { return repository, nil })
	registration := ctx.Value(managedJobContextKey{}).(managedJobRegistration)
	if err := registration.Register("worker", "codex"); err != nil {
		t.Fatal(err)
	}
	if err := registration.Update("submitted"); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, event := range db.snapshot() {
		if event.kind == "worker.job.updated" {
			var metadata struct {
				PerformedBy string `json:"performedBy"`
			}
			if json.Unmarshal([]byte(event.payload), &metadata) != nil || metadata.PerformedBy != "human" {
				t.Fatal("lifecycle attributed to credential owner")
			}
			count++
		}
	}
	if count != 2 {
		t.Fatal("expected allocation and submission evidence", count)
	}
}
