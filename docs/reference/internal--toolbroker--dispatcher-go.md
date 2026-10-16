# internal/toolbroker/dispatcher.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/dispatcher.go)

Go implementation for dispatcher; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Action`**: Action contains only admitted workspace input, never capability or launch options. Only an isolated credential-free executor may implement its operations. A source fork that supplies a host executor bypasses the required security boundary.
  Declaration: `struct { Operation string Path string Content string Command []string }`.
- **Type `Intent`**: Intent correlates the exact action digest with trusted scope and call ID. Audit implementations must durably persist it before acknowledging; returning nil without persistence invalidates the security contract. Raw tokens are excluded.
  Declaration: `struct { Scope Scope 'json:"scope"' CallID string 'json:"callId"' Operation string 'json:"operation"' PayloadDigest string 'json:"payloadDigest"' }`.
- **Type `Outcome`**: Outcome is evidence metadata, not raw tool output. Success is acknowledged only after completion persistence; failure may represent an ambiguous side effect.
  Declaration: `struct { Status string 'json:"status"' OutputBytes int 'json:"outputBytes"' OutputDigest string 'json:"outputDigest"' }`.
- **Type `Auditor`**: Auditor is the trusted controller persistence boundary. Implementations must honor context cancellation and must never insert capability or provider secrets.
  Declaration: `interface { RecordIntent(context.Context, Intent) error RecordOutcome(context.Context, Intent, Outcome) error }`.
- **Type `Executor`**: Executor must honor cancellation and execute exclusively inside the separate tool container. This package supplies no host, Docker or provider implementation. The supervisor must prove process/container shutdown; context cancellation alone cannot terminate an implementation that ignores it.
  Declaration: `interface { Execute(context.Context, Scope, Action, io.Writer) error }`.
- **Type `Dispatcher`**: Dispatcher gates injected work; it is not connected to any production endpoint. All dispatchers sharing an authority also share its single active-work slot.
  Declaration: `struct { authority *Authority auditor Auditor executor Executor }`.
- **`NewDispatcher`** — NewDispatcher requires explicit dependencies, without automatic fallback.
  Signature: `func(a *Authority, audit Auditor, exec Executor) (*Dispatcher, error)`.
- **`(*Dispatcher).Dispatch`** — Dispatch consumes an admitted call, persists intent, runs bounded work and persists its outcome. Repeated/ambiguous work is never automatically retried. The caller owns input until invocation and must not concurrently mutate it.
  Signature: `func(parent context.Context, request Request) ([]byte, error)`.
  Calls: `append`, `cancel`, `ctx.Err`, `d.auditor.RecordIntent`, `d.authority.Reserve`, `d.authority.Revoke`, `d.authority.executionContext`, `d.executor.Execute`, `d.recordOutcome`, `hex.EncodeToString`, `json.Marshal`, `output.snapshot`, `parent.Err`, `sha256.Sum256`, `time.Now`, `validRequest`.
- **Type `boundedOutput`**: boundedOutput refuses oversized writes, cancels work and never buffers beyond the cap. Executors must stop writing before returning; the mutex supports parallel stdout/stderr writes without an unsynchronized buffer.
  Declaration: `struct { mu sync.Mutex buffer bytes.Buffer overflow bool closed bool maxBytes int cancel context.CancelFunc }`.
- **`(*boundedOutput).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `len`, `w.buffer.Len`, `w.buffer.Write`, `w.cancel`, `w.mu.Lock`, `w.mu.Unlock`.
- **`(*boundedOutput).snapshot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() ([]byte, bool)`.
  Calls: `append`, `w.buffer.Bytes`, `w.mu.Lock`, `w.mu.Unlock`.
- **`(*Dispatcher).recordOutcome`** — recordOutcome withholds success and fences authority if persistence cannot acknowledge. Canceled audit contexts may leave an unresolved durable intent; recovery must never automatically replay such a call.
  Signature: `func(ctx context.Context, intent Intent, status string, raw []byte) error`.
  Calls: `d.auditor.RecordOutcome`, `d.authority.Revoke`, `hex.EncodeToString`, `len`, `sha256.Sum256`.
