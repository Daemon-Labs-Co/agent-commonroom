# cmd/server/worker_audit_session.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_audit_session.go)

Go implementation for worker audit session; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `watchedWorkerAuditLink`**: watchedWorkerAuditLink has one read pump. AuditClient remains the sole frame consumer; the pump only forwards bounded chunks through an unbuffered pipe. It observes idle EOF without competing with acknowledgement reads. Any link loss revokes the original supervisor lifetime, never releases its leases or claims container shutdown. This is not peer authentication: construction MUST follow the supervisor-owned private transport handshake. A source fork CAN replace the link/pump, BUT invalidates mandatory durable-audit guarantees.
  Declaration: `struct { raw io.ReadWriteCloser reader *io.PipeReader writer *io.PipeWriter failed func() once sync.Once done chan struct{} }`.
- **`newWatchedWorkerAuditLink`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, raw io.ReadWriteCloser, failed func()) *watchedWorkerAuditLink`.
  Calls: `close`, `context.AfterFunc`, `io.CopyBuffer`, `io.Pipe`, `link.closeIO`, `make`, `stop`.
- **`(*watchedWorkerAuditLink).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `l.reader.Read`.
- **`(*watchedWorkerAuditLink).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `l.raw.Write`.
- **`(*watchedWorkerAuditLink).closeIO`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `l.failed`, `l.once.Do`, `l.raw.Close`, `l.reader.CloseWithError`, `l.writer.CloseWithError`.
- **`(*watchedWorkerAuditLink).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `l.closeIO`.
- **`(*workerPairSupervisor).AcceptToolSession`** — AcceptToolSession consumes the one supervisor-owned authenticated socket and builds the actual shared tool session. Scope comes from the listener previously opened under this supervisor's lease, never from an engine callback. Trusted operator image/root policy and the approved Docker runner remain required. The private capability is returned once, not journalled or placed in stdout. Production launch must still supply a reviewed engine and fence both roles after any failure. Cancelling is NOT shutdown or permission to discard source.
  Signature: `func(ctx context.Context, listener *workerAuditListener, allowed []string, image, dockerRoot string, runner toolbroker.DockerRunner) (*toolbroker.ToolSession, string, error)`.
  Calls: `context.WithValue`, `ctx.Err`, `lifetime.Err`, `listener.Accept`, `listener.Close`, `newWatchedWorkerAuditLink`, `newWorkerToolSession`, `s.available`, `s.cancel`, `s.mu.Lock`, `s.mu.Unlock`, `session.Close`, `time.Now`, `validWorkerAuditScope`, `watched.Close`.
