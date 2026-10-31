package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"time"
)

const objectReadControlALPN = "commonroom-object-read-v1"

var errObjectReadPeer = errors.New("artifact reader peer authentication refused")

// objectReadPeerIdentity comes exclusively from trusted deployment provisioning,
// never from a worker request or model metadata. Roots authenticate the issuing
// authority; the exact leaf pin also separates workers sharing that authority.
// Certificate rotation requires a fresh one-shot session with a newly approved pin.
type objectReadPeerIdentity struct {
	Certificate tls.Certificate
	Roots       *x509.CertPool
	PeerDNSName string
	PeerSHA256  [sha256.Size]byte
}

func validObjectReadPeerIdentity(identity objectReadPeerIdentity) bool {
	return identity.Roots != nil && identity.PeerDNSName != "" && identity.PeerSHA256 != ([sha256.Size]byte{}) && len(identity.Certificate.Certificate) > 0 && !nilWorkerObservationDependency(identity.Certificate.PrivateKey)
}

// authenticateObjectReadStream owns raw on every path. It authenticates both ends
// before any authority frame is sent. It does not dial, discover identities, grant
// export rights, or replace the pinned human/creator/intent session. Production
// provisioning must keep keys and this stream outside all model containers. No
// plaintext, host-execution, or alternate-protocol fallback is permitted. A source
// fork replacing this constructor can bypass the boundary.
func authenticateObjectReadStream(ctx context.Context, raw net.Conn, identity objectReadPeerIdentity, server bool) (*tls.Conn, error) {
	return authenticatePinnedWorkerStream(ctx, raw, identity, server, objectReadControlALPN)
}

// Only fixed internal protocol wrappers may select these two disjoint ALPNs.
// This shares certificate/cipher policy, never operation or account authority.
func authenticatePinnedWorkerStream(ctx context.Context, raw net.Conn, identity objectReadPeerIdentity, server bool, protocol string) (*tls.Conn, error) {
	if raw == nil {
		return nil, errObjectReadPeer
	}
	refuse := func() (*tls.Conn, error) { _ = raw.Close(); return nil, errObjectReadPeer }
	if ctx == nil || ctx.Err() != nil || !validObjectReadPeerIdentity(identity) || (protocol != objectReadControlALPN && protocol != captureControlALPN) {
		return refuse()
	}
	// Clone roots so subsequent caller mutations cannot change this session's trust.
	config := &tls.Config{
		MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS13,
		// TLS 1.2 is restricted to forward-secret AEAD suites. TLS 1.3
		// cipher suites are selected securely by the standard library.
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		Certificates: []tls.Certificate{identity.Certificate},
		RootCAs:      identity.Roots.Clone(), ClientCAs: identity.Roots.Clone(),
		ServerName: identity.PeerDNSName, NextProtos: []string{protocol},
		SessionTicketsDisabled: true,
	}
	config.VerifyConnection = func(state tls.ConnectionState) error {
		if (state.Version != tls.VersionTLS12 && state.Version != tls.VersionTLS13) || state.NegotiatedProtocol != protocol || len(state.VerifiedChains) == 0 || len(state.PeerCertificates) == 0 {
			return errObjectReadPeer
		}
		leaf := state.PeerCertificates[0]
		if sha256.Sum256(leaf.Raw) != identity.PeerSHA256 || leaf.VerifyHostname(identity.PeerDNSName) != nil {
			return errObjectReadPeer
		}
		return nil
	}
	var conn *tls.Conn
	if server {
		config.ClientAuth = tls.RequireAndVerifyClientCert
		conn = tls.Server(raw, config)
	} else {
		conn = tls.Client(raw, config)
	}
	handshake, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if conn.HandshakeContext(handshake) != nil {
		return refuse()
	}
	return conn, nil
}
