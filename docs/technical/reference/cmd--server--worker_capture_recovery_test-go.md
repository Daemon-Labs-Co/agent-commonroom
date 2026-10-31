# cmd/server/worker_capture_recovery_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_recovery_test.go)

Automated regression checks for worker_capture_recovery; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRetainedCaptureRecoveryUsesDurableKnownIDAndNoStarter`** — Regression contract: TestRetainedCaptureRecoveryUsesDurableKnownIDAndNoStarter.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `cancel`, `captureRecordName`, `captureRecordSettled`, `context.WithCancel`, `errors.Is`, `j.BindCapture`, `j.FinishCapture`, `ownedCaptureSetup`, `readWorkerCaptureRecords`, `recoverRetainedWorkerCapture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerCaptureSlots`.
