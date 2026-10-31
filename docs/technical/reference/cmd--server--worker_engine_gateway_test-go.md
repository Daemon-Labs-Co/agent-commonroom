# cmd/server/worker_engine_gateway_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_engine_gateway_test.go)

Automated regression checks for worker_engine_gateway; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`shortGatewayRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) string`.
  Calls: `os.MkdirTemp`, `os.RemoveAll`, `t.Cleanup`, `t.Fatal`, `t.Helper`.
- **`engineGatewayFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, ctx context.Context, scope toolbroker.Scope) *workerEngineGateway`.
  Calls: `g.Close`, `newWorkerEngineGateway`, `shortGatewayRoot`, `t.Cleanup`, `t.Fatal`, `t.Helper`.
- **`TestEngineGatewayFreshPrivateScopedAndCancelled`** — Regression contract: TestEngineGatewayFreshPrivateScopedAndCancelled.
  Signature: `func(t *testing.T)`.
  Calls: `bufio.NewReader`, `bufio.NewReader(conn).ReadString`, `cancel`, `conn.Close`, `conn.SetDeadline`, `context.WithCancel`, `context.WithValue`, `first.Close`, `first.ready`, `io.WriteString`, `net.DialTimeout`, `newWorkerEngineGateway`, `os.IsNotExist`, `os.Lstat`, `sampleToolScope`, `second.Close`, `shortGatewayRoot`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestEngineGatewayRefusesUnsafeRootAndChangedSocket`** — Regression contract: TestEngineGatewayRefusesUnsafeRootAndChangedSocket.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `filepath.Join`, `g.Close`, `g.ready`, `newWorkerEngineGateway`, `os.Chmod`, `os.ReadFile`, `os.Remove`, `os.Stat`, `os.Symlink`, `os.WriteFile`, `sampleToolScope`, `shortGatewayRoot`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
