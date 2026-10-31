# cmd/server/worker_probe_journal_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_probe_journal_test.go)

Automated regression checks for worker_probe_journal; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticProbeGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { deny, stopped bool failStage string events []workerProbeEvent revokeStage string }`.
- **`(*syntheticProbeGuard).Authorize`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ toolbroker.WorkspaceProbeIntent) error`.
  Calls: `ctx.Err`, `errors.New`.
- **`(*syntheticProbeGuard).Acknowledge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `append`, `ctx.Err`, `errors.New`, `validWorkerProbeIntent`.
- **`(*syntheticProbeGuard).StopAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`probeJournalFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*workerProbeJournal, toolbroker.WorkspaceProbeIntent, *syntheticProbeGuard)`.
  Calls: `newWorkerProbeJournal`, `os.Chmod`, `retainedObservationFixture`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`TestProbeJournalDurableRestartAndExactBinding`** — Regression contract: TestProbeJournalDurableRestartAndExactBinding.
  Signature: `func(t *testing.T)`.
  Calls: `j.Begin`, `len`, `newWorkerProbeJournal`, `probeJournalFixture`, `probeRecordName`, `readWorkerProbeRecords`, `restarted.Begin`, `restarted.Bound`, `restarted.Finish`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestProbeJournalRetainsUncertaintyAndRevocation`** — Regression contract: TestProbeJournalRetainsUncertaintyAndRevocation.
  Signature: `func(t *testing.T)`.
  Calls: `errors.New`, `j.Begin`, `j.Bound`, `j.Finish`, `len`, `probeJournalFixture`, `probeRecordName`, `readWorkerProbeRecords`, `replacePrivateDurably`, `string`, `strings.Contains`, `strings.HasPrefix`, `strings.HasSuffix`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerProbeStopped`.
- **`TestProbeJournalPrivateReplayAndCorruption`** — Regression contract: TestProbeJournalPrivateReplayAndCorruption.
  Signature: `func(t *testing.T)`.
  Calls: `boundedWorkerRecords`, `filepath.Join`, `j.Begin`, `json.Marshal`, `json.Unmarshal`, `len`, `os.Chmod`, `os.Link`, `os.ReadFile`, `os.Rename`, `os.Symlink`, `os.WriteFile`, `probeJournalFixture`, `probeRecordName`, `readWorkerProbeRecords`, `string`, `strings.Repeat`, `strings.Replace`, `strings.ReplaceAll`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workerProbeCapacity`.
- **`TestProbeJournalSerializesCompetingInstances`** — Regression contract: TestProbeJournalSerializesCompetingInstances.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `j.Begin`, `make`, `newWorkerProbeJournal`, `probeJournalFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestProbeJournalStopsNormalAdmission`** — Regression contract: TestProbeJournalStopsNormalAdmission.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerRecord`, `fakeWorkerRecord`, `j.Begin`, `leases.Close`, `probeJournalFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestProbeJournalRequiresPrivateDependencies`** — Regression contract: TestProbeJournalRequiresPrivateDependencies.
  Signature: `func(t *testing.T)`.
  Calls: `newWorkerProbeJournal`, `os.Chmod`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestProbeJournalCapacityIsRetainedUntilConfirmed`** — Regression contract: TestProbeJournalCapacityIsRetainedUntilConfirmed.
  Signature: `func(t *testing.T)`.
  Calls: `boundedWorkerRecords`, `j.Begin`, `j.Finish`, `probeJournalFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerProbeCapacity`.
- **`TestProbeJournalAmbiguousDurabilityStopsRestart`** — Regression contract: TestProbeJournalAmbiguousDurabilityStopsRestart.
  Signature: `func(t *testing.T)`.
  Calls: `errors.New`, `j.Begin`, `len`, `newWorkerProbeJournal`, `probeJournalFixture`, `readWorkerProbeRecords`, `replacePrivateDurably`, `restarted.Begin`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestProbeJournalCurrentAuthorityAtEveryBoundary`** — Regression contract: TestProbeJournalCurrentAuthorityAtEveryBoundary.
  Signature: `func(t *testing.T)`.
  Calls: `j.Begin`, `j.Bound`, `j.Finish`, `len`, `probeJournalFixture`, `probeRecordName`, `readWorkerProbeRecords`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestProbeJournalRevocationDuringConfirmationWithholdsAcknowledgment`** — Regression contract: TestProbeJournalRevocationDuringConfirmationWithholdsAcknowledgment.
  Signature: `func(t *testing.T)`.
  Calls: `boundedWorkerRecords`, `j.Begin`, `j.Bound`, `j.Finish`, `len`, `probeJournalFixture`, `probeRecordName`, `readWorkerProbeRecords`, `replacePrivateDurably`, `string`, `strings.Contains`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerProbeCapacity`.
