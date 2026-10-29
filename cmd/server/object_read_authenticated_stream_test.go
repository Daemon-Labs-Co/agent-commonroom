package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"
)

func objectReadTestPeers(t *testing.T) (objectReadPeerIdentity, objectReadPeerIdentity) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	ca := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "synthetic reader CA"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	der, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	root, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(root)
	issue := func(name string, serial int64) tls.Certificate {
		p, k, e := ed25519.GenerateKey(rand.Reader)
		if e != nil {
			t.Fatal(e)
		}
		cert := &x509.Certificate{SerialNumber: big.NewInt(serial), DNSNames: []string{name}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}}
		d, e := x509.CreateCertificate(rand.Reader, cert, root, p, key)
		if e != nil {
			t.Fatal(e)
		}
		private, e := x509.MarshalPKCS8PrivateKey(k)
		if e != nil {
			t.Fatal(e)
		}
		pair, e := tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: d}), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: private}))
		if e != nil {
			t.Fatal(e)
		}
		return pair
	}
	client, server := issue("worker.test", 2), issue("controller.test", 3)
	return objectReadPeerIdentity{client, roots, "controller.test", sha256.Sum256(server.Certificate[0])}, objectReadPeerIdentity{server, roots, "worker.test", sha256.Sum256(client.Certificate[0])}
}
func TestObjectReadAuthenticatedStream(t *testing.T) {
	t.Parallel()
	client, server := objectReadTestPeers(t)
	for _, tc := range []struct {
		name   string
		change func(*objectReadPeerIdentity, *objectReadPeerIdentity)
		valid  bool
	}{
		{"mutual identity", func(c, s *objectReadPeerIdentity) {}, true},
		{"wrong controller pin", func(c, s *objectReadPeerIdentity) { c.PeerSHA256[0] ^= 1 }, false},
		{"wrong worker pin", func(c, s *objectReadPeerIdentity) { s.PeerSHA256[0] ^= 1 }, false},
		{"wrong controller name", func(c, s *objectReadPeerIdentity) { c.PeerDNSName = "foreign.test" }, false},
		{"wrong worker name", func(c, s *objectReadPeerIdentity) { s.PeerDNSName = "foreign.test" }, false},
		{"untrusted controller", func(c, s *objectReadPeerIdentity) { c.Roots = x509.NewCertPool() }, false},
		{"untrusted worker", func(c, s *objectReadPeerIdentity) { s.Roots = x509.NewCertPool() }, false},
		{"missing worker certificate", func(c, s *objectReadPeerIdentity) { c.Certificate = tls.Certificate{} }, false},
		{"missing pin", func(c, s *objectReadPeerIdentity) { s.PeerSHA256 = [32]byte{} }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, s := client, server
			tc.change(&c, &s)
			a, b := net.Pipe()
			defer a.Close()
			defer b.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			type result struct {
				conn *tls.Conn
				err  error
			}
			done := make(chan result, 1)
			go func() { conn, err := authenticateObjectReadStream(ctx, b, s, true); done <- result{conn, err} }()
			conn, err := authenticateObjectReadStream(ctx, a, c, false)
			other := <-done
			if !tc.valid {
				if err == nil && other.err == nil {
					t.Fatal("unapproved peer authenticated")
				}
				return
			}
			if err != nil || other.err != nil {
				t.Fatalf("mutual authentication: %v / %v", err, other.err)
			}
			wrote := make(chan error, 1)
			go func() { _, e := conn.Write([]byte("x")); wrote <- e }()
			var out [1]byte
			if _, e := other.conn.Read(out[:]); e != nil || out[0] != 'x' {
				t.Fatalf("authenticated data: %v", e)
			}
			if e := <-wrote; e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestObjectReadAuthenticatedStreamCanceledAndPlaintext(t *testing.T) {
	t.Parallel()
	client, _ := objectReadTestPeers(t)
	for _, canceled := range []bool{true, false} {
		t.Run(map[bool]string{true: "canceled", false: "unresponsive plaintext peer"}[canceled], func(t *testing.T) {
			t.Parallel()
			a, b := net.Pipe()
			defer b.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			if canceled {
				cancel()
			}
			conn, err := authenticateObjectReadStream(ctx, a, client, false)
			if err == nil || conn != nil {
				t.Fatal("unauthenticated peer accepted")
			}
			if _, err := a.Write([]byte("x")); err == nil {
				t.Fatal("failed authentication retained raw stream")
			}
		})
	}
}

func TestObjectReadAuthenticatedStreamRejectsOtherProtocols(t *testing.T) {
	t.Parallel()
	client, server := objectReadTestPeers(t)
	for _, tc := range []struct {
		name, protocol string
		version        uint16
	}{
		{"other authority protocol", "commonroom-capture-v1", tls.VersionTLS13},
		{"no protocol", "", tls.VersionTLS13},
		{"older TLS", objectReadControlALPN, tls.VersionTLS11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, b := net.Pipe()
			defer a.Close()
			defer b.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			config := &tls.Config{Certificates: []tls.Certificate{server.Certificate}, MinVersion: tc.version, MaxVersion: tc.version}
			if tc.protocol != "" {
				config.NextProtos = []string{tc.protocol}
			}
			done := make(chan struct{})
			go func() {
				defer close(done)
				peer := tls.Server(b, config)
				_ = peer.HandshakeContext(ctx)
				_ = b.Close()
			}()
			conn, err := authenticateObjectReadStream(ctx, a, client, false)
			if err == nil || conn != nil {
				t.Fatal("foreign protocol accepted")
			}
			<-done
		})
	}
}

func TestObjectReadAuthenticatedJournalLifecycle(t *testing.T) {
	t.Parallel()
	for _, denied := range []bool{false, true} {
		t.Run(map[bool]string{false: "settled", true: "denied retains charge"}[denied], func(t *testing.T) {
			t.Parallel()
			j, g := objectReadTransportFixture(t)
			if denied {
				g.failAck = "begin"
			}
			client, server := objectReadTestPeers(t)
			a, b := net.Pipe()
			defer a.Close()
			defer b.Close()
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				stream, err := authenticateObjectReadStream(ctx, b, server, true)
				if err != nil {
					done <- err
					return
				}
				done <- serveObjectReadAuthority(ctx, stream, j.record, j.intent, g)
			}()
			stream, err := authenticateObjectReadStream(ctx, a, client, false)
			if err != nil {
				t.Fatal(err)
			}
			c, err := newObjectReadAuthorityClient(stream, j.record, j.intent, func() {})
			if err != nil {
				t.Fatal(err)
			}
			defer c.StopObjectReadAuthority()
			j.guard = c
			err = j.BeginObjectRead(ctx, j.intent.Request, j.intent.Objects, j.intent.Image)
			if denied {
				if err == nil {
					t.Fatal("denied lifecycle accepted over authenticated channel")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				id := strings.Repeat("1", 64)
				if err = j.BindObjectRead(ctx, j.intent.Request, id); err != nil {
					t.Fatal(err)
				}
				if err = j.FinishObjectRead(ctx, j.intent.Request, id, true); err != nil {
					t.Fatal(err)
				}
			}
			expected := 0
			if denied {
				expected = 1
			}
			if n, err := workerObjectReadSlots(j.dir); err != nil || n != expected {
				t.Fatalf("charge %d, want %d: %v", n, expected, err)
			}
			_ = a.Close()
			c.StopObjectReadAuthority()
			<-done
		})
	}
}

// Force each supported version at the independently configured peer. Both
// adapter roles must accept mutual TLS 1.2 as well as TLS 1.3.
func TestObjectReadAuthenticatedStreamSupportedVersions(t *testing.T) {
	t.Parallel()
	client, server := objectReadTestPeers(t)
	for _, version := range []uint16{tls.VersionTLS12, tls.VersionTLS13} {
		for _, adapterServer := range []bool{false, true} {
			t.Run(fmt.Sprintf("%x/server=%t", version, adapterServer), func(t *testing.T) {
				t.Parallel()
				a, b := net.Pipe()
				defer a.Close()
				defer b.Close()
				ctx, cancel := context.WithTimeout(t.Context(), time.Second)
				defer cancel()
				local, peer := client, server
				if adapterServer {
					local, peer = server, client
				}
				config := &tls.Config{MinVersion: version, MaxVersion: version, Certificates: []tls.Certificate{peer.Certificate}, RootCAs: peer.Roots, ClientCAs: peer.Roots, ServerName: peer.PeerDNSName, NextProtos: []string{objectReadControlALPN}, ClientAuth: tls.RequireAndVerifyClientCert}
				done := make(chan error, 1)
				go func() {
					var conn *tls.Conn
					if adapterServer {
						conn = tls.Client(b, config)
					} else {
						conn = tls.Server(b, config)
					}
					done <- conn.HandshakeContext(ctx)
				}()
				conn, err := authenticateObjectReadStream(ctx, a, local, adapterServer)
				other := <-done
				if err != nil || other != nil {
					t.Fatalf("supported mutual TLS refused: %v / %v", err, other)
				}
				if conn.ConnectionState().Version != version {
					t.Fatal("wrong negotiated version")
				}
			})
		}
	}
}
