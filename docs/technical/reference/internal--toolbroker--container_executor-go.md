# internal/toolbroker/container_executor.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/container_executor.go)

Go implementation for container executor; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `DockerRunner`**: DockerRunner is a trusted isolated-worker dependency. It must invoke only the approved local rootless Docker daemon without shell parsing, inherited provider credentials or context overrides; stdout/stderr must honor writer errors and cancellation. Tests inject mocks. No real runner or automatic fallback exists. A source fork CAN supply a host/shell runner, BUT breaks the execution boundary.
  Declaration: `interface { Run(context.Context, []string, []byte, io.Writer) error }`.
- **Type `ContainerExecutor`**: ContainerExecutor adapts audited broker actions to a fixed container-side file helper or fixed terminal Go profile. It starts no container and requires a separately initialized tool pair. Use only through Dispatcher: this executor is not an audit/authentication API.
  Declaration: `struct { plan *ToolContainerPlan runner DockerRunner dockerRoot string }`.
- **`NewContainerExecutor`** — NewContainerExecutor binds trusted policy, daemon data root and explicit runner. The future supervisor must verify the rootless daemon and never source these dependencies from a model/plugin request. There is no default Docker context.
  Signature: `func(plan *ToolContainerPlan, runner DockerRunner, dockerRoot string) (*ContainerExecutor, error)`.
  Calls: `len`, `path.Clean`, `path.IsAbs`.
- **`(*ContainerExecutor).Execute`** — Execute verifies reported runtime/volume policy before invoking the fixed Node helper via immutable container ID. Action JSON contains no capability or Docker launch options. Failed/ambiguous work fences authority and requires tool-container absence proof; the future supervisor must additionally stop/verify the engine.
  Signature: `func(ctx context.Context, scope Scope, action Action, out io.Writer) error`.
  Calls: `append`, `cancel`, `cancelAuthority`, `context.AfterFunc`, `context.WithCancel`, `ctx.Deadline`, `ctx.Err`, `deadline.After`, `decodeInspection`, `e.capture`, `e.executeCommand`, `e.failed`, `e.plan.RunArgs`, `e.plan.VerifyInspection`, `e.plan.WorkspaceVolume`, `e.plan.authority.Revoke`, `e.plan.authority.executionContext`, `e.plan.authority.permitsAction`, `e.plan.commandUsed.CompareAndSwap`, `e.plan.commandUsed.Load`, `e.plan.workspace.Load`, `e.plan.workspaceReady`, `e.runner.Run`, `execCtx.Err`, `json.Marshal`, `len`, `out.Write`, `output.snapshot`, `ready.active`, `scope.Deadline.UTC`, `scope.Deadline.UTC().Round`, `stop`, `time.Now`, `toolcommand.Valid`, `validHelperResult`.
- **`(*ContainerExecutor).capture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, limit int) ([]byte, error)`.
  Calls: `cancel`, `child.Err`, `context.WithCancel`, `e.runner.Run`, `output.snapshot`.
- **`(*ContainerExecutor).failed`** — failed never retries the tool action. Cleanup gets a separate bounded context because shutdown must still be attempted after the execution deadline expires.
  Signature: `func(id string, overflow bool) error`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `e.plan.authority.Revoke`, `e.stopAndConfirm`.
- **`(*ContainerExecutor).stopAndConfirm`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `StopContainer`.
