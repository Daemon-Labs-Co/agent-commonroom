# cmd/server/workspace_capture_ledger_evidence_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_capture_ledger_evidence_test.go)

Automated regression checks for workspace_capture_ledger_evidence; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticLifecycleWitness`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls int deny bool revoke bool }`.
- **`(*syntheticLifecycleWitness).ValidateCaptureLifecycle`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ room.State, _ workerJobRecord, _ workerCaptureRecord, _ workerProbeEvent) error`.
  Calls: `ctx.Err`.
- **`TestWorkspaceCaptureLedgerPhysicalStages`** — Regression contract: TestWorkspaceCaptureLedgerPhysicalStages.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `cancel`, `captureControllerEvents`, `captureJournalFixtureScope`, `captureRecordName`, `context.WithCancel`, `copyCaptureIntent`, `e.ValidateCaptureTransition`, `filepath.Join`, `j.BeginCaptureRecovery`, `j.BindCapture`, `j.FinishCapture`, `len`, `newWorkspaceCaptureLedgerEvidence`, `os.WriteFile`, `physicalControllerFixture`, `probeControllerState`, `readWorkerCaptureRecords`, `releasedProbeEvents`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkspaceCaptureLedgerMissingScope`** — Regression contract: TestWorkspaceCaptureLedgerMissingScope.
  Signature: `func(t *testing.T)`.
  Calls: `missing.ValidateCaptureTransition`, `newWorkspaceCaptureLedgerEvidence`, `t.Context`, `t.Fatal`.
