# cmd/server/container_quota.go

[Module index](cmd.md) · [Open source](../../../cmd/server/container_quota.go)

Go implementation for container quota; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`managedContainerSlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(job room.ManagedJob) int`.
  Calls: `pairedLifecycleVersion`.
- **`admitManagedContainerQuota`** — admitManagedContainerQuota is evaluated against the replayed account journal before allocation. The allocation append MUST use that state's EventCount as an expected revision: a process-local lock alone cannot serialize controllers. Every unresolved allocation reserves its full size, independent of worker, room, container inventory or deadline. Only durable shutdown proof releases it. Legacy records with no company attribution conservatively count for this account; malformed evidence is never permission to allocate more containers.
  Signature: `func(st room.State, next room.ManagedJob) error`.
  Calls: `controllerHelperSlots`, `errors.New`, `managedContainerSlots`, `validManagedJob`.
