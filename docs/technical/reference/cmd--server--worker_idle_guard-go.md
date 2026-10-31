# cmd/server/worker_idle_guard.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_idle_guard.go)

Go implementation for worker idle guard; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerIdlePolicy`**: Worker-owner deployment policy only. Never read these values from model tools, prompts, room messages or untrusted launch arguments. Disabling this optional guard cannot disable hard deadlines, isolation, fencing or source retention.
  Declaration: `struct { Enabled bool Minutes int }`.
- **`readWorkerIdlePolicy`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(getenv func(string) string) (workerIdlePolicy, error)`.
  Calls: `errors.New`, `getenv`, `strconv.Atoi`.
- **Type `workerIdleAnnounce`**: workerIdleAnnounce is a trusted private control-stream adapter. Success MUST mean durable controller acknowledgment, never a model claim. It must honor context cancellation. A source fork supplying a no-op loses audit guarantees; this callback grants no provider launch, disposal or host-execution authority.
  Declaration: `func(context.Context, workerJobRecord) error`.
- **Type `workerIdleWatch`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { announce workerIdleAnnounce requireAnnouncement bool policy workerIdlePolicy runner toolbroker.DockerRunner since time.Time armed bool }`.
- **`(*workerPairSupervisor).configureIdle`** — configureIdle snapshots trusted deployment policy before startup. Explicit injection is for reviewed worker assembly/tests, never model/plugin adapters. A source fork can omit this guard, but cannot claim automatic idle shutdown.
  Signature: `func(p workerIdlePolicy, runner toolbroker.DockerRunner) error`.
  Calls: `errors.New`, `s.mu.Lock`, `s.mu.Unlock`.
- **`(*workerPairSupervisor).ensureIdle`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(runner toolbroker.DockerRunner) error`.
  Calls: `readWorkerIdlePolicy`, `s.configureIdle`, `s.mu.Lock`, `s.mu.Unlock`.
- **`(*workerPairSupervisor).armIdleLocked`** — Called under mu after successful startup. No idle timer covers preparation.
  Signature: `func()`.
  Calls: `s.watchIdle`, `time.Now`.
- **`(*workerPairSupervisor).watchIdle`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `s.ctx.Done`, `s.expireIdle`, `s.idle.since.Add`, `s.mu.Lock`, `s.mu.Unlock`, `time.Duration`, `time.NewTimer`, `time.Now`, `time.Until`, `timer.Stop`.
- **`(*workerPairSupervisor).expireIdle`** — expireIdle atomically wins against startup, active turns and queued turns. It journals intention before cancellation and uses only the existing bounded non-destructive Shutdown. A retained workspace keeps full job proof unresolved.
  Signature: `func(now time.Time) (bool, error)`.
  Calls: `callIdleAnnouncer`, `cancel`, `context.Background`, `context.WithTimeout`, `errors.New`, `now.Before`, `now.UTC`, `now.UTC().Round`, `s.Close`, `s.Shutdown`, `s.ctx.Err`, `s.idle.since.Add`, `s.idle.since.IsZero`, `s.mu.Lock`, `s.mu.Unlock`, `s.observeIdleBusy`, `s.startup.TryLock`, `s.startup.Unlock`, `s.turn.TryLock`, `s.turn.Unlock`, `time.Duration`, `time.Now`, `time.Now().UTC`, `writeWorkerRecord`.
- **`(*workerPairSupervisor).observeIdleBusy`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(now time.Time)`.
  Calls: `now.After`, `s.mu.Lock`, `s.mu.Unlock`.
- **`(*workerPairSupervisor).queueTurn`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `errors.New`, `s.ctx.Err`, `s.mu.Lock`, `s.mu.Unlock`.
- **`(*workerPairSupervisor).finishQueuedTurn`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `s.mu.Lock`, `s.mu.Unlock`, `time.Now`.
- **`callIdleAnnouncer`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, record workerJobRecord, announce workerIdleAnnounce) (err error)`.
  Calls: `announce`, `recover`.
