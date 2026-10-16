# cmd/server/managed_jobs_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/managed_jobs_test.go)

Automated regression checks for managed_jobs; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`syntheticManagedJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler, managedJobRegistration)`.
  Calls: `context.Background`, `ctx.Value`, `db.seed`, `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `registration.Register`, `s.actorGuard`, `s.auth.protect`, `s.managedJobContext`, `t.Fatal`, `t.Helper`.
- **`TestJobJournalIsMandatoryWithDiagnosticsOffAndFailureStopsExecution`** — Regression contract: TestJobJournalIsMandatoryWithDiagnosticsOffAndFailureStopsExecution.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `context.WithCancel`, `db.seed`, `other.Err`, `reg.Update`, `registration.Update`, `s.load`, `stop`, `syntheticManagedJob`, `t.Fatal`, `t.Parallel`.
- **`TestJobAllocationCannotRepeatOrOverrideUnresolvedMembership`** — Regression contract: TestJobAllocationCannotRepeatOrOverrideUnresolvedMembership.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `next.Register`, `registration.Register`, `registration.Update`, `s.managedJobContext`, `s.managedJobContext(context.Background(), p).Value`, `syntheticManagedJob`, `t.Fatal`, `t.Parallel`.
- **`TestJobRecoveryPermissionsFlagsProofAndActiveTurnIdentity`** — Regression contract: TestJobRecoveryPermissionsFlagsProofAndActiveTurnIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `errors.New`, `managedJobFence`, `mockRequest`, `request.record`, `s.load`, `strings.Repeat`, `syntheticManagedJob`, `t.Fatal`, `t.Run`, `t.Setenv`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestJobDiagnosticsNeverExposeRawWorkerFieldsAndRespectRoom`** — Regression contract: TestJobDiagnosticsNeverExposeRawWorkerFieldsAndRespectRoom.
  Signature: `func(t *testing.T)`.
  Calls: `boundedJobDiagnostics`, `context.Background`, `errors.New`, `len`, `managedJobFence`, `managedJobFence(job).record`, `s.load`, `syntheticManagedJob`, `t.Fatal`, `t.Parallel`.
- **`TestStartupAndResumeRefuseUnconfirmedControllerJournalEvenIfWorkerIdle`** — Regression contract: TestStartupAndResumeRefuseUnconfirmedControllerJournalEvenIfWorkerIdle.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `http.NewServeMux`, `mockRequest`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `s.verifyWorkerAtStartup`, `syntheticManagedJob`, `t.Fatal`, `t.Parallel`.
- **`TestJobRecoveryRejectsMissingCSRFReasonAndWrongRoomBeforeWorker`** — Regression contract: TestJobRecoveryRejectsMissingCSRFReasonAndWrongRoomBeforeWorker.
  Signature: `func(t *testing.T)`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `mockRequest`, `r.AddCookie`, `strings.NewReader`, `syntheticManagedJob`, `t.Fatal`, `t.Parallel`.
