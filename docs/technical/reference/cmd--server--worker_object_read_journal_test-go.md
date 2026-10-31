# cmd/server/worker_object_read_journal_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_object_read_journal_test.go)

Automated regression checks for worker_object_read_journal; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadJournalGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { f *assembledProbeFixture deny, failAck string stopped bool seen []string }`.
- **`(*objectReadJournalGuard).AuthorizeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord, i workerObjectReadIntent, e workerProbeEvent) error`.
  Calls: `ctx.Err`, `g.f.checkLock`.
- **`(*objectReadJournalGuard).AcknowledgeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerObjectReadRecord, e workerProbeEvent) error`.
  Calls: `append`, `ctx.Err`, `errors.New`, `g.f.checkLock`, `g.f.t.Fatal`, `len`, `objectReadRecordName`, `readWorkerObjectReadRecords`.
- **`(*objectReadJournalGuard).StopObjectReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`objectReadJournalFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*workerObjectReadJournal, *objectReadJournalGuard, *workerWorkspaceObservation, *workerCaptureJournal)`.
  Calls: `captureJournalFixtureScope`, `newWorkerObjectReadJournal`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`beginObjectReadJournal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, j *workerObjectReadJournal)`.
  Calls: `j.BeginObjectRead`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`TestWorkerObjectReadJournalDurableLifecycle`** — Regression contract: TestWorkerObjectReadJournalDurableLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `beginObjectReadJournal`, `boundedWorkerRecords`, `capture.BeginCapture`, `j.BeginObjectRead`, `j.BindObjectRead`, `j.FinishObjectRead`, `len`, `objectReadJournalFixture`, `objectReadRecordName`, `objectReadRecordSettled`, `readWorkerObjectReadRecords`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerObjectReadSlots`, `workerProbeCapacity`.
- **`TestWorkerObjectReadJournalKnownRecoveryAndUnknownRetention`** — Regression contract: TestWorkerObjectReadJournalKnownRecoveryAndUnknownRetention.
  Signature: `func(t *testing.T)`.
  Calls: `beginObjectReadJournal`, `j.BeginObjectReadRecovery`, `j.BindObjectRead`, `j.FinishObjectRead`, `j.FinishObjectReadRecovery`, `len`, `newWorkerObjectReadJournal`, `objectReadJournalFixture`, `objectReadRecordName`, `readWorkerObjectReadRecords`, `retry.BeginObjectReadRecovery`, `retry.FinishObjectReadRecovery`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerObjectReadSlots`.
- **`TestWorkerObjectReadJournalFailuresRetainCapacity`** — Regression contract: TestWorkerObjectReadJournalFailuresRetainCapacity.
  Signature: `func(t *testing.T)`.
  Calls: `beginObjectReadJournal`, `cancel`, `context.WithCancel`, `errors.New`, `j.BeginObjectRead`, `j.BindObjectRead`, `j.FinishObjectRead`, `len`, `objectReadJournalFixture`, `objectReadRecordName`, `persist`, `readWorkerObjectReadRecords`, `strings.Contains`, `strings.Join`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerObjectReadSlots`, `workerProbeStopped`.
- **`TestWorkerObjectReadJournalRefusesSubstitutionAndConflicts`** — Regression contract: TestWorkerObjectReadJournalRefusesSubstitutionAndConflicts.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `beginObjectReadJournal`, `cancel`, `context.WithCancel`, `filepath.Join`, `j.BeginObjectRead`, `j.BindObjectRead`, `j.FinishObjectRead`, `objectReadJournalFixture`, `os.WriteFile`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkerObjectReadJournalCannotEscapeCreatorLease`** — Regression contract: TestWorkerObjectReadJournalCannotEscapeCreatorLease.
  Signature: `func(t *testing.T)`.
  Calls: `beginObjectReadJournal`, `j.FinishObjectRead`, `l.mu.Lock`, `l.mu.Unlock`, `len`, `objectReadJournalFixture`, `objectReadRecordName`, `readWorkerObjectReadRecords`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerObjectReadJournalRequiresFreshAttemptAndPinnedLease`** — Regression contract: TestWorkerObjectReadJournalRequiresFreshAttemptAndPinnedLease.
  Signature: `func(t *testing.T)`.
  Calls: `another.BeginObjectRead`, `beginObjectReadJournal`, `j.BindObjectRead`, `j.FinishObjectRead`, `newWorkerObjectReadJournal`, `objectReadJournalFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
