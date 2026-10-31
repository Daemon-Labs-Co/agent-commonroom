# internal/room/task_repository_links.go

[Module index](internal.md) · [Open source](../../../internal/room/task_repository_links.go)

Go implementation for task repository links; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `TaskRepositoryLink`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Kind string 'json:"kind"' URL string 'json:"url"' }`.
- **Type `TaskRepositoryLinks`**: These are human-maintained references, never repository or execution authority.
  Declaration: `struct { Version int 'json:"version"' CompanyID string 'json:"companyId"' RoomID string 'json:"roomId"' TaskID string 'json:"taskId"' TaskRevision int 'json:"taskRevision"' Revision int 'json:"revision"' Links []TaskRepositoryLink 'json:"links"' ChangedBy string 'json:"changedBy,omitempty"' ChangedAt time.Time 'json:"changedAt,omitempty"' }`.
