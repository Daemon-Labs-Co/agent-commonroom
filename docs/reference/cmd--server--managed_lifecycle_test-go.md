# cmd/server/managed_lifecycle_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/managed_lifecycle_test.go)

Automated regression checks for managed_lifecycle; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestManagedDispatchOwnsLifecycle`** — Regression contract: TestManagedDispatchOwnsLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `cancel`, `cleanup`, `cleanupCtx.Deadline`, `cleanupCtx.Err`, `context.Background`, `context.WithCancel`, `context.WithValue`, `decodeWorkerRequest`, `dispatchManagedWorker`, `errors.New`, `exec.CommandContext`, `now.Add`, `providerTurnArgs`, `reflect.DeepEqual`, `request.Deadline.Equal`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().UTC`, `time.Until`.
- **`TestManagedDispatchRejectsUnavailableContext`** — Regression contract: TestManagedDispatchRejectsUnavailableContext.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `dispatchManagedWorker`, `t.Fatal`, `t.Parallel`.
