# cmd/server/worker_workspace_probe_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_workspace_probe_test.go)

Automated regression checks for worker_workspace_probe; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `assembledProbeFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T dir string record workerJobRecord base *workspaceCleanupDocker mode string id, name, image, challenge string exists, started, removed, stopped bool creates int events []workerProbeEvent afterFinishObservations int cancel context.CancelFunc durable func(string) bool }`.
- **`(*assembledProbeFixture).checkLock`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `f.t.Fatal`, `f.t.Helper`, `lease.Close`, `workerTryLease`.
- **`(*assembledProbeFixture).AuthorizeWorkspaceObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord) error`.
  Calls: `ctx.Err`, `errors.New`, `f.checkLock`, `f.t.Fatal`, `len`.
- **`(*assembledProbeFixture).Authorize`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error`.
  Calls: `ctx.Err`, `errors.New`, `f.checkLock`, `f.t.Fatal`, `len`.
- **`(*assembledProbeFixture).Acknowledge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `append`, `ctx.Err`, `errors.New`, `f.checkLock`, `f.t.Fatal`, `len`, `probeRecordName`, `readWorkerProbeRecords`.
- **`(*assembledProbeFixture).StopAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`(*assembledProbeFixture).inspection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]any`.
- **`(*assembledProbeFixture).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `ctx.Err`, `errors.New`, `f.base.Run`, `f.cancel`, `f.checkLock`, `f.durable`, `f.inspection`, `f.t.Fatal`, `io.WriteString`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `pairContainerName`, `slices.Contains`, `slices.Equal`, `strings.HasPrefix`, `strings.Repeat`.
- **`newAssembledProbeFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mode string) *assembledProbeFixture`.
  Calls: `os.Chmod`, `retainedObservationFixture`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`TestGuardedWorkspaceProbeIntegratedLifecycle`** — Regression contract: TestGuardedWorkspaceProbeIntegratedLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `lease.Close`, `len`, `newAssembledProbeFixture`, `probeRetainedWorkerWorkspace`, `readWorkerProbeRecords`, `strings.HasPrefix`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `workerProbeStopped`, `workerTryLease`, `workspacemount.ValidProof`.
- **`TestGuardedWorkspaceProbeBusyAndInvalidDependencies`** — Regression contract: TestGuardedWorkspaceProbeBusyAndInvalidDependencies.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `lease.Close`, `newAssembledProbeFixture`, `os.IsNotExist`, `os.Stat`, `probeRetainedWorkerWorkspace`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerTryLease`.
- **`TestOrdinaryWorkspaceObservationStillRejectsProbeReference`** — Regression contract: TestOrdinaryWorkspaceObservationStillRejectsProbeReference.
  Signature: `func(t *testing.T)`.
  Calls: `l.Check`, `newAssembledProbeFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `withWorkerWorkspaceObservation`.
