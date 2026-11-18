package oidc

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const clientID = "commonroom-client"

type fakeIdP struct {
	t       *testing.T
	server  *httptest.Server
	rsaKey  *rsa.PrivateKey
	ecKey   *ecdsa.PrivateKey
	mu      sync.Mutex
	jwks    []map[string]any
	issuer  string // what discovery claims; defaults to the server URL
	codes   map[string]codeGrant
	jwksHit atomic.Int64
	tokenFn func(w http.ResponseWriter, r *http.Request) bool // optional override
	now     time.Time
}

type codeGrant struct{ challenge, redirect, nonce string }

func b64u(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func newIdP(t *testing.T) *fakeIdP {
	t.Helper()
	f := &fakeIdP{t: t, codes: map[string]codeGrant{}, now: time.Now()}
	f.rsaKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	f.ecKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	f.jwks = []map[string]any{f.rsaJWK("rsa-1", f.rsaKey), f.ecJWK("ec-1", f.ecKey)}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		issuer := f.issuer
		if issuer == "" {
			issuer = f.server.URL
		}
		json.NewEncoder(w).Encode(map[string]string{"issuer": issuer, "authorization_endpoint": f.server.URL + "/auth", "token_endpoint": f.server.URL + "/token", "jwks_uri": f.server.URL + "/jwks"})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		f.jwksHit.Add(1)
		f.mu.Lock()
		defer f.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]any{"keys": f.jwks})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if f.tokenFn != nil && f.tokenFn(w, r) {
			return
		}
		r.ParseForm()
		grant, ok := f.codes[r.Form.Get("code")]
		sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		if !ok || r.Form.Get("client_id") != clientID || r.Form.Get("client_secret") != "s3cret" || r.Form.Get("redirect_uri") != grant.redirect || b64u(sum[:]) != grant.challenge || r.Form.Get("grant_type") != "authorization_code" {
			http.Error(w, `{"error":"invalid_grant"}`, 400)
			return
		}
		delete(f.codes, r.Form.Get("code")) // codes are single use
		json.NewEncoder(w).Encode(map[string]string{"id_token": f.sign("RS256", "rsa-1", f.claims(grant.nonce)), "token_type": "Bearer"})
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeIdP) rsaJWK(kid string, k *rsa.PrivateKey) map[string]any {
	return map[string]any{"kty": "RSA", "kid": kid, "use": "sig", "n": b64u(k.N.Bytes()), "e": b64u([]byte{1, 0, 1})}
}
func (f *fakeIdP) ecJWK(kid string, k *ecdsa.PrivateKey) map[string]any {
	return map[string]any{"kty": "EC", "kid": kid, "crv": "P-256", "x": b64u(k.X.FillBytes(make([]byte, 32))), "y": b64u(k.Y.FillBytes(make([]byte, 32)))}
}

func (f *fakeIdP) claims(nonce string) map[string]any {
	return map[string]any{"iss": f.server.URL, "sub": "user-123", "aud": clientID, "exp": f.now.Add(time.Hour).Unix(), "iat": f.now.Unix(), "nonce": nonce, "email": "a@example.com", "email_verified": true, "name": "Ada"}
}

func (f *fakeIdP) sign(alg, kid string, claims map[string]any) string {
	header, _ := json.Marshal(map[string]any{"alg": alg, "kid": kid, "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	signingInput := b64u(header) + "." + b64u(payload)
	digest := sha256.Sum256([]byte(signingInput))
	var sig []byte
	switch alg {
	case "RS256":
		sig, _ = rsa.SignPKCS1v15(rand.Reader, f.rsaKey, crypto.SHA256, digest[:])
	case "ES256":
		r, s, _ := ecdsa.Sign(rand.Reader, f.ecKey, digest[:])
		sig = append(r.FillBytes(make([]byte, 32)), s.FillBytes(make([]byte, 32))...)
	}
	return signingInput + "." + b64u(sig)
}

func (f *fakeIdP) provider() *Provider {
	p, err := New(Config{Issuer: f.server.URL, ClientID: clientID, ClientSecret: "s3cret", HTTP: f.server.Client(), AllowLoopbackForTests: true})
	if err != nil {
		f.t.Fatal(err)
	}
	return p
}

func TestAuthorizationCodeFlowWithPKCE(t *testing.T) {
	t.Parallel()
	f := newIdP(t)
	p := f.provider()
	verifier, challenge := PKCE()
	sum := sha256.Sum256([]byte(verifier))
	if challenge != b64u(sum[:]) || len(verifier) < 43 || verifier == func() string { v, _ := PKCE(); return v }() {
		t.Fatal("PKCE values are wrong or not random")
	}
	state, nonce := Random(), Random()
	redirect := "https://acme.rooms.example.com/api/oidc/callback"
	authURL, err := p.AuthURL(t.Context(), redirect, state, nonce, challenge)
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(authURL)
	q := u.Query()
	for k, want := range map[string]string{"response_type": "code", "client_id": clientID, "redirect_uri": redirect, "state": state, "nonce": nonce, "code_challenge": challenge, "code_challenge_method": "S256", "scope": "openid email"} {
		if q.Get(k) != want {
			t.Fatal("authorization URL missing", k, q.Get(k))
		}
	}
	f.codes["abc"] = codeGrant{challenge: challenge, redirect: redirect, nonce: nonce}
	claims, err := p.Exchange(t.Context(), "abc", redirect, verifier, nonce)
	if err != nil || claims.Subject != "user-123" || claims.Issuer != f.server.URL || claims.Email != "a@example.com" || !claims.EmailVerified {
		t.Fatal(claims, err)
	}
	if _, err := p.Exchange(t.Context(), "abc", redirect, verifier, nonce); err == nil {
		t.Fatal("an authorization code was usable twice")
	}
	f.codes["def"] = codeGrant{challenge: challenge, redirect: redirect, nonce: nonce}
	if _, err := p.Exchange(t.Context(), "def", redirect, "a-different-verifier-that-is-long-enough-123456", nonce); err == nil {
		t.Fatal("the PKCE verifier was not checked")
	}
	f.codes["ghi"] = codeGrant{challenge: challenge, redirect: redirect, nonce: nonce}
	if _, err := p.Exchange(t.Context(), "ghi", "https://evil.example.com/cb", verifier, nonce); err == nil {
		t.Fatal("the redirect URI was not bound")
	}
	f.codes["jkl"] = codeGrant{challenge: challenge, redirect: redirect, nonce: "other-nonce"}
	if _, err := p.Exchange(t.Context(), "jkl", redirect, verifier, nonce); err == nil {
		t.Fatal("a token for another nonce was accepted")
	}
}

func TestIDTokenVerification(t *testing.T) {
	t.Parallel()
	f := newIdP(t)
	p := f.provider()
	ok := func(alg, kid string, mutate func(c map[string]any)) (Claims, error) {
		c := f.claims("n-1")
		if mutate != nil {
			mutate(c)
		}
		return p.Verify(t.Context(), f.sign(alg, kid, c), "n-1")
	}
	for _, alg := range []string{"RS256", "ES256"} {
		kid := map[string]string{"RS256": "rsa-1", "ES256": "ec-1"}[alg]
		if claims, err := ok(alg, kid, nil); err != nil || claims.Subject != "user-123" {
			t.Fatal("valid", alg, err)
		}
	}
	bad := map[string]func(c map[string]any){
		"wrong issuer":                  func(c map[string]any) { c["iss"] = "https://evil.example.com" },
		"issuer with slash":             func(c map[string]any) { c["iss"] = f.server.URL + "/" },
		"wrong audience":                func(c map[string]any) { c["aud"] = "someone-else" },
		"audience list without client":  func(c map[string]any) { c["aud"] = []string{"a", "b"} },
		"several audiences without azp": func(c map[string]any) { c["aud"] = []string{clientID, "b"} },
		"several audiences wrong azp":   func(c map[string]any) { c["aud"] = []string{clientID, "b"}; c["azp"] = "b" },
		"azp for another client":        func(c map[string]any) { c["azp"] = "b" },
		"missing audience":              func(c map[string]any) { delete(c, "aud") },
		"expired":                       func(c map[string]any) { c["exp"] = f.now.Add(-2 * time.Minute).Unix() },
		"no expiry":                     func(c map[string]any) { delete(c, "exp") },
		"issued in the future":          func(c map[string]any) { c["iat"] = f.now.Add(5 * time.Minute).Unix() },
		"not yet valid":                 func(c map[string]any) { c["nbf"] = f.now.Add(5 * time.Minute).Unix() },
		"wrong nonce":                   func(c map[string]any) { c["nonce"] = "n-2" },
		"missing nonce":                 func(c map[string]any) { delete(c, "nonce") },
		"no subject":                    func(c map[string]any) { c["sub"] = "" },
		"huge subject":                  func(c map[string]any) { c["sub"] = strings.Repeat("s", 256) },
	}
	for name, mutate := range bad {
		if _, err := ok("RS256", "rsa-1", mutate); err == nil {
			t.Fatal("accepted:", name)
		}
	}
	if _, err := ok("RS256", "rsa-1", func(c map[string]any) { c["exp"] = f.now.Add(-30 * time.Second).Unix() }); err != nil {
		t.Fatal("a token a few seconds past expiry is within clock skew", err)
	}
	if _, err := p.Verify(t.Context(), f.sign("RS256", "rsa-1", f.claims("n-1")), ""); err == nil {
		t.Fatal("an empty expected nonce must refuse")
	}
}

func TestIDTokenSignatureAndAlgorithmAttacks(t *testing.T) {
	t.Parallel()
	f := newIdP(t)
	p := f.provider()
	good := f.sign("RS256", "rsa-1", f.claims("n-1"))
	parts := strings.Split(good, ".")
	craft := func(header map[string]any, signature string) string {
		h, _ := json.Marshal(header)
		return b64u(h) + "." + parts[1] + "." + signature
	}
	// Algorithm confusion and unsigned tokens.
	hmacKey := f.rsaKey.N.Bytes()
	mac := hmac.New(sha256.New, hmacKey)
	hs, _ := json.Marshal(map[string]any{"alg": "HS256", "kid": "rsa-1"})
	mac.Write([]byte(b64u(hs) + "." + parts[1]))
	attacks := map[string]string{
		"alg none":                  craft(map[string]any{"alg": "none", "kid": "rsa-1"}, ""),
		"alg None":                  craft(map[string]any{"alg": "None", "kid": "rsa-1"}, ""),
		"HS256 with the public key": b64u(hs) + "." + parts[1] + "." + b64u(mac.Sum(nil)),
		"PS256":                     craft(map[string]any{"alg": "PS256", "kid": "rsa-1"}, parts[2]),
		"RS512":                     craft(map[string]any{"alg": "RS512", "kid": "rsa-1"}, parts[2]),
		"critical header":           craft(map[string]any{"alg": "RS256", "kid": "rsa-1", "crit": []string{"x"}}, parts[2]),
		"RS256 against the EC key":  craft(map[string]any{"alg": "RS256", "kid": "ec-1"}, parts[2]),
		"ES256 against the RSA key": craft(map[string]any{"alg": "ES256", "kid": "rsa-1"}, b64u(make([]byte, 64))),
		"unknown key id":            craft(map[string]any{"alg": "RS256", "kid": "nope"}, parts[2]),
		"tampered payload":          parts[0] + "." + b64u([]byte(`{"iss":"x"}`)) + "." + parts[2],
		"tampered signature":        parts[0] + "." + parts[1] + "." + b64u(make([]byte, 256)),
		"two segments":              parts[0] + "." + parts[1],
		"four segments":             good + ".x",
		"not base64":                "!!!.!!!.!!!",
		"empty":                     "",
		"oversized":                 strings.Repeat("a", maxTokenLength+1),
	}
	for name, token := range attacks {
		if _, err := p.Verify(t.Context(), token, "n-1"); err == nil {
			t.Fatal("accepted:", name)
		}
	}
	// A short ES256 signature is not padded or reinterpreted.
	es := f.sign("ES256", "ec-1", f.claims("n-1"))
	esParts := strings.Split(es, ".")
	raw, _ := base64.RawURLEncoding.DecodeString(esParts[2])
	for _, sig := range [][]byte{raw[:63], append(raw, 0), nil} {
		if _, err := p.Verify(t.Context(), esParts[0]+"."+esParts[1]+"."+b64u(sig), "n-1"); err == nil {
			t.Fatal("a malformed ES256 signature was accepted")
		}
	}
	// A token signed by an attacker's own key under a trusted kid fails.
	other, _ := rsa.GenerateKey(rand.Reader, 2048)
	forger := &fakeIdP{rsaKey: other, server: f.server, now: f.now}
	if _, err := p.Verify(t.Context(), forger.sign("RS256", "rsa-1", f.claims("n-1")), "n-1"); err == nil {
		t.Fatal("a token signed by another key was accepted")
	}
}

func TestKeySetHandling(t *testing.T) {
	t.Parallel()
	f := newIdP(t)
	small, _ := rsa.GenerateKey(rand.Reader, 1024)
	f.jwks = append(f.jwks, f.rsaJWK("small", small), map[string]any{"kty": "RSA", "kid": "enc", "use": "enc", "n": b64u(f.rsaKey.N.Bytes()), "e": "AQAB"}, map[string]any{"kty": "oct", "kid": "sym", "k": "AAAA"})
	p := f.provider()
	if _, err := p.Verify(t.Context(), f.sign("RS256", "rsa-1", f.claims("n")), "n"); err != nil {
		t.Fatal(err)
	}
	for _, kid := range []string{"small", "enc", "sym"} {
		if _, err := p.Verify(t.Context(), f.sign("RS256", kid, f.claims("n")), "n"); err == nil {
			t.Fatal("a weak or non-signing key was usable:", kid)
		}
	}
	// A key rotation is picked up once; repeated unknown ids cannot hammer the provider.
	hits := f.jwksHit.Load()
	for i := 0; i < 30; i++ {
		p.Verify(t.Context(), f.sign("RS256", "rotated", f.claims("n")), "n")
	}
	if extra := f.jwksHit.Load() - hits; extra > 1 {
		t.Fatal("unknown key ids caused repeated key fetches", extra)
	}
	// Without a kid a token is accepted only when the provider has exactly one key.
	if _, err := p.Verify(t.Context(), f.sign("RS256", "", f.claims("n")), "n"); err == nil {
		t.Fatal("a token without a key id was accepted against several keys")
	}
}

func TestDiscoveryAndTransportSafety(t *testing.T) {
	t.Parallel()
	f := newIdP(t)
	f.issuer = "https://someone-else.example.com"
	if _, err := f.provider().AuthURL(t.Context(), "https://a.example.com/cb", "s", "n", "c"); err == nil {
		t.Fatal("a discovery document claiming another issuer was accepted")
	}
	for _, issuer := range []string{"http://idp.example.com", "https://127.0.0.1", "https://[::1]", "https://localhost", "https://app.localhost", "https://user:pw@idp.example.com", "https://idp.example.com#frag", "https://idp.example.com?x=1", "", "ftp://idp.example.com"} {
		if _, err := New(Config{Issuer: issuer, ClientID: clientID}); err == nil {
			t.Fatal("accepted issuer", issuer)
		}
	}
	if _, err := New(Config{Issuer: "https://idp.example.com"}); err == nil {
		t.Fatal("client id is required")
	}
	if _, err := New(Config{Issuer: "https://idp.example.com", ClientID: clientID}); err != nil {
		t.Fatal(err)
	}
	// Redirects from the provider are not followed.
	var trapped atomic.Int64
	trap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { trapped.Add(1) }))
	defer trap.Close()
	g := newIdP(t)
	g.tokenFn = func(w http.ResponseWriter, r *http.Request) bool {
		http.Redirect(w, r, trap.URL+"/latest/meta-data", http.StatusTemporaryRedirect)
		return true
	}
	if _, err := g.provider().Exchange(t.Context(), "c", "https://a.example.com/cb", "v", "n"); err == nil || trapped.Load() != 0 {
		t.Fatal("a redirect from the token endpoint was followed or accepted", trapped.Load())
	}
	for _, path := range []string{"/jwks", "/.well-known/openid-configuration"} {
		trapped.Store(0)
		h := newIdP(t)
		mux := http.NewServeMux()
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, trap.URL, http.StatusFound) })
		mux.Handle("/", h.server.Config.Handler)
		redirecting := httptest.NewServer(mux)
		h.server = redirecting
		p, _ := New(Config{Issuer: redirecting.URL, ClientID: clientID, HTTP: redirecting.Client(), AllowLoopbackForTests: true})
		p.AuthURL(t.Context(), "https://a.example.com/cb", "s", "n", "c")
		p.Verify(t.Context(), h.sign("RS256", "rsa-1", h.claims("n")), "n")
		redirecting.Close()
		if trapped.Load() != 0 {
			t.Fatal("a redirect was followed for", path)
		}
	}
	// Oversized and malformed token responses are refused.
	g.tokenFn = func(w http.ResponseWriter, r *http.Request) bool {
		w.Write([]byte(strings.Repeat("a", maxResponse+10)))
		return true
	}
	if _, err := g.provider().Exchange(t.Context(), "c", "https://a.example.com/cb", "v", "n"); err == nil {
		t.Fatal("oversized response accepted")
	}
	g.tokenFn = func(w http.ResponseWriter, r *http.Request) bool {
		w.Write([]byte(`{"access_token":"x"}`))
		return true
	}
	if _, err := g.provider().Exchange(t.Context(), "c", "https://a.example.com/cb", "v", "n"); err == nil {
		t.Fatal("a response without an id_token was accepted")
	}
	// An unreachable provider is generic.
	dead := newIdP(t)
	p := dead.provider()
	dead.server.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	if _, err := p.AuthURL(ctx, "https://a.example.com/cb", "s", "n", "c"); err != ErrUnavailable {
		t.Fatal("provider outage must be generic", err)
	}
}
