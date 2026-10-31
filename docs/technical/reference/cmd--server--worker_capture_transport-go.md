# cmd/server/worker_capture_transport.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_transport.go)

Go implementation for worker capture transport; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureAuthorityRequest`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Sequence uint64 'json:"sequence"' Operation string 'json:"operation"' Intent workerCaptureIntent 'json:"intent"' Event workerProbeEvent 'json:"event"' Record *workerCaptureRecord 'json:"record"' }`.
- **Type `captureAuthorityResponse`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Sequence uint64 'json:"sequence"' Accepted bool 'json:"accepted"' }`.
- **`writeCaptureAuthorityFrame`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w io.Writer, value any) error`.
  Calls: `binary.BigEndian.PutUint32`, `json.Marshal`, `len`, `uint32`, `w.Write`.
- **`readCaptureAuthorityFrame`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r io.Reader, value any) error`.
  Calls: `binary.BigEndian.Uint32`, `bytes.Equal`, `int`, `io.ReadFull`, `json.Marshal`, `json.Unmarshal`, `make`.
- **`captureAuthorityBinding`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerJobRecord, i workerCaptureIntent) bool`.
  Calls: `retainedObservationRecord`, `validWorkerCaptureRecord`.
- **`captureAuthorityEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(e workerProbeEvent) bool`.
  Calls: `probeHex`.
- **`captureAuthorityRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerCaptureRecord, i workerCaptureIntent, e workerProbeEvent) bool`.
  Calls: `captureAuthorityEvent`, `captureIntentEqual`, `len`, `validWorkerCaptureRecord`.
- **Type `captureAuthorityClient`**: Any uncertain exchange poisons this one-shot stream and stops local authority. The supplied stop callback must be local and nonblocking, never reenter the link.
  Declaration: `struct { mu sync.Mutex stream io.ReadWriteCloser record workerJobRecord intent workerCaptureIntent sequence uint64 failed bool stop func() once sync.Once notify sync.Once lifetime context.Context cancel context.CancelFunc stopped atomic.Bool }`.
- **`newCaptureAuthorityClient`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(stream io.ReadWriteCloser, r workerJobRecord, i workerCaptureIntent, stop func()) (*captureAuthorityClient, error)`.
  Calls: `captureAuthorityBinding`, `context.Background`, `context.WithCancel`, `copyCaptureIntent`, `detachedProbeWorkerRecord`, `newWatchedWorkerAuditLink`, `nilWorkerObservationDependency`.
- **`(*captureAuthorityClient).markStopped`** — The sole bounded raw-stream pump revokes idle authority on EOF/error. This callback must never join its own pump or reenter the private control stream.
  Signature: `func()`.
  Calls: `c.cancel`, `c.notify.Do`, `c.stopped.Store`.
- **`(*captureAuthorityClient).CaptureAuthorityLifetime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() context.Context`.
- **`(*captureAuthorityClient).StopCaptureAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `c.markStopped`, `c.once.Do`, `c.stream.Close`.
- **`(*captureAuthorityClient).exchange`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, operation string, e workerProbeEvent, r *workerCaptureRecord) error`.
  Calls: `c.StopCaptureAuthority`, `c.mu.Lock`, `c.mu.Unlock`, `c.stopped.Load`, `cancel`, `context.AfterFunc`, `context.WithTimeout`, `ctx.Err`, `fail`, `readCaptureAuthorityFrame`, `stopIO`, `stopLifetime`, `writeCaptureAuthorityFrame`.
- **`(*captureAuthorityClient).AuthorizeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord, i workerCaptureIntent, e workerProbeEvent) error`.
  Calls: `bytes.Equal`, `c.StopCaptureAuthority`, `c.exchange`, `captureAuthorityEvent`, `captureIntentEqual`, `json.Marshal`.
- **`(*captureAuthorityClient).AcknowledgeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerCaptureRecord, e workerProbeEvent) error`.
  Calls: `c.StopCaptureAuthority`, `c.exchange`, `captureAuthorityRecord`, `copyCaptureRecord`.
- **`serveCaptureAuthority`** — The guard supplies authenticated current permission, physical/quota evidence and durable controller acknowledgment. It must never reenter the worker admission lock or call this stream while the worker is waiting for its response. No guard is synthesized here. Session failure revokes controller authority exactly once.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, record workerJobRecord, intent workerCaptureIntent, guard workerCaptureJournalGuard) error`.
  Calls: `cancel`, `captureAuthorityBinding`, `captureAuthorityEvent`, `captureAuthorityRecord`, `captureIntentEqual`, `context.AfterFunc`, `context.WithTimeout`, `copyCaptureIntent`, `copyCaptureRecord`, `ctx.Err`, `detachedProbeWorkerRecord`, `guard.AcknowledgeCapture`, `guard.AuthorizeCapture`, `nilWorkerObservationDependency`, `once.Do`, `readCaptureAuthorityFrame`, `stop`, `stopIO`, `stream.Close`, `uint64`, `writeCaptureAuthorityFrame`.
