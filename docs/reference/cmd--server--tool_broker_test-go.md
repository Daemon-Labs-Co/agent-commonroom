# cmd/server/tool_broker_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/tool_broker_test.go)

Automated regression checks for tool_broker; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestNewToolDispatcherRequiresWorkerAndAuditChannel`** — Regression contract: TestNewToolDispatcherRequiresWorkerAndAuditChannel.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `cleanup`, `client.Close`, `context.WithCancel`, `context.WithValue`, `len`, `net.Pipe`, `newToolDispatcher`, `peer.Close`, `sampleToolScope`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`.
- **Type `fakeToolExecutor`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ fail bool }`.
- **`(fakeToolExecutor).Execute`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, action toolbroker.Action, out io.Writer) error`.
  Calls: `errors.New`, `out.Write`.
- **`TestDispatchRoundTripsThroughScopedControllerChannel`** — Regression contract: TestDispatchRoundTripsThroughScopedControllerChannel.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `client.Close`, `context.WithCancel`, `dispatcher.Dispatch`, `len`, `make`, `net.Pipe`, `s.load`, `scopedAuditFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `time.Now`, `toolAuditKey`, `toolbroker.New`, `toolbroker.NewAuditClient`, `toolbroker.NewDispatcher`, `toolbroker.ServeAudit`.
- **`TestToolSeparationFlagCannotGrantOrDisableAudit`** — Credential separation and durable audit are mandatory, not rollout toggles.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `len`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkerCancellationClosesAuditDespiteFreshDispatchContext`** — Regression contract: TestWorkerCancellationClosesAuditDespiteFreshDispatchContext.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `cleanup`, `close`, `context.Background`, `context.WithCancel`, `context.WithValue`, `dispatcher.Dispatch`, `make`, `net.Pipe`, `newToolDispatcher`, `peer.Close`, `peer.Read`, `sampleToolScope`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`, `time.Now`.
