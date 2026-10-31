# cmd/server/worker_capture_runtime.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_runtime.go)

Go implementation for worker capture runtime; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ownedWorkerCapture`**: ownedWorkerCapture joins the retained creator lease, durable worker lifecycle journal and supervised isolated helper. The runner MUST also own its starter; callers cannot splice in a different daemon or provider/host process stream. The supplied authorities remain mandatory independent trusted infrastructure, not production quota proof by themselves. No public API or fallback exists.
  Declaration: `struct { owner *workerWorkspaceObservation runtime *toolbroker.CaptureRuntime guard workerCaptureRuntimeGuard authority context.Context }`.
- **Type `workerCaptureRuntimeGuard`**: Continuous authenticated authority is required even between protocol calls. A prior successful exchange or still-live creator lease cannot replace it.
  Declaration: `interface { workerCaptureJournalGuard CaptureAuthorityLifetime() context.Context }`.
- **`newOwnedWorkerCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, q workspacecapture.Request, image string, objects toolbroker.CaptureObjectVolume, guard workerCaptureRuntimeGuard, authority workspacecapture.Authority, custody objectcustody.Journal) (*ownedWorkerCapture, error)`.
  Calls: `copyCaptureRequest`, `guard.CaptureAuthorityLifetime`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `lifetime.Err`, `newWorkerCaptureJournal`, `nilWorkerObservationDependency`, `toolbroker.NewCaptureContainerPlan`, `toolbroker.NewCaptureRuntime`.
- **`(*ownedWorkerCapture).Run`** — Run is one-shot even after failure. Ownership teardown cancels and joins this complete operation, including exact-ID cleanup, before releasing its creator lock. Dependencies MUST honor cancellation and join their own work; source forks supplying dishonest/blocking adapters invalidate that contract.
  Signature: `func(ctx context.Context) (report workspacecapture.Report, err error)`.
  Calls: `c.authority.Err`, `c.guard.StopCaptureAuthority`, `c.runtime.Run`, `cancel`, `context.AfterFunc`, `context.WithCancel`, `ctx.Err`, `errors.Join`, `finish`, `l.beginOperation`, `l.operationCurrent`, `nilWorkerObservationDependency`, `stop`.
- **`recoverRetainedWorkerCapture`** — recoverRetainedWorkerCapture can reconcile only an already journaled exact ID. A fresh trusted guard must independently authorize cleanup and acknowledge it durably. There is no starter, frozen-source authority, object journal, report, volume deletion or refund capability. Unknown creation outcomes remain incidents.
  Signature: `func(ctx context.Context, l *workerWorkspaceObservation, intent workerCaptureIntent, id string, guard workerCaptureRuntimeGuard) (err error)`.
  Calls: `cancel`, `complete`, `context.AfterFunc`, `context.WithCancel`, `copyCaptureIntent`, `ctx.Err`, `errors.Join`, `guard.CaptureAuthorityLifetime`, `guard.StopCaptureAuthority`, `l.beginOperation`, `l.mu.Lock`, `l.mu.Unlock`, `l.operationCurrent`, `lifetime.Err`, `newWorkerCaptureJournal`, `nilWorkerObservationDependency`, `recovery.Reconcile`, `stop`, `toolbroker.NewCaptureContainerPlan`, `toolbroker.NewCaptureRecovery`.
