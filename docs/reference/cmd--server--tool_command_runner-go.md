# cmd/server/tool_command_runner.go

[Module index](cmd.md) · [Open source](../../cmd/server/tool_command_runner.go)

Go implementation for tool command runner; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(dockerCLIRunner).RunCommand`** — RunCommand executes only worker-built docker arguments. The Go compiler and repository code run INSIDE the inspected tool container, never on this host. Source forks CAN substitute executables/argv, invalidating mandatory isolation. Nonzero CLI status is deliberately not certified as a compiler exit status.
  Signature: `func(ctx context.Context, args []string, out io.Writer) (int, error)`.
  Calls: `ctx.Err`, `exec.CommandContext`, `len`, `runDockerCommand`, `verifiedLocalDockerContext`.
- **`runDockerCommand`** — Factory injection is process-test infrastructure, never public configuration.
  Signature: `func(ctx context.Context, args []string, out io.Writer, factory func(context.Context, []string) *exec.Cmd) (int, error)`.
  Calls: `cmd.ProcessState.ExitCode`, `cmd.Run`, `ctx.Err`, `errors.As`, `factory`, `os.Getenv`.
