# cmd/server/worker_workspace.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_workspace.go)

Go implementation for worker workspace; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerCleanupOwnershipKey`**: workerCleanupOwnershipKey is set only after acquiring the creator job lease. Killing a creator's containers while its lease is held does not authorize erasing its workspace or confirming completed cleanup.
  Declaration: `struct{}`.
- **Type `workerWorkspaceDiscardKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(*workerPairSupervisor).prepareWorkspace`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, adapter workerWorkspaceLaunchAdapter) error`.
  Calls: `adapter.PrepareVolume`, `adapter.WorkspaceIntent`, `errors.New`, `s.available`, `s.mu.Lock`, `s.mu.Unlock`, `s.publishWorkspace`, `toolbroker.ValidWorkspaceVolume`.
- **`(*workerPairSupervisor).publishWorkspace`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, workspace *workerPairWorkspace, bind bool) error`.
  Calls: `errors.New`, `s.available`, `s.cancel`, `s.mu.Lock`, `s.mu.Unlock`, `time.Now`, `time.Now().UTC`, `validWorkerRecord`, `writeWorkerRecord`.
- **`stopWorkerWorkspace`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, v workerJobRecord, runner toolbroker.DockerRunner) (workerJobRecord, error)`.
  Calls: `ctx.Value`, `errors.New`, `filepath.Clean`, `filepath.IsAbs`, `pairCapture`, `toolbroker.RemoveWorkspaceVolume`, `toolbroker.ValidWorkspaceVolume`.
- **`sameWorkspaceEvidence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(before, after workerJobRecord) bool`.
