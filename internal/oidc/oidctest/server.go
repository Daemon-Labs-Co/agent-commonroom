// Package oidctest is a fake OpenID Connect provider for tests. Never import it from
// product code.
package oidctest

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"agentroom/internal/oidc"
)

const ClientID = "commonroom-test"

type Server struct {
	*httptest.Server
	Subject, Email string
	// ReusableCodes makes the provider accept an authorization code more than once, so
	// tests can show the relying party refuses a replay on its own.
	ReusableCodes bool
	key           *rsa.PrivateKey
	mu            sync.Mutex
	codes         map[string]grant
}

type grant struct{ challenge, redirect, nonce, subject, email string }

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func New(t *testing.T) *Server {
	t.Helper()
	k, _ := rsa.GenerateKey(rand.Reader, 2048)
	s := &Server{Subject: "subject-1", Email: "person@example.com", key: k, codes: map[string]grant{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"issuer": s.URL, "authorization_endpoint": s.URL + "/auth", "token_endpoint": s.URL + "/token", "jwks_uri": s.URL + "/jwks"})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{{"kty": "RSA", "kid": "k1", "use": "sig", "n": b64(k.N.Bytes()), "e": "AQAB"}}})
	})
	// The provider approves automatically for whoever Subject names, as a signed-in user would.
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		code := oidc.Random()
		s.mu.Lock()
		s.codes[code] = grant{challenge: q.Get("code_challenge"), redirect: q.Get("redirect_uri"), nonce: q.Get("nonce"), subject: s.Subject, email: s.Email}
		s.mu.Unlock()
		http.Redirect(w, r, q.Get("redirect_uri")+"?code="+code+"&state="+q.Get("state"), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		s.mu.Lock()
		g, ok := s.codes[r.Form.Get("code")]
		if !s.ReusableCodes {
			delete(s.codes, r.Form.Get("code"))
		}
		s.mu.Unlock()
		sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		if !ok || b64(sum[:]) != g.challenge || r.Form.Get("redirect_uri") != g.redirect || r.Form.Get("client_id") != ClientID {
			http.Error(w, `{"error":"invalid_grant"}`, 400)
			return
		}
		claims := map[string]any{"iss": s.URL, "sub": g.subject, "aud": ClientID, "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "nonce": g.nonce, "email": g.email, "email_verified": true}
		header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "k1"})
		payload, _ := json.Marshal(claims)
		input := b64(header) + "." + b64(payload)
		digest := sha256.Sum256([]byte(input))
		sig, _ := rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, digest[:])
		json.NewEncoder(w).Encode(map[string]string{"id_token": input + "." + b64(sig)})
	})
	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func (s *Server) Provider(t *testing.T) *oidc.Provider {
	t.Helper()
	p, err := oidc.New(oidc.Config{Issuer: s.URL, ClientID: ClientID, HTTP: s.Client(), AllowLoopbackForTests: true})
	if err != nil {
		t.Fatal(err)
	}
	return p
}
