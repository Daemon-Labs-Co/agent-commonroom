# cmd/server/worker_workspace_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_workspace_test.go)

Automated regression checks for worker_workspace; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workspaceLaunchFake`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *pairLaunchFake failWorkspace string }`.
- **`(*workspaceLaunchFake).WorkspaceIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (toolbroker.WorkspaceVolumeIdentity, error)`.
  Calls: `append`, `strings.Repeat`.
- **`(*workspaceLaunchFake).PrepareVolume`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent toolbroker.WorkspaceVolumeIdentity) (toolbroker.WorkspaceVolumeIdentity, error)`.
  Calls: `append`, `errors.New`, `f.t.Fatal`, `readWorkerRecord`, `strings.Repeat`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Format`.
- **`workspacePair`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (string, *workerPairSupervisor, *workspaceLaunchFake)`.
  Calls: `admitWorkerPair`, `pairRequest`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`TestWorkspacePairJournalsVolumeBeforeContainers`** — Regression contract: TestWorkspacePairJournalsVolumeBeforeContainers.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `readWorkerRecord`, `s.Start`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `validWorkerRecord`, `workerJobSettled`, `workspacePair`.
- **`TestWorkspacePairFailureCannotStartContainers`** — Regression contract: TestWorkspacePairFailureCannotStartContainers.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `readWorkerRecord`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workspacePair`.
- **Type `workspaceCleanupDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *pairFakeDocker t *testing.T record workerJobRecord volume bool removed int mode string observedCreation string }`.
- **`(*workspaceCleanupDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `ctx.Value`, `errors.New`, `f.pairFakeDocker.Run`, `f.t.Fatal`, `io.WriteString`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `path.Join`, `strings.HasPrefix`, `strings.Repeat`, `time.Now`, `time.Now().Add`, `time.Now().Add(time.Hour).UTC`, `time.Now().Add(time.Hour).UTC().Format`, `time.Now().UTC`, `time.Now().UTC().Format`.
- **`workspaceFence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *workerPairSupervisor) workerFenceRequest`.
  Calls: `pairRequest`.
- **`newWorkspaceCleanup`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, s *workerPairSupervisor) *workspaceCleanupDocker`.
  Calls: `t.Helper`.
- **`fenceWorkspace`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, s *workerPairSupervisor, f *workspaceCleanupDocker) (workerJobRecord, error)`.
  Calls: `fenceWorkerJobContainers`, `stopWorkerPair`, `workspaceFence`.
- **`TestWorkspaceShutdownRequiresVolumeAndLeaseProof`** — Regression contract: TestWorkspaceShutdownRequiresVolumeAndLeaseProof.
  Signature: `func(t *testing.T)`.
  Calls: `fenceWorkspace`, `newWorkspaceCleanup`, `readWorkerRecord`, `s.Close`, `s.Start`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `validWorkerFenceAck`, `workerFenceAcknowledgement`, `workspaceFence`, `workspaceFence(s).record`, `workspacePair`.
- **`TestWorkspaceCleanupWaitsForCreatorRelease`** — Regression contract: TestWorkspaceCleanupWaitsForCreatorRelease.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithTimeout`, `fenceWorkspace`, `newWorkspaceCleanup`, `s.Close`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `workspacePair`.
- **`TestWorkspaceCancellationCannotDowngrade`** — Regression contract: TestWorkspaceCancellationCannotDowngrade.
  Signature: `func(t *testing.T)`.
  Calls: `fenceWorkerJobContainers`, `readWorkerRecord`, `s.Close`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `workspaceFence`, `workspacePair`.
- **`TestWorkspaceManagedRegistrationAndRecoveryBinding`** — Regression contract: TestWorkspaceManagedRegistrationAndRecoveryBinding.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `ctx.Value`, `db.seed`, `managedJobFence`, `mockServer`, `registration.Register`, `s.load`, `s.managedExecutionJobContextVersion`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `validManagedJob`.
- **`TestWorkspaceCrashAfterCreateDiscoversOriginalVolume`** — Regression contract: TestWorkspaceCrashAfterCreateDiscoversOriginalVolume.
  Signature: `func(t *testing.T)`.
  Calls: `fenceWorkspace`, `newWorkspaceCleanup`, `s.Close`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `workspacePair`.
- **`TestWorkspaceCancellationBeforeCreationRefusesUnknownData`** — Regression contract: TestWorkspaceCancellationBeforeCreationRefusesUnknownData.
  Signature: `func(t *testing.T)`.
  Calls: `fenceWorkspace`, `newWorkspaceCleanup`, `s.Close`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workspacePair`.
- **`TestWorkspaceOwnerRecoveryKeepsVersionAndProof`** — Regression contract: TestWorkspaceOwnerRecoveryKeepsVersionAndProof.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `errors.New`, `managedJobFence`, `managedJobFence(job).record`, `managedPairFixtureVersion`, `mockRequest`, `response.Body.String`, `s.load`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Run`, `t.Setenv`.
- **Type `containerOnlyAdapter`**: Deliberately hides workspace methods to test mixed-version adapters.
  Declaration: `struct{ workerPairLaunchAdapter }`.
- **`TestLegacyPairRecordsAreRecoveryOnly`** — Regression contract: TestLegacyPairRecordsAreRecoveryOnly.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `len`, `pairRequest`, `s.Close`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestWorkspaceRetainedWithoutExplicitDiscard`** — Regression contract: TestWorkspaceRetainedWithoutExplicitDiscard.
  Signature: `func(t *testing.T)`.
  Calls: `fenceWorkerJobContainers`, `newWorkspaceCleanup`, `s.Close`, `s.Start`, `stopWorkerPair`, `t.Context`, `t.Fatal`, `t.Parallel`, `workspaceFence`, `workspacePair`.
