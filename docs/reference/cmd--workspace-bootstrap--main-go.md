# cmd/workspace-bootstrap/main.go

[Module index](cmd.md) · [Open source](../../cmd/workspace-bootstrap/main.go)

Go implementation for main; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `gitRunner`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ home string }`.
- **`(gitRunner).command`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, args ...string) *exec.Cmd`.
  Calls: `append`, `exec.CommandContext`.
- **`(gitRunner).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, args ...string) ([]byte, error)`.
  Calls: `cmd.Run`, `g.command`, `out.Bytes`.
- **Type `limitedWriter`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ bytes.Buffer }`.
- **`(*limitedWriter).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `len`, `w.Buffer.Write`, `w.Len`.
- **`main`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `fmt.Fprintln`, `len`, `os.Exit`, `os.MkdirTemp`, `os.RemoveAll`, `signal.NotifyContext`, `stop`, `workspacebootstrap.Run`.
