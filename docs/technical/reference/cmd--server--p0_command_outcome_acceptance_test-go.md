# cmd/server/p0_command_outcome_acceptance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_command_outcome_acceptance_test.go)

Automated regression checks for p0_command_outcome_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `commandOutcomeDocker`**: Substitute only Docker operations. Real session/bootstrap/command capture, exact-container cleanup, audit framing and controller persistence run normally.
  Declaration: `struct { *codexIntegrationDocker t *testing.T server *server db *mockDB effects []string commandCalls int fenced bool failOutcome bool outputBytes int timeout bool }`.
- **`(*commandOutcomeDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, input []byte, out io.Writer) error`.
  Calls: `append`, `ctx.Deadline`, `ctx.Err`, `len`, `r.codexIntegrationDocker.Run`, `r.t.Fatal`, `slices.Equal`, `strings.Repeat`, `time.Until`.
- **`(*commandOutcomeDocker).RunCommand`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, out io.Writer) (int, error)`.
  Calls: `append`, `ctx.Done`, `ctx.Err`, `errors.Is`, `io.WriteString`, `len`, `r.server.load`, `r.t.Fatal`, `slices.Contains`, `slices.Equal`, `strings.Repeat`, `time.Now`, `time.Since`, `toolAuditKey`.
- **`driveP0CommandOutcomeFailure`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `observeCommandOutcome`, `t.Fatal`, `t.Helper`.
- **`observeCommandOutcome`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, fail bool) pendingP0Observation`.
  Calls: `observeCommandCapture`, `t.Helper`.
- **`observeCommandCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, fail bool, outputBytes int) pendingP0Observation`.
  Calls: `observeCommandExecution`, `t.Helper`.
- **`observeCommandExecution`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, fail bool, outputBytes int, timeout bool) pendingP0Observation`.
  Calls: `append`, `awaitAuditResult`, `cancel`, `client.Close`, `context.WithCancel`, `context.WithValue`, `db.snapshot`, `errors.Is`, `io.NopCloser`, `len`, `make`, `net.Pipe`, `newWorkerToolSessionWithCommands`, `peer.Close`, `s.load`, `scopedAuditFixtureWithLifetime`, `session.Close`, `session.Dispatcher.Dispatch`, `session.Runtime.Bootstrap`, `session.Runtime.Create`, `session.Runtime.PrepareVolume`, `session.Runtime.Start`, `session.Runtime.WorkspaceIntent`, `strings.Contains`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `time.Now`, `toolAuditKey`, `toolbroker.ServeAudit`, `workerCodexToolRequest`.
- **`driveP0CommandTimeout`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `observeCommandExecution`, `synctest.Test`, `t.Fatal`, `t.Helper`.
- **`TestCommandOutcomeAcceptanceHealthyControl`** — Regression contract: TestCommandOutcomeAcceptanceHealthyControl.
  Signature: `func(t *testing.T)`.
  Calls: `observeCommandOutcome`, `slices.Equal`, `t.Fatal`, `t.Parallel`.
- **`driveP0CommandOutputLimit`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `observeCommandCapture`, `t.Fatal`, `t.Helper`.
- **`TestCommandCaptureExactOutputLimitControl`** — Regression contract: TestCommandCaptureExactOutputLimitControl.
  Signature: `func(t *testing.T)`.
  Calls: `observeCommandCapture`, `slices.Equal`, `t.Fatal`, `t.Parallel`.
