# cmd/server/p0_storage_acceptance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_storage_acceptance_test.go)

Automated regression checks for p0_storage_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0StorageMeteringUnknown`** — Inventory failure must leave no authoritative ledger, including after replay. This exercises controller accounting only; no physical allocation is attempted.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `context.WithValue`, `controller.load`, `controller.reserveManagedStorage`, `ctx.Err`, `db.seed`, `db.snapshot`, `errors.Is`, `errors.New`, `len`, `min`, `mockServer`, `quotaJob`, `reflect.DeepEqual`, `s.initializeManagedStorage`, `storageInventoryFunc`, `storageRecord`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`TestStorageMeteringFailureRequiresExplicitCompleteInventoryForRecovery`** — Regression contract: TestStorageMeteringFailureRequiresExplicitCompleteInventoryForRecovery.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `context.WithValue`, `db.seed`, `db.snapshot`, `errors.Is`, `errors.New`, `managedStorageTotal`, `mockServer`, `quotaJob`, `reflect.DeepEqual`, `restarted.initializeManagedStorage`, `restarted.reserveManagedStorage`, `s.initializeManagedStorage`, `storageInventoryFunc`, `storageRecord`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`driveP0StorageReservation`** — Drives real controller reservation methods and journal replay with the in-memory database adapter. These cases establish accounting, not physical disk quotas. Observe committed events/results; never manufacture observations from Want.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `close`, `controller.reserveManagedStorage`, `db.seed`, `db.snapshot`, `fmt.Sprintf`, `int64`, `len`, `make`, `managedStorageTotal`, `s.load`, `s.reserveManagedStorage`, `storageFixture`, `storageRecord`, `strconv.FormatInt`, `strconv.Itoa`, `strconv.ParseInt`, `t.Fatal`, `t.Helper`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **Type `roomStorageFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { id int bytes int64 category string }`.
