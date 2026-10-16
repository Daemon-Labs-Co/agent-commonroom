# cmd/server/worker_pair_shutdown_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_pair_shutdown_test.go)

Automated regression checks for worker_pair_shutdown; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ownedShutdownRunner`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T runtime toolbroker.DockerRunner }`.
- **`(ownedShutdownRunner).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, input []byte, out io.Writer) error`.
  Calls: `ctx.Value`, `r.runtime.Run`, `r.t.Fatal`.
- **`TestSupervisorShutdownFencesBothRolesAfterReleasingCreator`** — Regression contract: TestSupervisorShutdownFencesBothRolesAfterReleasingCreator.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `filepath.Join`, `len`, `os.Stat`, `pairFixture`, `s.BeginCreate`, `s.Shutdown`, `s.ctx.Err`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestSupervisorShutdownFailureRetainsUnresolvedEvidence`** — Regression contract: TestSupervisorShutdownFailureRetainsUnresolvedEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `pairFixture`, `s.Shutdown`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestSupervisorShutdownAlwaysRetainsPresentWorkspace`** — Regression contract: TestSupervisorShutdownAlwaysRetainsPresentWorkspace.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `newWorkspaceCleanup`, `pairContainerName`, `s.Shutdown`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `workspacePair`.
- **`TestSupervisorShutdownCancelledCleanupDoesNotClaimAbsence`** — Regression contract: TestSupervisorShutdownCancelledCleanupDoesNotClaimAbsence.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `len`, `pairFixture`, `s.Shutdown`, `s.ctx.Err`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestSupervisorRunAlwaysReconcilesTurnResults`** — Regression contract: TestSupervisorRunAlwaysReconcilesTurnResults.
  Signature: `func(t *testing.T)`.
  Calls: `errors.New`, `len`, `newWorkspaceCleanup`, `pairContainerName`, `panic`, `readWorkerRecord`, `recover`, `s.Run`, `s.cancel`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `workspacePair`.
- **`TestShutdownWaitsForActiveTurnBeforeReleasingCreator`** — Regression contract: TestShutdownWaitsForActiveTurnBeforeReleasingCreator.
  Signature: `func(t *testing.T)`.
  Calls: `awaitAuditResult`, `close`, `ctx.Done`, `ctx.Err`, `lease.Close`, `make`, `newWorkspaceCleanup`, `s.Run`, `s.Shutdown`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`, `workerTryLease`, `workspacePair`.
