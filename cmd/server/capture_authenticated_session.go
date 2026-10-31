package main

import (
	"context"
	"errors"
	"net"
	"time"
)

const captureControlALPN = "commonroom-capture-control-v1"

var errCapturePeer = errors.New("capture peer authentication refused")

// Certificate policy is shared with readers; protocol and operation grants are
// NEVER shared. Independent trusted setup pins exact worker/record/intent values.
// No request/model metadata can provision an identity or choose another protocol.
type capturePeerBinding struct {
	Identity             objectReadPeerIdentity
	Worker, RecordDigest string
	Intent               workerCaptureIntent
}

func (b capturePeerBinding) matches(r workerJobRecord, i workerCaptureIntent) bool {
	return workerAliasPattern.MatchString(b.Worker) && captureAuthorityBinding(r, i) && b.RecordDigest == probeWorkerRecordDigest(r) && captureIntentEqual(b.Intent, i)
}

type captureSessionHello struct {
	Protocol     string              `json:"protocol"`
	Version      int                 `json:"version"`
	Worker       string              `json:"worker"`
	RecordDigest string              `json:"recordDigest"`
	Intent       workerCaptureIntent `json:"intent"`
}

func (b capturePeerBinding) hello() captureSessionHello {
	return captureSessionHello{"capture-session-v1", 1, b.Worker, b.RecordDigest, copyCaptureIntent(b.Intent)}
}
func sameCaptureHello(a, b captureSessionHello) bool {
	return a.Protocol == b.Protocol && a.Version == b.Version && a.Worker == b.Worker && a.RecordDigest == b.RecordDigest && captureIntentEqual(a.Intent, b.Intent)
}

// Consumes raw on every path. Mutual certificate chain, DNS and leaf-pin checks
// and the capture-only ALPN precede finite hello agreement. No listener, discovery,
// plaintext retry, host/SSH execution or peer-selected capability is introduced.
func authenticateBoundCaptureStream(ctx context.Context, raw net.Conn, binding capturePeerBinding, server bool) (*objectReadControlStream, error) {
	if nilWorkerObservationDependency(raw) {
		return nil, errCapturePeer
	}
	refuse := func() (*objectReadControlStream, error) { _ = raw.Close(); return nil, errCapturePeer }
	if ctx == nil || ctx.Err() != nil || !workerAliasPattern.MatchString(binding.Worker) || !repositoryDigestPattern.MatchString(binding.RecordDigest) || !validWorkerCaptureRecord(workerCaptureRecord{Version: 1, Intent: binding.Intent, Events: []workerProbeEvent{{Stage: "begin"}}}) {
		return refuse()
	}
	binding.Intent = copyCaptureIntent(binding.Intent)
	stop := context.AfterFunc(ctx, func() { _ = raw.Close() })
	defer stop()
	conn, err := authenticatePinnedWorkerStream(ctx, raw, binding.Identity, server, captureControlALPN)
	if err != nil {
		return nil, errCapturePeer
	}
	stream := &objectReadControlStream{Conn: conn}
	deadline := time.Now().Add(5 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if raw.SetDeadline(deadline) != nil {
		return refuse()
	}
	want := binding.hello()
	var got captureSessionHello
	if server {
		if readCaptureAuthorityFrame(stream, &got) != nil || !sameCaptureHello(got, want) || writeCaptureAuthorityFrame(stream, want) != nil {
			return refuse()
		}
	} else {
		if writeCaptureAuthorityFrame(stream, want) != nil || readCaptureAuthorityFrame(stream, &got) != nil || !sameCaptureHello(got, want) {
			return refuse()
		}
	}
	if ctx.Err() != nil || raw.SetDeadline(time.Time{}) != nil {
		return refuse()
	}
	return stream, nil
}

type authenticatedCaptureClient struct {
	*captureAuthorityClient
	stopParent func() bool
}

func (c *authenticatedCaptureClient) StopCaptureAuthority() {
	if c != nil {
		if c.stopParent != nil {
			c.stopParent()
		}
		c.captureAuthorityClient.StopCaptureAuthority()
	}
}

// The original setup context remains authoritative while idle. The caller owns
// the returned client and must stop/join its pump and controller serving call.
func newAuthenticatedCaptureClient(ctx context.Context, raw net.Conn, b capturePeerBinding, r workerJobRecord, i workerCaptureIntent, stop func()) (*authenticatedCaptureClient, error) {
	if !b.matches(r, i) || stop == nil {
		if !nilWorkerObservationDependency(raw) {
			_ = raw.Close()
		}
		return nil, errCapturePeer
	}
	b.Intent = copyCaptureIntent(b.Intent)
	stream, err := authenticateBoundCaptureStream(ctx, raw, b, false)
	if err != nil {
		return nil, errCapturePeer
	}
	client, err := newCaptureAuthorityClient(stream, r, i, stop)
	if err != nil {
		_ = stream.Close()
		return nil, errCapturePeer
	}
	owned := &authenticatedCaptureClient{client, context.AfterFunc(ctx, client.StopCaptureAuthority)}
	if ctx.Err() != nil {
		owned.StopCaptureAuthority()
		return nil, errCapturePeer
	}
	return owned, nil
}

// Trusted setup supplies the ORIGINAL human/worker/creator-pinned session. TLS
// identity and hello agreement do not supply physical evidence or permissions.
func serveAuthenticatedCaptureSession(ctx context.Context, raw net.Conn, b capturePeerBinding, g *controllerCaptureSession) error {
	if !nilWorkerObservationDependency(raw) {
		defer raw.Close()
	}
	if g == nil {
		return errCapturePeer
	}
	defer g.StopCaptureAuthority()
	if ctx == nil || ctx.Err() != nil || g.lifetime == nil || g.lifetime.Err() != nil || b.Worker != g.worker || !b.matches(g.record, g.intent) {
		return errCapturePeer
	}
	b.Intent = copyCaptureIntent(b.Intent)
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(g.lifetime, cancel)
	defer stop()
	stream, err := authenticateBoundCaptureStream(child, raw, b, true)
	if err != nil {
		return errCapturePeer
	}
	defer stream.Close()
	return serveCaptureAuthority(child, stream, g.record, g.intent, g)
}

var _ workerCaptureRuntimeGuard = (*authenticatedCaptureClient)(nil)
