# cmd/server/storage_disposal.go

[Module index](cmd.md) · [Open source](../../cmd/server/storage_disposal.go)

Go implementation for storage disposal; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`validStorageWorkspace`** — A reservation may prebind the worker-generated identity before allocation. This is accounting evidence, never permission to write without a hard quota. Legacy unbound charges cannot be inferred from job IDs or refunded retrospectively.
  Signature: `func(r room.StorageReservation) bool`.
  Calls: `storageWorkspace`, `toolbroker.ValidWorkspaceVolume`.
- **`storageWorkspace`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r room.StorageReservation) toolbroker.WorkspaceVolumeIdentity`.
- **`reservationMatchesDisposal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r room.StorageReservation, p room.WorkspaceDisposal) bool`.
  Calls: `disposalIdentity`, `storageWorkspace`, `validStorageReservation`.
- **`disposalStorageReservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, p room.WorkspaceDisposal) string`.
  Calls: `reservationMatchesDisposal`.
- **`validStorageRelease`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, release room.StorageRelease) bool`.
  Calls: `reservationMatchesDisposal`, `sameDisposalJob`, `validDisposal`, `validWorkerJobID`.
