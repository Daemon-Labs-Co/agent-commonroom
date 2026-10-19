package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func idleSettingsHTTP(t *testing.T) (*server, *mockDB, http.Handler) {
	t.Helper()
	s, db, _ := mockServer(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", s.events)
	return s, db, s.auth.protect(s.actorGuard(mux))
}
func TestOwnerIdleSettingsPermissionsAndRevision(t *testing.T) {
	t.Parallel()
	s, db, h := idleSettingsHTTP(t)
	for _, body := range []string{`{"minutes":0,"revision":0}`, `{"minutes":61,"revision":0}`, `{"minutes":1.5,"revision":0}`, `{"minutes":5}`, `{"minutes":null,"revision":0}`, `{"minutes":1,"minutes":5,"revision":0}`, `{"minutes":5,"revision":0,"enabled":false}`} {
		before := len(db.snapshot())
		r := mockRequest(h, "POST", "/api/events", `{"kind":"agents.idle_policy.updated","payload":`+body+`}`)
		if r.Code != 400 || len(db.snapshot()) != before {
			t.Fatal("invalid idle policy admitted", r.Code, body)
		}
	}
	body := `{"kind":"agents.idle_policy.updated","payload":{"minutes":2,"revision":0}}`
	if r := mockRequest(h, "POST", "/api/events", body); r.Code != 200 {
		t.Fatal(r.Code, r.Body.String())
	}
	if r := mockRequest(h, "POST", "/api/events", body); r.Code != 409 {
		t.Fatal("stale save accepted", r.Code)
	}
	st, err := (&server{db: db}).load(t.Context())
	if err != nil || st.AgentIdleSettings != (room.AgentIdleSettings{Revision: 1, Minutes: 2}) {
		t.Fatal("settings not durable", err)
	}
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	before := len(db.snapshot())
	r := mockRequest(h, "POST", "/api/events", `{"kind":"agents.idle_policy.updated","payload":{"minutes":1,"revision":1}}`)
	if r.Code != 403 || len(db.snapshot()) != before {
		t.Fatal("non-owner changed interval")
	}
}
func TestIdlePolicySnapshotAndWorkerBinding(t *testing.T) {
	t.Parallel()
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "on", false: "off"}[enabled], func(t *testing.T) {
			s, db, _ := mockServer(t)
			db.seed("agents.idle_policy.updated", room.AgentIdleSettings{Revision: 1, Minutes: 2})
			db.seed("feature.flag.updated", map[string]any{"name": "idleAgentShutdown", "enabled": enabled})
			p := room.Participant{ID: "pair-agent", RoomID: "engineering", ConnectionID: "pair-connection"}
			db.seed("participant.created", p)
			ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
			repo := strings.Repeat("b", 64)
			registration := s.managedExecutionJobContextVersion(ctx, p, repo, func() (string, error) { return repo, nil }, workerWorkspaceLifecycleVersion).Value(managedJobContextKey{}).(managedJobRegistration)
			if err := registration.Register("worker", "codex"); err != nil {
				t.Fatal(err)
			}
			st, err := s.load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			job := st.ManagedJobs[registration.JobID]
			if job.IdlePolicyVersion != 1 || job.IdleMinutes != 2 || job.IdleShutdownEnabled != enabled || job.IdlePolicyRevision != 1 {
				t.Fatal("policy not snapshotted", job)
			}
			db.seed("agents.idle_policy.updated", room.AgentIdleSettings{Revision: 2, Minutes: 10})
			db.seed("feature.flag.updated", map[string]any{"name": "idleAgentShutdown", "enabled": !enabled})
			st, err = s.load(ctx)
			if err != nil || st.ManagedJobs[job.JobID] != job {
				t.Fatal("existing job policy changed", err)
			}
			// Bind the exact registered job to an admitted synthetic worker record.
			request := workerRunRequest{Version: job.Version, IssuedAt: job.AllocatedAt, Deadline: job.Deadline, JobID: job.JobID, Provider: job.Provider, Repository: job.RepositoryDigest, Key: job.RoomID + "/" + job.ParticipantID + "/" + job.ConnectionID + "/worker"}
			supervisor, err := admitWorkerPair(ctx, t.TempDir(), request, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer supervisor.Close()
			runner := &pairFakeDocker{containers: map[string]pairFakeContainer{}}
			foreign := job
			foreign.RepositoryDigest = strings.Repeat("f", 64)
			if supervisor.configureJobIdle(foreign, runner) == nil {
				t.Fatal("foreign idle policy binding accepted")
			}
			if err := supervisor.configureJobIdle(job, runner); err != nil {
				t.Fatal(err)
			}
			if supervisor.idle.policy != (workerIdlePolicy{Enabled: enabled, Minutes: 2}) {
				t.Fatal("worker ignored controller snapshot")
			}
			// Replay cannot replace an allocation's immutable policy.
			job.IdleMinutes = 3
			raw, _ := json.Marshal(job)
			db.seed("worker.job.updated", json.RawMessage(raw))
			if _, err := s.load(ctx); err == nil {
				t.Fatal("replay rewrote admitted policy")
			}
		})
	}
}

func TestIdleSettingsAuditFailureAndCSRF(t *testing.T) {
	t.Parallel()
	s, db, h := idleSettingsHTTP(t)
	p, err := idlePolicyForState(room.State{})
	if err != nil || p != (workerIdlePolicy{Enabled: true, Minutes: 5}) {
		t.Fatal("default policy", p, err)
	}
	body := `{"kind":"agents.idle_policy.updated","payload":{"minutes":2,"revision":0}}`
	before := len(db.snapshot())
	request := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/events", strings.NewReader(body))
	request.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != 403 || len(db.snapshot()) != before {
		t.Fatal("missing CSRF changed policy")
	}
	db.failAt = "insert"
	if response := mockRequest(h, "POST", "/api/events", body); response.Code < 500 || len(db.snapshot()) != before {
		t.Fatal("failed audit changed policy", response.Code)
	}
	db.failAt = ""
	st, err := s.load(t.Context())
	if err != nil || st.AgentIdleSettings.Revision != 0 {
		t.Fatal("failed setting persisted", err)
	}
}

func TestCodexAssemblyConsumesRegisteredIdlePolicy(t *testing.T) {
	t.Parallel()
	for _, foreign := range []bool{false, true} {
		t.Run(map[bool]string{true: "foreign", false: "bound"}[foreign], func(t *testing.T) {
			_, s, _ := workspacePair(t)
			record := s.record
			scope := toolbroker.Scope{Version: 1, JobID: record.JobID, Room: record.RoomID, Participant: record.ParticipantID, Connection: record.ConnectionID, Provider: record.Provider, RepositoryDigest: record.Pair.Repository, Deadline: record.Deadline}
			channel, peer := net.Pipe()
			defer peer.Close()
			runner := newWorkspaceCleanup(t, s)
			runner.volume = false
			ctx := context.WithValue(s.ctx, workerLocalJobContextKey{}, scope.JobID)
			session, token, err := newWorkerToolSession(ctx, scope, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", time.Now(), channel, runner)
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			job := room.ManagedJob{Version: record.Version, Execution: record.Execution, RepositoryDigest: record.Pair.Repository, InitiatorID: ownerID, CompanyID: "company", JobID: record.JobID, RoomID: record.RoomID, ParticipantID: record.ParticipantID, ConnectionID: record.ConnectionID, Worker: "worker", Provider: record.Provider, Status: "allocated", AllocatedAt: record.StartedAt, UpdatedAt: record.UpdatedAt, Deadline: record.Deadline, IdlePolicyVersion: 1, IdleMinutes: 2, IdleShutdownEnabled: false, IdlePolicyRevision: 3}
			if foreign {
				job.RepositoryDigest = strings.Repeat("f", 64)
			}
			bundle := &pairBundleFixture{Reader: strings.NewReader("synthetic-bundle")}
			_, _, err = s.RunCodex(t.Context(), session, token, bundle, workerCodexRunConfig{ManagedJob: &job, AuthRoot: t.TempDir()}, runner)
			if err == nil {
				t.Fatal("missing credential preparation should fail")
			}
			if foreign {
				if s.closed || s.idle != nil || bundle.closes.Load() != 0 {
					t.Fatal("foreign controller policy consumed caller resources")
				}
			} else if s.idle == nil || s.idle.policy != (workerIdlePolicy{Enabled: false, Minutes: 2}) || !s.closed || bundle.closes.Load() != 1 {
				t.Fatal("assembly ignored registered off policy or cleanup")
			}
		})
	}
}

func TestIdleFeatureFlagCommandsAndOffConfiguration(t *testing.T) {
	t.Parallel()
	s, db, h := idleSettingsHTTP(t)
	for _, enabled := range []bool{false, true} {
		flag, _ := json.Marshal(map[string]any{"kind": "feature.flag.updated", "payload": map[string]any{"name": "idleAgentShutdown", "enabled": enabled}})
		if response := mockRequest(h, "POST", "/api/events", string(flag)); response.Code != 200 {
			t.Fatal("flag command rejected", response.Code)
		}
		st, err := s.load(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		policy, err := idlePolicyForState(st)
		if err != nil || policy.Enabled != enabled {
			t.Fatal("flag not enforced by allocation policy", err)
		}
		if !enabled {
			if response := mockRequest(h, "POST", "/api/events", `{"kind":"agents.idle_policy.updated","payload":{"minutes":3,"revision":0}}`); response.Code != 200 {
				t.Fatal("off setting not editable", response.Code)
			}
		}
	}
	db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true})
	s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	before := len(db.snapshot())
	response := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"idleAgentShutdown","enabled":false}}`)
	if response.Code != 403 || len(db.snapshot()) != before {
		t.Fatal("non-owner changed idle flag")
	}
}
func TestIdleSettingsConcurrentControllers(t *testing.T) {
	t.Parallel()
	s, db, h := idleSettingsHTTP(t)
	other := &server{db: db, auth: s.auth}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", other.events)
	handlers := []http.Handler{h, other.auth.protect(other.actorGuard(mux))}
	results := make(chan int, 2)
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(handler http.Handler) {
			defer wg.Done()
			results <- mockRequest(handler, "POST", "/api/events", `{"kind":"agents.idle_policy.updated","payload":{"minutes":2,"revision":0}}`).Code
		}(handler)
	}
	wg.Wait()
	close(results)
	success, conflicts := 0, 0
	for code := range results {
		if code == 200 {
			success++
		} else if code == 409 {
			conflicts++
		} else {
			t.Fatal("unexpected settings outcome", code)
		}
	}
	st, err := s.load(t.Context())
	if err != nil || success != 1 || conflicts != 1 || st.AgentIdleSettings.Revision != 1 {
		t.Fatal("concurrent settings corrupted journal", success, conflicts, err)
	}
}
