# cmd/server/worker_codex_turn.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_codex_turn.go)

Go implementation for worker codex turn; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`workerCodexToolRequest`** — workerCodexToolRequest supplies authority only from trusted session binding. Model arguments never contain capability, scope, call ID, Docker options or provider settings. The dispatcher still authorizes path, operation and budget.
  Signature: `func(scope toolbroker.Scope, token, callID string, arguments json.RawMessage) (toolbroker.Request, error)`.
  Calls: `bytes.NewReader`, `bytes.TrimSpace`, `decoder.Decode`, `decoder.More`, `decoder.Token`, `json.Delim`, `json.NewDecoder`, `json.Unmarshal`, `len`, `strictjson.ValidEncoding`, `string`, `toolcommand.Valid`.
- **`(*workerPairSupervisor).CodexTurn`** — CodexTurn drives only the reviewed isolated app-server stream after startup. Engine process creation, controlled configuration and fencing are separate required boundaries. No codex exec, host tool, resume or API billing fallback is offered. A source fork CAN substitute stream/dispatcher, BUT loses isolation.
  Signature: `func(ctx context.Context, session *toolbroker.ToolSession, token string, stream io.ReadWriteCloser, prompt, model string) (codexbridge.Result, error)`.
  Calls: `s.codexTurn`.
- **`(*workerPairSupervisor).codexTurn`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, session *toolbroker.ToolSession, token string, stream io.ReadWriteCloser, prompt, model string, finalize func(context.Context) error) (codexbridge.Result, error)`.
  Calls: `cancel`, `codexbridge.RunWithOptions`, `context.AfterFunc`, `context.WithDeadline`, `repositoryDigestPattern.MatchString`, `s.available`, `s.cancel`, `s.mu.Lock`, `s.mu.Unlock`, `session.CommandsEnabled`, `session.Dispatcher.Dispatch`, `session.Scope`, `stop`, `validWorkerAuditScope`, `workerCodexToolRequest`.
