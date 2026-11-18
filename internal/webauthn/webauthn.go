package webauthn

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"
)

// COSE algorithm identifiers accepted.
const (
	AlgES256 = -7
	AlgRS256 = -257
)

const (
	flagUP = 0x01
	flagUV = 0x04
	flagAT = 0x40
	flagED = 0x80

	maxCredentialIDLength = 1023
	minRSABits            = 2048
	maxRSABits            = 4096
)

var (
	ErrInvalid       = errors.New("passkey response is invalid")
	ErrSignCount     = errors.New("passkey signature counter went backwards")
	ErrUnsupportedAl = errors.New("passkey algorithm is not supported")
)

// RegistrationParams are the values a relying party verifies a new credential against.
type RegistrationParams struct {
	AttestationObject []byte
	ClientDataJSON    []byte
	Challenge         []byte
	Origin            string // exact origin the browser reported, e.g. https://acme.rooms.example.com
	RPID              string // host name without port
}

type RegistrationResult struct {
	CredentialID []byte
	PublicKeyDER []byte // PKIX
	Alg          int
	SignCount    uint32
	BackedUp     bool
}

type AssertionParams struct {
	AuthenticatorData []byte
	ClientDataJSON    []byte
	Signature         []byte
	Challenge         []byte
	Origin            string
	RPID              string
	PublicKeyDER      []byte
	Alg               int
	StoredSignCount   uint32
}

type clientData struct {
	Type        string `json:"type"`
	Challenge   string `json:"challenge"`
	Origin      string `json:"origin"`
	CrossOrigin *bool  `json:"crossOrigin"`
	TopOrigin   string `json:"topOrigin"`
}

func checkClientData(raw []byte, wantType string, challenge []byte, origin string) error {
	var cd clientData
	if len(raw) > 4096 || json.Unmarshal(raw, &cd) != nil {
		return ErrInvalid
	}
	got, err := base64.RawURLEncoding.DecodeString(cd.Challenge)
	if err != nil || len(challenge) < 16 || subtle.ConstantTimeCompare(got, challenge) != 1 {
		return ErrInvalid
	}
	if cd.Type != wantType || subtle.ConstantTimeCompare([]byte(cd.Origin), []byte(origin)) != 1 {
		return ErrInvalid
	}
	if (cd.CrossOrigin != nil && *cd.CrossOrigin) || cd.TopOrigin != "" {
		return ErrInvalid // embedded in another site
	}
	return nil
}

type authData struct {
	flags     byte
	signCount uint32
	credID    []byte
	coseKey   value
}

func parseAuthData(raw []byte, rpID string, expectAttested bool) (authData, error) {
	if len(raw) < 37 {
		return authData{}, ErrInvalid
	}
	want := sha256.Sum256([]byte(rpID))
	if subtle.ConstantTimeCompare(raw[:32], want[:]) != 1 {
		return authData{}, ErrInvalid
	}
	ad := authData{flags: raw[32], signCount: binary.BigEndian.Uint32(raw[33:37])}
	if ad.flags&flagUP == 0 || ad.flags&flagUV == 0 {
		return authData{}, ErrInvalid // require presence and verification of the person
	}
	rest := raw[37:]
	if (ad.flags&flagAT != 0) != expectAttested {
		return authData{}, ErrInvalid
	}
	if expectAttested {
		if len(rest) < 18 {
			return authData{}, ErrInvalid
		}
		idLen := int(binary.BigEndian.Uint16(rest[16:18]))
		rest = rest[18:]
		if idLen < 16 || idLen > maxCredentialIDLength || len(rest) < idLen {
			return authData{}, ErrInvalid
		}
		ad.credID = append([]byte(nil), rest[:idLen]...)
		key, after, err := decode(rest[idLen:])
		if err != nil || key.kind != kMap {
			return authData{}, ErrInvalid
		}
		ad.coseKey, rest = key, after
	}
	if ad.flags&flagED != 0 {
		ext, after, err := decode(rest)
		if err != nil || ext.kind != kMap {
			return authData{}, ErrInvalid
		}
		rest = after
	}
	if len(rest) != 0 {
		return authData{}, ErrInvalid
	}
	return ad, nil
}

// publicKeyFromCOSE converts a COSE key to a PKIX DER public key, accepting only ES256
// on P-256 and RS256 with 2048-4096 bit keys.
func publicKeyFromCOSE(k value) (der []byte, alg int, err error) {
	kty, ok1 := k.intKey(1)
	algv, ok2 := k.intKey(3)
	if !ok1 || !ok2 || kty.kind != kUint || (algv.kind != kNegInt && algv.kind != kUint) {
		return nil, 0, ErrInvalid
	}
	switch {
	case kty.n == 2 && algv.n == AlgES256:
		crv, ok := k.intKey(-1)
		x, okx := k.intKey(-2)
		y, oky := k.intKey(-3)
		if !ok || !okx || !oky || crv.kind != kUint || crv.n != 1 || x.kind != kBytes || y.kind != kBytes || len(x.bytes) != 32 || len(y.bytes) != 32 {
			return nil, 0, ErrInvalid
		}
		point := append([]byte{4}, append(append([]byte(nil), x.bytes...), y.bytes...)...)
		pub, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), point)
		if err != nil {
			return nil, 0, ErrInvalid // not a point on the curve
		}
		der, err = x509.MarshalPKIXPublicKey(pub)
		return der, AlgES256, err
	case kty.n == 3 && algv.n == AlgRS256:
		n, okn := k.intKey(-1)
		e, oke := k.intKey(-2)
		if !okn || !oke || n.kind != kBytes || e.kind != kBytes || len(e.bytes) == 0 || len(e.bytes) > 4 {
			return nil, 0, ErrInvalid
		}
		modulus := new(big.Int).SetBytes(n.bytes)
		exponent := int(new(big.Int).SetBytes(e.bytes).Int64())
		if modulus.BitLen() < minRSABits || modulus.BitLen() > maxRSABits || exponent < 65537 || exponent%2 == 0 || modulus.Bit(0) == 0 {
			return nil, 0, ErrInvalid
		}
		der, err = x509.MarshalPKIXPublicKey(&rsa.PublicKey{N: modulus, E: exponent})
		return der, AlgRS256, err
	}
	return nil, 0, ErrUnsupportedAl
}

// VerifyRegistration checks a new passkey and returns what to store.
func VerifyRegistration(p RegistrationParams) (RegistrationResult, error) {
	if err := checkClientData(p.ClientDataJSON, "webauthn.create", p.Challenge, p.Origin); err != nil {
		return RegistrationResult{}, err
	}
	obj, rest, err := decode(p.AttestationObject)
	if err != nil || len(rest) != 0 || obj.kind != kMap || len(p.AttestationObject) > 16384 {
		return RegistrationResult{}, ErrInvalid
	}
	format, okf := obj.textKey("fmt")
	stmt, oks := obj.textKey("attStmt")
	raw, okd := obj.textKey("authData")
	if !okf || !oks || !okd || format.kind != kText || format.text != "none" || stmt.kind != kMap || len(stmt.pairs) != 0 || raw.kind != kBytes || len(obj.pairs) != 3 {
		return RegistrationResult{}, ErrInvalid // only `none` attestation is accepted
	}
	ad, err := parseAuthData(raw.bytes, p.RPID, true)
	if err != nil {
		return RegistrationResult{}, err
	}
	der, alg, err := publicKeyFromCOSE(ad.coseKey)
	if err != nil {
		return RegistrationResult{}, err
	}
	return RegistrationResult{CredentialID: ad.credID, PublicKeyDER: der, Alg: alg, SignCount: ad.signCount, BackedUp: ad.flags&0x10 != 0}, nil
}

// VerifyAssertion checks a sign-in and returns the credential's new signature counter.
func VerifyAssertion(p AssertionParams) (uint32, error) {
	if err := checkClientData(p.ClientDataJSON, "webauthn.get", p.Challenge, p.Origin); err != nil {
		return 0, err
	}
	ad, err := parseAuthData(p.AuthenticatorData, p.RPID, false)
	if err != nil {
		return 0, err
	}
	// A counter that does not advance suggests a cloned authenticator. Many platform
	// passkeys always report zero, which is accepted only while the stored value is zero too.
	if ad.signCount != 0 || p.StoredSignCount != 0 {
		if ad.signCount <= p.StoredSignCount {
			return 0, ErrSignCount
		}
	}
	pub, err := x509.ParsePKIXPublicKey(p.PublicKeyDER)
	if err != nil {
		return 0, ErrInvalid
	}
	hash := sha256.Sum256(p.ClientDataJSON)
	signed := sha256.Sum256(append(append([]byte(nil), p.AuthenticatorData...), hash[:]...))
	switch key := pub.(type) {
	case *ecdsa.PublicKey:
		if p.Alg != AlgES256 || key.Curve != elliptic.P256() || !ecdsa.VerifyASN1(key, signed[:], p.Signature) {
			return 0, ErrInvalid
		}
	case *rsa.PublicKey:
		if p.Alg != AlgRS256 || key.N.BitLen() < minRSABits || rsa.VerifyPKCS1v15(key, crypto.SHA256, signed[:], p.Signature) != nil {
			return 0, ErrInvalid
		}
	default:
		return 0, ErrUnsupportedAl
	}
	return ad.signCount, nil
}
