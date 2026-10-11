package main

import (
	"crypto/rsa"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ownerSession struct {
	ActorID string
	CSRF    string
	Expires time.Time
}
type ownerAuth struct {
	mu                     sync.Mutex
	publicKey              *rsa.PublicKey
	setupPath, privatePath string
	setupPending           bool
	challenges             map[string]ownerChallenge
	key                    string
	sessions               map[string]ownerSession
	attempts               int
	window                 time.Time
	port                   string
}

// The local bootstrap code is never served through HTTP or included in audit payloads.
func loadOwnerKey(path string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err == nil {
		key := id() + id()
		_, e := f.WriteString(key + "\n")
		closeErr := f.Close()
		if e != nil {
			return "", e
		}
		if closeErr != nil {
			return "", closeErr
		}
		return key, nil
	}
	if !os.IsExist(err) {
		return "", err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return "", fmt.Errorf("owner key must be a regular file readable only by its owner")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	key := strings.TrimSpace(string(raw))
	if len(key) != 64 {
		return "", fmt.Errorf("invalid owner key file")
	}
	return key, nil
}
func newOwnerAuth(key, addr string) (*ownerAuth, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return nil, fmt.Errorf("only numeric loopback LISTEN_ADDR is supported until secure remote authentication is implemented")
	}
	return &ownerAuth{key: key, port: port, sessions: map[string]ownerSession{}, challenges: map[string]ownerChallenge{}}, nil
}
func (a *ownerAuth) trustedRequest(r *http.Request) bool {
	allowed := map[string]bool{net.JoinHostPort("127.0.0.1", a.port): true, net.JoinHostPort("localhost", a.port): true, net.JoinHostPort("::1", a.port): true}
	if !allowed[r.Host] {
		return false
	}
	if site := r.Header.Get("Sec-Fetch-Site"); site == "cross-site" {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
		return false
	}
	return true
}
func (a *ownerAuth) session(r *http.Request) (ownerSession, bool) {
	cookie, e := r.Cookie("commonroom_session")
	if e != nil {
		return ownerSession{}, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	session, ok := a.sessions[cookie.Value]
	if !ok || !time.Now().Before(session.Expires) {
		delete(a.sessions, cookie.Value)
		return ownerSession{}, false
	}
	return session, true
}
func (s *server) validCSRF(r *http.Request) bool {
	session, ok := s.auth.session(r)
	return ok && subtle.ConstantTimeCompare([]byte(session.CSRF), []byte(r.Header.Get("X-CSRF-Token"))) == 1
}
func (a *ownerAuth) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.trustedRequest(r) {
			http.Error(w, "Untrusted request origin or host", 403)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api/login" && r.URL.Path != "/api/challenge" {
			if _, ok := a.session(r); !ok {
				http.Error(w, "Sign in required", 401)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (s *server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var request struct {
		Key         string `json:"key"`
		ActorID     string `json:"actorId"`
		ChallengeID string `json:"challengeId"`
		Signature   string `json:"signature"`
	}
	if json.NewDecoder(r.Body).Decode(&request) != nil {
		http.Error(w, "Invalid sign-in request", 400)
		return
	}
	a := s.auth
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	if now.Sub(a.window) >= time.Minute {
		a.window = now
		a.attempts = 0
	}
	if a.attempts >= 5 {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "Too many sign-in attempts; try again in a minute", 429)
		return
	}
	a.attempts++
	actorID := request.ActorID
	if actorID == "" {
		actorID = ownerID
	}
	signed := false
	if request.ChallengeID != "" {
		if actorID == ownerID {
			signed = a.verifySignature(request.ChallengeID, request.Signature)
		} else {
			signed = s.verifyMemberLogin(r.Context(), actorID, request.ChallengeID, request.Signature)
		}
	}
	bootstrap := actorID == ownerID && a.key != "" && subtle.ConstantTimeCompare([]byte(request.Key), []byte(a.key)) == 1
	if !signed && !bootstrap {
		http.Error(w, "Invalid owner key", 401)
		return
	}
	if signed && actorID == ownerID && a.setupPending {
		if e := os.Remove(a.privatePath); e != nil {
			http.Error(w, "Unable to complete key setup", 500)
			return
		}
		a.setupPending = false
		a.key = ""
		a.sessions = map[string]ownerSession{}
		if e := os.Remove(a.setupPath); e != nil && !os.IsNotExist(e) {
			http.Error(w, "Unable to remove local setup code", 500)
			return
		}
	}
	// Keep the memory/session count bounded and expire old sessions.
	for token, session := range a.sessions {
		if !now.Before(session.Expires) {
			delete(a.sessions, token)
		}
	}
	if len(a.sessions) >= 32 {
		http.Error(w, "Too many active sessions; sign out of another session or restart the server", 429)
		return
	}
	if old, e := r.Cookie("commonroom_session"); e == nil {
		delete(a.sessions, old.Value)
	}
	token := id() + id()
	session := ownerSession{ActorID: actorID, CSRF: id() + id(), Expires: now.Add(8 * time.Hour)}
	a.sessions[token] = session
	http.SetCookie(w, &http.Cookie{Name: "commonroom_session", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60})
	reply(w, map[string]any{"csrf": session.CSRF, "actorId": actorID, "actorType": "human", "keySetupPending": a.setupPending && actorID == ownerID})
}
func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	if cookie, e := r.Cookie("commonroom_session"); e == nil {
		s.auth.mu.Lock()
		delete(s.auth.sessions, cookie.Value)
		s.auth.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "commonroom_session", Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
	reply(w, map[string]bool{"signedOut": true})
}
