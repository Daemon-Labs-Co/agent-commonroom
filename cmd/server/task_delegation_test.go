package main

import (
	"agentroom/internal/room"
	"agentroom/internal/toolbroker"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func delegationFixture(t *testing.T) (*server, *mockDB, *controllerToolAuditor, toolbroker.Scope, http.Handler) {
	t.Helper()
	return delegationRoleFixture(t, false)
}
func delegationRoleFixture(t *testing.T, role bool) (*server, *mockDB, *controllerToolAuditor, toolbroker.Scope, http.Handler) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("feature.flag.updated", map[string]any{"name": "taskDelegation", "enabled": true})
	for _, v := range []struct{ id, endpoint string }{{"source", "ssh://worker"}, {"target", "ssh://worker/claude"}} {
		db.seed("connection.created", room.Connection{ID: v.id, Protocol: "CLI", Endpoint: v.endpoint, Revision: 1})
		status := "Idle"
		if v.id == "source" {
			status = "Working"
		}
		db.seed("participant.created", room.Participant{ID: v.id, RoomID: "engineering", ConnectionID: v.id, Status: status, Revision: 1})
	}
	db.seed("task.revised", room.Task{ID: "task", RoomID: "engineering", Title: "Bounded synthetic task", Criteria: "Synthetic criteria", Assignee: "source", Status: "Working", Revision: 2, DispatchID: "parent-dispatch", DispatchStatus: "pending"})
	now := time.Now().UTC()
	scope := toolbroker.Scope{Version: 1, JobID: strings.Repeat("a", 32), Room: "engineering", Participant: "source", Connection: "source", Provider: "codex", RepositoryDigest: strings.Repeat("b", 64), Deadline: now.Add(2 * time.Minute)}
	job := room.ManagedJob{Version: 4, Execution: pairExecution(4), RepositoryDigest: scope.RepositoryDigest, InitiatorID: ownerID, CompanyID: "company", TaskID: "task", TaskDispatchID: "parent-dispatch", JobID: scope.JobID, RoomID: scope.Room, ParticipantID: scope.Participant, ConnectionID: scope.Connection, Worker: "worker", Provider: "codex", Status: "submitted", AllocatedAt: now, UpdatedAt: now, Deadline: scope.Deadline}
	if role {
		db.seed("feature.flag.updated", map[string]any{"name": "agentCoordination", "enabled": true})
		db.seed("governance.appointed", map[string]any{"companyId": "company", "roomId": "engineering", "participantId": "source", "revision": 1, "performedBy": ownerID})
		job.CoordinatorID = "source"
		job.GovernanceRevision = 1
	}
	db.seed("worker.job.updated", job)
	s.delegationRepository = func(string) (string, error) { return scope.RepositoryDigest, nil }
	a, err := newControllerToolAuditor(t.Context(), s, scope, ownerID, func() (string, error) { return scope.RepositoryDigest, nil })
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-delegations", s.taskDelegations)
	mux.HandleFunc("/api/task-dispatch", s.taskDispatch)
	mux.HandleFunc("/api/events", s.events)
	return s, db, a, scope, s.auth.protect(s.actorGuard(mux))
}
func proposeDelegation(t *testing.T, a *controllerToolAuditor, scope toolbroker.Scope) {
	t.Helper()
	left, right := net.Pipe()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	defer left.Close()
	done := make(chan error, 1)
	go func() { done <- toolbroker.ServeAudit(ctx, right, scope, a) }()
	client, err := toolbroker.NewAuditClient(left, scope)
	if err != nil {
		t.Fatal(err)
	}
	req := toolbroker.DelegationRequest{ParentOperations: 1, RequestID: "assignment", TaskID: "task", TaskRevision: 2, DelegateID: "target"}
	if err = client.ProposeDelegation(ctx, req); err != nil {
		t.Fatal(err)
	}
	if err = client.ProposeDelegation(ctx, req); err != nil {
		t.Fatal("durable retry refused", err)
	}
	left.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("audit link did not close")
	}
}
func delegationDecisionRequest(h http.Handler, id, action string) *httptest.ResponseRecorder {
	return delegationHTTPRequest(h, "owner", id, action)
}
func delegationHTTPRequest(h http.Handler, actor, id, action string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/task-delegations", strings.NewReader(fmt.Sprintf(`{"action":%q,"id":%q,"revision":1}`, action, id)))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: actor})
	r.Header.Set("X-CSRF-Token", "csrf")
	r.Header.Set("Idempotency-Key", "delegation-"+action)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func finishDelegationParent(t *testing.T, s *server, scope toolbroker.Scope) {
	t.Helper()
	s.mu.Lock()
	err := s.recordManagedJobLocked(scope.JobID, "shutdown_confirmed")
	s.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var p room.Participant
	for _, v := range st.Participants {
		if v.ID == "source" {
			p = v
		}
	}
	s.finishAgentTurnWithTask(context.Background(), p, "Synthetic parent outcome", "codex", nil, "task", "parent-dispatch")
}
func singleDelegation(t *testing.T, s *server) room.TaskDelegation {
	t.Helper()
	st, err := s.load(t.Context())
	if err != nil || len(st.Delegations) != 1 {
		t.Fatal("missing/duplicate proposal", err, len(st.Delegations))
	}
	for _, d := range st.Delegations {
		return d
	}
	panic("unreachable")
}
func TestAuthenticatedDelegationProposalHumanApprovalAndBoundedDispatch(t *testing.T) {
	t.Parallel()
	s, db, a, scope, h := delegationFixture(t)
	proposeDelegation(t, a, scope)
	d := singleDelegation(t, s)
	st, _ := s.load(t.Context())
	if st.Tasks[0].Assignee != "source" || len(s.active) != 0 || d.SourceParticipantID != scope.Participant || d.RepositoryDigest != scope.RepositoryDigest {
		t.Fatal("proposal granted assignment or wrong scope")
	}
	if w := delegationDecisionRequest(h, d.ID, "approve"); w.Code != 409 {
		t.Fatal("unreconciled parent approved", w.Code)
	}
	finishDelegationParent(t, s, scope)
	w := delegationDecisionRequest(h, d.ID, "approve")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	before := len(db.snapshot())
	if w = delegationDecisionRequest(h, d.ID, "approve"); w.Code != 200 || len(db.snapshot()) != before {
		t.Fatal("approval replay duplicated")
	}
	st, err := (&server{db: db}).load(t.Context())
	if err != nil || st.Tasks[0].Assignee != "target" || st.Tasks[0].DelegationID != d.ID || len(s.active) != 0 {
		t.Fatal("approval scope/restart failed", err)
	}
	task := st.Tasks[0]
	w = mockRequest(h, "POST", "/api/events", fmt.Sprintf(`{"kind":"task.revised","payload":{"id":"task","revision":%d,"title":"Expanded task","criteria":"Other scope","assignee":"target","status":"Assigned"}}`, task.Revision))
	if w.Code != 409 {
		t.Fatal("scope expanded", w.Code, w.Body.String())
	}
	// Exercise the actual dispatch path with a synthetic isolated-turn boundary.
	s.taskReady = func(string) error { return nil }
	finished := make(chan struct{})
	s.taskTurn = func(ctx context.Context, p room.Participant, _, _ string) (string, error) {
		defer close(finished)
		binding, _ := ctx.Value(taskJobContextKey{}).(taskJobBinding)
		if binding.DelegationID != d.ID {
			t.Error("queue stripped delegation")
		}
		jobCtx := s.managedExecutionJobContext(ctx, p, scope.RepositoryDigest, func() (string, error) { return scope.RepositoryDigest, nil })
		reg := jobCtx.Value(managedJobContextKey{}).(managedJobRegistration)
		if err := reg.Register("worker", "claude"); err != nil {
			t.Error(err)
			return "", err
		}
		if err := reg.Update("submitted"); err != nil {
			t.Error(err)
			return "", err
		}
		targetScope := toolbroker.Scope{Version: 1, JobID: reg.JobID, Room: p.RoomID, Participant: p.ID, Connection: p.ConnectionID, Provider: "claude", RepositoryDigest: scope.RepositoryDigest, Deadline: reg.Deadline}
		audit, err := newControllerToolAuditor(ctx, s, targetScope, ownerID, func() (string, error) { return scope.RepositoryDigest, nil })
		if err != nil {
			t.Error(err)
			return "", err
		}
		intent := toolbroker.Intent{Scope: targetScope, CallID: "read", Operation: "read", PayloadDigest: strings.Repeat("c", 64)}
		if audit.RecordIntent(ctx, intent) != nil || audit.RecordOutcome(ctx, intent, toolbroker.Outcome{Status: "succeeded", OutputDigest: strings.Repeat("d", 64)}) != nil {
			t.Error("bounded tool audit refused")
		}
		if err = reg.Update("shutdown_confirmed"); err != nil {
			t.Error(err)
		}
		return "Synthetic delegated result", err
	}
	w = dispatchRequest(h, "delegated-dispatch", fmt.Sprintf(`{"taskId":"task","revision":%d,"mode":"work"}`, task.Revision))
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("delegated turn did not finish")
	}
	deadline := time.Now().Add(time.Second)
	for {
		s.mu.Lock()
		busy := s.active["target"]
		s.mu.Unlock()
		if !busy {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("turn completion stalled")
		}
		time.Sleep(time.Millisecond)
	}
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, j := range st.ManagedJobs {
		if j.ParticipantID == "target" && (j.DelegationID != d.ID || j.Deadline.After(d.ExpiresAt) || j.RepositoryDigest != d.RepositoryDigest) {
			t.Fatal("child scope expanded")
		}
	}
}
func TestDelegationRefusesForgedStaleAndExpandedScope(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"off", "wrong-job", "wrong-task", "stale-task", "foreign-target", "self", "audit-failure"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, a, scope, _ := delegationFixture(t)
			req := toolbroker.DelegationRequest{ParentOperations: 1, RequestID: "request", TaskID: "task", TaskRevision: 2, DelegateID: "target"}
			switch mode {
			case "off":
				db.seed("feature.flag.updated", map[string]any{"name": "taskDelegation", "enabled": false})
			case "wrong-job":
				scope.JobID = strings.Repeat("f", 32)
			case "wrong-task":
				req.TaskID = "foreign"
			case "stale-task":
				req.TaskRevision = 3
			case "foreign-target":
				req.DelegateID = "foreign"
			case "self":
				req.DelegateID = "source"
			case "audit-failure":
				db.failAt = "commit"
			}
			before := len(db.snapshot())
			if a.RecordDelegation(t.Context(), scope, req) == nil || len(db.snapshot()) != before {
				t.Fatal("invalid proposal accepted")
			}
			if (mode == "audit-failure") != s.integrityFault {
				t.Fatal("incorrect integrity stop")
			}
		})
	}
}
func TestDelegationApprovalRevalidatesRevocationAndPreservesHistory(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"off-on", "target-revision", "source-removed", "task-edited", "repository", "audit-failure", "reject"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, a, scope, h := delegationFixture(t)
			proposeDelegation(t, a, scope)
			finishDelegationParent(t, s, scope)
			d := singleDelegation(t, s)
			switch mode {
			case "off-on":
				for _, on := range []bool{false, true} {
					db.seed("feature.flag.updated", map[string]any{"name": "taskDelegation", "enabled": on})
				}
			case "target-revision":
				st, _ := s.load(t.Context())
				for _, p := range st.Participants {
					if p.ID == "target" {
						p.Revision++
						db.seed("participant.updated", p)
					}
				}
			case "source-removed":
				db.seed("participant.removed", map[string]string{"id": "source"})
			case "task-edited":
				st, _ := s.load(t.Context())
				v := st.Tasks[0]
				v.Criteria = "Expanded"
				v.Revision++
				db.seed("task.revised", v)
			case "repository":
				s.delegationRepository = func(string) (string, error) { return strings.Repeat("f", 64), nil }
			case "audit-failure":
				db.failAt = "commit"
			}
			action := "approve"
			if mode == "reject" {
				action = "reject"
			}
			w := delegationDecisionRequest(h, d.ID, action)
			want := 409
			if mode == "audit-failure" {
				want = 500
			}
			if mode == "reject" {
				want = 200
			}
			if w.Code != want {
				t.Fatal(mode, w.Code, w.Body.String())
			}
			db.failAt = ""
			st, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			if st.Tasks[0].Assignee != "source" || st.Delegations[d.ID].Status == "Approved" {
				t.Fatal("invalid approval reassigned task")
			}
		})
	}
}

func TestDelegatedToolEffectsRevalidateRoleAndDelegatePermissions(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"off-on", "target-revision", "source-revision", "governor-transfer", "scope-expansion", "disabled-agent", "actor-revision", "healthy"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, a, scope, h := delegationRoleFixture(t, true)
			proposeDelegation(t, a, scope)
			finishDelegationParent(t, s, scope)
			d := singleDelegation(t, s)
			if w := delegationDecisionRequest(h, d.ID, "approve"); w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, err := s.load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			task := st.Tasks[0]
			task.Revision++
			task.DispatchID = "child-dispatch"
			task.DispatchStatus = "pending"
			task.Status = "Working"
			db.seed("task.revised", task)
			now := time.Now().UTC()
			childScope := toolbroker.Scope{Version: 1, JobID: strings.Repeat("e", 32), Room: "engineering", Participant: "target", Connection: "target", Provider: "claude", RepositoryDigest: scope.RepositoryDigest, Deadline: scope.Deadline}
			job := room.ManagedJob{Version: 4, Execution: pairExecution(4), DelegationID: d.ID, RepositoryDigest: scope.RepositoryDigest, InitiatorID: ownerID, CompanyID: "company", TaskID: "task", TaskDispatchID: "child-dispatch", JobID: childScope.JobID, RoomID: "engineering", ParticipantID: "target", ConnectionID: "target", Worker: "worker", Provider: "claude", Status: "submitted", AllocatedAt: now, UpdatedAt: now, Deadline: scope.Deadline}
			db.seed("worker.job.updated", job)
			audit, err := newControllerToolAuditor(t.Context(), s, childScope, ownerID, func() (string, error) { return scope.RepositoryDigest, nil })
			if err != nil {
				t.Fatal(err)
			}
			intent := toolbroker.Intent{Scope: childScope, CallID: "read", Operation: "read", PayloadDigest: strings.Repeat("c", 64)}
			expanded := intent
			expanded.CallID = "expanded-write"
			expanded.Operation = "write"
			if audit.RecordIntent(t.Context(), expanded) == nil {
				t.Fatal("child expanded beyond parent's read-only policy")
			}
			if audit.RecordIntent(t.Context(), intent) != nil {
				t.Fatal("healthy child denied")
			}
			// Child requests cannot turn into another delegation chain.
			if audit.RecordDelegation(t.Context(), childScope, toolbroker.DelegationRequest{ParentOperations: 1, RequestID: "nested", TaskID: "task", TaskRevision: task.Revision, DelegateID: "source"}) == nil {
				t.Fatal("subdelegation expanded authority")
			}
			switch mode {
			case "off-on":
				for _, on := range []bool{false, true} {
					db.seed("feature.flag.updated", map[string]any{"name": "taskDelegation", "enabled": on})
				}
			case "target-revision", "source-revision":
				st, _ = s.load(t.Context())
				pid := "target"
				if mode == "source-revision" {
					pid = "source"
				}
				for _, p := range st.Participants {
					if p.ID == pid {
						p.Revision++
						db.seed("participant.updated", p)
					}
				}
			case "governor-transfer":
				db.seed("governance.appointed", map[string]any{"companyId": "company", "roomId": "engineering", "participantId": "target", "revision": 2, "performedBy": ownerID})
			case "disabled-agent", "actor-revision":
				st, _ = s.load(t.Context())
				agent, _ := member(st, "bot-target")
				agent.Revision++
				if mode == "disabled-agent" {
					agent.Enabled = false
				}
				db.seed("actor.revised", agent)
			case "scope-expansion":
				intent.Scope.RepositoryDigest = strings.Repeat("f", 64)
			}
			before := len(db.snapshot())
			err = audit.RecordOutcome(t.Context(), intent, toolbroker.Outcome{Status: "succeeded", OutputDigest: strings.Repeat("d", 64)})
			if (err == nil) != (mode == "healthy") {
				t.Fatal("wrong child outcome decision", mode, err)
			}
			if mode != "healthy" && len(db.snapshot()) != before {
				t.Fatal("revoked child outcome persisted")
			}
			if mode == "healthy" {
				st, _ = s.load(t.Context())
				expired := d
				expired.ExpiresAt = time.Now().Add(-time.Second)
				if delegationCurrent(st, expired) {
					t.Fatal("expired grant revived")
				}
			}
		})
	}
}
func TestDelegationOffDoesNotRevokeIndependentToolWork(t *testing.T) {
	t.Parallel()
	_, db, a, intent := scopedAuditFixture(t)
	db.seed("feature.flag.updated", map[string]any{"name": "taskDelegation", "enabled": false})
	if a.RecordIntent(t.Context(), intent) != nil {
		t.Fatal("off revoked independent work")
	}
}

func TestDelegationHumanEndpointCannotImpersonateWorker(t *testing.T) {
	t.Parallel()
	s, db, _, _, h := delegationFixture(t)
	before := len(db.snapshot())
	r := httptest.NewRequest("POST", "http://127.0.0.1:8787/api/task-delegations", strings.NewReader(`{"action":"propose","id":"fake","revision":1}`))
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "owner"})
	r.Header.Set("X-CSRF-Token", "csrf")
	r.Header.Set("Idempotency-Key", "fake-proposal")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 || len(db.snapshot()) != before {
		t.Fatal("browser created worker proposal", w.Code)
	}
	w = mockRequest(h, "POST", "/api/events", `{"kind":"task.delegation.proposed","payload":{}}`)
	if w.Code != 400 {
		t.Fatal("generic event forged worker proposal")
	}
	db.seed("actor.revised", room.Actor{ID: "agent", Kind: "agent", Enabled: true})
	s.auth.sessions["agent"] = ownerSession{ActorID: "agent", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w = delegationHTTPRequest(h, "agent", "fake", "approve"); w.Code != 403 {
		t.Fatal("agent approved its own proposal", w.Code)
	}
}
func TestDelegationApprovalConflictDoesNotReassignOrTripIntegrity(t *testing.T) {
	t.Parallel()
	s, db, a, scope, h := delegationFixture(t)
	proposeDelegation(t, a, scope)
	finishDelegationParent(t, s, scope)
	d := singleDelegation(t, s)
	once := false
	s.db = &interleavedEventDB{eventDatabase: db, beforeBegin: func(context.Context) {
		if !once {
			once = true
			db.seed("feature.flag.updated", map[string]any{"name": "taskDelegation", "enabled": false})
		}
	}}
	w := delegationDecisionRequest(h, d.ID, "approve")
	if w.Code != 409 || s.integrityFault {
		t.Fatal("concurrent disable bypassed or falsely corrupted audit", w.Code)
	}
	st, err := s.load(t.Context())
	if err != nil || st.Tasks[0].Assignee != "source" || st.Delegations[d.ID].Status != "Proposed" {
		t.Fatal("conflicting approval partially applied", err)
	}
}

func TestDelegationFlagDefaultAndOwnerEnable(t *testing.T) {
	t.Parallel()
	s, _, _ := mockServer(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/task-delegations", s.taskDelegations)
	mux.HandleFunc("/api/events", s.events)
	h := s.auth.protect(s.actorGuard(mux))
	w := mockRequest(h, "GET", "/api/task-delegations", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"enabled":false`) {
		t.Fatal("delegation default not off", w.Code)
	}
	w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"taskDelegation","enabled":true}}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = mockRequest(h, "GET", "/api/task-delegations", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"enabled":true`) {
		t.Fatal("owner enable failed", w.Code)
	}
}
