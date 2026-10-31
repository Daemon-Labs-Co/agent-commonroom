# cmd/server/worker_object_read_transport.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_object_read_transport.go)

Go implementation for worker object read transport; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadAuthorityRequest`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Protocol string 'json:"protocol"' Version int 'json:"version"' Sequence uint64 'json:"sequence"' Operation string 'json:"operation"' Intent workerObjectReadIntent 'json:"intent"' Event workerProbeEvent 'json:"event"' Record *workerObjectReadRecord 'json:"record"' }`.
- **Type `objectReadAuthorityResponse`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Protocol string 'json:"protocol"' Version int 'json:"version"' Sequence uint64 'json:"sequence"' Accepted bool 'json:"accepted"' }`.
- **`writeObjectReadAuthorityFrame`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w io.Writer, value any) error`.
  Calls: `binary.BigEndian.PutUint32`, `json.Marshal`, `len`, `uint32`, `w.Write`.
- **`readObjectReadAuthorityFrame`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r io.Reader, value any) error`.
  Calls: `binary.BigEndian.Uint32`, `bytes.Equal`, `int`, `io.ReadFull`, `json.Marshal`, `json.Unmarshal`, `make`.
- **`objectReadAuthorityBinding`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerJobRecord, i workerObjectReadIntent) bool`.
  Calls: `retainedObservationRecord`, `validWorkerObjectReadRecord`.
- **`objectReadAuthorityEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(e workerProbeEvent) bool`.
  Calls: `captureAuthorityEvent`.
- **Type `objectReadAuthorityClient`**: The object-read-v1 discriminator prevents cross-protocol frame substitution. Any uncertain exchange poisons this one-shot stream and stops local authority. The supplied stop callback must be local and nonblocking, never reenter the link.
  Declaration: `struct { mu sync.Mutex stream io.ReadWriteCloser record workerJobRecord intent workerObjectReadIntent sequence uint64 failed bool stop func() once sync.Once notify sync.Once lifetime context.Context cancel context.CancelFunc stopped atomic.Bool }`.
- **`newObjectReadAuthorityClient`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(stream io.ReadWriteCloser, r workerJobRecord, i workerObjectReadIntent, stop func()) (*objectReadAuthorityClient, error)`.
  Calls: `context.Background`, `context.WithCancel`, `detachedProbeWorkerRecord`, `newWatchedWorkerAuditLink`, `nilWorkerObservationDependency`, `objectReadAuthorityBinding`.
- **`(*objectReadAuthorityClient).markStopped`** — EOF must revoke authority even between exchanges. The existing bounded private read pump is the sole raw-stream reader; it cannot race frame consumers. Its callback never joins its own pump or reenters the client.
  Signature: `func()`.
  Calls: `c.cancel`, `c.notify.Do`, `c.stopped.Store`.
- **`(*objectReadAuthorityClient).ObjectReadAuthorityLifetime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() context.Context`.
- **`(*objectReadAuthorityClient).StopObjectReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `c.markStopped`, `c.once.Do`, `c.stream.Close`.
- **`(*objectReadAuthorityClient).exchange`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, operation string, e workerProbeEvent, r *workerObjectReadRecord) error`.
  Calls: `c.StopObjectReadAuthority`, `c.mu.Lock`, `c.mu.Unlock`, `c.stopped.Load`, `cancel`, `context.AfterFunc`, `context.WithTimeout`, `ctx.Err`, `fail`, `readObjectReadAuthorityFrame`, `stopIO`, `writeObjectReadAuthorityFrame`.
- **`(*objectReadAuthorityClient).AuthorizeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord, i workerObjectReadIntent, e workerProbeEvent) error`.
  Calls: `bytes.Equal`, `c.StopObjectReadAuthority`, `c.exchange`, `json.Marshal`, `objectReadAuthorityEvent`.
- **`(*objectReadAuthorityClient).AcknowledgeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerObjectReadRecord, e workerProbeEvent) error`.
  Calls: `c.StopObjectReadAuthority`, `c.exchange`, `copyObjectReadRecord`, `objectReadAuthorityRecord`.
- **`serveObjectReadAuthority`** — The guard supplies authenticated current permission, physical/quota evidence and durable controller acknowledgment. It must never reenter the worker admission lock or call this stream while the worker is waiting for its response. No guard is synthesized here. Session failure revokes controller authority exactly once.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, record workerJobRecord, intent workerObjectReadIntent, guard workerObjectReadJournalGuard) error`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithTimeout`, `copyObjectReadRecord`, `ctx.Err`, `detachedProbeWorkerRecord`, `guard.AcknowledgeObjectRead`, `guard.AuthorizeObjectRead`, `nilWorkerObservationDependency`, `objectReadAuthorityBinding`, `objectReadAuthorityEvent`, `objectReadAuthorityRecord`, `once.Do`, `readObjectReadAuthorityFrame`, `stop`, `stopIO`, `stream.Close`, `uint64`, `writeObjectReadAuthorityFrame`.
