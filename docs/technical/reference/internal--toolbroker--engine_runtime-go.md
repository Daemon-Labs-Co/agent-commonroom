# internal/toolbroker/engine_runtime.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/engine_runtime.go)

Go implementation for engine runtime; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `EngineContainerRuntime`**: EngineContainerRuntime creates only a fresh reviewed Codex engine, never a provider on the host. The gateway socket path is private worker policy, NOT model input. The worker must operate its approved private provider gateway before constructing this. Credentials enter only through the owned protocol stdin. A source fork CAN substitute image/runner/paths, BUT violates credential separation. No automatic image, provider, network or uncontained fallback exists.
  Declaration: `struct { mu sync.Mutex ctx context.Context original context.Context cancel context.CancelFunc scope Scope image, gateway, imageID, id string runner DockerRunner attempted, verified, started, running bool attached bool credentialExported bool }`.
- **`NewEngineContainerRuntime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope Scope, image, gateway string, runner DockerRunner) (*EngineContainerRuntime, error)`.
  Calls: `ValidAuditScope`, `context.WithDeadline`, `ctx.Err`, `len`, `path.Clean`, `path.IsAbs`, `strings.ContainsAny`, `time.Now`, `validPath`, `validToolImage`.
- **`(*EngineContainerRuntime).Close`** — Close only cancels execution. It never proves container absence or discards data.
  Signature: `func()`.
  Calls: `r.cancel`.
- **`(*EngineContainerRuntime).name`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() string`.
- **`(*EngineContainerRuntime).labels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]string`.
- **`engineEnvironment`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]string`.
- **`(*EngineContainerRuntime).createArgs`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() []string`.
  Calls: `append`, `engineEnvironment`, `r.labels`, `r.name`, `sort.Strings`.
- **`(*EngineContainerRuntime).context`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (context.Context, func(), error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithDeadline`, `ctx.Err`, `r.ctx.Err`, `r.original.Err`, `stop`.
- **`(*EngineContainerRuntime).capture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string) ([]byte, error)`.
  Calls: `cancel`, `child.Err`, `context.WithCancel`, `output.snapshot`, `r.runner.Run`.
- **`(*EngineContainerRuntime).verifyImage`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) error`.
  Calls: `decodeInspection`, `digest`, `engineEnvironment`, `len`, `slices.Contains`, `slices.Equal`, `strings.Cut`, `strings.HasPrefix`, `strings.TrimPrefix`.
- **`(*EngineContainerRuntime).verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, running bool) error`.
  Calls: `decodeInspection`, `digest`, `engineEnvironment`, `isBool`, `isInt`, `len`, `r.labels`, `r.name`, `sameMap`, `slices.Equal`, `strings.Cut`.
- **`(*EngineContainerRuntime).inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) ([]byte, error)`.
  Calls: `r.capture`.
- **`(*EngineContainerRuntime).Create`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (string, error)`.
  Calls: `digest`, `done`, `r.capture`, `r.context`, `r.createArgs`, `r.inspect`, `r.mu.Lock`, `r.mu.Unlock`, `r.name`, `r.verify`, `r.verifyImage`, `string`, `strings.TrimSpace`.
- **`(*EngineContainerRuntime).VerifyCreated`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `done`, `r.context`, `r.inspect`, `r.mu.Lock`, `r.mu.Unlock`, `r.verify`.
- **`(*EngineContainerRuntime).Inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) ([]byte, error)`.
  Calls: `done`, `r.context`, `r.inspect`, `r.mu.Lock`, `r.mu.Unlock`, `r.verify`.
- **`(*EngineContainerRuntime).Start`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `done`, `r.capture`, `r.context`, `r.inspect`, `r.mu.Lock`, `r.mu.Unlock`, `r.verify`, `string`, `strings.TrimSpace`.
