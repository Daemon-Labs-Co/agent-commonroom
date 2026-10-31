# cmd/server/worker_probe_transport.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_probe_transport.go)

Go implementation for worker probe transport; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `probeAuthorityRequest`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Sequence uint64 'json:"sequence"' Operation string 'json:"operation"' Intent toolbroker.WorkspaceProbeIntent 'json:"intent"' Event workerProbeEvent 'json:"event"' }`.
- **Type `probeAuthorityResponse`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Sequence uint64 'json:"sequence"' Accepted bool 'json:"accepted"' }`.
- **`writeProbeAuthorityFrame`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w io.Writer, value any) error`.
  Calls: `binary.BigEndian.PutUint32`, `json.Marshal`, `len`, `uint32`, `w.Write`.
- **`readProbeAuthorityFrame`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r io.Reader, value any) error`.
  Calls: `binary.BigEndian.Uint32`, `bytes.Equal`, `int`, `io.ReadFull`, `json.Marshal`, `json.Unmarshal`, `make`.
- **Type `probeAuthorityClient`**: probeAuthorityClient serializes a one-shot link, so responses cannot acknowledge another request. Any denial, malformed frame or uncertain I/O poisons the stream and invokes the supplied local authority stop exactly once. Cancellation closes the owned stream, unblocking both peers; callers must not share it with other work.
  Declaration: `struct { mu sync.Mutex stream io.ReadWriteCloser intent toolbroker.WorkspaceProbeIntent record workerJobRecord sequence uint64 failed bool stop func() once sync.Once }`.
- **`newProbeAuthorityClient`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(stream io.ReadWriteCloser, record workerJobRecord, intent toolbroker.WorkspaceProbeIntent, stop func()) (*probeAuthorityClient, error)`.
  Calls: `nilWorkerObservationDependency`, `retainedObservationRecord`, `validWorkerProbeIntent`.
- **`(*probeAuthorityClient).StopAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `c.once.Do`, `c.stop`, `c.stream.Close`.
- **`(*probeAuthorityClient).exchange`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, operation string, event workerProbeEvent) error`.
  Calls: `c.StopAuthority`, `c.mu.Lock`, `c.mu.Unlock`, `cancel`, `cancelIO`, `context.AfterFunc`, `context.WithTimeout`, `ctx.Err`, `readProbeAuthorityFrame`, `writeProbeAuthorityFrame`.
- **`(*probeAuthorityClient).Authorize`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error`.
  Calls: `c.exchange`.
- **`(*probeAuthorityClient).Acknowledge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `c.exchange`.
- **`(*probeAuthorityClient).AuthorizeWorkspaceObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord) error`.
  Calls: `bytes.Equal`, `c.exchange`, `json.Marshal`.
- **`serveProbeAuthority`** — serveProbeAuthority invokes an independently assembled trusted controller guard. Its authority comes from the authenticated session, never request-supplied actor, worker or room IDs. The fixed record/intent cannot be selected by the peer. Guard errors close this link without disclosing diagnostics. This transport does not implement physical evidence, create a FrozenLease or release source bytes.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, record workerJobRecord, intent toolbroker.WorkspaceProbeIntent, guard workerWorkspaceProbeGuard) error`.
  Calls: `cancel`, `cancelIO`, `context.AfterFunc`, `context.WithTimeout`, `ctx.Err`, `guard.Acknowledge`, `guard.Authorize`, `guard.AuthorizeWorkspaceObservation`, `nilWorkerObservationDependency`, `readProbeAuthorityFrame`, `retainedObservationRecord`, `stream.Close`, `uint64`, `validWorkerProbeIntent`, `writeProbeAuthorityFrame`.
