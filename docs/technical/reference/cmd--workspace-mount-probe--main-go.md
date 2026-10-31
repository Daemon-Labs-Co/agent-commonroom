# cmd/workspace-mount-probe/main.go

[Module index](cmd.md) · [Open source](../../../cmd/workspace-mount-probe/main.go)

Go implementation for main; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, out io.Writer, observe func(context.Context, string) (workspacemount.Proof, error)) error`.
  Calls: `ctx.Err`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `observe`, `workspacemount.ValidChallenge`, `workspacemount.ValidProof`.
- **`main`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `fmt.Fprintln`, `os.Exit`, `run`, `signal.NotifyContext`, `stop`.
