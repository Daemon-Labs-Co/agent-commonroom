# cmd/server/integrity_guard_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/integrity_guard_test.go)

Automated regression checks for integrity_guard; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestIntegrityStopSurvivesRestartAndRequiresVerifiedRecovery`** — Regression contract: TestIntegrityStopSurvivesRestartAndRequiresVerifiedRecovery.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `info.Mode`, `info.Mode().Perm`, `mockRequest`, `mockServer`, `os.IsNotExist`, `os.Stat`, `restarted.restoreIntegrityStop`, `s.tripIntegrityLocked`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestIntegrityOutageCancelsJobsAndDoesNotAutomaticallyResume`** — Regression contract: TestIntegrityOutageCancelsJobsAndDoesNotAutomaticallyResume.
  Signature: `func(t *testing.T)`.
  Calls: `cancelFlow`, `cancelModel`, `context.Background`, `context.WithCancel`, `db.seed`, `errors.New`, `flow.Err`, `mockRequest`, `mockServer`, `model.Err`, `s.agentsBlockedLocked`, `s.checkIntegrityNow`, `s.load`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestIntegrityTamperingOrFailedRecoveryCannotClearLatch`** — Regression contract: TestIntegrityTamperingOrFailedRecoveryCannotClearLatch.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `fresh.tripIntegrityLocked`, `mockRequest`, `mockServer`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **`TestNonHumanAuthenticatedActorCannotReadOrMutateAPI`** — Regression contract: TestNonHumanAuthenticatedActorCannotReadOrMutateAPI.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestIntegrityRecoveryRequiresOwnerReasonAndCSRF`** — Regression contract: TestIntegrityRecoveryRequiresOwnerReasonAndCSRF.
  Signature: `func(t *testing.T)`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `mockRequest`, `mockServer`, `r.AddCookie`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
