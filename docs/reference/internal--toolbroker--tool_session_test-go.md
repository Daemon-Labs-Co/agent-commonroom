# internal/toolbroker/tool_session_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/tool_session_test.go)

Automated regression checks for tool_session; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `sessionDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *bootstrapScenario response []byte executed int auditIntents *atomic.Int32 }`.
- **`(*sessionDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `f.auditIntents.Load`, `f.runtimeScenario.Run`, `f.t.Fatal`, `len`, `out.Write`.
- **`TestToolSessionSharesBootstrapAndDurableDispatcher`** — Regression contract: TestToolSessionSharesBootstrapAndDurableDispatcher.
  Signature: `func(t *testing.T)`.
  Calls: `NewToolSession`, `ServeAudit`, `boolValue`, `bytes.Contains`, `bytes.Equal`, `cancel`, `client.Close`, `context.Background`, `context.WithCancel`, `inspectionFixture`, `intents.Add`, `intents.Load`, `io.NopCloser`, `json.Marshal`, `len`, `liveFixture`, `make`, `net.Pipe`, `outcomes.Add`, `outcomes.Load`, `peer.Close`, `session.Close`, `session.Dispatcher.Dispatch`, `session.Runtime.Bootstrap`, `session.Runtime.Create`, `session.Runtime.PrepareVolume`, `session.Runtime.Start`, `session.Runtime.VerifyCreated`, `session.Runtime.WorkspaceIntent`, `strings.NewReader`, `syntheticReadResult`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `time.After`, `time.Now`.
- **`TestToolSessionConstructionFailureClosesOwnedAudit`** — Regression contract: TestToolSessionConstructionFailureClosesOwnedAudit.
  Signature: `func(t *testing.T)`.
  Calls: `NewToolSession`, `errors.New`, `liveFixture`, `mockDocker`, `net.Pipe`, `peer.Close`, `peer.Read`, `peer.SetReadDeadline`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestSessionCommandFlagAndGrantAreIndependent`** — Regression contract: TestSessionCommandFlagAndGrantAreIndependent.
  Signature: `func(t *testing.T)`.
  Calls: `NewToolSessionWithCommands`, `append`, `client.Close`, `inspectionFixture`, `liveFixture`, `mockDocker`, `net.Pipe`, `peer.Close`, `session.Close`, `session.CommandsEnabled`, `strings.HasPrefix`, `strings.HasSuffix`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`.
