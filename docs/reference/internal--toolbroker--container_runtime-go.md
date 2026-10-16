# internal/toolbroker/container_runtime.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/container_runtime.go)

Go implementation for container runtime; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ToolContainerRuntime`**: ToolContainerRuntime is the constrained tool half of the paired launcher. Only the trusted isolated worker may construct it after journalled admission. PrepareVolume allocates a fresh local workspace after a durable intent; Bootstrap initializes it after idle startup. This module never mounts host files or accepts Docker flags from a model. Pair lifecycle MUST bind both IDs and inspect aggregate policy before Start. A source fork CAN bypass those steps, BUT loses crash/containment guarantees. No production pair endpoint currently calls this module.
  Declaration: `struct { mu sync.Mutex lifetime context.Context original context.Context executor *ContainerExecutor id string attempted, started bool verified bool running bool bootstrapAttempted bool volumeAttempted bool volumePrepared bool }`.
- **`NewToolContainerRuntime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(lifetime context.Context, plan *ToolContainerPlan, runner DockerRunner, dockerRoot string) (*ToolContainerRuntime, error)`.
  Calls: `NewContainerExecutor`, `lifetime.Err`, `plan.CreateArgs`, `time.Now`.
- **`(*ToolContainerRuntime).context`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (context.Context, func(), error)`.
  Calls: `cancel`, `child.Err`, `context.AfterFunc`, `ctx.Err`, `r.executor.plan.authority.executionContext`, `r.lifetime.Err`, `r.original.Err`, `stop`.
- **`(*ToolContainerRuntime).Create`** — Create issues only Docker create, verifies the returned immutable ID and the full stopped-container/volume policy, and retains uncertainty on any failure. The caller owns the durable creation intent before invoking this method.
  Signature: `func(ctx context.Context) (string, error)`.
  Calls: `cancel`, `decodeInspection`, `digest`, `p.ContainerName`, `p.CreateArgs`, `p.WorkspaceVolume`, `p.verifyVolume`, `r.context`, `r.executor.capture`, `r.mu.Lock`, `r.mu.Unlock`, `r.verify`, `string`, `strings.TrimSpace`, `time.Now`.
- **`(*ToolContainerRuntime).inspection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) ([]byte, []byte, error)`.
  Calls: `decodeInspection`, `r.executor.capture`, `r.executor.plan.WorkspaceVolume`.
- **`(*ToolContainerRuntime).verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string, running bool) error`.
  Calls: `r.executor.plan.verifyInspection`, `r.inspection`, `time.Now`.
- **`(*ToolContainerRuntime).Inspect`** — Inspect supplies fresh ID-bound metadata for the supervisor's aggregate gate.
  Signature: `func(ctx context.Context, id string) ([]byte, error)`.
  Calls: `cancel`, `r.context`, `r.inspection`, `r.mu.Lock`, `r.mu.Unlock`.
- **`(*ToolContainerRuntime).Start`** — Start rechecks full stopped policy before starting by immutable ID, then checks running policy. It never retries an uncertain start and never claims pair proof.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `cancel`, `r.context`, `r.executor.capture`, `r.mu.Lock`, `r.mu.Unlock`, `r.verify`, `string`, `strings.TrimSpace`.
