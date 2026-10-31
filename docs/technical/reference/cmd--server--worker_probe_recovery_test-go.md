# cmd/server/worker_probe_recovery_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_probe_recovery_test.go)

Automated regression checks for worker_probe_recovery; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticRecoveryFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *assembledProbeFixture intent toolbroker.WorkspaceProbeIntent deny bool removals, postRecovered int }`.
- **`(*syntheticRecoveryFixture).AuthorizeWorkspaceObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord) error`.
  Calls: `ctx.Err`, `errors.New`, `f.checkLock`, `len`, `probeRecordName`, `readWorkerProbeRecords`.
- **`(*syntheticRecoveryFixture).AuthorizeRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string) error`.
  Calls: `ctx.Err`, `errors.New`, `f.checkLock`, `f.t.Fatal`, `strings.Repeat`.
- **`(*syntheticRecoveryFixture).Acknowledge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `append`, `ctx.Err`, `errors.New`, `f.checkLock`, `f.t.Fatal`, `probeRecordName`, `readWorkerProbeRecords`.
- **`(*syntheticRecoveryFixture).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `ctx.Err`, `f.assembledProbeFixture.Run`, `f.cancel`, `f.checkLock`, `f.inspection`, `f.t.Fatal`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `probeRecordName`, `readWorkerProbeRecords`, `strings.Repeat`.
- **`recoveryFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mode string) *syntheticRecoveryFixture`.
  Calls: `j.Begin`, `j.Bound`, `j.Finish`, `newAssembledProbeFixture`, `newWorkerProbeJournal`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`TestRecordedProbeRecoveryLifecycle`** — Regression contract: TestRecordedProbeRecoveryLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `boundedWorkerRecords`, `cancel`, `context.WithCancel`, `lease.Close`, `len`, `probeRecordName`, `readWorkerProbeRecords`, `recoverRetainedWorkspaceProbe`, `recoveryFixture`, `strings.HasPrefix`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `workerProbeCapacity`, `workerProbeSettled`, `workerProbeStopped`, `workerTryLease`.
- **`TestRecordedProbeRecoveryExplicitRetryAfterRevocation`** — Regression contract: TestRecordedProbeRecoveryExplicitRetryAfterRevocation.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `probeRecordName`, `readWorkerProbeRecords`, `recoverRetainedWorkspaceProbe`, `recoveryFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerProbeSettled`.
- **`TestRecordedProbeRecoveryResumesCleanupOutcome`** — Regression contract: TestRecordedProbeRecoveryResumesCleanupOutcome.
  Signature: `func(t *testing.T)`.
  Calls: `recoverRetainedWorkspaceProbe`, `recoveryFixture`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestRecordedProbeRecoveryRefusesBusyStoppedAndWrongIntent`** — Regression contract: TestRecordedProbeRecoveryRefusesBusyStoppedAndWrongIntent.
  Signature: `func(t *testing.T)`.
  Calls: `lock.Close`, `recoverRetainedWorkspaceProbe`, `recoveryFixture`, `replacePrivateDurably`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerTryLease`.
