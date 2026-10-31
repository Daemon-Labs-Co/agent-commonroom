# cmd/server/capture_authenticated_session.go

[Module index](cmd.md) · [Open source](../../../cmd/server/capture_authenticated_session.go)

Go implementation for capture authenticated session; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `capturePeerBinding`**: Certificate policy is shared with readers; protocol and operation grants are NEVER shared. Independent trusted setup pins exact worker/record/intent values. No request/model metadata can provision an identity or choose another protocol.
  Declaration: `struct { Identity objectReadPeerIdentity Worker, RecordDigest string Intent workerCaptureIntent }`.
- **`(capturePeerBinding).matches`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerJobRecord, i workerCaptureIntent) bool`.
  Calls: `captureAuthorityBinding`, `captureIntentEqual`, `probeWorkerRecordDigest`, `workerAliasPattern.MatchString`.
- **Type `captureSessionHello`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Protocol string 'json:"protocol"' Version int 'json:"version"' Worker string 'json:"worker"' RecordDigest string 'json:"recordDigest"' Intent workerCaptureIntent 'json:"intent"' }`.
- **`(capturePeerBinding).hello`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() captureSessionHello`.
  Calls: `copyCaptureIntent`.
- **`sameCaptureHello`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a, b captureSessionHello) bool`.
  Calls: `captureIntentEqual`.
- **`authenticateBoundCaptureStream`** — Consumes raw on every path. Mutual certificate chain, DNS and leaf-pin checks and the capture-only ALPN precede finite hello agreement. No listener, discovery, plaintext retry, host/SSH execution or peer-selected capability is introduced.
  Signature: `func(ctx context.Context, raw net.Conn, binding capturePeerBinding, server bool) (*objectReadControlStream, error)`.
  Calls: `authenticatePinnedWorkerStream`, `binding.hello`, `context.AfterFunc`, `copyCaptureIntent`, `ctx.Deadline`, `ctx.Err`, `d.Before`, `nilWorkerObservationDependency`, `raw.Close`, `raw.SetDeadline`, `readCaptureAuthorityFrame`, `refuse`, `repositoryDigestPattern.MatchString`, `sameCaptureHello`, `stop`, `time.Now`, `time.Now().Add`, `validWorkerCaptureRecord`, `workerAliasPattern.MatchString`, `writeCaptureAuthorityFrame`.
- **Type `authenticatedCaptureClient`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *captureAuthorityClient stopParent func() bool }`.
- **`(*authenticatedCaptureClient).StopCaptureAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `c.captureAuthorityClient.StopCaptureAuthority`, `c.stopParent`.
- **`newAuthenticatedCaptureClient`** — The original setup context remains authoritative while idle. The caller owns the returned client and must stop/join its pump and controller serving call.
  Signature: `func(ctx context.Context, raw net.Conn, b capturePeerBinding, r workerJobRecord, i workerCaptureIntent, stop func()) (*authenticatedCaptureClient, error)`.
  Calls: `authenticateBoundCaptureStream`, `b.matches`, `context.AfterFunc`, `copyCaptureIntent`, `ctx.Err`, `newCaptureAuthorityClient`, `nilWorkerObservationDependency`, `owned.StopCaptureAuthority`, `raw.Close`, `stream.Close`.
- **`serveAuthenticatedCaptureSession`** — Trusted setup supplies the ORIGINAL human/worker/creator-pinned session. TLS identity and hello agreement do not supply physical evidence or permissions.
  Signature: `func(ctx context.Context, raw net.Conn, b capturePeerBinding, g *controllerCaptureSession) error`.
  Calls: `authenticateBoundCaptureStream`, `b.matches`, `cancel`, `context.AfterFunc`, `context.WithCancel`, `copyCaptureIntent`, `ctx.Err`, `g.StopCaptureAuthority`, `g.lifetime.Err`, `nilWorkerObservationDependency`, `raw.Close`, `serveCaptureAuthority`, `stop`, `stream.Close`.
