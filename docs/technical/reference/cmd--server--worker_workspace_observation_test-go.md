# cmd/server/worker_workspace_observation_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_workspace_observation_test.go)

Automated regression checks for worker_workspace_observation; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticWorkspaceObservationGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls int deny bool mutate func(workerJobRecord) }`.
- **`(*syntheticWorkspaceObservationGuard).AuthorizeWorkspaceObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord) error`.
  Calls: `ctx.Err`, `errors.New`, `g.mutate`.
- **`retainedObservationFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (string, workerJobRecord, *workspaceCleanupDocker)`.
  Calls: `len`, `newWorkspaceCleanup`, `pairContainerName`, `retainedObservationRecord`, `s.Shutdown`, `s.Start`, `t.Context`, `t.Fatal`, `t.Helper`, `workspacePair`.
- **`TestWorkspaceObservationHoldsCreatorLockAndExpires`** — Regression contract: TestWorkspaceObservationHoldsCreatorLockAndExpires.
  Signature: `func(t *testing.T)`.
  Calls: `competing.Close`, `escaped.Check`, `l.Check`, `lease.Close`, `readWorkerRecord`, `retainedObservationFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `withWorkerWorkspaceObservation`, `workerTryLease`.
- **`TestWorkspaceObservationRefusesUnsafeAdmission`** — Regression contract: TestWorkspaceObservationRefusesUnsafeAdmission.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `lease.Close`, `os.Remove`, `os.WriteFile`, `pairContainerName`, `retainedObservationFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `withWorkerWorkspaceObservation`, `workerTryLease`.
- **`TestWorkspaceObservationRechecksAfterUseAndRetainsFailures`** — Regression contract: TestWorkspaceObservationRechecksAfterUseAndRetainsFailures.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `changed.UpdatedAt.Add`, `context.Background`, `context.WithCancel`, `errors.New`, `escaped.Check`, `invoke`, `l.Check`, `lease.Close`, `panic`, `readWorkerRecord`, `recover`, `retainedObservationFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `withWorkerWorkspaceObservation`, `workerTryLease`, `writeWorkerRecord`.
- **`TestWorkspaceObservationGuardCannotMutateLeaseBinding`** — Regression contract: TestWorkspaceObservationGuardCannotMutateLeaseBinding.
  Signature: `func(t *testing.T)`.
  Calls: `l.Check`, `retainedObservationFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `withWorkerWorkspaceObservation`.
