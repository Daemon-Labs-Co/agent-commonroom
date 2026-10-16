# internal/toolbroker/tool_session.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/tool_session.go)

Go implementation for tool session; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ToolSession`**: ToolSession assembles one authority, plan, runtime and audited dispatcher. Runtime bootstrap opens readiness on the exact plan used by Dispatcher; callers cannot accidentally bootstrap one plan and execute through another. These are trusted worker objects, never serialized as model configuration. Close revokes access and closes audit I/O, NOT proof of container/volume stop. The paired supervisor must retain ownership through all calls and fence after failure. A source fork CAN substitute an unaudited runner/link, BUT invalidates isolation and durable audit guarantees. No host fallback is provided.
  Declaration: `struct { Runtime *ToolContainerRuntime Dispatcher *Dispatcher close func() scope Scope }`.
- **`(*ToolSession).Scope`** — Scope returns immutable nonsecret registration metadata, never the capability.
  Signature: `func() Scope`.
- **`(*ToolSession).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `s.close`.
- **`NewToolSession`** — NewToolSession requires a worker-authenticated private audit channel separated from engine stdout, a registered immutable scope and an approved local runner. It does not authenticate peers or grant controller/owner authority. The caller supplies operator-reviewed image/root policy, never model-provided Docker data. After accepting basic inputs, construction owns/closes the channel on failure. The raw capability is returned once for private engine delivery, never history.
  Signature: `func(ctx context.Context, scope Scope, allowed []string, image, dockerRoot string, now time.Time, channel io.ReadWriteCloser, runner DockerRunner) (*ToolSession, string, error)`.
  Calls: `New`, `NewAuditClient`, `NewDispatcher`, `NewToolContainerPlan`, `NewToolContainerRuntime`, `authority.Revoke`, `cancel`, `channel.Close`, `cleanup`, `closeLink`, `context.AfterFunc`, `context.WithDeadline`, `ctx.Err`, `lifetime.Err`, `once.Do`, `scope.Deadline.UTC`, `scope.Deadline.UTC().Round`, `stop`.
- **`(*ToolContainerRuntime).WorkspaceIntent`** — WorkspaceIntent is a lifetime-bound request for the supervisor to journal. It creates no Docker object and never becomes permission to skip that journal.
  Signature: `func(ctx context.Context) (WorkspaceVolumeIdentity, error)`.
  Calls: `cancel`, `child.Err`, `r.context`, `r.executor.plan.WorkspaceIntent`, `r.mu.Lock`, `r.mu.Unlock`, `time.Now`.
- **`(*ToolContainerRuntime).VerifyCreated`** — VerifyCreated repeats full stopped policy for the supervisor's durable bind.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `cancel`, `r.context`, `r.mu.Lock`, `r.mu.Unlock`, `r.verify`.
