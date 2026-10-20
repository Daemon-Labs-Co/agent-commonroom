package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type idleNoticeFixture struct {
	s        *server
	db       *mockDB
	job      room.ManagedJob
	worker   *workerPairSupervisor
	docker   *workspaceCleanupDocker
	announce workerIdleAnnounce
	now      time.Time
}

func newIdleNoticeFixture(t *testing.T, shortDirectory ...bool) *idleNoticeFixture {
	t.Helper()
	return newIdleNoticeIntervalFixture(t, 1, shortDirectory...)
}

func newIdleNoticeIntervalFixture(t *testing.T, minutes int, shortDirectory ...bool) *idleNoticeFixture {
	t.Helper()
	s, db, _ := mockServer(t)
	now := time.Now().UTC()
	job := quotaJob(1, true)
	job.InitiatorID = ownerID
	job.RoomID = "engineering"
	job.Worker = "worker"
	job.AllocatedAt = now.Add(-time.Duration(minutes)*time.Minute - 30*time.Second)
	job.UpdatedAt = job.AllocatedAt
	job.Deadline = job.AllocatedAt.Add(3 * time.Minute)
	job.IdlePolicyVersion = 1
	job.IdlePolicyRevision = 1
	job.IdleShutdownEnabled = true
	job.IdleMinutes = minutes
	db.seed("worker.job.updated", job)
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	announce, err := s.newIdleAnnouncer(ctx, job.JobID)
	if err != nil {
		t.Fatal(err)
	}
	request := workerRunRequest{Version: job.Version, IssuedAt: job.AllocatedAt, Deadline: job.Deadline, JobID: job.JobID, Provider: job.Provider, Repository: job.RepositoryDigest, Key: job.RoomID + "/" + job.ParticipantID + "/" + job.ConnectionID + "/worker"}
	dir := t.TempDir()
	if len(shortDirectory) > 0 && shortDirectory[0] {
		var tempErr error
		dir, tempErr = os.MkdirTemp("/tmp", "cr-idle-")
		if tempErr != nil {
			t.Fatal(tempErr)
		}
		t.Cleanup(func() { os.RemoveAll(dir) })
	}
	worker, err := admitWorkerPair(ctx, dir, request, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(worker.Close)
	worker.record.StartedAt = job.AllocatedAt.Add(time.Second)
	if err := writeWorkerRecord(worker.dir, worker.record, false); err != nil {
		t.Fatal(err)
	}
	docker := newWorkspaceCleanup(t, worker)
	f := &idleNoticeFixture{s: s, db: db, job: job, worker: worker, docker: docker, announce: announce, now: now}
	checked := func(ctx context.Context, r workerJobRecord) error {
		if docker.removed != 0 || len(docker.pairFakeDocker.removed) != 0 {
			t.Fatal("cleanup preceded announcement")
		}
		if worker.ctx.Err() != nil {
			t.Fatal("announcement inherited canceled lifetime")
		}
		return f.announce(ctx, r)
	}
	if err := worker.configureJobIdle(job, ownedShutdownRunner{t, docker}, checked); err != nil {
		t.Fatal(err)
	}
	// Drive virtual expiry explicitly; do not race a wall-clock timer.
	worker.mu.Lock()
	worker.idle.armed = true
	worker.mu.Unlock()
	adapter := &workspaceLaunchFake{pairLaunchFake: &pairLaunchFake{t: t, supervisor: worker}}
	if err := worker.Start(ctx, adapter); err != nil {
		t.Fatal(err)
	}
	docker.record = worker.record
	for role, id := range map[string]string{"engine": worker.record.Pair.Engine.ID, "tool": worker.record.Pair.Tool.ID} {
		docker.containers[id] = pairFakeContainer{name: pairContainerName(job.JobID, role), execution: "separated-" + role + "-v1", repository: job.RepositoryDigest, job: job.JobID}
	}
	worker.mu.Lock()
	worker.idle.since = now.Add(-time.Duration(minutes) * time.Minute)
	worker.mu.Unlock()
	return f
}
func TestIdleAnnouncementPrecedesCleanupAndRetainsEvidence(t *testing.T) {
	t.Parallel()
	f := newIdleNoticeFixture(t)
	expired, err := f.worker.expireIdle(f.now)
	if !expired || err == nil {
		t.Fatal("retained workspace must remain unresolved")
	}
	st, err := (&server{db: f.db}).load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	notice, ok := st.IdleStopNotices[f.job.JobID]
	if !ok || len(st.Messages) != 1 || st.Messages[0].ActorType != "system" || len(f.docker.pairFakeDocker.removed) != 2 || !f.docker.volume || st.ManagedJobs[f.job.JobID].ShutdownConfirmed {
		t.Fatal("notice became cleanup proof or source was lost")
	}
	if notice.Job.InitiatorID != ownerID || notice.Job.IdleMinutes != 1 {
		t.Fatal("attribution/policy lost")
	}
	count := len(f.db.snapshot())
	if err := f.announce(t.Context(), f.worker.record); err == nil {
		t.Fatal("post-cleanup state must not masquerade as pre-cleanup intention")
	}
	// A lost acknowledgement can be reconciled against exactly the original intent.
	record := f.docker.record
	record.DiagnosticCode = "idle_shutdown_requested"
	record.IdleShutdownRequestedAt = f.now
	record.IdleTimeoutMinutes = 1
	if err := f.announce(t.Context(), record); err != nil || len(f.db.snapshot()) != count {
		t.Fatal("exact announcement retry duplicated history", err)
	}
	record.IdleShutdownRequestedAt = record.IdleShutdownRequestedAt.Add(time.Second)
	if f.announce(t.Context(), record) == nil {
		t.Fatal("changed intention reused acknowledgment")
	}
}
func TestIdleAnnouncementFailuresDoNotAuthorizeCleanup(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"audit", "panic", "missing", "reply-lost"} {
		t.Run(mode, func(t *testing.T) {
			f := newIdleNoticeFixture(t)
			switch mode {
			case "audit":
				f.db.failAt = "insert"
			case "panic":
				f.announce = func(context.Context, workerJobRecord) error { panic("synthetic private failure") }
			case "missing":
				f.worker.mu.Lock()
				f.worker.idle.announce = nil
				f.worker.mu.Unlock()
			case "reply-lost":
				original := f.announce
				f.announce = func(ctx context.Context, r workerJobRecord) error {
					if err := original(ctx, r); err != nil {
						return err
					}
					return errors.New("synthetic lost acknowledgment")
				}
			}
			expired, err := f.worker.expireIdle(f.now)
			f.db.failAt = ""
			if !expired || err == nil || !f.worker.closed || len(f.docker.pairFakeDocker.removed) != 0 || !f.docker.volume {
				t.Fatal("unacknowledged announcement reached runtime", mode, err)
			}
			st, err := f.s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			if mode == "reply-lost" {
				want = 1
			}
			if len(st.IdleStopNotices) != want || st.ManagedJobs[f.job.JobID].ShutdownConfirmed {
				t.Fatal("failed announcement invented evidence")
			}
			if expired, _ := f.worker.expireIdle(f.now.Add(time.Minute)); expired {
				t.Fatal("unknown announcement retried")
			}
		})
	}
}
func TestIdleNoticeVisibilityCannotDisableAudit(t *testing.T) {
	t.Parallel()
	f := newIdleNoticeFixture(t)
	f.db.seed("feature.flag.updated", map[string]any{"name": "idleStopNotices", "enabled": false})
	f.db.seed("feature.flag.updated", map[string]any{"name": "idleAgentShutdown", "enabled": false})
	f.worker.expireIdle(f.now)
	st, err := f.s.load(t.Context())
	if err != nil || len(st.Messages) != 0 || len(st.IdleStopNotices) != 1 || len(f.docker.pairFakeDocker.removed) != 2 {
		t.Fatal("visibility altered audit or admitted cleanup", err, len(st.Messages), len(st.IdleStopNotices), len(f.docker.pairFakeDocker.removed))
	}
	f.db.seed("feature.flag.updated", map[string]any{"name": "idleStopNotices", "enabled": true})
	st, err = f.s.load(t.Context())
	if err != nil || len(st.Messages) != 1 || len(st.IdleStopNotices) != 1 {
		t.Fatal("history did not reappear", err)
	}
}
func TestPublicEventsCannotForgeIdleNotice(t *testing.T) {
	t.Parallel()
	_, db, h := idleSettingsHTTP(t)
	before := len(db.snapshot())
	response := mockRequest(h, "POST", "/api/events", `{"kind":"worker.idle_stop.announced","payload":{}}`)
	if response.Code != 400 || len(db.snapshot()) != before {
		t.Fatal("public forged worker intention")
	}
}
func TestIdleNoticeReplayRejectsChangedOrEarlyEvidence(t *testing.T) {
	t.Parallel()
	f := newIdleNoticeFixture(t)
	base := room.IdleStopNotice{Version: 1, ID: id(), Job: f.job, WorkerStartedAt: f.job.AllocatedAt.Add(time.Second), RequestedAt: f.now, RecordedAt: f.now}
	for _, mode := range []string{"early", "future", "deadline", "foreign", "disabled", "duplicate"} {
		t.Run(mode, func(t *testing.T) {
			st, err := f.s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			n := base
			switch mode {
			case "early":
				n.RequestedAt = n.WorkerStartedAt.Add(time.Second)
			case "future":
				n.RequestedAt = f.now.Add(6 * time.Second)
			case "deadline":
				n.RequestedAt = n.Job.Deadline
			case "foreign":
				n.Job.RepositoryDigest = "foreign"
			case "disabled":
				n.Job.IdleShutdownEnabled = false
			case "duplicate":
				raw, _ := json.Marshal(n)
				if err := projectIdleNotice(&st, raw); err != nil {
					t.Fatal(err)
				}
			}
			raw, _ := json.Marshal(n)
			if projectIdleNotice(&st, raw) == nil {
				t.Fatal("forged/duplicate notice replay", mode)
			}
		})
	}
}

func TestIdleNoticeConstructorPermissionsAndFlagCommands(t *testing.T) {
	t.Parallel()
	f := newIdleNoticeFixture(t)
	if _, err := f.s.newIdleAnnouncer(t.Context(), f.job.JobID); err == nil {
		t.Fatal("missing actor obtained announcer")
	}
	f.db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true, RoomIDs: []string{"engineering"}})
	ctx := context.WithValue(t.Context(), actorContextKey{}, "developer")
	if _, err := f.s.newIdleAnnouncer(ctx, f.job.JobID); err == nil {
		t.Fatal("foreign initiator obtained announcer")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", f.s.events)
	handler := f.s.auth.protect(f.s.actorGuard(mux))
	if response := mockRequest(handler, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"idleStopNotices","enabled":false}}`); response.Code != 200 {
		t.Fatal("owner notice flag refused", response.Code)
	}
	f.s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	before := len(f.db.snapshot())
	if response := mockRequest(handler, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"idleStopNotices","enabled":true}}`); response.Code != 403 || len(f.db.snapshot()) != before {
		t.Fatal("non-owner notice flag accepted", response.Code)
	}
}
func TestIdleNoticeConcurrentControllersDeduplicate(t *testing.T) {
	t.Parallel()
	f := newIdleNoticeFixture(t)
	other := &server{db: f.db}
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	second, err := other.newIdleAnnouncer(ctx, f.job.JobID)
	if err != nil {
		t.Fatal(err)
	}
	record := f.docker.record
	record.DiagnosticCode = "idle_shutdown_requested"
	record.IdleShutdownRequestedAt = f.now
	record.IdleTimeoutMinutes = 1
	before := len(f.db.snapshot())
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, callback := range []workerIdleAnnounce{f.announce, second} {
		wg.Add(1)
		go func(callback workerIdleAnnounce) { defer wg.Done(); results <- callback(t.Context(), record) }(callback)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	if success < 1 || len(f.db.snapshot()) != before+1 {
		t.Fatal("duplicate idle announcements", success)
	}
	if err := second(t.Context(), record); err != nil || len(f.db.snapshot()) != before+1 {
		t.Fatal("reconciled announcement duplicated history", err)
	}
}

func TestCodexRefusesUnregisteredIdleCallback(t *testing.T) {
	t.Parallel()
	_, worker, _ := workspacePair(t)
	record := worker.record
	scope := toolbroker.Scope{Version: 1, JobID: record.JobID, Room: record.RoomID, Participant: record.ParticipantID, Connection: record.ConnectionID, Provider: record.Provider, RepositoryDigest: record.Pair.Repository, Deadline: record.Deadline}
	channel, peer := net.Pipe()
	defer peer.Close()
	runner := newWorkspaceCleanup(t, worker)
	ctx := context.WithValue(worker.ctx, workerLocalJobContextKey{}, scope.JobID)
	session, token, err := newWorkerToolSession(ctx, scope, []string{"read"}, "ghcr.io/example/tool@sha256:"+strings.Repeat("a", 64), "/trusted/docker", time.Now(), channel, runner)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	bundle := &pairBundleFixture{Reader: strings.NewReader("synthetic-bundle")}
	calls := 0
	_, _, err = worker.RunCodex(t.Context(), session, token, bundle, workerCodexRunConfig{IdleAnnounce: func(context.Context, workerJobRecord) error { calls++; return nil }}, runner)
	if err == nil || calls != 0 || worker.closed || worker.idle != nil || bundle.closes.Load() != 0 {
		t.Fatal("unregistered callback consumed worker ownership")
	}
}
