# cmd/server/task_delegation_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_delegation_test.go)

Automated regression checks for task_delegation; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`delegationFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, *controllerToolAuditor, toolbroker.Scope, http.Handler)`.
  Calls: `delegationRoleFixture`, `t.Helper`.
- **`delegationRoleFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, role bool) (*server, *mockDB, *controllerToolAuditor, toolbroker.Scope, http.Handler)`.
  Calls: `db.seed`, `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `newControllerToolAuditor`, `now.Add`, `pairExecution`, `s.actorGuard`, `s.auth.protect`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `time.Now`, `time.Now().UTC`.
- **`proposeDelegation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, a *controllerToolAuditor, scope toolbroker.Scope)`.
  Calls: `cancel`, `client.ProposeDelegation`, `context.WithCancel`, `left.Close`, `make`, `net.Pipe`, `t.Context`, `t.Fatal`, `t.Helper`, `time.After`, `toolbroker.NewAuditClient`, `toolbroker.ServeAudit`.
- **`delegationDecisionRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, id, action string) *httptest.ResponseRecorder`.
  Calls: `delegationHTTPRequest`.
- **`delegationHTTPRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, actor, id, action string) *httptest.ResponseRecorder`.
  Calls: `fmt.Sprintf`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`.
- **`finishDelegationParent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, s *server, scope toolbroker.Scope)`.
  Calls: `context.Background`, `s.finishAgentTurnWithTask`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.recordManagedJobLocked`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`singleDelegation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, s *server) room.TaskDelegation`.
  Calls: `len`, `panic`, `s.load`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`TestAuthenticatedDelegationProposalHumanApprovalAndBoundedDispatch`** — Regression contract: TestAuthenticatedDelegationProposalHumanApprovalAndBoundedDispatch.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `audit.RecordIntent`, `audit.RecordOutcome`, `close`, `ctx.Value`, `db.snapshot`, `delegationDecisionRequest`, `delegationFixture`, `dispatchRequest`, `finishDelegationParent`, `fmt.Sprintf`, `j.Deadline.After`, `jobCtx.Value`, `len`, `make`, `mockRequest`, `newControllerToolAuditor`, `proposeDelegation`, `reg.Register`, `reg.Update`, `s.load`, `s.managedExecutionJobContext`, `s.mu.Lock`, `s.mu.Unlock`, `singleDelegation`, `strings.Repeat`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.After`, `time.Now`, `time.Now().Add`, `time.Now().After`, `time.Sleep`, `w.Body.String`.
- **`TestDelegationRefusesForgedStaleAndExpandedScope`** — Regression contract: TestDelegationRefusesForgedStaleAndExpandedScope.
  Signature: `func(t *testing.T)`.
  Calls: `a.RecordDelegation`, `db.seed`, `db.snapshot`, `delegationFixture`, `len`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestDelegationApprovalRevalidatesRevocationAndPreservesHistory`** — Regression contract: TestDelegationApprovalRevalidatesRevocationAndPreservesHistory.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `delegationDecisionRequest`, `delegationFixture`, `finishDelegationParent`, `proposeDelegation`, `s.load`, `singleDelegation`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestDelegatedToolEffectsRevalidateRoleAndDelegatePermissions`** — Regression contract: TestDelegatedToolEffectsRevalidateRoleAndDelegatePermissions.
  Signature: `func(t *testing.T)`.
  Calls: `audit.RecordDelegation`, `audit.RecordIntent`, `audit.RecordOutcome`, `db.seed`, `db.snapshot`, `delegationCurrent`, `delegationDecisionRequest`, `delegationRoleFixture`, `finishDelegationParent`, `len`, `member`, `newControllerToolAuditor`, `pairExecution`, `proposeDelegation`, `s.load`, `singleDelegation`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `time.Now().UTC`, `w.Body.String`.
- **`TestDelegationOffDoesNotRevokeIndependentToolWork`** — Regression contract: TestDelegationOffDoesNotRevokeIndependentToolWork.
  Signature: `func(t *testing.T)`.
  Calls: `a.RecordIntent`, `db.seed`, `scopedAuditFixture`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestDelegationHumanEndpointCannotImpersonateWorker`** — Regression contract: TestDelegationHumanEndpointCannotImpersonateWorker.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `delegationFixture`, `delegationHTTPRequest`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `mockRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestDelegationApprovalConflictDoesNotReassignOrTripIntegrity`** — Regression contract: TestDelegationApprovalConflictDoesNotReassignOrTripIntegrity.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `delegationDecisionRequest`, `delegationFixture`, `finishDelegationParent`, `proposeDelegation`, `s.load`, `singleDelegation`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestDelegationFlagDefaultAndOwnerEnable`** — Regression contract: TestDelegationFlagDefaultAndOwnerEnable.
  Signature: `func(t *testing.T)`.
  Calls: `http.NewServeMux`, `mockRequest`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestDistinctTaskReviewerManagedRegistration`** — Review registration must follow the durably dispatched reviewer, even when delegation is disabled and the assignee is a different participant.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `delegationFixture`, `dispatchRequest`, `fmt.Errorf`, `fmt.Sprint`, `jobCtx.Value`, `make`, `managedTaskCurrent`, `reg.Register`, `reg.Update`, `s.load`, `s.managedExecutionJobContext`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `w.Body.String`.
