# cmd/server/p0_storage_boundaries_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_storage_boundaries_test.go)

Automated regression checks for p0_storage_boundaries; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`observeStorageReservationEncoding`** — This exercises the real internal journal decoder, not a new public numeric API.
  Signature: `func(t *testing.T, encoded string) pendingP0Observation`.
  Calls: `append`, `bytes.Equal`, `bytes.Replace`, `db.snapshot`, `errors.Is`, `json.Marshal`, `projectManagedStorage`, `reflect.DeepEqual`, `s.load`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Helper`.
- **`driveP0StorageOverflow`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `observeStorageReservationEncoding`, `t.Helper`.
- **`TestStorageReservationDecoderRejectsUnrepresentableEvidence`** — Regression contract: TestStorageReservationDecoderRejectsUnrepresentableEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `observeStorageReservationEncoding`, `reflect.DeepEqual`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStorageOverflowEvidenceStopsJournalReplay`** — Regression contract: TestStorageOverflowEvidenceStopsJournalReplay.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Replace`, `db.seed`, `db.snapshot`, `json.Marshal`, `json.RawMessage`, `reflect.DeepEqual`, `s.load`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`.
- **`storageAccountBoundaryFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, company string, used int64) (*server, *mockDB, context.Context)`.
  Calls: `append`, `context.WithValue`, `db.seed`, `mockServer`, `quotaJob`, `s.initializeManagedStorage`, `storageInventoryFunc`, `storageRecord`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`driveP0StorageCrossAccount`** — Separate installation journals model independent accounts. This does not establish shared-database multi-tenancy or physical cross-worker quotas.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `(&server{db: account.db}).load`, `append`, `b.reserveManagedStorage`, `dbA.snapshot`, `dbB.snapshot`, `len`, `managedStorageTotal`, `reflect.DeepEqual`, `storageAccountBoundaryFixture`, `storageRecord`, `strconv.FormatInt`, `strconv.ParseInt`, `t.Fatal`, `t.Helper`.
- **`TestStorageReservationCannotBorrowAnotherAccountIdentity`** — Regression contract: TestStorageReservationCannotBorrowAnotherAccountIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `errors.Is`, `reflect.DeepEqual`, `s.reserveManagedStorage`, `storageAccountBoundaryFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestRetainedWorkspaceReconciliationKeepsStorageAndContainerReservations`** — Regression contract: TestRetainedWorkspaceReconciliationKeepsStorageAndContainerReservations.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `admitManagedContainerQuota`, `context.WithValue`, `db.seed`, `db.snapshot`, `fenceWorkerJobContainers`, `len`, `managedJobFence`, `managedStorageTotal`, `mockServer`, `newWorkspaceCleanup`, `pairContainerName`, `quotaJob`, `readWorkerRecord`, `reflect.DeepEqual`, `s.initializeManagedStorage`, `s.reconcileManagedJob`, `slices.Equal`, `stopWorkerPair`, `storageInventoryFunc`, `storageRecord`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `worker.Close`, `worker.Start`, `workspacePair`.
