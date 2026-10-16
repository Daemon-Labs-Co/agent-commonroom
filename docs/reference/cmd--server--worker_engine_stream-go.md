# cmd/server/worker_engine_stream.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_engine_stream.go)

Go implementation for worker engine stream; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(dockerCLIRunner).OpenEngine`** — OpenEngine attaches only to a checked immutable ID. Rootless daemon selection is revalidated per invocation, as for Run. No host provider, shell, restart or stderr forwarding is allowed. Closing this CLI proves no container absence; the supervisor must still fence both roles through its normal shutdown path.
  Signature: `func(ctx context.Context, id string) (io.ReadWriteCloser, error)`.
  Calls: `ctx.Err`, `exec.CommandContext`, `openEngineCommand`, `os.Getenv`, `repositoryDigestPattern.MatchString`, `verifiedLocalDockerContext`.
- **`openEngineCommand`** — The factory is internal trusted process policy, injected only by synthetic process tests. Never accept an executable or arguments from a model/request.
  Signature: `func(ctx context.Context, command func(context.Context) *exec.Cmd) (io.ReadWriteCloser, error)`.
  Calls: `cancel`, `close`, `cmd.Start`, `cmd.StdinPipe`, `cmd.Wait`, `command`, `context.WithCancel`, `ctx.Err`, `in.Close`, `make`, `os.Pipe`, `out.Close`, `outputWriter.Close`.
- **Type `engineCommandStream`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { in io.WriteCloser out io.ReadCloser cancel context.CancelFunc done chan struct{} once sync.Once }`.
- **`(*engineCommandStream).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `s.out.Read`.
- **`(*engineCommandStream).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `s.in.Write`.
- **`(*engineCommandStream).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `s.cancel`, `s.in.Close`, `s.once.Do`, `s.out.Close`.
