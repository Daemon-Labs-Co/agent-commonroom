package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type disposalFixture struct {
	s      *server
	db     *mockDB
	h      http.Handler
	job    room.ManagedJob
	view   workerJobView
	now    atomic.Int64
	calls  atomic.Int32
	probes atomic.Int32
}

func newDisposalFixture(t *testing.T, enabled bool) *disposalFixture {
	t.Helper()
	s, db, _ := mockServer(t)
	f := &disposalFixture{s: s, db: db}
	f.job = quotaJob(1, true)
	f.job.Worker = "worker"
	f.job.Status = "shutdown_unconfirmed"
	db.seed("worker.job.updated", f.job)
	db.seed("feature.flag.updated", map[string]any{"name": "workspaceDisposal", "enabled": enabled})
	r := storageRecord(10, 42, "workspace")
	r.Worker = "worker"
	db.seed("storage.inventory.recorded", managedStorageInventory{1, "company", strings.Repeat("d", 64), []room.StorageReservation{r}})
	record, err := managedJobFence(f.job).record()
	if err != nil {
		t.Fatal(err)
	}
	record.Deadline = f.job.Deadline
	record.Status = "shutdown_unconfirmed"
	record.Pair.Engine = workerPairContainer{CreationStarted: true, ID: strings.Repeat("a", 64), Absent: true}
	record.Pair.Tool = workerPairContainer{CreationStarted: true, ID: strings.Repeat("c", 64), Absent: true}
	record.Pair.Workspace = &workerPairWorkspace{CreationStarted: true, Identity: toolbroker.WorkspaceVolumeIdentity{Version: 1, Name: "commonroom-tool-work-" + f.job.JobID, Nonce: strings.Repeat("e", 32), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}}
	if !validWorkerRecord(record) {
		t.Fatal("invalid synthetic retained workspace")
	}
	f.view = workerJobView{workerJobRecord: record, LeaseState: "free"}
	f.now.Store(time.Now().UTC().UnixNano())
	s.disposalNow = func() time.Time { return time.Unix(0, f.now.Load()).UTC() }
	s.jobSnapshot = func(context.Context, string) (workerJobsSnapshot, error) {
		f.probes.Add(1)
		return workerJobsSnapshot{Version: 2, Jobs: []workerJobView{f.view}}, nil
	}
	s.jobFence = func(ctx context.Context, alias string, request workerFenceRequest) error {
		f.calls.Add(1)
		if alias != "worker" || !request.DiscardWorkspace || request.ExpectedWorkspace == nil || *request.ExpectedWorkspace != f.view.Pair.Workspace.Identity {
			t.Error("unbound destructive call")
		}
		st, err := s.load(ctx)
		if err != nil {
			t.Fatal(err)
		}
		claimed := false
		for _, p := range st.WorkspaceDisposals {
			if p.Job.JobID == request.JobID && p.Status == "claimed" {
				claimed = true
			}
		}
		if !claimed {
			t.Error("worker action preceded durable approval claim")
		}
		return nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/workspace-disposal", s.workspaceDisposal)
	mux.HandleFunc("/api/events", s.events)
	f.h = s.auth.protect(s.actorGuard(mux))
	return f
}
func (f *disposalFixture) post(t *testing.T, payload any) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return mockRequest(f.h, "POST", "/api/workspace-disposal", string(raw))
}
func (f *disposalFixture) propose(t *testing.T) room.WorkspaceDisposal {
	t.Helper()
	res := f.post(t, map[string]string{"action": "propose", "jobId": f.job.JobID})
	if res.Code != 200 {
		t.Fatalf("proposal %d %s", res.Code, res.Body.String())
	}
	var p room.WorkspaceDisposal
	if json.Unmarshal(res.Body.Bytes(), &p) != nil || p.Status != "pending" {
		t.Fatal("invalid proposal response")
	}
	return p
}
func (f *disposalFixture) approve(t *testing.T, p room.WorkspaceDisposal) *httptest.ResponseRecorder {
	return f.post(t, map[string]string{"action": "approve", "proposalId": p.ID, "confirm": p.Job.JobID})
}
func TestWorkspaceDisposalHTTP(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	t.Run("review-minute-exact-confirmation-and-one-shot", func(t *testing.T) {
		f := newDisposalFixture(t, true)
		p := f.propose(t)
		if f.calls.Load() != 0 {
			t.Fatal("proposal deleted data")
		}
		f.now.Store(p.NotBefore.Add(-time.Nanosecond).UnixNano())
		if r := f.approve(t, p); r.Code != 409 {
			t.Fatal("early approval accepted", r.Code)
		}
		f.now.Store(p.NotBefore.UnixNano())
		if r := f.post(t, map[string]string{"action": "approve", "proposalId": p.ID, "confirm": "wrong"}); r.Code != 409 {
			t.Fatal("wrong confirmation accepted")
		}
		if r := f.approve(t, p); r.Code != 200 {
			t.Fatal(r.Code, r.Body.String())
		}
		if r := f.approve(t, p); r.Code != 409 || f.calls.Load() != 1 {
			t.Fatal("disposal repeated")
		}
		st, err := f.s.load(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		if st.WorkspaceDisposals[p.ID].Status != "succeeded" || !st.ManagedJobs[p.Job.JobID].ShutdownConfirmed {
			t.Fatal("outcome/slots not durable")
		}
		total, err := managedStorageTotal(st)
		if err != nil || total != 42 {
			t.Fatal("disposal silently refunded unbound storage")
		}
		kinds := []string{}
		for _, e := range f.db.snapshot() {
			if strings.HasPrefix(e.kind, "workspace.disposal.") {
				kinds = append(kinds, e.kind)
			}
		}
		if strings.Join(kinds, ",") != "workspace.disposal.proposed,workspace.disposal.claimed,workspace.disposal.succeeded" {
			t.Fatal("wrong audit sequence", kinds)
		}
	})
	t.Run("off-and-current-permissions", func(t *testing.T) {
		for _, mode := range []string{"off", "non-owner", "disabled-owner", "no-csrf"} {
			t.Run(mode, func(t *testing.T) {
				f := newDisposalFixture(t, mode != "off")
				if mode == "non-owner" {
					f.db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true})
					f.s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
				}
				if mode == "disabled-owner" {
					f.db.seed("actor.revised", room.Actor{ID: ownerID, Kind: "human", Enabled: false})
				}
				var r *httptest.ResponseRecorder
				if mode == "no-csrf" {
					req := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/workspace-disposal", strings.NewReader(`{"action":"propose"}`))
					req.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
					r = httptest.NewRecorder()
					f.h.ServeHTTP(r, req)
				} else {
					r = f.post(t, map[string]string{"action": "propose", "jobId": f.job.JobID})
				}
				if r.Code != 403 || f.probes.Load() != 0 || f.calls.Load() != 0 {
					t.Fatal("permission/flag bypass", mode, r.Code)
				}
			})
		}
	})
	t.Run("worker-state-refusals", func(t *testing.T) {
		for _, mode := range []string{"held", "running", "unknown-time", "truncated", "missing", "foreign"} {
			t.Run(mode, func(t *testing.T) {
				f := newDisposalFixture(t, true)
				switch mode {
				case "held":
					f.view.LeaseState = "held"
				case "running":
					f.view.Pair.Tool.Absent = false
				case "unknown-time":
					f.view.Pair.Workspace.Identity.CreatedAt = ""
				case "foreign":
					f.view.Pair.Repository = strings.Repeat("f", 64)
				case "truncated":
					f.s.jobSnapshot = func(context.Context, string) (workerJobsSnapshot, error) {
						return workerJobsSnapshot{Version: 2, Truncated: true}, nil
					}
				case "missing":
					f.s.jobSnapshot = func(context.Context, string) (workerJobsSnapshot, error) { return workerJobsSnapshot{Version: 2}, nil }
				}
				if r := f.post(t, map[string]string{"action": "propose", "jobId": f.job.JobID}); r.Code < 400 || f.calls.Load() != 0 {
					t.Fatal("unsafe worker admitted", mode, r.Code)
				}
			})
		}
	})
	t.Run("expiry-revocation-and-toggle-preserve-history", func(t *testing.T) {
		f := newDisposalFixture(t, true)
		p := f.propose(t)
		f.now.Store(p.ExpiresAt.UnixNano())
		if r := f.approve(t, p); r.Code != 409 {
			t.Fatal("expired approval accepted")
		}
		f.db.seed("feature.flag.updated", map[string]any{"name": "workspaceDisposal", "enabled": false})
		if r := f.post(t, map[string]string{"action": "revoke", "proposalId": p.ID}); r.Code != 200 {
			t.Fatal("disabled flag blocked safe revocation", r.Code)
		}
		if r := f.approve(t, p); r.Code != 403 {
			t.Fatal("flag off permitted approval")
		}
		st, err := f.s.load(t.Context())
		if err != nil || st.WorkspaceDisposals[p.ID].Status != "revoked" || f.calls.Load() != 0 {
			t.Fatal("revocation/history lost")
		}
	})
	t.Run("stale-identity-and-access-change-after-probe", func(t *testing.T) {
		for _, mode := range []string{"identity", "flag", "owner"} {
			t.Run(mode, func(t *testing.T) {
				f := newDisposalFixture(t, true)
				p := f.propose(t)
				f.now.Store(p.NotBefore.UnixNano())
				if mode == "identity" {
					f.view.Pair.Workspace.Identity.Nonce = strings.Repeat("f", 32)
				} else {
					f.s.jobSnapshot = func(context.Context, string) (workerJobsSnapshot, error) {
						if mode == "flag" {
							f.db.seed("feature.flag.updated", map[string]any{"name": "workspaceDisposal", "enabled": false})
						} else {
							f.db.seed("actor.revised", room.Actor{ID: ownerID, Kind: "human", Enabled: false})
						}
						return workerJobsSnapshot{Version: 2, Jobs: []workerJobView{f.view}}, nil
					}
				}
				if r := f.approve(t, p); r.Code != 409 || f.calls.Load() != 0 {
					t.Fatal("stale approval executed", mode, r.Code)
				}
			})
		}
	})
	t.Run("audit-and-worker-failures-never-retry", func(t *testing.T) {
		for _, mode := range []string{"proposal-audit", "claim-audit", "outcome-audit", "worker"} {
			t.Run(mode, func(t *testing.T) {
				f := newDisposalFixture(t, true)
				if mode == "proposal-audit" {
					f.db.failAt = "insert"
					if r := f.post(t, map[string]string{"action": "propose", "jobId": f.job.JobID}); r.Code != 503 || f.calls.Load() != 0 {
						t.Fatal("proposal audit bypass")
					}
					return
				}
				p := f.propose(t)
				f.now.Store(p.NotBefore.UnixNano())
				if mode == "claim-audit" {
					f.db.failAt = "insert"
				} else {
					f.s.jobFence = func(context.Context, string, workerFenceRequest) error {
						f.calls.Add(1)
						if mode == "outcome-audit" {
							f.db.failAt = "insert"
							return nil
						}
						return errors.New("synthetic private failure")
					}
				}
				r := f.approve(t, p)
				if r.Code != 503 || strings.Contains(r.Body.String(), "synthetic private") {
					t.Fatal("unsafe failure response", r.Code)
				}
				f.db.failAt = ""
				st, err := f.s.load(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				expected := "claimed"
				if mode == "worker" {
					expected = "unconfirmed"
				}
				if mode == "claim-audit" {
					expected = "pending"
				}
				if st.WorkspaceDisposals[p.ID].Status != expected {
					t.Fatal("lost uncertainty", st.WorkspaceDisposals[p.ID].Status)
				}
				if mode == "claim-audit" {
					if f.calls.Load() != 0 {
						t.Fatal("failed claim reached worker")
					}
					return
				}
				if r = f.approve(t, p); r.Code != 409 || f.calls.Load() != 1 {
					t.Fatal("uncertain action repeated", r.Code)
				}
			})
		}
	})
	t.Run("concurrent-clicks-and-other-controller", func(t *testing.T) {
		f := newDisposalFixture(t, true)
		p := f.propose(t)
		f.now.Store(p.NotBefore.UnixNano())
		other := &server{db: f.db, auth: f.s.auth, disposalNow: f.s.disposalNow, jobSnapshot: f.s.jobSnapshot, jobFence: f.s.jobFence}
		mux := http.NewServeMux()
		mux.HandleFunc("/api/workspace-disposal", other.workspaceDisposal)
		h := other.auth.protect(other.actorGuard(mux))
		// Both controllers observe committed publication before the review minute.
		f.now.Store(p.CreatedAt.UnixNano())
		if response := mockRequest(h, "GET", "/api/workspace-disposal", ""); response.Code != 200 {
			t.Fatal(response.Code)
		}
		f.now.Store(p.NotBefore.UnixNano())
		var wg sync.WaitGroup
		responses := make(chan int, 2)
		body := `{"action":"approve","proposalId":"` + p.ID + `","confirm":"` + p.Job.JobID + `"}`
		for _, handler := range []http.Handler{f.h, h} {
			wg.Add(1)
			go func(handler http.Handler) {
				defer wg.Done()
				responses <- mockRequest(handler, "POST", "/api/workspace-disposal", body).Code
			}(handler)
		}
		wg.Wait()
		close(responses)
		success := 0
		for code := range responses {
			if code == 200 {
				success++
			}
		}
		if success != 1 || f.calls.Load() != 1 {
			t.Fatal("duplicate destructive dispatch", success, f.calls.Load())
		}
	})
	t.Run("review-minute-starts-after-probe", func(t *testing.T) {
		f := newDisposalFixture(t, true)
		f.s.jobSnapshot = func(context.Context, string) (workerJobsSnapshot, error) {
			f.now.Add(int64(20 * time.Second))
			return workerJobsSnapshot{Version: 2, Jobs: []workerJobView{f.view}}, nil
		}
		p := f.propose(t)
		if p.CreatedAt.UnixNano() != f.now.Load() || p.NotBefore.Sub(time.Unix(0, f.now.Load())) != time.Minute {
			t.Fatal("probe consumed review minute")
		}
	})
}

func TestDisposalProjectionRefusesForgedTransitions(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	for _, mode := range []string{"early", "changed-volume", "changed-job", "skip-claim", "rewrite", "flag-off-claim"} {
		t.Run(mode, func(t *testing.T) {
			f := newDisposalFixture(t, true)
			p := f.propose(t)
			st, err := f.s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "early":
				p.Status = "claimed"
				p.ClaimedAt = p.NotBefore.Add(-time.Nanosecond)
			case "changed-volume":
				p.Status = "claimed"
				p.ClaimedAt = p.NotBefore
				p.WorkspaceNonce = strings.Repeat("f", 32)
			case "changed-job":
				p.Status = "claimed"
				p.ClaimedAt = p.NotBefore
				p.Job.Worker = "foreign"
			case "skip-claim":
				p.Status = "succeeded"
				p.ClaimedAt = p.NotBefore
				p.FinishedAt = p.NotBefore
			case "rewrite":
				p.WorkspaceName = "foreign"
			case "flag-off-claim":
				st.FeatureFlags["workspaceDisposal"] = false
				p.Status = "claimed"
				p.ClaimedAt = p.NotBefore
			}
			raw, _ := json.Marshal(p)
			if projectWorkspaceDisposal(&st, disposalEvent(p), raw) == nil {
				t.Fatal("forged approval transition accepted", mode)
			}
		})
	}
}
func TestPublicEventsCannotForgeDisposalApproval(t *testing.T) {
	for _, kind := range []string{"workspace.disposal.proposed", "workspace.disposal.claimed", "workspace.disposal.succeeded"} {
		_, db, h := mockServer(t)
		before := len(db.snapshot())
		response := mockRequest(h, "POST", "/api/events", `{"kind":"`+kind+`","payload":{}}`)
		if response.Code < 400 || len(db.snapshot()) != before {
			t.Fatal("public event forged approval", kind)
		}
	}
}

func TestClaimedDisposalSurvivesDisconnectAndFlagDisable(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true)
	p := f.propose(t)
	f.now.Store(p.NotBefore.UnixNano())
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	f.s.jobFence = func(action context.Context, _ string, _ workerFenceRequest) error {
		f.calls.Add(1)
		cancel()
		f.db.seed("feature.flag.updated", map[string]any{"name": "workspaceDisposal", "enabled": false})
		if action.Err() != nil {
			t.Fatal("admitted cleanup inherited disconnected request")
		}
		return nil
	}
	body := `{"action":"approve","proposalId":"` + p.ID + `","confirm":"` + p.Job.JobID + `"}`
	req := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/workspace-disposal", strings.NewReader(body)).WithContext(ctx)
	req.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	req.Header.Set("X-CSRF-Token", "csrf")
	response := httptest.NewRecorder()
	f.h.ServeHTTP(response, req)
	if response.Code != 200 || f.calls.Load() != 1 {
		t.Fatal("admitted outcome lost", response.Code)
	}
	st, err := f.s.load(t.Context())
	if err != nil || st.WorkspaceDisposals[p.ID].Status != "succeeded" || st.FeatureFlags["workspaceDisposal"] {
		t.Fatal("in-flight audit/flag state lost", err)
	}
}
func TestDisposalAdapterPanicBecomesDurableUncertainty(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true)
	p := f.propose(t)
	f.now.Store(p.NotBefore.UnixNano())
	f.s.jobFence = func(context.Context, string, workerFenceRequest) error {
		f.calls.Add(1)
		panic("synthetic private panic")
	}
	response := f.approve(t, p)
	if response.Code != 503 || strings.Contains(response.Body.String(), "private panic") {
		t.Fatal("unsafe panic response")
	}
	st, err := f.s.load(t.Context())
	if err != nil || st.WorkspaceDisposals[p.ID].Status != "unconfirmed" {
		t.Fatal("panic made approval replayable", err)
	}
	if response = f.approve(t, p); response.Code != 409 || f.calls.Load() != 1 {
		t.Fatal("panicked operation repeated")
	}
}

func TestDisposalOutcomeAndSlotReleaseAreAtomic(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true)
	p := f.propose(t)
	f.now.Store(p.NotBefore.UnixNano())
	f.db.failInsertAt = 2
	response := f.approve(t, p)
	if response.Code != 503 || f.calls.Load() != 1 {
		t.Fatal("expected outcome transaction failure")
	}
	f.db.failInsertAt = 0
	st, err := f.s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if st.WorkspaceDisposals[p.ID].Status != "claimed" || st.ManagedJobs[p.Job.JobID].ShutdownConfirmed {
		t.Fatal("partial outcome/slot credit committed")
	}
	if response = f.approve(t, p); response.Code != 409 || f.calls.Load() != 1 {
		t.Fatal("failed outcome transaction repeated disposal")
	}
}
func TestDisposalFeatureSettingUsesOwnerEventPermissions(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, false)
	r := mockRequest(f.h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"workspaceDisposal","enabled":true}}`)
	if r.Code != 200 {
		t.Fatal("owner cannot enable disposal", r.Code, r.Body.String())
	}
	p := f.propose(t)
	r = mockRequest(f.h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"workspaceDisposal","enabled":false}}`)
	if r.Code != 200 {
		t.Fatal("owner cannot disable disposal", r.Code)
	}
	f.now.Store(p.NotBefore.UnixNano())
	if r = f.approve(t, p); r.Code != 403 || f.calls.Load() != 0 {
		t.Fatal("disabled setting bypassed")
	}
	f.db.seed("actor.revised", room.Actor{ID: "developer", Kind: "human", Enabled: true})
	f.s.auth.sessions["owner"] = ownerSession{ActorID: "developer", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	r = mockRequest(f.h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"workspaceDisposal","enabled":true}}`)
	if r.Code != 403 {
		t.Fatal("non-owner enabled destructive feature")
	}
}

func TestDisposalReviewRestartsConservativelyAfterControllerRestart(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true)
	p := f.propose(t)
	f.now.Store(p.NotBefore.UnixNano())
	restarted := &server{db: f.db, auth: f.s.auth, disposalNow: f.s.disposalNow, jobSnapshot: f.s.jobSnapshot, jobFence: f.s.jobFence}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/workspace-disposal", restarted.workspaceDisposal)
	handler := restarted.auth.protect(restarted.actorGuard(mux))
	body := `{"action":"approve","proposalId":"` + p.ID + `","confirm":"` + p.Job.JobID + `"}`
	if response := mockRequest(handler, "POST", "/api/workspace-disposal", body); response.Code != 409 || f.calls.Load() != 0 {
		t.Fatal("restart shortened publication review", response.Code)
	}
	response := mockRequest(handler, "GET", "/api/workspace-disposal", "")
	var listed struct {
		Proposals []disposalReviewView `json:"proposals"`
	}
	if json.Unmarshal(response.Body.Bytes(), &listed) != nil || len(listed.Proposals) != 1 {
		t.Fatal("missing restart review view")
	}
	ready := listed.Proposals[0].ReviewReadyAt
	if !ready.Equal(p.NotBefore.Add(time.Minute)) {
		t.Fatal("wrong conservative restart delay", ready)
	}
	f.now.Store(ready.UnixNano())
	if response = mockRequest(handler, "POST", "/api/workspace-disposal", body); response.Code != 200 || f.calls.Load() != 1 {
		t.Fatal("restarted controller failed reviewed disposal", response.Code, response.Body.String())
	}
}
func TestDisposalReviewStartsAfterAuditPublication(t *testing.T) {
	t.Setenv("COMMONROOM_REMOTE_WORKER", "true")
	t.Setenv("COMMONROOM_REMOTE_WORKER_ALIAS", "worker")
	f := newDisposalFixture(t, true)
	initial := time.Unix(0, f.now.Load()).UTC()
	var calls atomic.Int32
	f.s.disposalNow = func() time.Time {
		// Initial request, pre-append intention, then committed-state observation.
		if calls.Add(1) >= 3 {
			return initial.Add(20 * time.Second)
		}
		return initial
	}
	response := f.post(t, map[string]string{"action": "propose", "jobId": f.job.JobID})
	var review disposalReviewView
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &review) != nil {
		t.Fatal(response.Code, response.Body.String())
	}
	if !review.ReviewReadyAt.Equal(initial.Add(80 * time.Second)) {
		t.Fatal("audit delay consumed review minute", review.ReviewReadyAt)
	}
	f.now.Store(initial.Add(60 * time.Second).UnixNano())
	f.s.disposalNow = func() time.Time { return time.Unix(0, f.now.Load()).UTC() }
	if response = f.approve(t, review.WorkspaceDisposal); response.Code != 409 || f.calls.Load() != 0 {
		t.Fatal("pre-publication time counted")
	}
	f.now.Store(review.ReviewReadyAt.UnixNano())
	if response = f.approve(t, review.WorkspaceDisposal); response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
}
