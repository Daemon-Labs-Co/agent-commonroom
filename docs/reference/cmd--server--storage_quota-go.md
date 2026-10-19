# cmd/server/storage_quota.go

[Module index](cmd.md) · [Open source](../../cmd/server/storage_quota.go)

Go implementation for storage quota; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `managedStorageInventory`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' CompanyID string 'json:"companyId"' Digest string 'json:"digest"' Reservations []room.StorageReservation 'json:"reservations"' }`.
- **Type `managedStorageInventorySource`**: This trusted adapter must freeze writers and inventory ALL retained volumes, staged bundles and artifacts across the account's approved workers. It cannot use model claims, polling samples or a partial worker list as complete evidence. No production adapter exists yet: unknown inventory must block reservations. Source forks CAN supply a dishonest adapter, invalidating accounting/isolation.
  Declaration: `interface { Inventory(context.Context, string) (managedStorageInventory, error) }`.
- **`validStorageReservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r room.StorageReservation) bool`.
  Calls: `repositoryDigestPattern.MatchString`, `validStorageWorkspace`, `validWorkerJobID`, `workerAliasPattern.MatchString`.
- **`managedStorageTotal`** — Storage totals charge reserved capacity, including retained/uncertain data, without trusting current process/container status. Never sum unchecked int64s.
  Signature: `func(st room.State) (int64, error)`.
  Calls: `repositoryDigestPattern.MatchString`, `validStorageRelease`, `validStorageReservation`, `workerAliasPattern.MatchString`.
- **`admitManagedStorage`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, r room.StorageReservation) (int64, error)`.
  Calls: `managedStorageTotal`, `validStorageReservation`.
- **`(*server).initializeManagedStorage`** — initializeManagedStorage records one complete trusted inventory before capacity may be reserved. It is internal infrastructure, NOT an HTTP/model operation or a way to reset usage. Over-limit existing data is preserved and blocks growth. No resources are created here, and no physical quota guarantee is implied.
  Signature: `func(ctx context.Context, company string, source managedStorageInventorySource) error`.
  Calls: `cancel`, `context.WithValue`, `ctx.Err`, `ctx.Value`, `json.Marshal`, `member`, `projectManagedStorage`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.storageContext`, `source.Inventory`.
- **`(*server).reserveManagedStorage`** — reserveManagedStorage commits capacity before any external allocation. The expected journal revision is checked under the transactional audit-chain lock, so separate controllers cannot both reserve the last byte from stale state. Returning success is not permission to write without a physical quota backend. No refund/retry is inferred after timeout, failed append or container removal.
  Signature: `func(ctx context.Context, r room.StorageReservation) (int64, error)`.
  Calls: `admitManagedStorage`, `canAccess`, `cancel`, `context.WithValue`, `ctx.Err`, `ctx.Value`, `member`, `s.agentsBlockedLocked`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.storageContext`, `time.Now`, `time.Now().Before`, `validManagedJob`.
- **`projectManagedStorage`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, kind string, raw []byte) error`.
  Calls: `admitManagedStorage`, `json.Unmarshal`, `managedStorageTotal`, `repositoryDigestPattern.MatchString`, `validStorageRelease`.
