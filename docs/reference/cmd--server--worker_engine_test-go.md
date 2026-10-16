# cmd/server/worker_engine_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_engine_test.go)

Automated regression checks for worker_engine; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWorkerCodexEngineRequiresOriginalWorkerJobAndRole`** — Regression contract: TestWorkerCodexEngineRequiresOriginalWorkerJobAndRole.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `engine.Create`, `engine.Verify`, `engine.runtime.Close`, `engineGatewayFixture`, `newWorkerCodexEngine`, `sampleToolScope`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestPairAssemblyRejectsEngineFromDifferentCompleteScope`** — Regression contract: TestPairAssemblyRejectsEngineFromDifferentCompleteScope.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `engine.Turn`, `engine.runtime.Close`, `engineGatewayFixture`, `input.closes.Load`, `net.Pipe`, `newWorkerCodexEngine`, `newWorkerToolSession`, `peer.Close`, `s.AssemblePair`, `s.ctx.Err`, `session.Close`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `time.Now`, `workspacePair`.
