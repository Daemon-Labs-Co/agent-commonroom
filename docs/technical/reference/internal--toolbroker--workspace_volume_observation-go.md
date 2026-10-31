# internal/toolbroker/workspace_volume_observation.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/workspace_volume_observation.go)

Go implementation for workspace volume observation; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`VerifyRetainedWorkspace`** — VerifyRetainedWorkspace observes an exact retained volume with no container references. The caller MUST hold its creator job lease and prevent all managed writers for the entire observation/use interval. runner MUST enforce the same approved rootless daemon on every call. Daemon administrators remain trusted. This is NOT a read-only mount attestation, source root or FrozenLease. It never mounts, executes, adopts, deletes or refunds data. Source forks CAN omit these checks, BUT then cannot claim managed writer fencing. No host fallback exists.
  Signature: `func(ctx context.Context, runner DockerRunner, job, repository, dockerRoot string, identity WorkspaceVolumeIdentity) error`.
  Calls: `ValidWorkspaceVolume`, `absentWorkspaceRunner`, `capture`, `ctx.Err`, `decodeInspection`, `digest`, `e.capture`, `path.Clean`, `path.IsAbs`, `string`, `strings.TrimSpace`, `validObservedWorkspaceVolume`.
- **`absentWorkspaceRunner`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(runner DockerRunner) bool`.
  Calls: `r.IsNil`, `r.Kind`, `reflect.ValueOf`.
- **`validObservedWorkspaceVolume`** — Shared with disposal: observation and removal must bind the same immutable creator metadata even though only disposal may authorize deletion.
  Signature: `func(v toolVolumeInspection, job, repository, dockerRoot string, identity WorkspaceVolumeIdentity) bool`.
  Calls: `ValidWorkspaceVolume`, `len`, `path.Join`, `sameMap`.
