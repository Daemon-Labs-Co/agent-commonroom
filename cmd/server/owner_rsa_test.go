package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRSAOwnerLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setup")
	code, e := loadOwnerKey(path)
	if e != nil {
		t.Fatal(e)
	}
	a, _ := newOwnerAuth(code, "127.0.0.1:8787")
	if e = a.prepareRSA(path); e != nil {
		t.Fatal(e)
	}
	if !a.setupPending || a.publicKey.N.BitLen() != 3072 {
		t.Fatal("RSA setup missing")
	}
	raw, e := readPrivateFile(a.privatePath)
	if e != nil {
		t.Fatal(e)
	}
	block, _ := pem.Decode(raw)
	parsed, e := x509.ParsePKCS8PrivateKey(block.Bytes)
	if e != nil {
		t.Fatal(e)
	}
	key := parsed.(*rsa.PrivateKey)
	s := &server{auth: a}
	a.sessions["setup-session"] = ownerSession{CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	r := httptest.NewRequest("POST", "/api/owner-key/download", nil)
	r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: "setup-session"})
	w := httptest.NewRecorder()
	s.downloadOwnerKey(w, r)
	if w.Code != 403 {
		t.Fatal("download without CSRF")
	}
	r.Header.Set("X-CSRF-Token", "csrf")
	w = httptest.NewRecorder()
	s.downloadOwnerKey(w, r)
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || w.Body.String() != string(raw) {
		t.Fatal("download failed")
	}
	a.challenges["login"] = ownerChallenge{Value: "challenge", Expires: time.Now().Add(time.Minute)}
	digest := sha256.Sum256([]byte("challenge"))
	sig, e := rsa.SignPSS(rand.Reader, key, crypto.SHA256, digest[:], &rsa.PSSOptions{SaltLength: 32, Hash: crypto.SHA256})
	if e != nil {
		t.Fatal(e)
	}
	body, _ := json.Marshal(map[string]string{"challengeId": "login", "signature": base64.StdEncoding.EncodeToString(sig)})
	r = httptest.NewRequest("POST", "/api/login", strings.NewReader(string(body)))
	w = httptest.NewRecorder()
	s.login(w, r)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if a.setupPending || a.key != "" {
		t.Fatal("bootstrap not revoked")
	}
	for _, path := range []string{a.privatePath, a.setupPath} {
		if _, e = os.Stat(path); !os.IsNotExist(e) {
			t.Fatal("private setup material retained")
		}
	}
	if a.verifySignature("login", base64.StdEncoding.EncodeToString(sig)) {
		t.Fatal("challenge replay")
	}
	b, _ := newOwnerAuth("", "127.0.0.1:8787")
	if e = b.prepareRSA(path); e != nil || b.setupPending {
		t.Fatal("public-only restart failed", e)
	}
	w = httptest.NewRecorder()
	s.downloadOwnerKey(w, r)
	if w.Code == 200 {
		t.Fatal("key downloadable after setup")
	}
	a.challenges["expired"] = ownerChallenge{Value: "challenge", Expires: time.Now().Add(-time.Minute)}
	if a.verifySignature("expired", base64.StdEncoding.EncodeToString(sig)) {
		t.Fatal("expired challenge accepted")
	}
	a.challenges["wrong"] = ownerChallenge{Value: "different", Expires: time.Now().Add(time.Minute)}
	if a.verifySignature("wrong", base64.StdEncoding.EncodeToString(sig)) {
		t.Fatal("wrong signature accepted")
	}
}
