package webauthn

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// A minimal CBOR encoder for building authenticator output in tests.
func cbHead(major byte, n uint64) []byte {
	switch {
	case n < 24:
		return []byte{major<<5 | byte(n)}
	case n < 256:
		return []byte{major<<5 | 24, byte(n)}
	case n < 65536:
		return []byte{major<<5 | 25, byte(n >> 8), byte(n)}
	}
	b := make([]byte, 5)
	b[0] = major<<5 | 26
	binary.BigEndian.PutUint32(b[1:], uint32(n))
	return b
}
func cbInt(n int64) []byte {
	if n >= 0 {
		return cbHead(0, uint64(n))
	}
	return cbHead(1, uint64(-1-n))
}
func cbBytes(b []byte) []byte { return append(cbHead(2, uint64(len(b))), b...) }
func cbText(s string) []byte  { return append(cbHead(3, uint64(len(s))), s...) }
func cbMap(pairs ...[]byte) []byte {
	out := cbHead(5, uint64(len(pairs)/2))
	for _, p := range pairs {
		out = append(out, p...)
	}
	return out
}

const testRP, testOrigin = "acme.rooms.example.com", "https://acme.rooms.example.com"

type authenticator struct {
	ec     *ecdsa.PrivateKey
	rsa    *rsa.PrivateKey
	credID []byte
	count  uint32
}

func newES256() *authenticator {
	k, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	id := make([]byte, 32)
	rand.Read(id)
	return &authenticator{ec: k, credID: id}
}

func newRS256() *authenticator {
	k, _ := rsa.GenerateKey(rand.Reader, 2048)
	id := make([]byte, 32)
	rand.Read(id)
	return &authenticator{rsa: k, credID: id}
}

func (a *authenticator) cose() []byte {
	if a.ec != nil {
		x, y := a.ec.PublicKey.X.FillBytes(make([]byte, 32)), a.ec.PublicKey.Y.FillBytes(make([]byte, 32))
		return cbMap(cbInt(1), cbInt(2), cbInt(3), cbInt(AlgES256), cbInt(-1), cbInt(1), cbInt(-2), cbBytes(x), cbInt(-3), cbBytes(y))
	}
	return cbMap(cbInt(1), cbInt(3), cbInt(3), cbInt(AlgRS256), cbInt(-1), cbBytes(a.rsa.N.Bytes()), cbInt(-2), cbBytes([]byte{1, 0, 1}))
}

func authDataBytes(rp string, flags byte, count uint32, attested []byte) []byte {
	h := sha256.Sum256([]byte(rp))
	out := append(append([]byte(nil), h[:]...), flags)
	out = binary.BigEndian.AppendUint32(out, count)
	return append(out, attested...)
}

func (a *authenticator) attested() []byte {
	out := make([]byte, 16) // zero AAGUID
	out = binary.BigEndian.AppendUint16(out, uint16(len(a.credID)))
	out = append(out, a.credID...)
	return append(out, a.cose()...)
}

func clientJSON(typ string, challenge []byte, origin string, extra map[string]any) []byte {
	m := map[string]any{"type": typ, "challenge": base64.RawURLEncoding.EncodeToString(challenge), "origin": origin, "crossOrigin": false, "other_keys_can_be_added_here": "do not compare clientDataJSON against a template"}
	for k, v := range extra {
		m[k] = v
	}
	raw, _ := json.Marshal(m)
	return raw
}

func attestationObject(format string, authData []byte, stmt []byte) []byte {
	if stmt == nil {
		stmt = cbMap()
	}
	return cbMap(cbText("fmt"), cbText(format), cbText("attStmt"), stmt, cbText("authData"), cbBytes(authData))
}

func challenge() []byte {
	c := make([]byte, 32)
	rand.Read(c)
	return c
}

func (a *authenticator) register(ch []byte) RegistrationParams {
	ad := authDataBytes(testRP, flagUP|flagUV|flagAT, a.count, a.attested())
	return RegistrationParams{AttestationObject: attestationObject("none", ad, nil), ClientDataJSON: clientJSON("webauthn.create", ch, testOrigin, nil), Challenge: ch, Origin: testOrigin, RPID: testRP}
}

func (a *authenticator) assert(ch []byte, mutate func(ad []byte) []byte) AssertionParams {
	a.count++
	ad := authDataBytes(testRP, flagUP|flagUV, a.count, nil)
	if mutate != nil {
		ad = mutate(ad)
	}
	cd := clientJSON("webauthn.get", ch, testOrigin, nil)
	hash := sha256.Sum256(cd)
	digest := sha256.Sum256(append(append([]byte(nil), ad...), hash[:]...))
	var sig []byte
	if a.ec != nil {
		sig, _ = ecdsa.SignASN1(rand.Reader, a.ec, digest[:])
	} else {
		sig, _ = rsa.SignPKCS1v15(rand.Reader, a.rsa, crypto.SHA256, digest[:])
	}
	return AssertionParams{AuthenticatorData: ad, ClientDataJSON: cd, Signature: sig, Challenge: ch, Origin: testOrigin, RPID: testRP}
}

func TestRegisterAndSignInWithES256AndRS256(t *testing.T) {
	t.Parallel()
	for name, a := range map[string]*authenticator{"ES256": newES256(), "RS256": newRS256()} {
		reg, err := VerifyRegistration(a.register(challenge()))
		if err != nil {
			t.Fatal(name, err)
		}
		if !reflect.DeepEqual(reg.CredentialID, a.credID) || len(reg.PublicKeyDER) == 0 || (name == "ES256") != (reg.Alg == AlgES256) {
			t.Fatal(name, reg)
		}
		for i := 0; i < 3; i++ {
			p := a.assert(challenge(), nil)
			p.PublicKeyDER, p.Alg, p.StoredSignCount = reg.PublicKeyDER, reg.Alg, uint32(i)
			count, err := VerifyAssertion(p)
			if err != nil || count != uint32(i+1) {
				t.Fatal(name, i, count, err)
			}
		}
	}
}

func TestRegistrationRefusals(t *testing.T) {
	t.Parallel()
	a := newES256()
	good := a.register(challenge())
	attested := func(flags byte, rp string, tail []byte) []byte {
		return attestationObject("none", authDataBytes(rp, flags, 0, append(a.attested(), tail...)), nil)
	}
	mutations := map[string]func(p *RegistrationParams){
		"wrong challenge": func(p *RegistrationParams) { p.Challenge = challenge() },
		"wrong origin":    func(p *RegistrationParams) { p.Origin = "https://evil.example.com" },
		"wrong rp":        func(p *RegistrationParams) { p.RPID = "evil.example.com" },
		"get instead of create": func(p *RegistrationParams) {
			p.ClientDataJSON = clientJSON("webauthn.get", p.Challenge, testOrigin, nil)
		},
		"cross origin": func(p *RegistrationParams) {
			p.ClientDataJSON = clientJSON("webauthn.create", p.Challenge, testOrigin, map[string]any{"crossOrigin": true})
		},
		"top origin": func(p *RegistrationParams) {
			p.ClientDataJSON = clientJSON("webauthn.create", p.Challenge, testOrigin, map[string]any{"topOrigin": "https://evil.test"})
		},
		"short challenge": func(p *RegistrationParams) {
			p.Challenge = []byte("short")
			p.ClientDataJSON = clientJSON("webauthn.create", p.Challenge, testOrigin, nil)
		},
		"packed attestation": func(p *RegistrationParams) {
			p.AttestationObject = attestationObject("packed", authDataBytes(testRP, flagUP|flagUV|flagAT, 0, a.attested()), cbMap(cbText("alg"), cbInt(-7), cbText("sig"), cbBytes([]byte("x"))))
		},
		"none with statement": func(p *RegistrationParams) {
			p.AttestationObject = attestationObject("none", authDataBytes(testRP, flagUP|flagUV|flagAT, 0, a.attested()), cbMap(cbText("x"), cbInt(1)))
		},
		"no user verification":  func(p *RegistrationParams) { p.AttestationObject = attested(flagUP|flagAT, testRP, nil) },
		"no user presence":      func(p *RegistrationParams) { p.AttestationObject = attested(flagUV|flagAT, testRP, nil) },
		"no attested data flag": func(p *RegistrationParams) { p.AttestationObject = attested(flagUP|flagUV, testRP, nil) },
		"wrong rp hash": func(p *RegistrationParams) {
			p.AttestationObject = attested(flagUP|flagUV|flagAT, "evil.example.com", nil)
		},
		"trailing bytes":         func(p *RegistrationParams) { p.AttestationObject = attested(flagUP|flagUV|flagAT, testRP, []byte{0}) },
		"extension flag no data": func(p *RegistrationParams) { p.AttestationObject = attested(flagUP|flagUV|flagAT|flagED, testRP, nil) },
		"extra attestation field": func(p *RegistrationParams) {
			ad := authDataBytes(testRP, flagUP|flagUV|flagAT, 0, a.attested())
			p.AttestationObject = cbMap(cbText("fmt"), cbText("none"), cbText("attStmt"), cbMap(), cbText("authData"), cbBytes(ad), cbText("extra"), cbInt(1))
		},
		"truncated authData": func(p *RegistrationParams) {
			p.AttestationObject = attestationObject("none", authDataBytes(testRP, flagUP|flagUV|flagAT, 0, a.attested()[:20]), nil)
		},
		"credential id too short": func(p *RegistrationParams) {
			short := &authenticator{ec: a.ec, credID: []byte("short")}
			p.AttestationObject = attestationObject("none", authDataBytes(testRP, flagUP|flagUV|flagAT, 0, short.attested()), nil)
		},
		"credential id too long": func(p *RegistrationParams) {
			long := &authenticator{ec: a.ec, credID: make([]byte, 1024)}
			p.AttestationObject = attestationObject("none", authDataBytes(testRP, flagUP|flagUV|flagAT, 0, long.attested()), nil)
		},
		"empty":                func(p *RegistrationParams) { p.AttestationObject = nil },
		"garbage":              func(p *RegistrationParams) { p.AttestationObject = []byte{0xff, 0xff, 0xff} },
		"client data not json": func(p *RegistrationParams) { p.ClientDataJSON = []byte("nope") },
		"client data too large": func(p *RegistrationParams) {
			p.ClientDataJSON = []byte(`{"type":"` + strings.Repeat("a", 5000) + `"}`)
		},
	}
	for name, mutate := range mutations {
		p := good
		mutate(&p)
		if _, err := VerifyRegistration(p); err == nil {
			t.Fatal("accepted:", name)
		}
	}
	if _, err := VerifyRegistration(good); err != nil {
		t.Fatal("the unmodified registration must pass", err)
	}
}

func TestCredentialKeyRefusals(t *testing.T) {
	t.Parallel()
	es := newES256()
	x, y := es.ec.PublicKey.X.FillBytes(make([]byte, 32)), es.ec.PublicKey.Y.FillBytes(make([]byte, 32))
	build := func(cose []byte) RegistrationParams {
		id := make([]byte, 32)
		out := make([]byte, 16)
		out = binary.BigEndian.AppendUint16(out, 32)
		out = append(out, id...)
		out = append(out, cose...)
		ch := challenge()
		return RegistrationParams{AttestationObject: attestationObject("none", authDataBytes(testRP, flagUP|flagUV|flagAT, 0, out), nil), ClientDataJSON: clientJSON("webauthn.create", ch, testOrigin, nil), Challenge: ch, Origin: testOrigin, RPID: testRP}
	}
	zero := make([]byte, 32)
	small, _ := rsa.GenerateKey(rand.Reader, 1024)
	big := newRS256().rsa.N.Bytes()
	cases := map[string][]byte{
		"EdDSA":              cbMap(cbInt(1), cbInt(1), cbInt(3), cbInt(-8), cbInt(-1), cbInt(6), cbInt(-2), cbBytes(x)),
		"ES384":              cbMap(cbInt(1), cbInt(2), cbInt(3), cbInt(-35), cbInt(-1), cbInt(2), cbInt(-2), cbBytes(x), cbInt(-3), cbBytes(y)),
		"P-256 off curve":    cbMap(cbInt(1), cbInt(2), cbInt(3), cbInt(AlgES256), cbInt(-1), cbInt(1), cbInt(-2), cbBytes(x), cbInt(-3), cbBytes(zero)),
		"wrong curve id":     cbMap(cbInt(1), cbInt(2), cbInt(3), cbInt(AlgES256), cbInt(-1), cbInt(2), cbInt(-2), cbBytes(x), cbInt(-3), cbBytes(y)),
		"short coordinate":   cbMap(cbInt(1), cbInt(2), cbInt(3), cbInt(AlgES256), cbInt(-1), cbInt(1), cbInt(-2), cbBytes(x[:31]), cbInt(-3), cbBytes(y)),
		"missing alg":        cbMap(cbInt(1), cbInt(2), cbInt(-1), cbInt(1), cbInt(-2), cbBytes(x), cbInt(-3), cbBytes(y)),
		"RSA 1024":           cbMap(cbInt(1), cbInt(3), cbInt(3), cbInt(AlgRS256), cbInt(-1), cbBytes(small.N.Bytes()), cbInt(-2), cbBytes([]byte{1, 0, 1})),
		"RSA exponent 3":     cbMap(cbInt(1), cbInt(3), cbInt(3), cbInt(AlgRS256), cbInt(-1), cbBytes(big), cbInt(-2), cbBytes([]byte{3})),
		"RSA with ES alg":    cbMap(cbInt(1), cbInt(3), cbInt(3), cbInt(AlgES256), cbInt(-1), cbBytes(big), cbInt(-2), cbBytes([]byte{1, 0, 1})),
		"EC with RSA alg":    cbMap(cbInt(1), cbInt(2), cbInt(3), cbInt(AlgRS256), cbInt(-1), cbInt(1), cbInt(-2), cbBytes(x), cbInt(-3), cbBytes(y)),
		"octet key pair":     cbMap(cbInt(1), cbInt(4), cbInt(3), cbInt(AlgES256)),
		"duplicate map keys": cbMap(cbInt(1), cbInt(2), cbInt(1), cbInt(2)),
		"not a map":          cbBytes([]byte("x")),
	}
	for name, cose := range cases {
		if _, err := VerifyRegistration(build(cose)); err == nil {
			t.Fatal("accepted a bad credential key:", name)
		}
	}
}

func TestAssertionRefusals(t *testing.T) {
	t.Parallel()
	a := newES256()
	reg, _ := VerifyRegistration(a.register(challenge()))
	other := newES256()
	otherReg, _ := VerifyRegistration(other.register(challenge()))
	rsaReg, _ := VerifyRegistration(newRS256().register(challenge()))
	ready := func(mutate func(ad []byte) []byte) AssertionParams {
		p := a.assert(challenge(), mutate)
		p.PublicKeyDER, p.Alg = reg.PublicKeyDER, reg.Alg
		return p
	}
	if _, err := VerifyAssertion(ready(nil)); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(p *AssertionParams){
		"wrong challenge": func(p *AssertionParams) { p.Challenge = challenge() },
		"wrong origin":    func(p *AssertionParams) { p.Origin = "https://evil.example.com" },
		"wrong rp":        func(p *AssertionParams) { p.RPID = "evil.example.com" },
		"create type": func(p *AssertionParams) {
			p.ClientDataJSON = clientJSON("webauthn.create", p.Challenge, testOrigin, nil)
		},
		"tampered client data": func(p *AssertionParams) {
			p.ClientDataJSON = append(append([]byte(nil), p.ClientDataJSON[:len(p.ClientDataJSON)-1]...), []byte(`,"x":1}`)...)
		},
		"tampered signature": func(p *AssertionParams) {
			p.Signature = append([]byte(nil), p.Signature...)
			p.Signature[len(p.Signature)-1] ^= 1
		},
		"empty signature":     func(p *AssertionParams) { p.Signature = nil },
		"other authenticator": func(p *AssertionParams) { p.PublicKeyDER = otherReg.PublicKeyDER },
		"wrong algorithm":     func(p *AssertionParams) { p.Alg = AlgRS256 },
		"rsa key as ec":       func(p *AssertionParams) { p.PublicKeyDER, p.Alg = rsaReg.PublicKeyDER, AlgES256 },
		"garbage key":         func(p *AssertionParams) { p.PublicKeyDER = []byte("nope") },
		"replayed counter":    func(p *AssertionParams) { p.StoredSignCount = 99 },
		"flags tampered": func(p *AssertionParams) {
			p.AuthenticatorData = append([]byte(nil), p.AuthenticatorData...)
			p.AuthenticatorData[32] = flagUP
		},
	}
	for name, mutate := range cases {
		p := ready(nil)
		mutate(&p)
		if _, err := VerifyAssertion(p); err == nil {
			t.Fatal("accepted:", name)
		}
	}
	p := ready(nil)
	p.StoredSignCount = 50
	if _, err := VerifyAssertion(p); err != ErrSignCount {
		t.Fatal(err)
	}
	// Authenticators that never count (zero) work; a drop to zero after counting is refused.
	zero := &authenticator{ec: a.ec, credID: a.credID}
	zp := zero.assert(challenge(), nil)
	zp.AuthenticatorData = authDataBytes(testRP, flagUP|flagUV, 0, nil)
	hash := sha256.Sum256(zp.ClientDataJSON)
	d := sha256.Sum256(append(append([]byte(nil), zp.AuthenticatorData...), hash[:]...))
	zp.Signature, _ = ecdsa.SignASN1(rand.Reader, a.ec, d[:])
	zp.PublicKeyDER, zp.Alg, zp.StoredSignCount = reg.PublicKeyDER, reg.Alg, 0
	if count, err := VerifyAssertion(zp); err != nil || count != 0 {
		t.Fatal("authenticators that never count must work", count, err)
	}
	zp.StoredSignCount = 4
	if _, err := VerifyAssertion(zp); err != ErrSignCount {
		t.Fatal("a counter that drops to zero after counting must be refused", err)
	}
	withAT := ready(func(ad []byte) []byte {
		ad = append([]byte(nil), ad...)
		ad[32] |= flagAT
		return append(ad, a.attested()...)
	})
	if _, err := VerifyAssertion(withAT); err == nil {
		t.Fatal("assertion with attested data accepted")
	}
	for _, mutate := range []func(ad []byte) []byte{func(ad []byte) []byte { return append(ad, 0) }, func(ad []byte) []byte { return ad[:36] }} {
		if _, err := VerifyAssertion(ready(mutate)); err == nil {
			t.Fatal("malformed authenticator data accepted")
		}
	}
}

func TestCBORVectors(t *testing.T) {
	t.Parallel()
	// RFC 8949 appendix A examples.
	good := map[string]func(v value) bool{
		"00":                 func(v value) bool { return v.kind == kUint && v.n == 0 },
		"17":                 func(v value) bool { return v.kind == kUint && v.n == 23 },
		"1818":               func(v value) bool { return v.n == 24 },
		"1903e8":             func(v value) bool { return v.n == 1000 },
		"1a000f4240":         func(v value) bool { return v.n == 1000000 },
		"1b000000e8d4a51000": func(v value) bool { return v.n == 1000000000000 },
		"20":                 func(v value) bool { return v.kind == kNegInt && v.n == -1 },
		"29":                 func(v value) bool { return v.n == -10 },
		"3863":               func(v value) bool { return v.n == -100 },
		"3903e7":             func(v value) bool { return v.n == -1000 },
		"4401020304":         func(v value) bool { return v.kind == kBytes && reflect.DeepEqual(v.bytes, []byte{1, 2, 3, 4}) },
		"6449455446":         func(v value) bool { return v.kind == kText && v.text == "IETF" },
		"83010203":           func(v value) bool { return v.kind == kArray && len(v.array) == 3 && v.array[2].n == 3 },
		"a201020304":         func(v value) bool { x, ok := v.intKey(3); return ok && x.n == 4 },
		"f4":                 func(v value) bool { return v.kind == kBool && v.n == 0 },
		"f5":                 func(v value) bool { return v.kind == kBool && v.n == 1 },
		"f6":                 func(v value) bool { return v.kind == kNull },
	}
	for h, check := range good {
		raw, _ := hex.DecodeString(h)
		v, rest, err := decode(raw)
		if err != nil || len(rest) != 0 || !check(v) {
			t.Fatal("vector", h, v, err)
		}
	}
	bad := map[string]string{
		"half float":          "f93c00",
		"double":              "fb3ff199999999999a",
		"tag":                 "c11a514b67b0",
		"indefinite array":    "9fff",
		"indefinite bytes":    "5f42010243030405ff",
		"indefinite map":      "bf616101ff",
		"reserved info":       "1c",
		"undefined":           "f7",
		"simple value":        "f0",
		"stray break":         "ff",
		"beyond int64":        "1b8000000000000000",
		"truncated bytes":     "4401",
		"truncated array":     "830102",
		"map key is a map":    "a1a1010201",
		"duplicate key":       "a201020103",
		"empty":               "",
		"length beyond input": "5affffffff00",
	}
	for name, h := range bad {
		raw, _ := hex.DecodeString(h)
		if _, _, err := decode(raw); err == nil {
			t.Fatal("accepted", name)
		}
	}
	deep, _ := hex.DecodeString(strings.Repeat("81", 20) + "00")
	if _, _, err := decode(deep); err == nil {
		t.Fatal("deeply nested input accepted")
	}
	many := append([]byte{0x98, 100}, make([]byte, 100)...)
	if _, _, err := decode(many); err == nil {
		t.Fatal("array with too many items accepted")
	}
}

func FuzzDecodeCBOR(f *testing.F) {
	for _, seed := range []string{"a201020304", "83010203", "6449455446", "bf6161ff", "a1a10102"} {
		raw, _ := hex.DecodeString(seed)
		f.Add(raw)
	}
	f.Fuzz(func(t *testing.T, data []byte) { decode(data) })
}

func FuzzVerifyRegistration(f *testing.F) {
	a := newES256()
	p := a.register(challenge())
	f.Add(p.AttestationObject, p.ClientDataJSON)
	f.Fuzz(func(t *testing.T, attestation, clientData []byte) {
		VerifyRegistration(RegistrationParams{AttestationObject: attestation, ClientDataJSON: clientData, Challenge: challenge(), Origin: testOrigin, RPID: testRP})
	})
}
