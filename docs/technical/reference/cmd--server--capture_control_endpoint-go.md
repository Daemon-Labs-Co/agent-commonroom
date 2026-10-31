# cmd/server/capture_control_endpoint.go

[Module index](cmd.md) · [Open source](../../../cmd/server/capture_control_endpoint.go)

Go implementation for capture control endpoint; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureControlEndpoint`**: An operator-provisioned literal IP:port, never a model/request-selected URL or address. DNS discovery, redirects, alternate endpoints and reconnect are absent. Identity/binding still independently pin the TLS peer and retained capture session. Exposing this constructor to model tools would violate the deployment contract.
  Declaration: `struct { Address string Binding capturePeerBinding }`.
- **`(captureControlEndpoint).valid`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerJobRecord, i workerCaptureIntent) bool`.
  Calls: `e.Binding.matches`, `validObjectReadControlAddress`, `validObjectReadPeerIdentity`.
- **`connectCaptureControl`** — The dialer and literal-address validator are shared transport primitives only; no reader grant or protocol is reused. connectCaptureControl uses only one approved TCP endpoint and consumes any resulting connection. Deployment supplies a net.Dialer; a source fork replacing it can bypass network policy, never TLS/session/permission checks. Dial has a five-second bound and follows the original setup context. No shell/SSH, host capture, plaintext retry or optional security flag exists.
  Signature: `func(ctx context.Context, e captureControlEndpoint, r workerJobRecord, i workerCaptureIntent, stop func(), dialer objectReadControlDialer) (*authenticatedCaptureClient, error)`.
  Calls: `cancel`, `context.WithTimeout`, `ctx.Err`, `dialCtx.Err`, `dialer.DialContext`, `e.valid`, `newAuthenticatedCaptureClient`, `nilWorkerObservationDependency`, `raw.Close`.
- **`serveCaptureControlOnce`** — serveCaptureControlOnce owns the already provisioned listener on entry and consumes exactly one connection. Close MUST unblock Accept (as net.TCPListener does). The listener is closed before TLS/session setup, so failure cannot accept a replacement peer or retry an uncertain session. The caller must join this call before relinquishing its creator lease. No permission or physical evidence is synthesized from the accepted socket; the pinned controller session is required.
  Signature: `func(ctx context.Context, listener net.Listener, b capturePeerBinding, g *controllerCaptureSession) error`.
  Calls: `acceptCtx.Err`, `b.matches`, `cancel`, `context.AfterFunc`, `context.WithTimeout`, `ctx.Err`, `g.StopCaptureAuthority`, `g.lifetime.Err`, `listener.Accept`, `listener.Close`, `nilWorkerObservationDependency`, `raw.Close`, `serveAuthenticatedCaptureSession`, `stopListener`, `stopSession`, `validObjectReadPeerIdentity`.
