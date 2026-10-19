# internal/room/storage.go

[Module index](internal.md) · [Open source](../../internal/room/storage.go)

Go implementation for storage; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `StorageReservation`**: StorageReservation is durable capacity ownership, not a measurement or a grant to write. Bytes remain charged after execution/container shutdown. There is no automatic release on job shutdown. Exact approved disposal is required. Paths, credentials, model output and artifact contents never belong here.
  Declaration: `struct { Version int 'json:"version"' ID string 'json:"id"' CompanyID string 'json:"companyId"' JobID string 'json:"jobId"' RoomID string 'json:"roomId"' Worker string 'json:"worker"' RepositoryDigest string 'json:"repositoryDigest"' Category string 'json:"category"' Bytes int64 'json:"bytes"' WorkspaceName string 'json:"workspaceName,omitempty"' WorkspaceNonce string 'json:"workspaceNonce,omitempty"' WorkspaceCreatedAt string 'json:"workspaceCreatedAt,omitempty"' }`.
- **Type `ManagedStorage`**: ManagedStorage is internal replayed accounting. A nil ledger is unknown usage, not an empty account. InventoryDigest binds the trusted initial inventory; it is not a current-filesystem attestation or kernel quota certificate.
  Declaration: `struct { Version int 'json:"version"' CompanyID string 'json:"companyId"' InventoryDigest string 'json:"inventoryDigest"' Reservations map[string]StorageReservation 'json:"reservations"' Releases map[string]StorageRelease 'json:"releases,omitempty"' }`.
- **Type `StorageRelease`**: StorageRelease retains the original charge and links its single refund to a verified, owner-approved disposal. It never erases reservation or audit history.
  Declaration: `struct { Version int 'json:"version"' ReservationID string 'json:"reservationId"' DisposalID string 'json:"disposalId"' }`.
