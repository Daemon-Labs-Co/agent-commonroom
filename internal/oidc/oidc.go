// Package oidc implements OpenID Connect sign-in (authorization code flow with PKCE and a
// nonce) using only the Go standard library (ADR 0006). ID tokens must be signed RS256 or
// ES256 and are checked against the configured issuer, audience, expiry and nonce.
package oidc

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrToken       = errors.New("identity token is invalid")
	ErrUnavailable = errors.New("identity provider unavailable")
)

const (
	clockSkew      = 60 * time.Second
	maxResponse    = 1 << 20
	maxTokenLength = 8192
	jwksTTL        = time.Hour
	jwksMinRefresh = time.Minute
)

type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string // empty for a public client (PKCE only)
	HTTP         *http.Client
	Now          func() time.Time
	// AllowLoopbackForTests lets tests use a local http or IP-literal provider. The server
	// never sets it and no setting can.
	AllowLoopbackForTests bool
}

type Claims struct {
	Issuer, Subject, Email, Name string
	EmailVerified                bool
}

type Provider struct {
	cfg Config

	mu        sync.Mutex
	discovery *discovery
	discAt    time.Time
	keys      map[string]any // kid -> *rsa.PublicKey | *ecdsa.PublicKey
	keysAt    time.Time
}

type discovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

// New validates the configuration. The issuer must be an https URL with a DNS host name.
func New(cfg Config) (*Provider, error) {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.HTTP == nil {
		cfg.HTTP = &http.Client{Timeout: 10 * time.Second}
	}
	// Redirects from an identity provider are never followed: a redirect could point the
	// server at an internal address.
	client := *cfg.HTTP
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	cfg.HTTP = &client
	p := &Provider{cfg: cfg}
	if cfg.ClientID == "" || len(cfg.ClientID) > 256 || !p.allowedURL(cfg.Issuer) || strings.Contains(cfg.Issuer, "?") || strings.Contains(cfg.Issuer, "#") {
		return nil, errors.New("OIDC issuer must be an https URL with a DNS host name, and a client id is required")
	}
	return p, nil
}

func (p *Provider) allowedURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Host == "" || u.Fragment != "" {
		return false
	}
	host := u.Hostname()
	if p.cfg.AllowLoopbackForTests {
		return u.Scheme == "https" || u.Scheme == "http"
	}
	return u.Scheme == "https" && host != "" && net.ParseIP(host) == nil && host != "localhost" && !strings.HasSuffix(host, ".localhost")
}

func (p *Provider) get(ctx context.Context, target string) ([]byte, error) {
	if !p.allowedURL(target) {
		return nil, ErrUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return nil, ErrUnavailable
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.cfg.HTTP.Do(req)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil || resp.StatusCode != 200 || len(body) > maxResponse {
		return nil, ErrUnavailable
	}
	return body, nil
}

func (p *Provider) discover(ctx context.Context) (*discovery, error) {
	p.mu.Lock()
	if p.discovery != nil && p.cfg.Now().Sub(p.discAt) < jwksTTL {
		d := p.discovery
		p.mu.Unlock()
		return d, nil
	}
	p.mu.Unlock()
	body, err := p.get(ctx, strings.TrimRight(p.cfg.Issuer, "/")+"/.well-known/openid-configuration")
	if err != nil {
		return nil, err
	}
	var d discovery
	if json.Unmarshal(body, &d) != nil || d.Issuer != p.cfg.Issuer || !p.allowedURL(d.AuthorizationEndpoint) || !p.allowedURL(d.TokenEndpoint) || !p.allowedURL(d.JWKSURI) {
		return nil, ErrUnavailable
	}
	p.mu.Lock()
	p.discovery, p.discAt = &d, p.cfg.Now()
	p.mu.Unlock()
	return &d, nil
}

// PKCE returns a code verifier and its S256 challenge.
func PKCE() (verifier, challenge string) {
	raw := make([]byte, 32)
	rand.Read(raw)
	verifier = base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	return verifier, base64.RawURLEncoding.EncodeToString(sum[:])
}

// Random returns an unguessable URL-safe value for state and nonce.
func Random() string {
	raw := make([]byte, 32)
	rand.Read(raw)
	return base64.RawURLEncoding.EncodeToString(raw)
}

// AuthURL is the address to send the browser to.
func (p *Provider) AuthURL(ctx context.Context, redirectURI, state, nonce, challenge string) (string, error) {
	d, err := p.discover(ctx)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(d.AuthorizationEndpoint)
	if err != nil {
		return "", ErrUnavailable
	}
	q := u.Query()
	for k, v := range map[string]string{"response_type": "code", "client_id": p.cfg.ClientID, "redirect_uri": redirectURI, "scope": "openid email", "state": state, "nonce": nonce, "code_challenge": challenge, "code_challenge_method": "S256"} {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Exchange trades an authorization code for an ID token and verifies it.
func (p *Provider) Exchange(ctx context.Context, code, redirectURI, verifier, nonce string) (Claims, error) {
	d, err := p.discover(ctx)
	if err != nil {
		return Claims{}, err
	}
	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirectURI}, "client_id": {p.cfg.ClientID}, "code_verifier": {verifier}}
	if p.cfg.ClientSecret != "" {
		form.Set("client_secret", p.cfg.ClientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", d.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Claims{}, ErrUnavailable
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := p.cfg.HTTP.Do(req)
	if err != nil {
		return Claims{}, ErrUnavailable
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil || resp.StatusCode != 200 || len(body) > maxResponse {
		return Claims{}, ErrToken
	}
	var token struct {
		IDToken string `json:"id_token"`
	}
	if json.Unmarshal(body, &token) != nil || token.IDToken == "" {
		return Claims{}, ErrToken
	}
	return p.Verify(ctx, token.IDToken, nonce)
}

func b64(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }

func (p *Provider) loadKeys(ctx context.Context, d *discovery, force bool) error {
	p.mu.Lock()
	fresh := p.keys != nil && p.cfg.Now().Sub(p.keysAt) < jwksTTL
	recent := p.keys != nil && p.cfg.Now().Sub(p.keysAt) < jwksMinRefresh
	p.mu.Unlock()
	if (fresh && !force) || (force && recent) {
		return nil
	}
	body, err := p.get(ctx, d.JWKSURI)
	if err != nil {
		return err
	}
	var set struct {
		Keys []struct {
			Kty, Kid, Use, Crv, N, E, X, Y string
		} `json:"keys"`
	}
	if json.Unmarshal(body, &set) != nil || len(set.Keys) > 64 {
		return ErrUnavailable
	}
	keys := map[string]any{}
	for _, k := range set.Keys {
		if k.Use != "" && k.Use != "sig" {
			continue
		}
		switch k.Kty {
		case "RSA":
			n, errN := b64(k.N)
			e, errE := b64(k.E)
			if errN != nil || errE != nil || len(e) == 0 || len(e) > 4 {
				continue
			}
			modulus := new(big.Int).SetBytes(n)
			if modulus.BitLen() < 2048 || modulus.BitLen() > 8192 {
				continue
			}
			keys[k.Kid] = &rsa.PublicKey{N: modulus, E: int(new(big.Int).SetBytes(e).Int64())}
		case "EC":
			x, errX := b64(k.X)
			y, errY := b64(k.Y)
			if k.Crv != "P-256" || errX != nil || errY != nil || len(x) != 32 || len(y) != 32 {
				continue
			}
			pub, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), append([]byte{4}, append(x, y...)...))
			if err != nil {
				continue
			}
			keys[k.Kid] = pub
		}
	}
	p.mu.Lock()
	p.keys, p.keysAt = keys, p.cfg.Now()
	p.mu.Unlock()
	return nil
}

func (p *Provider) keyFor(ctx context.Context, d *discovery, kid, alg string) (any, error) {
	for attempt := 0; attempt < 2; attempt++ {
		if err := p.loadKeys(ctx, d, attempt == 1); err != nil {
			return nil, err
		}
		p.mu.Lock()
		var found any
		if kid != "" {
			found = p.keys[kid]
		} else if len(p.keys) == 1 {
			for _, k := range p.keys {
				found = k
			}
		}
		p.mu.Unlock()
		switch k := found.(type) {
		case *rsa.PublicKey:
			if alg == "RS256" {
				return k, nil
			}
			return nil, ErrToken
		case *ecdsa.PublicKey:
			if alg == "ES256" {
				return k, nil
			}
			return nil, ErrToken
		}
	}
	return nil, ErrToken // unknown signing key, even after one refresh
}

// Verify checks an ID token and returns its claims.
func (p *Provider) Verify(ctx context.Context, raw, nonce string) (Claims, error) {
	if len(raw) > maxTokenLength || nonce == "" {
		return Claims{}, ErrToken
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return Claims{}, ErrToken
	}
	headerRaw, err1 := b64(parts[0])
	payloadRaw, err2 := b64(parts[1])
	signature, err3 := b64(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return Claims{}, ErrToken
	}
	var header struct {
		Alg  string `json:"alg"`
		Kid  string `json:"kid"`
		Crit any    `json:"crit"`
	}
	if json.Unmarshal(headerRaw, &header) != nil || header.Crit != nil || (header.Alg != "RS256" && header.Alg != "ES256") {
		return Claims{}, ErrToken // never none, HS*, PS* or unknown critical extensions
	}
	d, err := p.discover(ctx)
	if err != nil {
		return Claims{}, err
	}
	key, err := p.keyFor(ctx, d, header.Kid, header.Alg)
	if err != nil {
		return Claims{}, err
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	switch k := key.(type) {
	case *rsa.PublicKey:
		if rsa.VerifyPKCS1v15(k, crypto.SHA256, digest[:], signature) != nil {
			return Claims{}, ErrToken
		}
	case *ecdsa.PublicKey:
		if len(signature) != 64 || !ecdsa.Verify(k, digest[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:])) {
			return Claims{}, ErrToken
		}
	}
	var c struct {
		Iss           string          `json:"iss"`
		Sub           string          `json:"sub"`
		Aud           json.RawMessage `json:"aud"`
		Azp           string          `json:"azp"`
		Exp           *float64        `json:"exp"`
		Iat           *float64        `json:"iat"`
		Nbf           *float64        `json:"nbf"`
		Nonce         string          `json:"nonce"`
		Email         string          `json:"email"`
		EmailVerified bool            `json:"email_verified"`
		Name          string          `json:"name"`
	}
	if json.Unmarshal(payloadRaw, &c) != nil || c.Exp == nil {
		return Claims{}, ErrToken
	}
	now := p.cfg.Now()
	if c.Iss != p.cfg.Issuer || c.Sub == "" || len(c.Sub) > 255 || subtle.ConstantTimeCompare([]byte(c.Nonce), []byte(nonce)) != 1 {
		return Claims{}, ErrToken
	}
	if now.After(time.Unix(int64(*c.Exp), 0).Add(clockSkew)) || (c.Iat != nil && time.Unix(int64(*c.Iat), 0).After(now.Add(clockSkew))) || (c.Nbf != nil && time.Unix(int64(*c.Nbf), 0).After(now.Add(clockSkew))) {
		return Claims{}, ErrToken
	}
	var audiences []string
	if json.Unmarshal(c.Aud, &audiences) != nil {
		var one string
		if json.Unmarshal(c.Aud, &one) != nil {
			return Claims{}, ErrToken
		}
		audiences = []string{one}
	}
	matched := false
	for _, a := range audiences {
		if a == p.cfg.ClientID {
			matched = true
		}
	}
	if !matched || (len(audiences) > 1 && c.Azp != p.cfg.ClientID) || (c.Azp != "" && c.Azp != p.cfg.ClientID) {
		return Claims{}, ErrToken
	}
	return Claims{Issuer: c.Iss, Subject: c.Sub, Email: c.Email, EmailVerified: c.EmailVerified, Name: c.Name}, nil
}
