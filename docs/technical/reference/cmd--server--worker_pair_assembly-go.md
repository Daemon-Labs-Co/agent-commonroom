# cmd/server/worker_pair_assembly.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_pair_assembly.go)

Go implementation for worker pair assembly; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerEngineLaunch`**: workerEngineLaunch is a reviewed credential-bearing engine policy, never a model-selected command. Implementations MUST create fresh stopped containers, verify immutable image/mount/network/credential policy, start only by ID and respect cancellation. No local/SSH provider fallback is permitted. A source fork CAN substitute host execution here, BUT invalidates mandatory isolation. Codex supplies reported-policy checks; reviewed images, private mount staging and live acceptance are still required before production admission.
  Declaration: `interface { Scope() toolbroker.Scope Create(context.Context, workerPairIdentity) (string, error) Verify(context.Context, workerPairIdentity, string) error Inspect(context.Context, string) ([]byte, error) Start(context.Context, string) error }`.
- **Type `workerPairToolRuntime`**: Behavior contract; required method signatures are shown in the declaration.
  Declaration: `interface { WorkspaceIntent(context.Context) (toolbroker.WorkspaceVolumeIdentity, error) PrepareVolume(context.Context, toolbroker.WorkspaceVolumeIdentity) (toolbroker.WorkspaceVolumeIdentity, error) Create(context.Context) (string, error) VerifyCreated(context.Context, string) error Inspect(context.Context, string) ([]byte, error) Start(context.Context, string) error Bootstrap(context.Context, io.ReadCloser) (workspacebootstrap.Receipt, error) }`.
- **Type `workerAssembledPair`**: workerAssembledPair routes the credential-free tool role through the exact audited session's runtime. Engine and tool never share images, credentials or dispatchers. Bootstrap owns the approved bundle, not a model-controlled path. This is launch assembly, not authentication or shutdown proof. The supervisor retains creation journals/leases and owns fencing after any failure.
  Declaration: `struct { mu sync.Mutex ctx context.Context scope toolbroker.Scope engine workerEngineLaunch tool workerPairToolRuntime bundle io.ReadCloser engineID, toolID string bootstrapAttempted bool bootstrapped bool toolCreateAttempted, engineCreateAttempted bool toolVerified, engineVerified bool toolStarted bool toolStartAttempted, engineStartAttempted bool }`.
- **`(*workerPairSupervisor).AssemblePair`** — AssemblePair may only bind a live workspace-aware supervisor to its exact private audited session. It does not expose a public launch or relax admission. The caller owns session lifetime and must close it when the supervisor ends.
  Signature: `func(session *toolbroker.ToolSession, engine workerEngineLaunch, bundle io.ReadCloser) (*workerAssembledPair, error)`.
  Calls: `context.AfterFunc`, `engine.Scope`, `errors.New`, `input.Close`, `s.available`, `s.mu.Lock`, `s.mu.Unlock`, `session.Close`, `session.Scope`, `validWorkerAuditScope`.
- **Type `workerPairBundle`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { io.ReadCloser once sync.Once err error }`.
- **`(*workerPairBundle).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `b.ReadCloser.Close`, `b.once.Do`.
- **`(*workerAssembledPair).context`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (context.Context, func(), error)`.
  Calls: `a.ctx.Err`, `cancel`, `context.AfterFunc`, `context.WithDeadline`, `ctx.Err`, `errors.New`, `stop`.
- **`(*workerAssembledPair).identity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v workerPairIdentity) bool`.
- **`(*workerAssembledPair).WorkspaceIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (toolbroker.WorkspaceVolumeIdentity, error)`.
  Calls: `a.context`, `a.tool.WorkspaceIntent`, `done`.
- **`(*workerAssembledPair).PrepareVolume`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent toolbroker.WorkspaceVolumeIdentity) (toolbroker.WorkspaceVolumeIdentity, error)`.
  Calls: `a.context`, `a.tool.PrepareVolume`, `done`.
- **`(*workerAssembledPair).Create`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, identity workerPairIdentity) (string, error)`.
  Calls: `a.context`, `a.engine.Create`, `a.identity`, `a.mu.Lock`, `a.mu.Unlock`, `a.tool.Create`, `child.Err`, `done`, `errors.New`, `validPairContainerID`.
- **`(*workerAssembledPair).Verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, identity workerPairIdentity, id string) error`.
  Calls: `a.context`, `a.engine.Verify`, `a.identity`, `a.mu.Lock`, `a.mu.Unlock`, `a.tool.VerifyCreated`, `child.Err`, `done`, `errors.New`, `validPairContainerID`.
- **`(*workerAssembledPair).Inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) ([]byte, error)`.
  Calls: `a.context`, `a.engine.Inspect`, `a.mu.Lock`, `a.mu.Unlock`, `a.tool.Inspect`, `done`, `errors.New`, `validPairContainerID`.
- **`(*workerAssembledPair).Start`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `a.context`, `a.engine.Start`, `a.mu.Lock`, `a.mu.Unlock`, `a.tool.Start`, `child.Err`, `done`, `errors.New`, `validPairContainerID`.
- **`(*workerAssembledPair).Bootstrap`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (workspacebootstrap.Receipt, error)`.
  Calls: `a.bundle.Close`, `a.context`, `a.mu.Lock`, `a.mu.Unlock`, `a.tool.Bootstrap`, `child.Err`, `done`, `errors.New`, `validPairContainerID`, `workspacebootstrap.ValidReceipt`.
