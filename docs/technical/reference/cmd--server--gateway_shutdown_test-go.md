# cmd/server/gateway_shutdown_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/gateway_shutdown_test.go)

Automated regression checks for gateway_shutdown; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestGatewayCloseCancelsPendingResolution`** — Regression contract: TestGatewayCloseCancelsPendingResolution.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `context.Background`, `ctx.Done`, `ctx.Err`, `filepath.Join`, `g.Close`, `g.connect`, `g.ctx.Err`, `httptest.NewRecorder`, `httptest.NewRequest`, `make`, `newProviderGateway`, `os.MkdirTemp`, `os.RemoveAll`, `t.Cleanup`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.After`.
