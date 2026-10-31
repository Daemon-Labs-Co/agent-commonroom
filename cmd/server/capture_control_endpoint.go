package main

import (
	"context"
	"net"
	"time"
)

// An operator-provisioned literal IP:port, never a model/request-selected URL or
// address. DNS discovery, redirects, alternate endpoints and reconnect are absent.
// Identity/binding still independently pin the TLS peer and retained capture session.
// Exposing this constructor to model tools would violate the deployment contract.
type captureControlEndpoint struct {
	Address string
	Binding capturePeerBinding
}

func (e captureControlEndpoint) valid(r workerJobRecord, i workerCaptureIntent) bool {
	return validObjectReadControlAddress(e.Address) && e.Binding.matches(r, i) && validObjectReadPeerIdentity(e.Binding.Identity)
}

// The dialer and literal-address validator are shared transport primitives only;
// no reader grant or protocol is reused.
// connectCaptureControl uses only one approved TCP endpoint and consumes any
// resulting connection. Deployment supplies a net.Dialer; a source fork replacing
// it can bypass network policy, never TLS/session/permission checks. Dial has a
// five-second bound and follows the original setup context. No shell/SSH, host
// capture, plaintext retry or optional security flag exists.
func connectCaptureControl(ctx context.Context, e captureControlEndpoint, r workerJobRecord, i workerCaptureIntent, stop func(), dialer objectReadControlDialer) (*authenticatedCaptureClient, error) {
	if ctx == nil || ctx.Err() != nil || stop == nil || nilWorkerObservationDependency(dialer) || !e.valid(r, i) {
		return nil, errCapturePeer
	}
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	raw, err := dialer.DialContext(dialCtx, "tcp", e.Address)
	expired := dialCtx.Err() != nil
	cancel()
	if err != nil || expired || ctx.Err() != nil || nilWorkerObservationDependency(raw) {
		if !nilWorkerObservationDependency(raw) {
			_ = raw.Close()
		}
		return nil, errCapturePeer
	}
	// The dial timeout is not the session lifetime. Subsequent authority remains
	// bound to the original context, certificate and retained job/intent.
	return newAuthenticatedCaptureClient(ctx, raw, e.Binding, r, i, stop)
}

// serveCaptureControlOnce owns the already provisioned listener on entry and
// consumes exactly one connection. Close MUST unblock Accept (as net.TCPListener
// does). The listener is closed before TLS/session setup, so failure cannot accept
// a replacement peer or retry an uncertain session. The caller must join this call
// before relinquishing its creator lease. No permission or physical evidence is
// synthesized from the accepted socket; the pinned controller session is required.
func serveCaptureControlOnce(ctx context.Context, listener net.Listener, b capturePeerBinding, g *controllerCaptureSession) error {
	if nilWorkerObservationDependency(listener) {
		if g != nil {
			g.StopCaptureAuthority()
		}
		return errCapturePeer
	}
	defer listener.Close()
	if g == nil {
		return errCapturePeer
	}
	defer g.StopCaptureAuthority()
	if ctx == nil || ctx.Err() != nil || g.lifetime == nil || g.lifetime.Err() != nil || b.Worker != g.worker || !b.matches(g.record, g.intent) || !validObjectReadPeerIdentity(b.Identity) {
		return errCapturePeer
	}
	acceptCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	stopSession := context.AfterFunc(g.lifetime, cancel)
	stopListener := context.AfterFunc(acceptCtx, func() { _ = listener.Close() })
	raw, err := listener.Accept()
	expired := acceptCtx.Err() != nil
	stopListener()
	stopSession()
	cancel()
	closeErr := listener.Close()
	if closeErr != nil || err != nil || expired || ctx.Err() != nil || g.lifetime.Err() != nil || nilWorkerObservationDependency(raw) {
		if !nilWorkerObservationDependency(raw) {
			_ = raw.Close()
		}
		return errCapturePeer
	}
	return serveAuthenticatedCaptureSession(ctx, raw, b, g)
}
