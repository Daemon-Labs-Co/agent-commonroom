package main

import (
	"agentroom/internal/room"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func loginAvailability(t *testing.T) (*server, *ownerAuth) {
	t.Helper()
	a, err := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
	if err != nil {
		t.Fatal(err)
	}
	return &server{auth: a}, a
}
func badLoginFlood(s *server) {
	for i := 0; i < 6; i++ {
		s.login(httptest.NewRecorder(), httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"key":"wrong"}`)))
	}
}
func TestValidBootstrapLoginSurvivesFailedAttemptFlood(t *testing.T) {
	t.Parallel()
	s, a := loginAvailability(t)
	badLoginFlood(s)
	w := httptest.NewRecorder()
	s.login(w, httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"key":"`+a.key+`"}`)))
	if w.Code != 200 {
		t.Fatal("bad attempts locked valid owner out", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	s.login(w, httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"key":"wrong"}`)))
	if w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatal("valid login cleared failed-attempt protection")
	}
}
func TestValidSignaturesSurviveFailedAttemptFlood(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		t.Fatal(err)
	}
	for _, actor := range []string{ownerID, "invited-human"} {
		t.Run(actor, func(t *testing.T) {
			s, a := loginAvailability(t)
			s.db = newMockDB(t)
			a.key = ""
			a.publicKey = &key.PublicKey
			if actor != ownerID {
				raw, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
				if err != nil {
					t.Fatal(err)
				}
				s.db.(*mockDB).seed("actor.revised", room.Actor{ID: actor, Kind: "human", Enabled: true, PublicKey: string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: raw}))})
			}
			badLoginFlood(s)
			a.challenges["synthetic"] = ownerChallenge{Value: "synthetic proof", Expires: time.Now().Add(time.Minute)}
			digest := sha256.Sum256([]byte("synthetic proof"))
			signature, err := rsa.SignPSS(rand.Reader, key, crypto.SHA256, digest[:], &rsa.PSSOptions{SaltLength: 32, Hash: crypto.SHA256})
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(map[string]string{"actorId": actor, "challengeId": "synthetic", "signature": base64.StdEncoding.EncodeToString(signature)})
			w := httptest.NewRecorder()
			s.login(w, httptest.NewRequest("POST", "/api/login", strings.NewReader(string(raw))))
			if w.Code != 200 {
				t.Fatal("valid signature blocked", actor, w.Code, w.Body.String())
			}
			w = httptest.NewRecorder()
			s.login(w, httptest.NewRequest("POST", "/api/login", strings.NewReader(string(raw))))
			if w.Code == 200 {
				t.Fatal("replayed signature accepted")
			}
		})
	}
}
func TestSessionRenewalAtCapacityAndFailureRollback(t *testing.T) {
	for _, hashed := range []bool{false, true} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("hashed=%t,failure=%t", hashed, fail), func(t *testing.T) {
				t.Parallel()
				s, a := loginAvailability(t)
				token := "synthetic-cookie"
				oldkey := token
				if hashed {
					oldkey = sessionDigest(token)
				}
				original := ownerSession{ActorID: ownerID, CSRF: strings.Repeat("b", 64), Expires: time.Now().Add(time.Hour)}
				a.sessions[oldkey] = original
				for i := 0; i < 31; i++ {
					a.sessions[fmt.Sprintf("other-%d", i)] = original
				}
				if fail {
					a.sessionPath = filepath.Join(t.TempDir(), "sessions")
					if err := os.Mkdir(a.sessionPath, 0700); err != nil {
						t.Fatal(err)
					}
				}
				r := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"key":"`+a.key+`"}`))
				r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: token})
				w := httptest.NewRecorder()
				s.login(w, r)
				if fail {
					if w.Code != 503 || a.sessions[oldkey] != original || len(a.sessions) != 32 || len(w.Result().Cookies()) != 0 {
						t.Fatal("failed replacement lost old session", w.Code)
					}
				} else {
					if w.Code != 200 || len(a.sessions) != 32 {
						t.Fatal("renewal at capacity rejected", w.Code)
					}
					if _, ok := a.sessions[oldkey]; ok {
						t.Fatal("old session retained after rotation")
					}
				}
			})
		}
	}
}
func TestUnknownChallengeDoesNotReadMemberDatabase(t *testing.T) {
	t.Parallel()
	s, _ := loginAvailability(t) // nil DB intentionally proves the precheck.
	for _, body := range []string{`{"actorId":"unknown","challengeId":"absent","signature":"bad"}`, `{"actorId":"unknown","challengeId":"absent","signature":"bad"} {}`, `{"key":"x","unexpected":"x"}`} {
		w := httptest.NewRecorder()
		s.login(w, httptest.NewRequest("POST", "/api/login", strings.NewReader(body)))
		if w.Code != 401 && w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
}

func TestMalformedMemberProofConsumesChallengeBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()
	s, a := loginAvailability(t)
	a.challenges["synthetic"] = ownerChallenge{Value: "synthetic", Expires: time.Now().Add(time.Minute)}
	// A nil DB makes any unexpected database replay fail this regression.
	r := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"actorId":"unknown","challengeId":"synthetic","signature":"malformed"}`))
	w := httptest.NewRecorder()
	s.login(w, r)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
	if _, exists := a.challenges["synthetic"]; exists {
		t.Fatal("failed challenge remained reusable")
	}
}
