# cmd/server/worker_idle.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_idle.go)

Go implementation for worker idle; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`workerRecordsIdle`** — workerRecordsIdle rejects unknown/nonterminal records instead of inferring that a missing container proves a suspended worker cannot launch it later.
  Signature: `func(dir string) (bool, error)`.
  Calls: `entry.Name`, `errors.New`, `f.Close`, `f.ReadDir`, `filepath.Join`, `info.Mode`, `info.Mode().IsRegular`, `len`, `os.Lstat`, `os.Open`, `readWorkerRecord`, `strings.HasSuffix`, `strings.TrimSuffix`, `validWorkerJobID`, `workerJobSettled`.
- **`workerIdleProbe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (bool, error)`.
  Calls: `cmd.Run`, `configureWorker`, `dockerSecurityReady`, `errors.New`, `exec.CommandContext`, `ids.String`, `info.Bytes`, `strings.TrimSpace`, `workerSnapshot`, `workerStateDirectory`.
