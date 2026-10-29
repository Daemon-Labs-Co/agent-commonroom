package main

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"
)

// Provisioning binds the authenticated peer certificate to exactly one worker,
// complete retained record and finite read intent. This binding is trusted local
// configuration, never filled from a received frame. TLS identity alone does not
// grant account, creator, physical-quota or artifact-export authority.
type objectReadPeerBinding struct {
	Identity     objectReadPeerIdentity
	Worker       string
	RecordDigest string
	Intent       workerObjectReadIntent
}

func (b objectReadPeerBinding) matches(r workerJobRecord, i workerObjectReadIntent) bool {
	return workerAliasPattern.MatchString(b.Worker) && objectReadAuthorityBinding(r, i) && b.RecordDigest == probeWorkerRecordDigest(r) && b.Intent == i
}

type objectReadSessionHello struct {
	Protocol     string                 `json:"protocol"`
	Version      int                    `json:"version"`
	Worker       string                 `json:"worker"`
	RecordDigest string                 `json:"recordDigest"`
	Intent       workerObjectReadIntent `json:"intent"`
}

func (b objectReadPeerBinding) hello() objectReadSessionHello {
	return objectReadSessionHello{"object-read-session-v1", 1, b.Worker, b.RecordDigest, b.Intent}
}

// Close aborts the underlying connection before joining the authority read pump;
// TLS close-notify cannot delay cancellation or stand in for durable cleanup.
// The complete framed acknowledgments, never EOF, establish journal transitions.
type objectReadControlStream struct {
	*tls.Conn
	once sync.Once
}

func (s *objectReadControlStream) Close() error {
	var err error
	s.once.Do(func() { err = s.NetConn().Close(); _ = s.Conn.Close() })
	return err
}

// authenticateBoundObjectReadStream consumes raw on every path. Both endpoints
// must independently agree on the provisioned binding before authority frames.
// No listener/discovery, model-selected identity, host/SSH fallback or off switch.
// A source fork bypassing this setup invalidates the authentication contract.
func authenticateBoundObjectReadStream(ctx context.Context, raw net.Conn, b objectReadPeerBinding, server bool) (*objectReadControlStream, error) {
	if nilWorkerObservationDependency(raw) {
		return nil, errObjectReadPeer
	}
	refuse := func() (*objectReadControlStream, error) { _ = raw.Close(); return nil, errObjectReadPeer }
	if ctx == nil || ctx.Err() != nil || !workerAliasPattern.MatchString(b.Worker) || len(b.RecordDigest) != 64 {
		return refuse()
	}
	stop := context.AfterFunc(ctx, func() { _ = raw.Close() })
	defer stop()
	conn, err := authenticateObjectReadStream(ctx, raw, b.Identity, server)
	if err != nil {
		return nil, errObjectReadPeer
	}
	stream := &objectReadControlStream{Conn: conn}
	deadline := time.Now().Add(5 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if raw.SetDeadline(deadline) != nil {
		return refuse()
	}
	want := b.hello()
	var got objectReadSessionHello
	if server {
		if readObjectReadAuthorityFrame(stream, &got) != nil || got != want || writeObjectReadAuthorityFrame(stream, want) != nil {
			return refuse()
		}
	} else {
		if writeObjectReadAuthorityFrame(stream, want) != nil || readObjectReadAuthorityFrame(stream, &got) != nil || got != want {
			return refuse()
		}
	}
	if ctx.Err() != nil || raw.SetDeadline(time.Time{}) != nil {
		return refuse()
	}
	return stream, nil
}

// The worker owns the returned client and must StopObjectReadAuthority to join
// its read pump on every outcome. Its original setup context remains authoritative
// while idle, not just during a protocol call. Closure cannot release capacity.
type authenticatedObjectReadClient struct {
	*objectReadAuthorityClient
	stopParent func() bool
}

func (c *authenticatedObjectReadClient) StopObjectReadAuthority() {
	if c != nil {
		if c.stopParent != nil {
			c.stopParent()
		}
		c.objectReadAuthorityClient.StopObjectReadAuthority()
	}
}
func newAuthenticatedObjectReadClient(ctx context.Context, raw net.Conn, b objectReadPeerBinding, r workerJobRecord, i workerObjectReadIntent, stop func()) (*authenticatedObjectReadClient, error) {
	if !b.matches(r, i) || stop == nil {
		if !nilWorkerObservationDependency(raw) {
			_ = raw.Close()
		}
		return nil, errObjectReadPeer
	}
	stream, err := authenticateBoundObjectReadStream(ctx, raw, b, false)
	if err != nil {
		return nil, errObjectReadPeer
	}
	c, err := newObjectReadAuthorityClient(stream, r, i, stop)
	if err != nil {
		_ = stream.Close()
		return nil, errObjectReadPeer
	}
	owned := &authenticatedObjectReadClient{objectReadAuthorityClient: c, stopParent: context.AfterFunc(ctx, c.StopObjectReadAuthority)}
	if ctx.Err() != nil {
		owned.StopObjectReadAuthority()
		return nil, errObjectReadPeer
	}
	return owned, nil
}

// Trusted controller setup supplies the already human/worker/creator-pinned session.
// The serving call owns and joins this single connection; any exit revokes only
// this session. No evidence, permissions or account identity comes from TLS frames.
func serveAuthenticatedObjectReadSession(ctx context.Context, raw net.Conn, b objectReadPeerBinding, g *controllerObjectReadSession) error {
	if !nilWorkerObservationDependency(raw) {
		defer raw.Close()
	}
	if g == nil {
		return errObjectReadPeer
	}
	defer g.StopObjectReadAuthority()
	if ctx == nil || ctx.Err() != nil || g.lifetime == nil || g.lifetime.Err() != nil || b.Worker != g.worker || !b.matches(g.record, g.intent) {
		return errObjectReadPeer
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(g.lifetime, cancel)
	defer stop()
	stream, err := authenticateBoundObjectReadStream(child, raw, b, true)
	if err != nil {
		return errObjectReadPeer
	}
	defer stream.Close()
	return serveObjectReadAuthority(child, stream, g.record, g.intent, g)
}

var _ workerObjectReadRuntimeGuard = (*authenticatedObjectReadClient)(nil)
