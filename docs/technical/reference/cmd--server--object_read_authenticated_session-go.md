# cmd/server/object_read_authenticated_session.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_authenticated_session.go)

Go implementation for object read authenticated session; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadPeerBinding`**: Provisioning binds the authenticated peer certificate to exactly one worker, complete retained record and finite read intent. This binding is trusted local configuration, never filled from a received frame. TLS identity alone does not grant account, creator, physical-quota or artifact-export authority.
  Declaration: `struct { Identity objectReadPeerIdentity Worker string RecordDigest string Intent workerObjectReadIntent }`.
- **`(objectReadPeerBinding).matches`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerJobRecord, i workerObjectReadIntent) bool`.
  Calls: `objectReadAuthorityBinding`, `probeWorkerRecordDigest`, `workerAliasPattern.MatchString`.
- **Type `objectReadSessionHello`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Protocol string 'json:"protocol"' Version int 'json:"version"' Worker string 'json:"worker"' RecordDigest string 'json:"recordDigest"' Intent workerObjectReadIntent 'json:"intent"' }`.
- **`(objectReadPeerBinding).hello`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() objectReadSessionHello`.
- **Type `objectReadControlStream`**: Close aborts the underlying connection before joining the authority read pump; TLS close-notify cannot delay cancellation or stand in for durable cleanup. The complete framed acknowledgments, never EOF, establish journal transitions.
  Declaration: `struct { *tls.Conn once sync.Once }`.
- **`(*objectReadControlStream).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `s.Conn.Close`, `s.NetConn`, `s.NetConn().Close`, `s.once.Do`.
- **`authenticateBoundObjectReadStream`** — authenticateBoundObjectReadStream consumes raw on every path. Both endpoints must independently agree on the provisioned binding before authority frames. No listener/discovery, model-selected identity, host/SSH fallback or off switch. A source fork bypassing this setup invalidates the authentication contract.
  Signature: `func(ctx context.Context, raw net.Conn, b objectReadPeerBinding, server bool) (*objectReadControlStream, error)`.
  Calls: `authenticateObjectReadStream`, `b.hello`, `context.AfterFunc`, `ctx.Deadline`, `ctx.Err`, `d.Before`, `len`, `nilWorkerObservationDependency`, `raw.Close`, `raw.SetDeadline`, `readObjectReadAuthorityFrame`, `refuse`, `stop`, `time.Now`, `time.Now().Add`, `workerAliasPattern.MatchString`, `writeObjectReadAuthorityFrame`.
- **Type `authenticatedObjectReadClient`**: The worker owns the returned client and must StopObjectReadAuthority to join its read pump on every outcome. Its original setup context remains authoritative while idle, not just during a protocol call. Closure cannot release capacity.
  Declaration: `struct { *objectReadAuthorityClient stopParent func() bool }`.
- **`(*authenticatedObjectReadClient).StopObjectReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `c.objectReadAuthorityClient.StopObjectReadAuthority`, `c.stopParent`.
- **`newAuthenticatedObjectReadClient`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, raw net.Conn, b objectReadPeerBinding, r workerJobRecord, i workerObjectReadIntent, stop func()) (*authenticatedObjectReadClient, error)`.
  Calls: `authenticateBoundObjectReadStream`, `b.matches`, `context.AfterFunc`, `ctx.Err`, `newObjectReadAuthorityClient`, `nilWorkerObservationDependency`, `owned.StopObjectReadAuthority`, `raw.Close`, `stream.Close`.
- **`serveAuthenticatedObjectReadSession`** — Trusted controller setup supplies the already human/worker/creator-pinned session. The serving call owns and joins this single connection; any exit revokes only this session. No evidence, permissions or account identity comes from TLS frames.
  Signature: `func(ctx context.Context, raw net.Conn, b objectReadPeerBinding, g *controllerObjectReadSession) error`.
  Calls: `authenticateBoundObjectReadStream`, `b.matches`, `cancel`, `context.AfterFunc`, `context.WithCancel`, `ctx.Err`, `g.StopObjectReadAuthority`, `g.lifetime.Err`, `nilWorkerObservationDependency`, `raw.Close`, `serveObjectReadAuthority`, `stop`, `stream.Close`.
