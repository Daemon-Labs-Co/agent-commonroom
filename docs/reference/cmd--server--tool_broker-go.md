# cmd/server/tool_broker.go

[Module index](cmd.md) · [Open source](../../cmd/server/tool_broker.go)

Go implementation for tool broker; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`newToolDispatcher`** — newToolDispatcher assembles the worker-side broker with mandatory controller acknowledgements. The supervisor MUST authenticate channel and keep it separate from model/stdout traffic. There is no local persistence or no-op audit fallback. This component does not launch a pair or authenticate a transport by itself. A source fork CAN substitute channel, BUT invalidates durable admission.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, allowed []string, image, dockerRoot string, now time.Time, channel io.ReadWriteCloser) (*toolbroker.Dispatcher, string, func(), error)`.
  Calls: `newWorkerToolSession`.
- **`newWorkerToolSession`** — newWorkerToolSession preserves the worker-job check while sharing readiness, authority and lifetime across runtime and durable audit. Runner injection is trusted assembly/test work, never a public/model choice or host fallback.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, allowed []string, image, dockerRoot string, now time.Time, channel io.ReadWriteCloser, runner toolbroker.DockerRunner) (*toolbroker.ToolSession, string, error)`.
  Calls: `isolatedCommandsSetting`, `newWorkerToolSessionWithCommands`, `os.Getenv`.
- **`newWorkerToolSessionWithCommands`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, allowed []string, image, dockerRoot string, now time.Time, channel io.ReadWriteCloser, runner toolbroker.DockerRunner, enabled bool) (*toolbroker.ToolSession, string, error)`.
  Calls: `ctxJob`, `toolbroker.NewToolSessionWithCommands`.
- **`ctxJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (string, bool)`.
  Calls: `ctx.Err`, `ctx.Value`, `validWorkerJobID`.
- **`isolatedCommandsSetting`** — Only exact deployment values are accepted; malformed opt-ins fail closed.
  Signature: `func(value string) (enabled, valid bool)`.
