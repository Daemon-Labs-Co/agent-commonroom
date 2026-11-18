package main

// Passkey, single sign-on and invitation sign-in for a workspace (ADR 0006). Every method
// ends in the same session as the key-file sign-in; none can reach another workspace
// because credentials, invitations and challenges live in this workspace's own server
// and history. Unauthenticated endpoints share the sign-in failure budget and answer
// every refusal the same way.

import (
	"agentroom/internal/oidc"
	"agentroom/internal/room"
	"agentroom/internal/webauthn"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	identityChallengeTTL = 5 * time.Minute
	oidcPendingTTL       = 10 * time.Minute
	maxPendingIdentity   = 512
	maxInviteHours       = 168
	defaultInviteHours   = 72
	flagPasskeys         = "passkeySignIn"
	flagOIDC             = "oidcSignIn"
	flagInvitations      = "memberInvitations"
)

type identityChallenge struct {
	Purpose   string // login, register, invite
	Challenge []byte
	ActorID   string // register: the signed-in actor; invite: the new actor's id
	Invite    string // invite: invitation id
	Name      string // invite: the name the person chose
	Expires   time.Time
}

type oidcPending struct {
	Purpose  string // login, link, invite
	Nonce    string
	Verifier string
	Binding  string // hash of the browser-binding cookie
	ActorID  string // link: the signed-in actor
	Token    string // invite: held only here, in memory
	Name     string
	Expires  time.Time
}

// identityStore holds short-lived challenges for one workspace.
type identityStore struct {
	mu         sync.Mutex
	challenges map[string]identityChallenge
	pending    map[string]oidcPending
}

func (st *identityStore) putChallenge(c identityChallenge, now time.Time) (string, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.challenges == nil {
		st.challenges = map[string]identityChallenge{}
	}
	for id, old := range st.challenges {
		if !now.Before(old.Expires) {
			delete(st.challenges, id)
		}
	}
	if len(st.challenges) >= maxPendingIdentity {
		// Never refuse: anonymous requests create challenges, so refusing would let
		// a flood block every sign-in. Drop the oldest challenge of the purpose that
		// holds the most, which is the flood and not the few people mid sign-in.
		counts := map[string]int{}
		for _, old := range st.challenges {
			counts[old.Purpose]++
		}
		busiest := ""
		for purpose, n := range counts {
			if busiest == "" || n > counts[busiest] {
				busiest = purpose
			}
		}
		oldest := ""
		for id, old := range st.challenges {
			if old.Purpose == busiest && (oldest == "" || old.Expires.Before(st.challenges[oldest].Expires)) {
				oldest = id
			}
		}
		delete(st.challenges, oldest)
	}
	id := oidc.Random()
	c.Expires = now.Add(identityChallengeTTL)
	st.challenges[id] = c
	return id, true
}

// takeChallenge always consumes the challenge: it is single use whatever happens next.
func (st *identityStore) takeChallenge(id, purpose string, now time.Time) (identityChallenge, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	c, ok := st.challenges[id]
	delete(st.challenges, id)
	if !ok || c.Purpose != purpose || !now.Before(c.Expires) {
		return identityChallenge{}, false
	}
	return c, true
}

func (st *identityStore) putPending(p oidcPending, now time.Time) (string, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.pending == nil {
		st.pending = map[string]oidcPending{}
	}
	for id, old := range st.pending {
		if !now.Before(old.Expires) {
			delete(st.pending, id)
		}
	}
	if len(st.pending) >= maxPendingIdentity {
		// Same reasoning as putChallenge: replace the oldest instead of refusing.
		oldest := ""
		for id, old := range st.pending {
			if oldest == "" || old.Expires.Before(st.pending[oldest].Expires) {
				oldest = id
			}
		}
		delete(st.pending, oldest)
	}
	id := oidc.Random()
	p.Expires = now.Add(oidcPendingTTL)
	st.pending[id] = p
	return id, true
}

func (st *identityStore) takePending(id string, now time.Time) (oidcPending, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	p, ok := st.pending[id]
	delete(st.pending, id)
	if !ok || !now.Before(p.Expires) {
		return oidcPending{}, false
	}
	return p, true
}

// identitySettings are installation-wide; the flags below are per workspace.
type identitySettings struct {
	provider *oidc.Provider
	name     string
	// hostedDefault turns the identity flags on for workspaces that never set them.
	hostedDefault bool
}

func (s *server) identityFlag(st room.State, name string) bool {
	if enabled, set := st.FeatureFlags[name]; set {
		return enabled
	}
	return s.identity.hostedDefault
}

func (s *server) oidcAvailable(st room.State) bool {
	return s.identity.provider != nil && s.identityFlag(st, flagOIDC)
}

// expectedOrigin is the exact origin and relying-party id the browser must report.
func (a *ownerAuth) expectedOrigin(r *http.Request) (origin, rp string) {
	if a.publicOrigin != "" {
		if u, err := url.Parse(a.publicOrigin); err == nil && r.Host == u.Host {
			return a.publicOrigin, u.Hostname()
		}
	}
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return "http://" + r.Host, host
}

// throttled reports whether sign-in failures have used up this minute's budget. It
// shares the counter with the key-file sign-in.
func (a *ownerAuth) throttled(now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if now.Sub(a.window) >= time.Minute {
		a.window, a.attempts = now, 0
	}
	return a.attempts >= 5
}

func (a *ownerAuth) countFailure() {
	a.mu.Lock()
	a.attempts++
	a.mu.Unlock()
}

func readIdentityBody(w http.ResponseWriter, r *http.Request, into any, limit int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if dec.Decode(into) != nil || dec.Decode(new(any)) != io.EOF {
		http.Error(w, "Invalid request", 400)
		return false
	}
	return true
}

func (s *server) requirePost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return false
	}
	return true
}

// authMethods tells the sign-in page what this workspace offers. It reveals only
// configuration, never accounts.
func (s *server) authMethods(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "Unavailable", 503)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	reply(w, map[string]any{"passkeys": s.identityFlag(st, flagPasskeys), "oidc": s.oidcAvailable(st), "oidcConfigured": s.identity.provider != nil, "oidcName": s.identity.name, "invitations": s.identityFlag(st, flagInvitations)})
}

func attestationDeny(w http.ResponseWriter) { http.Error(w, "Passkey could not be registered", 400) }

// ---- passkey sign-in ----

func (s *server) passkeyLoginOptions(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	st, err := s.load(r.Context())
	if err != nil || !s.identityFlag(st, flagPasskeys) {
		http.Error(w, "Passkey sign-in is not available", 403)
		return
	}
	now := time.Now()
	if s.auth.throttled(now) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "Too many failed sign-in attempts; try again in a minute", 429)
		return
	}
	challenge := randomChallenge()
	_, rp := s.auth.expectedOrigin(r)
	id, ok := s.identityStore.putChallenge(identityChallenge{Purpose: "login", Challenge: challenge}, now)
	if !ok {
		http.Error(w, "Too many sign-ins in progress", 429)
		return
	}
	reply(w, map[string]any{"challengeId": id, "publicKey": map[string]any{"challenge": b64url(challenge), "rpId": rp, "timeout": 60000, "userVerification": "required"}})
}

func (s *server) passkeyLoginVerify(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	var req struct {
		ChallengeID       string `json:"challengeId"`
		ID                string `json:"id"`
		AuthenticatorData string `json:"authenticatorData"`
		ClientDataJSON    string `json:"clientDataJSON"`
		Signature         string `json:"signature"`
		UserHandle        string `json:"userHandle"`
	}
	if !readIdentityBody(w, r, &req, 16384) {
		return
	}
	now := time.Now()
	if s.auth.throttled(now) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "Too many failed sign-in attempts; try again in a minute", 429)
		return
	}
	fail := func() {
		s.auth.countFailure()
		http.Error(w, "Passkey sign-in failed", 401)
	}
	challenge, ok := s.identityStore.takeChallenge(req.ChallengeID, "login", now)
	authData, ok1 := decodeB64(req.AuthenticatorData, 2048)
	clientData, ok2 := decodeB64(req.ClientDataJSON, 4096)
	signature, ok3 := decodeB64(req.Signature, 1024)
	if !ok || !ok1 || !ok2 || !ok3 || len(req.ID) > 1400 {
		fail()
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil || !s.identityFlag(st, flagPasskeys) {
		fail()
		return
	}
	actor, cred, found := findCredential(st, req.ID)
	if !found || !actor.Enabled || actor.Kind != "human" {
		fail()
		return
	}
	if req.UserHandle != "" {
		if handle, ok := decodeB64(req.UserHandle, 128); !ok || string(handle) != actor.ID {
			fail()
			return
		}
	}
	der, ok := decodeStdB64(cred.PublicKey)
	if !ok {
		fail()
		return
	}
	origin, rp := s.auth.expectedOrigin(r)
	newCount, err := webauthn.VerifyAssertion(webauthn.AssertionParams{AuthenticatorData: authData, ClientDataJSON: clientData, Signature: signature, Challenge: challenge.Challenge, Origin: origin, RPID: rp, PublicKeyDER: der, Alg: cred.Alg, StoredSignCount: cred.SignCount})
	if err != nil {
		fail()
		return
	}
	if err := s.append(context.WithValue(r.Context(), actorContextKey{}, actor.ID), "actor.credential.used", credentialUsed{ActorID: actor.ID, CredentialID: cred.ID, SignCount: newCount}); err != nil {
		http.Error(w, "Sign-in could not be recorded", 503)
		return
	}
	s.auth.mu.Lock()
	defer s.auth.mu.Unlock()
	session, ok := s.auth.startSessionLocked(w, r, actor.ID, now)
	if !ok {
		return
	}
	reply(w, map[string]any{"csrf": session.CSRF, "actorId": actor.ID, "actorType": "human", "keySetupPending": false})
}

// ---- passkey registration for the signed-in person ----

func (s *server) passkeyRegisterOptions(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) || !s.validCSRF(r) {
		if r.Method == "POST" {
			http.Error(w, "Invalid request token", 403)
		}
		return
	}
	s.mu.Lock()
	st, err := s.load(r.Context())
	s.mu.Unlock()
	actor, ok := member(st, requestActor(r))
	if err != nil || !ok || actor.Kind != "human" || !s.identityFlag(st, flagPasskeys) {
		http.Error(w, "Passkeys are not available", 403)
		return
	}
	if len(actor.Credentials) >= maxCredentialsPerActor {
		http.Error(w, "Remove a passkey before adding another", 409)
		return
	}
	challenge := randomChallenge()
	_, rp := s.auth.expectedOrigin(r)
	id, ok := s.identityStore.putChallenge(identityChallenge{Purpose: "register", Challenge: challenge, ActorID: actor.ID}, time.Now())
	if !ok {
		http.Error(w, "Too many sign-ins in progress", 429)
		return
	}
	exclude := []map[string]string{}
	for _, c := range actor.Credentials {
		exclude = append(exclude, map[string]string{"type": "public-key", "id": c.ID})
	}
	reply(w, map[string]any{"challengeId": id, "publicKey": creationOptions(challenge, rp, actor.ID, actor.Name, exclude)})
}

func creationOptions(challenge []byte, rp, userID, userName string, exclude []map[string]string) map[string]any {
	return map[string]any{
		"challenge": b64url(challenge), "rp": map[string]string{"id": rp, "name": "Agent CommonRoom"},
		"user":                   map[string]string{"id": b64url([]byte(userID)), "name": userName, "displayName": userName},
		"pubKeyCredParams":       []map[string]any{{"type": "public-key", "alg": webauthn.AlgES256}, {"type": "public-key", "alg": webauthn.AlgRS256}},
		"authenticatorSelection": map[string]string{"residentKey": "required", "userVerification": "required"},
		"attestation":            "none", "timeout": 60000, "excludeCredentials": exclude,
	}
}

type registrationBody struct {
	ChallengeID       string `json:"challengeId"`
	AttestationObject string `json:"attestationObject"`
	ClientDataJSON    string `json:"clientDataJSON"`
	Name              string `json:"name"`
}

func (s *server) verifyRegistration(r *http.Request, body registrationBody, challenge identityChallenge) (webauthn.RegistrationResult, bool) {
	attestation, ok1 := decodeB64(body.AttestationObject, 8192)
	clientData, ok2 := decodeB64(body.ClientDataJSON, 4096)
	if !ok1 || !ok2 {
		return webauthn.RegistrationResult{}, false
	}
	origin, rp := s.auth.expectedOrigin(r)
	res, err := webauthn.VerifyRegistration(webauthn.RegistrationParams{AttestationObject: attestation, ClientDataJSON: clientData, Challenge: challenge.Challenge, Origin: origin, RPID: rp})
	return res, err == nil
}

func (s *server) passkeyRegisterVerify(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	var body registrationBody
	if !readIdentityBody(w, r, &body, 32768) {
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "Passkey"
	}
	challenge, ok := s.identityStore.takeChallenge(body.ChallengeID, "register", time.Now())
	if !ok || challenge.ActorID != requestActor(r) || !validWorkspaceName(name) || len(name) > 60 {
		attestationDeny(w)
		return
	}
	res, ok := s.verifyRegistration(r, body, challenge)
	if !ok {
		attestationDeny(w)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	actor, found := member(st, challenge.ActorID)
	if err != nil || !found || !actor.Enabled || len(actor.Credentials) >= maxCredentialsPerActor || !s.identityFlag(st, flagPasskeys) {
		attestationDeny(w)
		return
	}
	if _, _, exists := findCredential(st, b64url(res.CredentialID)); exists {
		http.Error(w, "That passkey is already registered", 409)
		return
	}
	cred := room.Credential{ID: b64url(res.CredentialID), Name: name, PublicKey: stdB64(res.PublicKeyDER), Alg: res.Alg, SignCount: res.SignCount, Created: time.Now().UTC()}
	if err := s.append(context.WithValue(r.Context(), actorContextKey{}, actor.ID), "actor.credential.added", credentialAdded{ActorID: actor.ID, Credential: cred}); err != nil {
		http.Error(w, "Passkey could not be saved", 503)
		return
	}
	reply(w, map[string]any{"added": true, "id": cred.ID, "name": cred.Name})
}

func (s *server) passkeyRemove(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	var req struct {
		CredentialID string `json:"credentialId"`
	}
	if !readIdentityBody(w, r, &req, 2048) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	actor, ok := member(st, requestActor(r))
	if err != nil || !ok {
		http.Error(w, "Account unavailable", 503)
		return
	}
	found := false
	for _, c := range actor.Credentials {
		found = found || c.ID == req.CredentialID
	}
	if !found {
		http.Error(w, "Passkey not found", 404)
		return
	}
	if !hasOtherSignInMethod(actor, req.CredentialID) {
		http.Error(w, "Add another way to sign in before removing your last one", 409)
		return
	}
	if err := s.append(r.Context(), "actor.credential.removed", credentialRemoved{ActorID: actor.ID, CredentialID: req.CredentialID}); err != nil {
		http.Error(w, "Passkey could not be removed", 503)
		return
	}
	reply(w, map[string]any{"removed": true})
}
