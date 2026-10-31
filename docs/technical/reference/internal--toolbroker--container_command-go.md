# internal/toolbroker/container_command.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/container_command.go)

Go implementation for container command; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `DockerCommandRunner`**: DockerCommandRunner captures combined command stdout/stderr and the observed Docker CLI exit status. Zero indicates successful exec completion; nonzero may be a command failure OR a client/daemon error and MUST remain unconfirmed. It runs only approved rootless Docker, never a host compiler/shell/provider. No model may supply this adapter. Source replacement violates isolation.
  Declaration: `interface { DockerRunner RunCommand(context.Context, []string, io.Writer) (int, error) }`.
- **Type `isolatedCommandAdapter`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { runner DockerCommandRunner id string }`.
- **`(isolatedCommandAdapter).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, spec toolcommand.Spec, out io.Writer) (int, error)`.
  Calls: `a.runner.Run`, `a.runner.RunCommand`, `append`.
- **`(*ContainerExecutor).executeCommand`** — executeCommand observes the receipt in trusted worker memory, not a process alongside untrusted tests. It always consumes the tool container: no result is released before verified exact-ID absence, including escaped/background child processes. The supervisor must still reconcile both roles and retain source.
  Signature: `func(ctx context.Context, id string, action Action, out io.Writer) error`.
  Calls: `context.Background`, `context.WithTimeout`, `ctx.Err`, `done`, `e.failed`, `e.plan.RunArgs`, `e.plan.authority.Revoke`, `e.stopAndConfirm`, `errors.Is`, `json.Marshal`, `len`, `out.Write`, `time.Now`, `toolcommand.Run`, `validCommandReceipt`.
