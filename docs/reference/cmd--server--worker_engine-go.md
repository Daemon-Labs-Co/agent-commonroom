# cmd/server/worker_engine.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_engine.go)

Go implementation for worker engine; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerCodexEngine`**: workerCodexEngine implements the reviewed engine role only. Private socket setup and connection credential leasing must precede construction; a model never supplies paths, image, runner or Docker flags. The runtime verifies reported isolation policy. A source fork CAN replace these dependencies, BUT invalidates isolation. There is no CLI/host fallback, implicit image, automatic pull or native-tool exception.
  Declaration: `struct { scope toolbroker.Scope runtime *toolbroker.EngineContainerRuntime credential *workerEngineCredential gateway *workerEngineGateway }`.
- **`newWorkerCodexEngine`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, image string, gateway *workerEngineGateway, credential *workerEngineCredential, runner toolbroker.DockerRunner) (*workerCodexEngine, error)`.
  Calls: `credential.mu.Lock`, `credential.mu.Unlock`, `credential.original.Err`, `ctxJob`, `gateway.ready`, `scope.Deadline.UTC`, `scope.Deadline.UTC().Round`, `toolbroker.NewEngineContainerRuntime`.
- **`(*workerCodexEngine).identity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v workerPairIdentity) bool`.
- **`(*workerCodexEngine).Create`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, identity workerPairIdentity) (string, error)`.
  Calls: `e.gateway.ready`, `e.identity`, `e.runtime.Create`.
- **`(*workerCodexEngine).Verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, identity workerPairIdentity, id string) error`.
  Calls: `e.gateway.ready`, `e.identity`, `e.runtime.VerifyCreated`.
- **`(*workerCodexEngine).Inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) ([]byte, error)`.
  Calls: `e.gateway.ready`, `e.runtime.Inspect`.
- **`(*workerCodexEngine).Start`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `e.gateway.ready`, `e.runtime.Start`.
- **`(*workerCodexEngine).Scope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() toolbroker.Scope`.
- **`(*workerCodexEngine).Turn`** — Turn connects the inspected running engine to the original tool session. Invoke inside supervisor.Run's turn callback so all outcomes receive fenced shutdown. Neither model arguments nor public requests select the stream/ID.
  Signature: `func(ctx context.Context, s *workerPairSupervisor, session *toolbroker.ToolSession, token, prompt, model string) (codexbridge.Result, error)`.
  Calls: `clear`, `e.credential.Provision`, `e.credential.StageRefresh`, `e.gateway.ready`, `e.runtime.Attach`, `e.runtime.ExportCredential`, `s.available`, `s.cancel`, `s.codexTurn`, `s.mu.Lock`, `s.mu.Unlock`, `session.Scope`, `stream.Close`, `validWorkerAuditScope`.
