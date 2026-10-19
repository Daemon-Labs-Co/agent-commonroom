# internal/codexbridge/bridge_test.go

[Module index](internal.md) · [Open source](../../internal/codexbridge/bridge_test.go)

Automated regression checks for bridge; read test names for individual contracts.

Package: `codexbridge`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`serverSetup`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, conn net.Conn) *json.Decoder`.
  Calls: `serverSetupCommands`.
- **`serverSetupCommands`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, conn net.Conn, commands bool) *json.Decoder`.
  Calls: `bufio.NewReader`, `decoder.Decode`, `json.NewDecoder`, `json.NewEncoder`, `json.NewEncoder(conn).Encode`, `json.Unmarshal`, `len`, `receive`, `send`, `string`, `strings.Contains`, `t.Error`, `t.Errorf`, `t.Helper`.
- **`callback`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]any`.
- **`completed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]any`.
- **`TestRunDrivesBoundedCallbackOnlyTurn`** — Regression contract: TestRunDrivesBoundedCallbackOnlyTurn.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `callback`, `cancel`, `close`, `completed`, `context.WithTimeout`, `decoder.Decode`, `json.NewEncoder`, `json.NewEncoder(server).Encode`, `make`, `net.Pipe`, `server.Close`, `serverSetup`, `string`, `strings.Contains`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **`TestRunRejectsEscapesReplaysAndWrongCorrelation`** — Regression contract: TestRunRejectsEscapesReplaysAndWrongCorrelation.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `callback`, `calls.Add`, `calls.Load`, `cancel`, `client.Close`, `close`, `completed`, `context.WithTimeout`, `decoder.Decode`, `delete`, `errors.New`, `int32`, `io.WriteString`, `json.NewEncoder`, `json.NewEncoder(server).Encode`, `make`, `net.Pipe`, `server.Close`, `serverSetup`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestRunCancellationInterruptsProviderIO`** — Regression contract: TestRunCancellationInterruptsProviderIO.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `cancel`, `context.WithCancel`, `make`, `net.Pipe`, `server.Close`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestRunInvalidInputStillClosesOwnedStream`** — Regression contract: TestRunInvalidInputStillClosesOwnedStream.
  Signature: `func(t *testing.T)`.
  Calls: `Handler`, `Run`, `make`, `net.Pipe`, `peer.Close`, `peer.Read`, `peer.SetReadDeadline`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`.
- **`TestReadOnlySandboxRejectsAmbiguousOrBroadenedPolicy`** — Regression contract: TestReadOnlySandboxRejectsAmbiguousOrBroadenedPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `readOnlySandbox`, `t.Fatal`, `t.Parallel`.
- **`TestRunRefusesCallbackBeforeTurnSubmission`** — Regression contract: TestRunRefusesCallbackBeforeTurnSubmission.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `callback`, `cancel`, `close`, `context.WithTimeout`, `json.NewDecoder`, `json.NewDecoder(peer).Decode`, `json.NewEncoder`, `json.NewEncoder(peer).Encode`, `make`, `net.Pipe`, `peer.Close`, `peer.Read`, `peer.SetReadDeadline`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **Type `finalizerStream`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn closed atomic.Bool }`.
- **`(*finalizerStream).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `s.Conn.Close`, `s.closed.Store`.
- **`TestFinalizerOwnsPostTurnBeforeStreamClose`** — Regression contract: TestFinalizerOwnsPostTurnBeforeStreamClose.
  Signature: `func(t *testing.T)`.
  Calls: `RunWithFinalizer`, `cancel`, `close`, `completed`, `context.WithTimeout`, `errors.New`, `json.NewEncoder`, `json.NewEncoder(server).Encode`, `make`, `net.Pipe`, `server.Close`, `serverSetup`, `stream.closed.Load`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCommandSchemaRequiresTrustedOptions`** — Regression contract: TestCommandSchemaRequiresTrustedOptions.
  Signature: `func(t *testing.T)`.
  Calls: `RunWithOptions`, `cancel`, `close`, `completed`, `context.WithTimeout`, `json.NewEncoder`, `json.NewEncoder(server).Encode`, `make`, `net.Pipe`, `server.Close`, `serverSetupCommands`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`.
