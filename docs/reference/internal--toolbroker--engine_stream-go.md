# internal/toolbroker/engine_stream.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/engine_stream.go)

Go implementation for engine stream; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `EngineStreamRunner`**: EngineStreamRunner attaches only protocol stdin/stdout to an existing immutable container ID on the same approved rootless daemon as Run. Implementations must discard stderr, disable detach keys and signal proxying, close blocked I/O on cancellation, and reap the attachment process on Close. It must never start, exec, adopt or reconnect a provider. A source replacement can violate this contract; this interface is not protection against a hostile worker host.
  Declaration: `interface { DockerRunner OpenEngine(context.Context, string) (io.ReadWriteCloser, error) }`.
- **`(*EngineContainerRuntime).Attach`** — Attach consumes one attachment attempt after rechecking the running policy. Failure is uncertain: callers must fence the job rather than reconnect. The returned stream belongs to the callback driver and remains lifetime-bound.
  Signature: `func(ctx context.Context, id string) (io.ReadWriteCloser, error)`.
  Calls: `child.Err`, `closeStream`, `context.AfterFunc`, `done`, `once.Do`, `owned.Close`, `r.context`, `r.ctx.Err`, `r.inspect`, `r.mu.Lock`, `r.mu.Unlock`, `r.original.Err`, `r.verify`, `runner.OpenEngine`, `stop`, `stream.Close`.
- **Type `engineStream`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { io.ReadWriteCloser ctx, original, lifetime context.Context close func() error }`.
- **`(*engineStream).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `clear`, `s.ReadWriteCloser.Read`, `s.ctx.Err`, `s.lifetime.Err`, `s.original.Err`.
- **`(*engineStream).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `s.ReadWriteCloser.Write`, `s.ctx.Err`, `s.lifetime.Err`, `s.original.Err`.
- **`(*engineStream).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `s.close`.
