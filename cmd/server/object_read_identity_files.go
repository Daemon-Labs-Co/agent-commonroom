package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"path/filepath"
	"strings"
	"time"
)

// These paths and the peer pin are trusted deployment input, never model/request
// values or files mounted into agent containers. Parent directories must remain
// operator-controlled. The existing private-file boundary checks the opened inode:
// owner-only, current UID, regular, single-link, no-follow, nonblocking and bounded.
// No key is generated, written, logged, sent to a provider or stored in the audit.
type objectReadIdentityFiles struct {
	Certificate string
	PrivateKey  string
	Authorities string
	PeerDNSName string
	PeerSHA256  string
}

func objectReadCertificateBundle(raw []byte, limit int) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for len(bytes.TrimSpace(raw)) > 0 {
		// pem.Decode otherwise skips leading junk. Configuration must be exact PEM.
		raw = bytes.TrimSpace(raw)
		if !bytes.HasPrefix(raw, []byte("-----BEGIN CERTIFICATE-----")) {
			return nil, errObjectReadPeer
		}
		block, rest := pem.Decode(raw)
		if block == nil || block.Type != "CERTIFICATE" || len(block.Headers) != 0 || len(certs) >= limit || bytes.Count(raw[:len(raw)-len(rest)], []byte("-----BEGIN ")) != 1 {
			return nil, errObjectReadPeer
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errObjectReadPeer
		}
		certs = append(certs, cert)
		raw = rest
	}
	if len(certs) == 0 {
		return nil, errObjectReadPeer
	}
	return certs, nil
}
func objectReadCertificateCurrent(c *x509.Certificate, now time.Time) bool {
	return c != nil && !now.Before(c.NotBefore) && now.Before(c.NotAfter)
}

// loadObjectReadPeerIdentity loads a fresh independent snapshot for one endpoint
// role (server=false means worker/client). Failure returns no partial identity and
// a sanitized error. Rotation requires new provisioning and a fresh session; this
// loader never changes a running session or relaxes TLS/peer/account authorization.
// It is mandatory security infrastructure, without a flag or host/SSH fallback.
// A source fork weakening the loader or its trusted parent boundary voids custody.
func loadObjectReadPeerIdentity(files objectReadIdentityFiles, server bool) (objectReadPeerIdentity, error) {
	fail := func() (objectReadPeerIdentity, error) { return objectReadPeerIdentity{}, errObjectReadPeer }
	paths := []string{files.Certificate, files.PrivateKey, files.Authorities}
	seen := map[string]bool{}
	for _, p := range paths {
		if !filepath.IsAbs(p) || filepath.Clean(p) != p || seen[p] {
			return fail()
		}
		seen[p] = true
	}
	if files.PeerDNSName == "" || strings.TrimSpace(files.PeerDNSName) != files.PeerDNSName || strings.ContainsAny(files.PeerDNSName, "*\r\n\t /\\") {
		return fail()
	}
	pin, err := hex.DecodeString(files.PeerSHA256)
	if err != nil || len(pin) != sha256.Size || files.PeerSHA256 != strings.ToLower(files.PeerSHA256) {
		return fail()
	}
	var digest [sha256.Size]byte
	copy(digest[:], pin)
	if digest == ([sha256.Size]byte{}) {
		return fail()
	}
	certPEM, err := readPrivateBounded(files.Certificate, 64*1024)
	if err != nil {
		return fail()
	}
	keyPEM, err := readPrivateBounded(files.PrivateKey, 64*1024)
	if err != nil {
		return fail()
	}
	defer clear(keyPEM)
	caPEM, err := readPrivateBounded(files.Authorities, 256*1024)
	if err != nil {
		return fail()
	}
	certificates, err := objectReadCertificateBundle(certPEM, 8)
	if err != nil {
		return fail()
	}
	keyBlock, rest := pem.Decode(bytes.TrimSpace(keyPEM))
	if keyBlock == nil || len(keyBlock.Headers) != 0 || len(bytes.TrimSpace(rest)) != 0 || !bytes.HasPrefix(bytes.TrimSpace(keyPEM), []byte("-----BEGIN "+keyBlock.Type+"-----")) || bytes.Count(keyPEM, []byte("-----BEGIN ")) != 1 {
		return fail()
	}
	switch keyBlock.Type {
	case "PRIVATE KEY", "EC PRIVATE KEY", "RSA PRIVATE KEY":
	default:
		return fail()
	}
	defer clear(keyBlock.Bytes)
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fail()
	}
	now := time.Now()
	leaf := certificates[0]
	if leaf.IsCA || !objectReadCertificateCurrent(leaf, now) {
		return fail()
	}
	usage := x509.ExtKeyUsageClientAuth
	if server {
		usage = x509.ExtKeyUsageServerAuth
	}
	allowed := len(leaf.ExtKeyUsage) == 0
	for _, u := range leaf.ExtKeyUsage {
		if u == usage || u == x509.ExtKeyUsageAny {
			allowed = true
		}
	}
	if !allowed {
		return fail()
	}
	pair.Leaf = leaf
	authorities, err := objectReadCertificateBundle(caPEM, 16)
	if err != nil {
		return fail()
	}
	roots := x509.NewCertPool()
	caSeen := map[[sha256.Size]byte]bool{}
	for _, ca := range authorities {
		hash := sha256.Sum256(ca.Raw)
		if caSeen[hash] || !ca.IsCA || ca.KeyUsage&x509.KeyUsageCertSign == 0 || !objectReadCertificateCurrent(ca, now) {
			return fail()
		}
		caSeen[hash] = true
		roots.AddCert(ca)
	}
	identity := objectReadPeerIdentity{Certificate: pair, Roots: roots, PeerDNSName: files.PeerDNSName, PeerSHA256: digest}
	if !validObjectReadPeerIdentity(identity) {
		return fail()
	}
	return identity, nil
}
