# cmd/server/storage_quota_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/storage_quota_test.go)

Automated regression checks for storage_quota; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `storageInventoryFunc`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, string) (managedStorageInventory, error)`.
- **`(storageInventoryFunc).Inventory`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, s string) (managedStorageInventory, error)`.
  Calls: `f`.
- **`storageRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i int, bytes int64, category string) room.StorageReservation`.
  Calls: `fmt.Sprintf`, `quotaJob`.
- **`storageFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, initial ...room.StorageReservation) (*server, *mockDB, context.Context)`.
  Calls: `context.WithValue`, `db.seed`, `mockServer`, `quotaJob`, `s.initializeManagedStorage`, `storageInventoryFunc`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`TestStorageReservationExactBoundaryAndRetainedCategories`** — Regression contract: TestStorageReservationExactBoundaryAndRetainedCategories.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `db.seed`, `db.snapshot`, `len`, `managedStorageTotal`, `s.reserveManagedStorage`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStorageReservationsSurviveContainerShutdown`** — Regression contract: TestStorageReservationsSurviveContainerShutdown.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `managedStorageTotal`, `quotaJob`, `s.load`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`.
- **`TestStorageUnknownInventoryAndInvalidAmountsRefuseWrites`** — Regression contract: TestStorageUnknownInventoryAndInvalidAmountsRefuseWrites.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedStorage`, `context.WithValue`, `db.seed`, `db.snapshot`, `errors.New`, `len`, `mockServer`, `quotaJob`, `s.initializeManagedStorage`, `s.reserveManagedStorage`, `storageInventoryFunc`, `storageRecord`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestStorageLastByteIsSerializedAcrossControllers`** — Regression contract: TestStorageLastByteIsSerializedAcrossControllers.
  Signature: `func(t *testing.T)`.
  Calls: `c.reserveManagedStorage`, `close`, `make`, `managedStorageTotal`, `s.load`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestStorageStaleRevisionCannotCommitLastByte`** — Regression contract: TestStorageStaleRevisionCannotCommitLastByte.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedStorage`, `context.WithValue`, `other.append`, `s.append`, `s.load`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`.
- **`TestStoragePermissionAndBindingRefusals`** — Regression contract: TestStoragePermissionAndBindingRefusals.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `context.WithValue`, `db.seed`, `db.snapshot`, `job.AllocatedAt.Add`, `job.Deadline.Add`, `len`, `mockServer`, `newDB.seed`, `s.load`, `s.reserveManagedStorage`, `storageFixture`, `storageRecord`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStorageProjectionRejectsCorruptOrRewrittenEvidence`** — Regression contract: TestStorageProjectionRejectsCorruptOrRewrittenEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `json.Marshal`, `projectManagedStorage`, `s.load`, `storageFixture`, `storageRecord`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStorageAppendFailureDoesNotGrantCapacity`** — Regression contract: TestStorageAppendFailureDoesNotGrantCapacity.
  Signature: `func(t *testing.T)`.
  Calls: `managedStorageTotal`, `s.load`, `s.reserveManagedStorage`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStorageInventoryCannotBeResetOrProvidedWithoutOwner`** — Regression contract: TestStorageInventoryCannotBeResetOrProvidedWithoutOwner.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `db.snapshot`, `len`, `mockServer`, `s.initializeManagedStorage`, `storageFixture`, `storageInventoryFunc`, `storageRecord`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStorageEventsCannotBeInjectedThroughPublicEvents`** — Regression contract: TestStorageEventsCannotBeInjectedThroughPublicEvents.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `fmt.Sprintf`, `len`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Parallel`.
- **`TestStorageInitiatingHumanRetainsRoomAndJobPermissionChecks`** — Regression contract: TestStorageInitiatingHumanRetainsRoomAndJobPermissionChecks.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `db.seed`, `s.reserveManagedStorage`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestStorageOverLimitInventoryPreservesDataAndBlocksGrowth`** — Regression contract: TestStorageOverLimitInventoryPreservesDataAndBlocksGrowth.
  Signature: `func(t *testing.T)`.
  Calls: `managedStorageTotal`, `s.load`, `s.reserveManagedStorage`, `storageFixture`, `storageRecord`, `t.Fatal`, `t.Parallel`.
