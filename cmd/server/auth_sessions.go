package main

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Session persistence saves hashes of opaque cookies, never PEM keys or cookie
// tokens. The private server file is bound to the owner public key; normal
// expiry, CSRF, origin and actor authorization checks still apply after restart.
type savedHumanSessions struct {
	Version  int                     `json:"version"`
	Identity string                  `json:"identity"`
	Sessions map[string]ownerSession `json:"sessions"`
}

func sessionDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "sha256:" + hex.EncodeToString(sum[:])
}
func (a *ownerAuth) sessionIdentity() string {
	raw := []byte(a.key)
	if a.publicKey != nil {
		raw = x509.MarshalPKCS1PublicKey(a.publicKey)
	}
	return sessionDigest(string(raw))
}

// configureSessionPersistence is called once after RSA setup. Unsafe files or
// corrupt content fail startup closed. Disabled persistence invalidates the
// previous store so switching it back on cannot resurrect signed-out sessions.
func (a *ownerAuth) configureSessionPersistence(path string, enabled bool) error {
	if !enabled {
		a.sessionPath = ""
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	a.sessionPath = path
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > 65536 {
		return errors.New("Session store must be a private, bounded regular file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var saved savedHumanSessions
	if json.Unmarshal(raw, &saved) != nil || saved.Version != 1 || len(saved.Sessions) > 32 {
		return errors.New("Invalid private session store")
	}
	if saved.Identity != a.sessionIdentity() {
		return a.persistSessionsLocked()
	}
	now := time.Now()
	for hash, session := range saved.Sessions {
		if len(hash) != 71 || !strings.HasPrefix(hash, "sha256:") {
			return errors.New("Invalid session token digest")
		}
		if _, err := hex.DecodeString(hash[7:]); err != nil {
			return errors.New("Invalid session token digest")
		}
		if session.ActorID == "" || len(session.CSRF) != 64 || session.Expires.After(now.Add(8*time.Hour)) {
			return errors.New("Invalid saved session")
		}
		if now.Before(session.Expires) && !a.setupPending {
			a.sessions[hash] = session
		}
	}
	return nil
}

// persistSessionsLocked requires a.mu; atomic replacement makes revocation
// durable before a successful logout response. Storage failure fails the action.
func (a *ownerAuth) persistSessionsLocked() error {
	if a.sessionPath == "" {
		return nil
	}
	saved := savedHumanSessions{Version: 1, Identity: a.sessionIdentity(), Sessions: map[string]ownerSession{}}
	if !a.setupPending {
		for token, session := range a.sessions {
			if time.Now().Before(session.Expires) {
				hash := token
				if !strings.HasPrefix(token, "sha256:") {
					hash = sessionDigest(token)
				}
				saved.Sessions[hash] = session
			}
		}
	}
	raw, err := json.Marshal(saved)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(a.sessionPath), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(a.sessionPath), ".sessions-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), a.sessionPath)
}
func persistentHumanSessionsEnabled() (bool, error) {
	value := os.Getenv("COMMONROOM_PERSISTENT_SESSIONS")
	if value == "" {
		return true, nil
	}
	return strconv.ParseBool(value)
}
