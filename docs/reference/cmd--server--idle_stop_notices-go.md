# cmd/server/idle_stop_notices.go

[Module index](cmd.md) · [Open source](../../cmd/server/idle_stop_notices.go)

Go implementation for idle stop notices; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`validIdleNotice`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(n room.IdleStopNotice) bool`.
  Calls: `j.AllocatedAt.Add`, `n.RecordedAt.Add`, `n.RecordedAt.IsZero`, `n.RequestedAt.After`, `n.RequestedAt.Before`, `n.WorkerStartedAt.Add`, `n.WorkerStartedAt.Before`, `n.WorkerStartedAt.IsZero`, `time.Duration`, `validManagedJob`, `validWorkerJobID`.
- **`projectIdleNotice`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `append`, `idleNoticeMessage`, `json.Unmarshal`, `sameDisposalJob`, `validIdleNotice`.
- **`(*server).newIdleAnnouncer`** — newIdleAnnouncer binds a registered job and initiating human to a trusted private callback. This is not a public/model/plugin API or peer authentication. Production assembly must carry it over a reviewed authenticated control stream.
  Signature: `func(ctx context.Context, jobID string) (workerIdleAnnounce, error)`.
  Calls: `action.Err`, `canAccess`, `ctx.Err`, `ctx.Value`, `member`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.recordIdleStop`, `sameWorkerBinding`, `validManagedJob`, `validWorkerRecord`.
- **`idleNoticeMessage`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(n room.IdleStopNotice) room.Message`.
  Calls: `fmt.Sprintf`.
- **`projectIdleNoticeMessages`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State)`.
  Calls: `append`.
- **`(*server).recordIdleStop`** — recordIdleStop is shared by the local callback and the authenticated private stream adapter; all authority comes from its captured registered allocation.
  Signature: `func(action context.Context, job room.ManagedJob, started, requested time.Time) error`.
  Calls: `action.Err`, `context.WithValue`, `errors.Is`, `id`, `prior.RequestedAt.Equal`, `prior.WorkerStartedAt.Equal`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `sameDisposalJob`, `time.Now`, `time.Now().UTC`, `validIdleNotice`.
