# internal/toolbroker/workspace_volume.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/workspace_volume.go)

Go implementation for workspace volume; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `WorkspaceVolumeIdentity`**: WorkspaceVolumeIdentity is non-secret creation evidence, not authority. Named Docker volumes have no immutable deletion ID; nonce and daemon creation time detect replacement, while the worker's exclusive creator lease prevents other managed code from replacing the name during verification/removal.
  Declaration: `struct { Version int 'json:"version"' Name string 'json:"name"' Nonce string 'json:"nonce"' CreatedAt string 'json:"createdAt,omitempty"' }`.
- **`ValidWorkspaceVolume`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(job string, v WorkspaceVolumeIdentity, created bool) bool`.
  Calls: `digest`, `len`, `time.Parse`, `timestamp.IsZero`.
- **`(*ToolContainerPlan).WorkspaceIntent`** — WorkspaceIntent mints the identity that MUST be durably journalled before PrepareVolume. It accepts no model names, paths, mounts, labels or driver flags.
  Signature: `func(now time.Time) (WorkspaceVolumeIdentity, error)`.
  Calls: `hex.EncodeToString`, `p.RunArgs`, `p.WorkspaceIntent`, `p.WorkspaceVolume`, `p.volume.CompareAndSwap`, `p.volume.Load`, `rand.Read`.
- **`(*ToolContainerPlan).volumeLabels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(identity WorkspaceVolumeIdentity) map[string]string`.
  Calls: `p.labels`.
- **`(*ToolContainerPlan).verifyVolumeIdentity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v toolVolumeInspection, identity WorkspaceVolumeIdentity, root string) error`.
  Calls: `ValidWorkspaceVolume`, `p.verifyVolumeShape`, `p.volumeLabels`.
- **`(*ToolContainerRuntime).PrepareVolume`** — PrepareVolume refuses existing volumes and attempts creation once. The worker must hold creator/admission ownership and journal intent before entering here. Any daemon/receipt/inspection failure is uncertain; it never adopts, deletes or retries the name. No automatic reconciliation or host directory mount exists.
  Signature: `func(ctx context.Context, intent WorkspaceVolumeIdentity) (WorkspaceVolumeIdentity, error)`.
  Calls: `ValidWorkspaceVolume`, `cancel`, `child.Err`, `decodeInspection`, `p.VolumeCreateArgs`, `p.verifyVolumeIdentity`, `p.volume.Load`, `p.volume.Store`, `r.context`, `r.executor.capture`, `r.lifetime.Err`, `r.mu.Lock`, `r.mu.Unlock`, `string`, `strings.TrimSpace`, `time.Now`.
