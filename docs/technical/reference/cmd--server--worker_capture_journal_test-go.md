# cmd/server/worker_capture_journal_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_journal_test.go)

Automated regression checks for worker_capture_journal; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureJournalGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { f *assembledProbeFixture deny, failAck string stopped bool seen []string }`.
- **`(*captureJournalGuard).AuthorizeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord, i workerCaptureIntent, e workerProbeEvent) error`.
  Calls: `ctx.Err`, `g.f.checkLock`.
- **`(*captureJournalGuard).AcknowledgeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerCaptureRecord, e workerProbeEvent) error`.
  Calls: `append`, `captureRecordName`, `ctx.Err`, `errors.New`, `g.f.checkLock`, `g.f.t.Fatal`, `len`, `readWorkerCaptureRecords`.
- **`(*captureJournalGuard).StopCaptureAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`(*captureJournalGuard).CaptureAuthorityLifetime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() context.Context`.
  Calls: `g.f.t.Context`.
- **`captureJournalFixtureScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*workerCaptureJournal, *captureJournalGuard, *workerWorkspaceObservation)`.
  Calls: `cancel`, `context.WithCancel`, `lease.Close`, `newAssembledProbeFixture`, `newWorkerCaptureJournal`, `strings.Repeat`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`, `workerTryLease`.
- **`captureJournalFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*workerCaptureJournal, *captureJournalGuard)`.
  Calls: `captureJournalFixtureScope`.
- **`beginCaptureJournal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, j *workerCaptureJournal)`.
  Calls: `j.BeginCapture`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`TestWorkerCaptureJournalDurableLifecycleAndCapacity`** — Regression contract: TestWorkerCaptureJournalDurableLifecycleAndCapacity.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `boundedWorkerRecords`, `captureJournalFixture`, `captureRecordName`, `captureRecordSettled`, `j.BeginCapture`, `j.BindCapture`, `j.FinishCapture`, `len`, `readWorkerCaptureRecords`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerCaptureSlots`, `workerProbeCapacity`.
- **`TestWorkerCaptureJournalKnownRecoveryAndUnknownRetention`** — Regression contract: TestWorkerCaptureJournalKnownRecoveryAndUnknownRetention.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `captureJournalFixture`, `j.BeginCaptureRecovery`, `j.BindCapture`, `j.FinishCapture`, `j.FinishCaptureRecovery`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerCaptureSlots`.
- **`TestWorkerCaptureJournalAmbiguousPersistenceStopsAuthority`** — Regression contract: TestWorkerCaptureJournalAmbiguousPersistenceStopsAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `captureJournalFixture`, `errors.New`, `j.BeginCapture`, `j.BindCapture`, `j.FinishCapture`, `real`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerProbeCapacity`, `workerProbeStopped`.
- **`TestWorkerCaptureJournalRefusesSubstitutionAndUnsafeRecords`** — Regression contract: TestWorkerCaptureJournalRefusesSubstitutionAndUnsafeRecords.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `beginCaptureJournal`, `captureJournalFixture`, `captureRecordName`, `copyCaptureRequest`, `filepath.Join`, `j.BeginCapture`, `j.BindCapture`, `os.Chmod`, `os.ReadFile`, `os.Remove`, `os.Symlink`, `os.WriteFile`, `readWorkerCaptureRecords`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workerProbeCapacity`.
- **`TestWorkerCaptureJournalAuthorityDenialRetainsCharges`** — Regression contract: TestWorkerCaptureJournalAuthorityDenialRetainsCharges.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `captureJournalFixture`, `j.BeginCapture`, `j.BindCapture`, `j.FinishCapture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerCaptureSlots`.
- **`TestWorkerCaptureJournalBlocksSecondUnsettledCapture`** — Regression contract: TestWorkerCaptureJournalBlocksSecondUnsettledCapture.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `captureJournalFixture`, `copyCaptureIntent`, `other.BeginCapture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
