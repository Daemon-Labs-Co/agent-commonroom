// Package webauthntest is a software passkey authenticator for tests. It builds the exact
// responses a browser forwards from a real authenticator. Never import it from product code.
package webauthntest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
)

type Authenticator struct {
	key    *ecdsa.PrivateKey
	id     []byte
	Count  uint32
	Flags  byte // override for refusal tests; zero means presence + verification
	NoZero bool
}

func New() *Authenticator {
	k, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	id := make([]byte, 32)
	rand.Read(id)
	return &Authenticator{key: k, id: id}
}

func (a *Authenticator) CredentialID() string { return b64(a.id) }

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func head(major byte, n uint64) []byte {
	switch {
	case n < 24:
		return []byte{major<<5 | byte(n)}
	case n < 256:
		return []byte{major<<5 | 24, byte(n)}
	}
	return []byte{major<<5 | 25, byte(n >> 8), byte(n)}
}
func cint(n int64) []byte {
	if n >= 0 {
		return head(0, uint64(n))
	}
	return head(1, uint64(-1-n))
}
func cbytes(b []byte) []byte { return append(head(2, uint64(len(b))), b...) }
func ctext(s string) []byte  { return append(head(3, uint64(len(s))), s...) }

func clientData(typ string, challenge []byte, origin string) []byte {
	raw, _ := json.Marshal(map[string]any{"type": typ, "challenge": b64(challenge), "origin": origin, "crossOrigin": false})
	return raw
}

func (a *Authenticator) flags(extra byte) byte {
	if a.Flags != 0 {
		return a.Flags | extra
	}
	return 0x01 | 0x04 | extra
}

func authData(rp string, flags byte, count uint32, tail []byte) []byte {
	h := sha256.Sum256([]byte(rp))
	out := append(append([]byte(nil), h[:]...), flags)
	out = binary.BigEndian.AppendUint32(out, count)
	return append(out, tail...)
}

// Create answers navigator.credentials.create: base64url credential id, attestation object
// and client data, as the browser would forward them.
func (a *Authenticator) Create(challenge []byte, rp, origin string) (id, attestation, clientDataJSON string) {
	x, y := a.key.X.FillBytes(make([]byte, 32)), a.key.Y.FillBytes(make([]byte, 32))
	cose := append(append([]byte{0xa5}, cint(1)...), cint(2)...)
	cose = append(cose, cint(3)...)
	cose = append(cose, cint(-7)...)
	cose = append(cose, cint(-1)...)
	cose = append(cose, cint(1)...)
	cose = append(cose, cint(-2)...)
	cose = append(cose, cbytes(x)...)
	cose = append(cose, cint(-3)...)
	cose = append(cose, cbytes(y)...)
	attested := append(make([]byte, 16), byte(len(a.id)>>8), byte(len(a.id)))
	attested = append(append(attested, a.id...), cose...)
	ad := authData(rp, a.flags(0x40), a.Count, attested)
	obj := append([]byte{0xa3}, ctext("fmt")...)
	obj = append(obj, ctext("none")...)
	obj = append(obj, ctext("attStmt")...)
	obj = append(obj, 0xa0)
	obj = append(obj, ctext("authData")...)
	obj = append(obj, cbytes(ad)...)
	return b64(a.id), b64(obj), b64(clientData("webauthn.create", challenge, origin))
}

// Get answers navigator.credentials.get.
func (a *Authenticator) Get(challenge []byte, rp, origin string) (authenticatorData, clientDataJSON, signature string) {
	if !a.NoZero {
		a.Count++
	}
	ad := authData(rp, a.flags(0), a.Count, nil)
	cd := clientData("webauthn.get", challenge, origin)
	hash := sha256.Sum256(cd)
	digest := sha256.Sum256(append(append([]byte(nil), ad...), hash[:]...))
	sig, _ := ecdsa.SignASN1(rand.Reader, a.key, digest[:])
	return b64(ad), b64(cd), b64(sig)
}

// Decode converts a base64url challenge from an options response.
func Decode(s string) []byte {
	raw, _ := base64.RawURLEncoding.DecodeString(s)
	return raw
}
