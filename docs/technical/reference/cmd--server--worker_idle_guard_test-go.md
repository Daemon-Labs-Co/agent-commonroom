# cmd/server/worker_idle_guard_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_idle_guard_test.go)

Automated regression checks for worker_idle_guard; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWorkerIdleDeploymentPolicy`** — Regression contract: TestWorkerIdleDeploymentPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `readWorkerIdlePolicy`, `t.Fatal`, `t.Parallel`.
- **`idleWorkspace`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, enabled bool) (string, *workerPairSupervisor, *workspaceCleanupDocker, time.Time)`.
  Calls: `newWorkspaceCleanup`, `pairContainerName`, `s.Start`, `s.configureIdle`, `s.mu.Lock`, `s.mu.Unlock`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`, `time.Now`, `workspacePair`.
- **`TestWorkerIdleGuardExactBoundaryAndRetention`** — Regression contract: TestWorkerIdleGuardExactBoundaryAndRetention.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `filepath.Join`, `idleWorkspace`, `len`, `os.Stat`, `pairRoleSettled`, `readWorkerRecord`, `record.IdleShutdownRequestedAt.IsZero`, `s.BeginCreate`, `s.ctx.Err`, `s.expireIdle`, `since.Add`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerIdleGuardDisabledAndBusy`** — Regression contract: TestWorkerIdleGuardDisabledAndBusy.
  Signature: `func(t *testing.T)`.
  Calls: `idleWorkspace`, `s.expireIdle`, `s.finishQueuedTurn`, `s.queueTurn`, `s.startup.Lock`, `s.startup.Unlock`, `s.turn.Lock`, `s.turn.Unlock`, `since.Add`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkerIdlePersistenceFailureHasNoRuntimeEffects`** — Regression contract: TestWorkerIdlePersistenceFailureHasNoRuntimeEffects.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `idleWorkspace`, `len`, `os.Mkdir`, `os.Remove`, `s.expireIdle`, `since.Add`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerIdlePolicyCannotBeRebound`** — Regression contract: TestWorkerIdlePolicyCannotBeRebound.
  Signature: `func(t *testing.T)`.
  Calls: `idleWorkspace`, `s.configureIdle`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerIdleQueuedArrivalResetsFullInterval`** — Regression contract: TestWorkerIdleQueuedArrivalResetsFullInterval.
  Signature: `func(t *testing.T)`.
  Calls: `finished.Add`, `idleWorkspace`, `s.expireIdle`, `s.finishQueuedTurn`, `s.mu.Lock`, `s.mu.Unlock`, `s.queueTurn`, `since.Add`, `t.Fatal`, `t.Parallel`.
- **`TestRunSnapshotsWorkerIdleConfiguration`** — Regression contract: TestRunSnapshotsWorkerIdleConfiguration.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `newWorkspaceCleanup`, `s.Run`, `s.mu.Lock`, `s.mu.Unlock`, `t.Context`, `t.Fatal`, `t.Run`, `t.Setenv`, `workspacePair`.
- **`TestWorkerIdleCleanupUncertaintyCannotRevivePair`** — Regression contract: TestWorkerIdleCleanupUncertaintyCannotRevivePair.
  Signature: `func(t *testing.T)`.
  Calls: `idleWorkspace`, `readWorkerRecord`, `s.expireIdle`, `s.queueTurn`, `since.Add`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerIdleEvidenceValidation`** — Regression contract: TestWorkerIdleEvidenceValidation.
  Signature: `func(t *testing.T)`.
  Calls: `idleWorkspace`, `record.StartedAt.Add`, `s.mu.Lock`, `s.mu.Unlock`, `t.Fatal`, `t.Parallel`, `validWorkerRecord`.
