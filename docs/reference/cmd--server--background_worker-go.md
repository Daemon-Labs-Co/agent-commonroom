# cmd/server/background_worker.go

[Module index](cmd.md) · [Open source](../../cmd/server/background_worker.go)

Go implementation for background worker; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`backgroundWorkerStage`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, job, provider string) string`.
  Calls: `json.Unmarshal`, `len`.
- **`monitorRemoteBackground`** — Poll only the approved SSH worker and the job allocated by this controller. Same-state observations are bounded to fifteen-second confirmations. They do not extend deadlines, authorize tools or prove useful model progress.
  Signature: `func(ctx context.Context, done <-chan struct{}, alias, job, provider string)`.
  Calls: `backgroundWorkerStage`, `cancel`, `cmd.Run`, `context.WithTimeout`, `ctx.Done`, `emitBackground`, `output.Bytes`, `ticker.Stop`, `time.NewTicker`, `time.Now`, `time.Since`, `workerSSHCommand`, `workerTerminal`.
