# cmd/server/agent_safety_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/agent_safety_test.go)

Automated regression checks for agent_safety; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`safetyHarness`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler)`.
  Calls: `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`.
- **`TestSafetyResumeRefusesUnknownWorkerOrIntegrityStop`** — Regression contract: TestSafetyResumeRefusesUnknownWorkerOrIntegrityStop.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `len`, `mockRequest`, `safetyHarness`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestApplicationStopCancelsWorkflowEvenWhenDatabaseReadFails`** — Regression contract: TestApplicationStopCancelsWorkflowEvenWhenDatabaseReadFails.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `ctx.Err`, `mockRequest`, `safetyHarness`, `t.Fatal`, `t.Parallel`.
- **`TestEmergencyStopAndExplicitResume`** — Regression contract: TestEmergencyStopAndExplicitResume.
  Signature: `func(t *testing.T)`.
  Calls: `cancelDev`, `cancelQA`, `context.Background`, `context.WithCancel`, `db.seed`, `dev.Err`, `fresh.agentsBlockedLocked`, `mockRequest`, `qa.Err`, `s.agentsBlockedLocked`, `s.load`, `safetyHarness`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestEmergencyStopFailureAndAccess`** — Regression contract: TestEmergencyStopFailureAndAccess.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `context.WithCancel`, `ctx.Err`, `db.seed`, `db.snapshot`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `mockRequest`, `r.AddCookie`, `safetyHarness`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`.
- **`TestCancelledTurnCannotPostAfterResume`** — Regression contract: TestCancelledTurnCannotPostAfterResume.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `db.seed`, `len`, `mockServer`, `s.finishAgentTurn`, `s.load`, `t.Fatal`, `t.Parallel`.
- **`TestEmergencyUIFlagDoesNotDisableBackend`** — Regression contract: TestEmergencyUIFlagDoesNotDisableBackend.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `httptest.NewRecorder`, `httptest.NewRequest`, `mockRequest`, `optionalScriptHandler`, `optionalScriptHandler(enabled, root, "agent-safety.js").ServeHTTP`, `os.WriteFile`, `safetyHarness`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestSafetyResumeVerifiesChainBeforeWorkerOrApproval`** — Regression contract: TestSafetyResumeVerifiesChainBeforeWorkerOrApproval.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `mockRequest`, `safetyHarness`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **`TestFailedEmergencyStopAuditSurvivesRestart`** — Regression contract: TestFailedEmergencyStopAuditSurvivesRestart.
  Signature: `func(t *testing.T)`.
  Calls: `cancelModel`, `cancelWorkflow`, `context.Background`, `context.WithCancel`, `filepath.Join`, `fresh.agentsBlockedLocked`, `fresh.restoreIntegrityStop`, `info.Mode`, `info.Mode().Perm`, `mockRequest`, `model.Err`, `os.Stat`, `s.checkIntegrityNow`, `safetyHarness`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workflow.Err`.
