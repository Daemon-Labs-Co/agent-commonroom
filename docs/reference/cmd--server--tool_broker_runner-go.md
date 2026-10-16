# cmd/server/tool_broker_runner.go

[Module index](cmd.md) · [Open source](../../cmd/server/tool_broker_runner.go)

Go implementation for tool broker runner; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `dockerCLIRunner`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(dockerCLIRunner).Run`** — Run verifies the same local-only, resource-enforcing Docker context required of the main agent container path (verifiedLocalDockerContext, isolation.go) before every invocation — this runner is called several times per tool action (inspect, inspect, exec, and on failure ps/rm/ps), and each one is a point where a tampered or remote Docker context must still be caught, not just the first. Checks are never cached as permission for a later call.
  Signature: `func(ctx context.Context, args []string, stdin []byte, out io.Writer) error`.
  Calls: `(dockerCLIRunner{}).RunInput`, `bytes.NewReader`, `len`.
- **`(dockerCLIRunner).RunInput`** — RunInput streams only worker-approved snapshot input; it applies the same per-invocation rootless checks and fixed environment as bounded tool actions.
  Signature: `func(ctx context.Context, args []string, input io.Reader, out io.Writer) error`.
  Calls: `cmd.Run`, `ctx.Err`, `errors.New`, `exec.CommandContext`, `len`, `os.Getenv`, `verifiedLocalDockerContext`.
