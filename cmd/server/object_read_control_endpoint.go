package main

import (
	"context"
	"net"
	"net/netip"
	"time"
)

// An operator-provisioned literal IP:port, never a model/request-selected URL or
// address. DNS discovery, redirects, alternate endpoints and reconnect are absent.
// Identity/binding still independently pin the TLS peer and retained read session.
// Exposing this constructor to model tools would violate the deployment contract.
type objectReadControlEndpoint struct {
	Address string
	Binding objectReadPeerBinding
}
type objectReadControlDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

func (e objectReadControlEndpoint) valid(r workerJobRecord, i workerObjectReadIntent) bool {
	address, err := netip.ParseAddrPort(e.Address)
	return err == nil && address.Port() != 0 && !address.Addr().IsUnspecified() && !address.Addr().IsMulticast() && address.Addr().Zone() == "" && !address.Addr().Is4In6() && e.Binding.matches(r, i) && validObjectReadPeerIdentity(e.Binding.Identity)
}

// connectObjectReadControl uses only one approved TCP endpoint and consumes any
// resulting connection. Deployment supplies a net.Dialer; a source fork replacing
// it can bypass network policy, never TLS/session/permission checks. Dial has a
// five-second bound and follows the original setup context. No shell/SSH, host
// reader, plaintext retry or optional security flag exists.
func connectObjectReadControl(ctx context.Context, e objectReadControlEndpoint, r workerJobRecord, i workerObjectReadIntent, stop func(), dialer objectReadControlDialer) (*authenticatedObjectReadClient, error) {
	if ctx == nil || ctx.Err() != nil || stop == nil || nilWorkerObservationDependency(dialer) || !e.valid(r, i) {
		return nil, errObjectReadPeer
	}
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	raw, err := dialer.DialContext(dialCtx, "tcp", e.Address)
	expired := dialCtx.Err() != nil
	cancel()
	if err != nil || expired || ctx.Err() != nil || nilWorkerObservationDependency(raw) {
		if !nilWorkerObservationDependency(raw) {
			_ = raw.Close()
		}
		return nil, errObjectReadPeer
	}
	// The dial timeout is not the session lifetime. Subsequent authority remains
	// bound to the original context, certificate and retained job/intent.
	return newAuthenticatedObjectReadClient(ctx, raw, e.Binding, r, i, stop)
}

// serveObjectReadControlOnce owns the already provisioned listener on entry and
// consumes exactly one connection. Close MUST unblock Accept (as net.TCPListener
// does). The listener is closed before TLS/session setup, so failure cannot accept
// a replacement peer or retry an uncertain session. The caller must join this call
// before relinquishing its creator lease. No permission or physical evidence is
// synthesized from the accepted socket; the pinned controller session is required.
func serveObjectReadControlOnce(ctx context.Context, listener net.Listener, b objectReadPeerBinding, g *controllerObjectReadSession) error {
	if nilWorkerObservationDependency(listener) {
		if g != nil {
			g.StopObjectReadAuthority()
		}
		return errObjectReadPeer
	}
	defer listener.Close()
	if g == nil {
		return errObjectReadPeer
	}
	defer g.StopObjectReadAuthority()
	if ctx == nil || ctx.Err() != nil || g.lifetime == nil || g.lifetime.Err() != nil || b.Worker != g.worker || !b.matches(g.record, g.intent) || !validObjectReadPeerIdentity(b.Identity) {
		return errObjectReadPeer
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
		return errObjectReadPeer
	}
	return serveAuthenticatedObjectReadSession(ctx, raw, b, g)
}
