# cmd/server/workspace_capture_session_evidence_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_capture_session_evidence_test.go)

Automated regression checks for workspace_capture_session_evidence; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticCaptureOperationWitness`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { syntheticLifecycleWitness operations int mode string change func() }`.
- **`(*syntheticCaptureOperationWitness).ValidateCaptureOperation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ room.State, r workerJobRecord, i workerCaptureIntent, _ workerProbeEvent) error`.
  Calls: `ctx.Err`, `w.change`.
- **`TestWorkspaceCaptureOperationPhysicalHistory`** — Regression contract: TestWorkspaceCaptureOperationPhysicalHistory.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `beginCaptureJournal`, `cancel`, `captureJournalFixtureScope`, `captureRecordName`, `context.WithCancel`, `detachedProbeWorkerRecord`, `e.ValidateCaptureOperation`, `filepath.Join`, `j.BeginCaptureRecovery`, `j.BindCapture`, `j.FinishCapture`, `json.Marshal`, `newWorkspaceCaptureSessionEvidence`, `os.WriteFile`, `physicalControllerFixture`, `probeControllerState`, `readWorkerCaptureRecords`, `strings.HasPrefix`, `strings.HasSuffix`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkspaceCaptureSessionEvidenceRequiresScope`** — Regression contract: TestWorkspaceCaptureSessionEvidenceRequiresScope.
  Signature: `func(t *testing.T)`.
  Calls: `captureJournalFixtureScope`, `e.ValidateCaptureOperation`, `newWorkspaceCaptureSessionEvidence`, `t.Context`, `t.Fatal`.
