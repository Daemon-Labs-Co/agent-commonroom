# cmd/server/worker_pair_shutdown.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_pair_shutdown.go)

Go implementation for worker pair shutdown; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*workerPairSupervisor).Shutdown`** — Shutdown is the supervisor's bounded stop-and-reconcile path. It cancels first, waits for in-flight startup to unwind before releasing creator ownership, then uses the existing durable tombstone/lease/identity fences to remove both roles. The caller supplies a fresh cleanup context, not the cancelled model lifetime. No default runner, direct host execution, restart or workspace discard authority exists here. A source fork CAN skip this method, BUT cannot claim shutdown proof. A misbehaving launch adapter that ignores cancellation keeps Close waiting; never release its lease early or turn a timeout into absence evidence.
  Signature: `func(ctx context.Context, runner toolbroker.DockerRunner) (workerJobRecord, error)`.
  Calls: `cancel`, `context.WithTimeout`, `credential.Release`, `credential.mu.Lock`, `credential.mu.Unlock`, `ctx.Err`, `errors.Join`, `errors.New`, `fenceWorkerJobContainers`, `listener.Close`, `s.Close`, `s.mu.Lock`, `s.mu.Unlock`, `s.reconcile.Lock`, `s.reconcile.Unlock`, `sameWorkerFenceBinding`, `stopWorkerPair`, `strings.Join`, `validWorkerRecord`.
- **`(*workerPairSupervisor).Run`** — Run owns one admitted isolated model turn and always reconciles after startup or turn failure, cancellation, normal return, or panic. turn is trusted worker driver code for the reviewed engine, never a model-selected host function. It MUST honor its context and return before calling Close/Shutdown; those methods wait for the turn to unwind. This method provides no host execution fallback. No result is a successful lifecycle result unless fenced cleanup also succeeds. Present retained workspaces deliberately keep that proof unresolved.
  Signature: `func(ctx context.Context, adapter workerPairLaunchAdapter, runner toolbroker.DockerRunner, turn func(context.Context) error) (record workerJobRecord, err error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.Background`, `context.WithDeadline`, `ctx.Err`, `errors.Join`, `errors.New`, `lifetime.Err`, `s.Shutdown`, `s.Start`, `s.ctx.Err`, `s.mu.Lock`, `s.mu.Unlock`, `s.turn.Lock`, `s.turn.Unlock`, `stop`, `turn`.
