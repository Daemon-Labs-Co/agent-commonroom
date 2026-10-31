# cmd/server/worker_object_read_runtime.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_object_read_runtime.go)

Go implementation for worker object read runtime; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerObjectReadRuntimeGuard`**: workerObjectReadRuntime assembles the mandatory local journal with the fixed isolated reader runtime. A real retained-workspace creator lease, authenticated operation guard, same-daemon starter and independently authorized byte-release guard are all required. None is inferred from model metadata or substituted with host/SSH execution. There is no optional isolation switch or default authority. A source fork can replace these dependencies and invalidate the boundary. The runtime requires continuous authority, not just successful past exchanges.
  Declaration: `interface { workerObjectReadJournalGuard ObjectReadAuthorityLifetime() context.Context }`.
- **Type `workerObjectReadRuntime`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { runtime *toolbroker.ObjectReadRuntime guard workerObjectReadRuntimeGuard creator context.Context owner *workerWorkspaceObservation release objectcustody.ReadReleaseGuard }`.
- **`newWorkerObjectReadRuntime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, i workerObjectReadIntent, g workerObjectReadRuntimeGuard, release objectcustody.ReadReleaseGuard) (*workerObjectReadRuntime, error)`.
  Calls: `g.ObjectReadAuthorityLifetime`, `g.ObjectReadAuthorityLifetime().Err`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `newWorkerObjectReadJournal`, `nilWorkerObservationDependency`, `toolbroker.NewObjectReadContainerPlan`, `toolbroker.NewObjectReadRuntime`.
- **`(*workerObjectReadRuntime).Run`** — The caller owns and must join/close its private control transport on all paths. Runtime errors revoke operation authority as well as the runtime's byte-release authority. Success still requires exit, exact-ID cleanup, both durable journals and current release permission; it never means user delivery or task acceptance.
  Signature: `func(ctx context.Context) (raw []byte, err error)`.
  Calls: `cancel`, `clear`, `complete`, `context.AfterFunc`, `context.WithCancel`, `ctx.Err`, `errors.Join`, `lifetime.Err`, `nilWorkerObservationDependency`, `r.creator.Err`, `r.guard.ObjectReadAuthorityLifetime`, `r.guard.StopObjectReadAuthority`, `r.owner.beginOperation`, `r.owner.operationCurrent`, `r.release.StopReadAuthority`, `r.runtime.Run`, `stop`.
- **`recoverRetainedWorkerObjectRead`** — recoverRetainedWorkerObjectRead assembles cleanup only. It requires fresh scoped authority and a local durable known-ID history; it has no starter, byte-release guard, reader, object deletion, storage refund or unknown-ID adoption capability.
  Signature: `func(ctx context.Context, l *workerWorkspaceObservation, i workerObjectReadIntent, id string, g workerObjectReadRuntimeGuard) (err error)`.
  Calls: `cancel`, `complete`, `context.AfterFunc`, `context.WithCancel`, `creator.Err`, `ctx.Err`, `errors.Join`, `g.ObjectReadAuthorityLifetime`, `g.ObjectReadAuthorityLifetime().Err`, `g.StopObjectReadAuthority`, `l.beginOperation`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `lifetime.Err`, `newWorkerObjectReadJournal`, `nilWorkerObservationDependency`, `recovery.Reconcile`, `stop`, `stopCreator`, `toolbroker.NewObjectReadContainerPlan`, `toolbroker.NewObjectReadRecovery`.
