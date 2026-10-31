# internal/toolbroker/workspace_volume_cleanup.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/workspace_volume_cleanup.go)

Go implementation for workspace volume cleanup; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`RemoveWorkspaceVolume`** — RemoveWorkspaceVolume is explicit ephemeral-workspace teardown, not retention policy or audit deletion. The caller MUST hold the original creator lease and prevent other managed creation until this returns. Only an exact owner-approved disposal workflow may set discard; ordinary cancellation retains data. Docker supplies no immutable volume-removal ID: creation nonce/time distinguish replacements, but an administrator with daemon access can still race the name. Such administrators and the daemon remain trusted. Models get neither access. Source changes CAN skip this contract, BUT invalidate safe cleanup claims.
  Signature: `func(ctx context.Context, runner DockerRunner, job, repository, dockerRoot string, identity *WorkspaceVolumeIdentity, discard bool) (*WorkspaceVolumeIdentity, error)`.
  Calls: `ValidWorkspaceVolume`, `capture`, `ctx.Err`, `decodeInspection`, `digest`, `e.capture`, `inventory`, `path.Clean`, `path.IsAbs`, `string`, `strings.TrimSpace`, `validObservedWorkspaceVolume`, `verify`.
