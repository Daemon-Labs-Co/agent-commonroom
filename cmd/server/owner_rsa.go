package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ownerChallenge struct {
	Value   string
	Expires time.Time
}

func writePrivateFile(path string, data []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(data)
	closeErr := f.Close()
	if e != nil {
		return e
	}
	return closeErr
}
func readPrivateFile(path string) ([]byte, error) {
	return readPrivateBounded(path, 65536)
}
func (a *ownerAuth) prepareRSA(path string) error {
	a.setupPath = path
	a.privatePath = path + ".private.pem"
	publicPath := path + ".public.pem"
	raw, e := readPrivateFile(publicPath)
	if os.IsNotExist(e) {
		// Resume a pending generation if public-key persistence previously failed.
		privateRaw, readErr := readPrivateFile(a.privatePath)
		var key *rsa.PrivateKey
		if os.IsNotExist(readErr) {
			key, e = rsa.GenerateKey(rand.Reader, 3072)
			if e != nil {
				return e
			}
			der, e := x509.MarshalPKCS8PrivateKey(key)
			if e != nil {
				return e
			}
			privateRaw = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
			if e = writePrivateFile(a.privatePath, privateRaw); e != nil {
				return e
			}
		} else if readErr != nil {
			return readErr
		} else {
			block, _ := pem.Decode(privateRaw)
			if block == nil {
				return fmt.Errorf("invalid pending private key")
			}
			parsed, e := x509.ParsePKCS8PrivateKey(block.Bytes)
			if e != nil {
				return e
			}
			var ok bool
			key, ok = parsed.(*rsa.PrivateKey)
			if !ok {
				return fmt.Errorf("pending key is not RSA")
			}
		}
		der, e := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if e != nil {
			return e
		}
		raw = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
		if e = writePrivateFile(publicPath, raw); e != nil {
			return e
		}
	} else if e != nil {
		return e
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return fmt.Errorf("invalid owner public key")
	}
	parsed, e := x509.ParsePKIXPublicKey(block.Bytes)
	if e != nil {
		return e
	}
	var ok bool
	a.publicKey, ok = parsed.(*rsa.PublicKey)
	if !ok || a.publicKey.N.BitLen() < 3072 {
		return fmt.Errorf("owner RSA key must be at least 3072 bits")
	}
	_, e = readPrivateFile(a.privatePath)
	a.setupPending = e == nil
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	if !a.setupPending {
		a.key = ""
		if e = os.Remove(path); e != nil && !os.IsNotExist(e) {
			return e
		}
	}
	return nil
}
func (s *server) challenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	a := s.auth
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for id, v := range a.challenges {
		if !now.Before(v.Expires) {
			delete(a.challenges, id)
		}
	}
	if len(a.challenges) >= 128 {
		http.Error(w, "Too many pending sign-in requests; try again shortly", 429)
		return
	}
	id, value := id(), "CommonRoom owner login: "+id()+id()
	a.challenges[id] = ownerChallenge{Value: value, Expires: now.Add(2 * time.Minute)}
	reply(w, map[string]string{"challengeId": id, "challenge": value})
}
func (a *ownerAuth) verifySignature(challengeID, signature string) bool {
	v, ok := a.challenges[challengeID]
	delete(a.challenges, challengeID)
	if !ok || !time.Now().Before(v.Expires) || a.publicKey == nil {
		return false
	}
	sig, e := base64.StdEncoding.DecodeString(signature)
	if e != nil {
		return false
	}
	digest := sha256.Sum256([]byte(v.Value))
	return rsa.VerifyPSS(a.publicKey, crypto.SHA256, digest[:], sig, &rsa.PSSOptions{SaltLength: 32, Hash: crypto.SHA256}) == nil
}
func (s *server) downloadOwnerKey(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	a := s.auth
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.setupPending {
		http.Error(w, "Owner key setup is already complete", 410)
		return
	}
	raw, e := readPrivateFile(a.privatePath)
	if e != nil {
		http.Error(w, "Owner key unavailable", 500)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="commonroom-owner-private.pem"`)
	w.Write(raw)
}
