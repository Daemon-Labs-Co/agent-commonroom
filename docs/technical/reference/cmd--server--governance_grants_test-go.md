# cmd/server/governance_grants_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/governance_grants_test.go)

Automated regression checks for governance_grants; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`coordinationFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler, room.Participant, room.GovernanceAppointment)`.
  Calls: `db.seed`, `governanceFixture`, `s.load`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`coordinatedContext`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.GovernanceAppointment) context.Context`.
  Calls: `context.Background`, `context.WithValue`, `withInitiatingActor`.
- **`replaceCoordinator`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler)`.
  Calls: `governanceRequest`, `t.Fatal`, `t.Helper`, `w.Body.String`.
- **`TestCoordinationAdmissionAndSubmissionRevalidateRevision`** — Regression contract: TestCoordinationAdmissionAndSubmissionRevalidateRevision.
  Signature: `func(t *testing.T)`.
  Calls: `coordinatedContext`, `coordinationFixture`, `db.seed`, `db.snapshot`, `len`, `registration.Register`, `registration.Update`, `replaceCoordinator`, `s.load`, `s.managedJobContext`, `s.managedJobContext(ctx, p).Value`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `withInitiatingActor`.
- **`TestCoordinationToolAuditRevalidatesAfterTransferAndRestart`** — Regression contract: TestCoordinationToolAuditRevalidatesAfterTransferAndRestart.
  Signature: `func(t *testing.T)`.
  Calls: `auditor.RecordIntent`, `auditor.RecordOutcome`, `coordinationFixture`, `db.seed`, `db.snapshot`, `len`, `newControllerToolAuditor`, `now.Add`, `replaceCoordinator`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().UTC`.
- **`TestCoordinationTransferCancelsOnlyRoleBoundTurnsAndWithholdsStaleResults`** — Regression contract: TestCoordinationTransferCancelsOnlyRoleBoundTurnsAndWithholdsStaleResults.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `coordinatedContext`, `coordinationFixture`, `governanceRequest`, `independent.Err`, `len`, `role.Err`, `s.finishAgentTurn`, `s.load`, `stop`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCoordinationSendHandlerCapturesDependencyAndRejectsStaleInput`** — Regression contract: TestCoordinationSendHandlerCapturesDependencyAndRejectsStaleInput.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `close`, `coordinationFixture`, `ctx.Done`, `ctx.Err`, `ctx.Value`, `db.seed`, `db.snapshot`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `make`, `r.AddCookie`, `r.Header.Clone`, `r.Header.Set`, `r2.AddCookie`, `replaceCoordinator`, `replay.Header`, `replay.Header().Get`, `s.mu.Lock`, `s.mu.Unlock`, `strings.NewReader`, `strings.Replace`, `t.Fatal`, `t.Run`, `t.Setenv`, `time.After`, `time.Now`, `time.Now().Add`, `time.Now().After`, `time.Sleep`, `w.Body.String`.
- **`TestCoordinationDisableCancelsRoleTurnsAndFreshEnableDoesNotReviveOldWork`** — Regression contract: TestCoordinationDisableCancelsRoleTurnsAndFreshEnableDoesNotReviveOldWork.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `cancel`, `context.Background`, `context.WithCancel`, `context.WithValue`, `coordinatedContext`, `coordinationContextAllowed`, `coordinationFixture`, `coordinationJobAllowed`, `db.seed`, `independent.Err`, `mockRequest`, `registration.Register`, `registration.Update`, `role.Err`, `s.load`, `s.managedJobContext`, `s.managedJobContext(fresh, p).Value`, `stop`, `t.Context`, `t.Fatal`, `t.Parallel`, `w.Body.String`, `withInitiatingActor`.
