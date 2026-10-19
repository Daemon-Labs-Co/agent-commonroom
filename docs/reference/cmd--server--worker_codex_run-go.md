# cmd/server/worker_codex_run.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_codex_run.go)

Go implementation for worker codex run; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerCodexRunConfig`**: All paths/images are trusted worker deployment configuration. Prompt/model still require the controller's normal authorization and feature admission. No launch flag or model argument can replace isolation, runner or gateway.
  Declaration: `struct { // Trusted authenticated controller allocation, never a model-selected override. ManagedJob *room.ManagedJob IdleAnnounce workerIdleAnnounce EngineImage, AuthRoot, GatewayRoot string Prompt, Model string }`.
- **`(*workerPairSupervisor).RunCodex`** — RunCodex owns one already-admitted supervisor/session/bundle through credential acquisition, private gateway creation, paired assembly, callback turn and fenced shutdown. It exposes no public endpoint and does not grant admission or audit authority. A source fork CAN bypass it, BUT loses the isolation contract. Validation refusal retains caller ownership; after admission all exit paths, including panic, close owned resources and reconcile container identities.
  Signature: `func(ctx context.Context, session *toolbroker.ToolSession, token string, bundle io.ReadCloser, config workerCodexRunConfig, runner toolbroker.DockerRunner) (result codexbridge.Result, record workerJobRecord, err error)`.
  Calls: `acquireWorkerEngineCredential`, `cancel`, `context.AfterFunc`, `context.Background`, `context.WithDeadline`, `context.WithValue`, `credential.Release`, `ctx.Err`, `engine.Turn`, `engine.runtime.Close`, `errors.Join`, `gateway.Close`, `input.Close`, `newWorkerCodexEngine`, `newWorkerEngineGateway`, `repositoryDigestPattern.MatchString`, `s.AssemblePair`, `s.Run`, `s.Shutdown`, `s.available`, `s.configureJobIdle`, `s.mu.Lock`, `s.mu.Unlock`, `session.Close`, `session.Scope`, `stop`, `validWorkerAuditScope`.
