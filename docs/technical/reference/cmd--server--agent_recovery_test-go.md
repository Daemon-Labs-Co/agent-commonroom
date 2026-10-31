# cmd/server/agent_recovery_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/agent_recovery_test.go)

Automated regression checks for agent_recovery; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`recoveryServer`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler)`.
  Calls: `db.seed`, `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`.
- **`TestRecoveryPreservesOrClearsOnlySessionBinding`** — Regression contract: TestRecoveryPreservesOrClearsOnlySessionBinding.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.snapshot`, `len`, `mockRequest`, `recoveryServer`, `s.load`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestRecoveryFailurePathsLeaveSessionUntouched`** — Regression contract: TestRecoveryFailurePathsLeaveSessionUntouched.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `errors.New`, `len`, `mockRequest`, `recoveryServer`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestWorkerIdleScanFailsClosed`** — Regression contract: TestWorkerIdleScanFailsClosed.
  Signature: `func(t *testing.T)`.
  Calls: `fakeWorkerRecord`, `filepath.Join`, `os.WriteFile`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `workerRecordsIdle`, `writeWorkerRecord`.
- **`TestRecoveryAuthenticationAndCSRF`** — Regression contract: TestRecoveryAuthenticationAndCSRF.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `mockRequest`, `r.AddCookie`, `recoveryServer`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestRecoveryAuditFailureDoesNotResetSession`** — Regression contract: TestRecoveryAuditFailureDoesNotResetSession.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.snapshot`, `len`, `mockRequest`, `recoveryServer`, `s.load`, `t.Fatal`, `t.Parallel`.
