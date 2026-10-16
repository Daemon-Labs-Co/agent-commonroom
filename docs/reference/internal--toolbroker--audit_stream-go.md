# internal/toolbroker/audit_stream.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/audit_stream.go)

Go implementation for audit stream; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `AuditClient`**: AuditClient implements Auditor across a private, authenticated controller link. The link MUST be separate from model/stdout traffic. Close must interrupt I/O; the pair supervisor must stop both containers after any channel failure. A source fork CAN feed self-acknowledgements, BUT invalidates durable admission.
  Declaration: `struct { channel io.ReadWriteCloser reader *bufio.Reader scope Scope mu sync.Mutex sequence int session string pending *Intent failed bool }`.
- **`NewAuditClient`** — NewAuditClient binds one authenticated, non-reusable link to a job and fresh random session nonce. It performs no peer authentication or model launch.
  Signature: `func(channel io.ReadWriteCloser, scope Scope) (*AuditClient, error)`.
  Calls: `bufio.NewReaderSize`, `hex.EncodeToString`, `rand.Read`, `time.Now`, `validScope`.
- **`(*AuditClient).RecordIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent Intent) error`.
  Calls: `ValidAuditIntent`, `client.exchange`, `client.fail`, `client.mu.TryLock`, `client.mu.Unlock`, `time.Now`.
- **`(*AuditClient).RecordOutcome`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent Intent, outcome Outcome) error`.
  Calls: `ValidAuditOutcome`, `client.exchange`, `client.fail`, `client.mu.TryLock`, `client.mu.Unlock`.
- **`(*AuditClient).fail`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `client.channel.Close`.
- **`(*AuditClient).exchange`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(parent context.Context, frame AuditFrame) error`.
  Calls: `auditFrameDigest`, `auditRead`, `auditWrite`, `cancel`, `client.channel.Close`, `client.fail`, `context.AfterFunc`, `context.WithDeadline`, `ctx.Err`, `parent.Err`, `stop`.
- **`ServeAudit`** — ServeAudit acknowledges only after the supplied controller auditor returns durable success. It rejects reordered/repeated calls and never invokes an executor. Scope and peer authentication come from trusted job registration, not a frame. EOF/ambiguity is failure; pending intents are not auto-replayed.
  Signature: `func(parent context.Context, channel io.ReadWriteCloser, scope Scope, auditor Auditor) error`.
  Calls: `ValidAuditIntent`, `ValidAuditOutcome`, `auditFrameDigest`, `auditRead`, `auditWrite`, `auditor.RecordIntent`, `auditor.RecordOutcome`, `bufio.NewReaderSize`, `cancel`, `channel.Close`, `context.AfterFunc`, `context.WithDeadline`, `ctx.Err`, `digest`, `stop`, `time.Now`, `validScope`.
