# cmd/server/startup_worker_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/startup_worker_test.go)

Automated regression checks for startup_worker; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestStartupWorkerRequiresVerifiedIdleWithoutLaunchingOrReplaying`** — Regression contract: TestStartupWorkerRequiresVerifiedIdleWithoutLaunchingOrReplaying.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.snapshot`, `len`, `mockServer`, `s.verifyWorkerAtStartup`, `t.Fatal`, `t.Parallel`, `t.Run`.
