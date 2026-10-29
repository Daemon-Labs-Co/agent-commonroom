package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeObjectReadTestIdentity(t *testing.T, identity objectReadPeerIdentity) objectReadIdentityFiles {
	t.Helper()
	dir := t.TempDir()
	cert := []byte{}
	for _, der := range identity.Certificate.Certificate {
		cert = append(cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})...)
	}
	key, err := x509.MarshalPKCS8PrivateKey(identity.Certificate.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	// The synthetic peer helper creates a common issuing root. Its generated
	// issuer is supplied separately by the fixture below, never read from user files.
	files := objectReadIdentityFiles{filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"), filepath.Join(dir, "ca.pem"), identity.PeerDNSName, hex.EncodeToString(identity.PeerSHA256[:])}
	for p, raw := range map[string][]byte{files.Certificate: cert, files.PrivateKey: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key})} {
		if err := os.WriteFile(p, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return files
}
func provisionedObjectReadTestFiles(t *testing.T) (objectReadIdentityFiles, objectReadIdentityFiles) {
	return provisionedObjectReadTestFilesWith(t, nil)
}
func provisionedObjectReadTestFilesWith(t *testing.T, mutate func(*x509.Certificate, bool)) (objectReadIdentityFiles, objectReadIdentityFiles) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "synthetic provisioning root"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	if mutate != nil {
		mutate(template, true)
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	ca, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	issue := func(name string, serial int64) tls.Certificate {
		p, k, e := ed25519.GenerateKey(rand.Reader)
		if e != nil {
			t.Fatal(e)
		}
		leaf := &x509.Certificate{SerialNumber: big.NewInt(serial), DNSNames: []string{name}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}, KeyUsage: x509.KeyUsageDigitalSignature}
		if mutate != nil {
			mutate(leaf, false)
		}
		d, e := x509.CreateCertificate(rand.Reader, leaf, ca, p, key)
		if e != nil {
			t.Fatal(e)
		}
		return tls.Certificate{Certificate: [][]byte{d}, PrivateKey: k}
	}
	client, server := issue("worker.test", 2), issue("controller.test", 3)
	c := writeObjectReadTestIdentity(t, objectReadPeerIdentity{Certificate: client, PeerDNSName: "controller.test", PeerSHA256: sha256.Sum256(server.Certificate[0])})
	s := writeObjectReadTestIdentity(t, objectReadPeerIdentity{Certificate: server, PeerDNSName: "worker.test", PeerSHA256: sha256.Sum256(client.Certificate[0])})
	for _, path := range []string{c.Authorities, s.Authorities} {
		if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return c, s
}
func TestObjectReadIdentityFilesMutualAuthentication(t *testing.T) {
	t.Parallel()
	client, server := provisionedObjectReadTestFiles(t)
	c, err := loadObjectReadPeerIdentity(client, false)
	if err != nil {
		t.Fatal(err)
	}
	s, err := loadObjectReadPeerIdentity(server, true)
	if err != nil {
		t.Fatal(err)
	}
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := authenticateObjectReadStream(ctx, b, s, true); done <- err }()
	conn, err := authenticateObjectReadStream(ctx, a, c, false)
	other := <-done
	if err != nil || other != nil || conn == nil {
		t.Fatal("provisioned identity failed mutual authentication", err, other)
	}
	// Files are snapshots, not a live trust source or automatic rotation hook.
	if err := os.WriteFile(client.Authorities, []byte("invalid replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if !validObjectReadPeerIdentity(c) {
		t.Fatal("file mutation changed active identity")
	}
	if _, err := loadObjectReadPeerIdentity(client, false); err != errObjectReadPeer {
		t.Fatal("invalid replacement loaded")
	}
}
func TestObjectReadIdentityFilesRefuseUnsafeInputs(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"relative", "alias", "pin", "zero pin", "wildcard", "key mismatch", "junk cert", "junk key", "junk CA", "leaf as CA", "duplicate CA", "symlink", "hardlink", "public key file", "directory", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			files, other := provisionedObjectReadTestFiles(t)
			read := func(p string) []byte {
				v, e := os.ReadFile(p)
				if e != nil {
					t.Fatal(e)
				}
				return v
			}
			write := func(p string, v []byte) {
				if e := os.WriteFile(p, v, 0600); e != nil {
					t.Fatal(e)
				}
			}
			switch kind {
			case "relative":
				files.Certificate = "cert.pem"
			case "alias":
				files.PrivateKey = files.Certificate
			case "pin":
				files.PeerSHA256 = "not a fingerprint"
			case "zero pin":
				files.PeerSHA256 = strings.Repeat("0", 64)
			case "wildcard":
				files.PeerDNSName = "*.test"
			case "key mismatch":
				write(files.PrivateKey, read(other.PrivateKey))
			case "junk cert":
				write(files.Certificate, append([]byte("junk\n"), read(files.Certificate)...))
			case "junk key":
				write(files.PrivateKey, append(read(files.PrivateKey), []byte("junk")...))
			case "junk CA":
				write(files.Authorities, append(read(files.Authorities), []byte("junk")...))
			case "leaf as CA":
				write(files.Authorities, read(files.Certificate))
			case "duplicate CA":
				raw := read(files.Authorities)
				write(files.Authorities, append(raw, raw...))
			case "symlink":
				if e := os.Remove(files.PrivateKey); e != nil {
					t.Fatal(e)
				}
				if e := os.Symlink(other.PrivateKey, files.PrivateKey); e != nil {
					t.Fatal(e)
				}
			case "hardlink":
				if e := os.Link(files.PrivateKey, filepath.Join(t.TempDir(), "alias")); e != nil {
					t.Fatal(e)
				}
			case "public key file":
				if e := os.Chmod(files.PrivateKey, 0644); e != nil {
					t.Fatal(e)
				}
			case "directory":
				files.PrivateKey = t.TempDir()
			case "oversized":
				write(files.PrivateKey, []byte(strings.Repeat("x", 64*1024+1)))
			}
			identity, err := loadObjectReadPeerIdentity(files, false)
			want := errObjectReadPeer
			if kind == "junk cert" {
				want = errObjectReadCertificatePEM
			}
			if err != want || validObjectReadPeerIdentity(identity) {
				t.Fatal("unsafe identity accepted or unsanitized error", err)
			}
		})
	}
}

func TestObjectReadIdentityFilesCertificatePolicy(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"expired leaf", "future leaf", "CA leaf", "wrong role", "wrong server role", "expired CA", "future CA", "non-signing CA"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			files, _ := provisionedObjectReadTestFilesWith(t, func(c *x509.Certificate, ca bool) {
				if ca {
					switch kind {
					case "expired CA":
						c.NotAfter = time.Now().Add(-time.Minute)
					case "future CA":
						c.NotBefore = time.Now().Add(time.Minute)
					case "non-signing CA":
						c.KeyUsage = x509.KeyUsageDigitalSignature
					}
				}
				if !ca {
					switch kind {
					case "expired leaf":
						c.NotAfter = time.Now().Add(-time.Minute)
					case "future leaf":
						c.NotBefore = time.Now().Add(time.Minute)
					case "CA leaf":
						c.IsCA = true
						c.BasicConstraintsValid = true
					case "wrong role":
						c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
					case "wrong server role":
						c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
					}
				}
			})
			identity, err := loadObjectReadPeerIdentity(files, kind == "wrong server role")
			if err != errObjectReadPeer || validObjectReadPeerIdentity(identity) {
				t.Fatal("invalid certificate policy accepted", err)
			}
		})
	}
}

func TestObjectReadIdentityFilesMissingPEMMarkersExplained(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"missing begin", "missing end", "missing both", "DER", "empty", "wrong marker", "missing key too"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			files, _ := provisionedObjectReadTestFiles(t)
			raw, err := os.ReadFile(files.Certificate)
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "missing begin":
				raw = []byte(strings.ReplaceAll(string(raw), "-----BEGIN CERTIFICATE-----", ""))
			case "missing end":
				raw = []byte(strings.ReplaceAll(string(raw), "-----END CERTIFICATE-----", ""))
			case "missing both", "missing key too":
				raw = []byte(strings.ReplaceAll(strings.ReplaceAll(string(raw), "-----BEGIN CERTIFICATE-----", ""), "-----END CERTIFICATE-----", ""))
				if kind == "missing key too" {
					if err := os.Remove(files.PrivateKey); err != nil {
						t.Fatal(err)
					}
				}
			case "DER":
				block, _ := pem.Decode(raw)
				raw = block.Bytes
			case "empty":
				raw = nil
			case "wrong marker":
				raw = []byte(strings.ReplaceAll(string(raw), "CERTIFICATE", "PUBLIC KEY"))
			}
			if err := os.WriteFile(files.Certificate, raw, 0600); err != nil {
				t.Fatal(err)
			}
			identity, err := loadObjectReadPeerIdentity(files, false)
			const want = "invalid certificate format: expected PEM with BEGIN CERTIFICATE and END CERTIFICATE markers"
			if err != errObjectReadCertificatePEM || err.Error() != want || validObjectReadPeerIdentity(identity) {
				t.Fatal("missing PEM markers not explained safely", err)
			}
		})
	}
}
