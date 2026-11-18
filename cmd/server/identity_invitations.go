package main

import (
	"agentroom/internal/oidc"
	"agentroom/internal/room"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"time"
)

// refuseInvitation is the single answer for every invitation that cannot be used, so a
// caller learns nothing about whether a secret ever existed.
func refuseInvitation(w http.ResponseWriter, a *ownerAuth) {
	a.countFailure()
	http.Error(w, "This invitation is not valid", 404)
}

// createInvitation: the owner makes a single-use link for one new member.
func (s *server) createInvitation(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	var req struct {
		Role    string   `json:"role"`
		RoomIDs []string `json:"roomIds"`
		Hours   int      `json:"hours"`
	}
	if !readIdentityBody(w, r, &req, 8192) {
		return
	}
	if req.Role != "Developer" && req.Role != "QA" {
		http.Error(w, "Choose Developer or QA", 400)
		return
	}
	if req.Hours == 0 {
		req.Hours = defaultInviteHours
	}
	if req.Hours < 1 || req.Hours > maxInviteHours || len(req.RoomIDs) > 64 {
		http.Error(w, "Invitations last 1 to 168 hours and name at most 64 rooms", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	if !s.identityFlag(st, flagInvitations) {
		http.Error(w, "Invitations are disabled", 403)
		return
	}
	for _, rid := range req.RoomIDs {
		found := false
		for _, rm := range st.Rooms {
			found = found || (rm.ID == rid && rm.PersonalActorID == "")
		}
		if !found {
			http.Error(w, "Only shared rooms can be assigned", 400)
			return
		}
	}
	now := time.Now().UTC()
	if openInvitationCount(st, now) >= maxOpenInvitations {
		http.Error(w, "Revoke an unused invitation before creating another", 409)
		return
	}
	if msg := s.limitRefusal("members", humanMembers(st, now)); msg != "" {
		http.Error(w, msg, 409)
		return
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		http.Error(w, "Invitation could not be created", 500)
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	inv := room.Invitation{ID: invitationID(token), Role: req.Role, RoomIDs: append([]string{}, req.RoomIDs...), Created: now, Expires: now.Add(time.Duration(req.Hours) * time.Hour), CreatedBy: ownerID}
	if err := s.append(r.Context(), "invitation.created", inv); err != nil {
		http.Error(w, "Invitation could not be saved", 503)
		return
	}
	origin, _ := s.auth.expectedOrigin(r)
	reply(w, map[string]any{"id": inv.ID, "link": origin + "/#invite=" + token, "expires": inv.Expires, "role": inv.Role})
}

func (s *server) revokeInvitation(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if !readIdentityBody(w, r, &req, 512) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	for _, inv := range st.Invitations {
		if inv.ID == req.ID && inv.Status == invitationOpen {
			if err := s.append(r.Context(), "invitation.revoked", invitationRevokedEvent{ID: inv.ID}); err != nil {
				http.Error(w, "Invitation could not be revoked", 503)
				return
			}
			reply(w, map[string]any{"revoked": true})
			return
		}
	}
	http.Error(w, "Invitation not found", 404)
}

func validTokenShape(token string) bool {
	if len(token) < 40 || len(token) > 64 {
		return false
	}
	_, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil
}

// lookupInvitation lets the invitee's browser confirm the link before asking for a name.
func (s *server) lookupInvitation(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if !readIdentityBody(w, r, &req, 512) {
		return
	}
	now := time.Now()
	if s.auth.throttled(now) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "Too many attempts; try again in a minute", 429)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "Unavailable", 503)
		return
	}
	inv, ok := openInvitation(st, req.Token, now)
	if !validTokenShape(req.Token) || !ok || !s.identityFlag(st, flagInvitations) {
		refuseInvitation(w, s.auth)
		return
	}
	reply(w, map[string]any{"valid": true, "role": inv.Role, "expires": inv.Expires, "passkeys": s.identityFlag(st, flagPasskeys), "oidc": s.oidcAvailable(st), "oidcName": s.identity.name})
}

func validPersonName(name string) bool { return validWorkspaceName(name) }

// invitationPasskeyOptions starts passkey registration for a new member.
func (s *server) invitationPasskeyOptions(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	var req struct {
		Token string `json:"token"`
		Name  string `json:"name"`
	}
	if !readIdentityBody(w, r, &req, 1024) {
		return
	}
	now := time.Now()
	if s.auth.throttled(now) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "Too many attempts; try again in a minute", 429)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "Unavailable", 503)
		return
	}
	inv, ok := openInvitation(st, req.Token, now)
	if !validTokenShape(req.Token) || !ok || !s.identityFlag(st, flagInvitations) || !s.identityFlag(st, flagPasskeys) {
		refuseInvitation(w, s.auth)
		return
	}
	if !validPersonName(req.Name) {
		http.Error(w, "Enter your name (1-100 characters)", 400)
		return
	}
	newActor := id()
	challenge := randomChallenge()
	_, rp := s.auth.expectedOrigin(r)
	cid, ok := s.identityStore.putChallenge(identityChallenge{Purpose: "invite", Challenge: challenge, ActorID: newActor, Invite: inv.ID, Name: req.Name}, now)
	if !ok {
		http.Error(w, "Too many sign-ins in progress", 429)
		return
	}
	reply(w, map[string]any{"challengeId": cid, "publicKey": creationOptions(challenge, rp, newActor, req.Name, []map[string]string{})})
}

// redeemInvitationLocked creates the member, records how they sign in and spends the
// invitation in one atomic batch. The caller holds s.mu.
func (s *server) redeemInvitationLocked(ctx context.Context, st room.State, inviteID, actorID, name string, extra auditChange, now time.Time) (room.Actor, int, string) {
	var inv room.Invitation
	found := false
	for _, candidate := range st.Invitations {
		if candidate.ID == inviteID && candidate.Status == invitationOpen && now.Before(candidate.Expires) {
			inv, found = candidate, true
		}
	}
	if !found {
		return room.Actor{}, 404, "This invitation is not valid"
	}
	rooms := []string{}
	for _, rid := range inv.RoomIDs {
		for _, rm := range st.Rooms {
			if rm.ID == rid && rm.PersonalActorID == "" {
				rooms = append(rooms, rid)
			}
		}
	}
	actor := room.Actor{ID: actorID, Name: name, Kind: "human", Role: inv.Role, Enabled: true, RoomIDs: rooms, Revision: 1, ChangedBy: inv.CreatedBy}
	ctx = context.WithValue(ctx, actorContextKey{}, actorID)
	err := s.appendBatch(ctx, []auditChange{{Kind: "actor.revised", Payload: actor}, extra, {Kind: "invitation.redeemed", Payload: invitationRedeemedEvent{ID: inv.ID, ActorID: actorID}}})
	if err != nil {
		return room.Actor{}, 503, "Your account could not be created"
	}
	return actor, 200, ""
}

// invitationPasskeyComplete finishes a new member's passkey registration.
func (s *server) invitationPasskeyComplete(w http.ResponseWriter, r *http.Request) {
	if !s.requirePost(w, r) {
		return
	}
	var body registrationBody
	if !readIdentityBody(w, r, &body, 32768) {
		return
	}
	now := time.Now()
	if s.auth.throttled(now) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "Too many attempts; try again in a minute", 429)
		return
	}
	challenge, ok := s.identityStore.takeChallenge(body.ChallengeID, "invite", now)
	if !ok {
		refuseInvitation(w, s.auth)
		return
	}
	res, ok := s.verifyRegistration(r, body, challenge)
	if !ok {
		refuseInvitation(w, s.auth)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil || !s.identityFlag(st, flagInvitations) || !s.identityFlag(st, flagPasskeys) {
		refuseInvitation(w, s.auth)
		return
	}
	if _, _, exists := findCredential(st, b64url(res.CredentialID)); exists {
		refuseInvitation(w, s.auth)
		return
	}
	cred := room.Credential{ID: b64url(res.CredentialID), Name: "Passkey", PublicKey: stdB64(res.PublicKeyDER), Alg: res.Alg, SignCount: res.SignCount, Created: now.UTC()}
	actor, code, msg := s.redeemInvitationLocked(r.Context(), st, challenge.Invite, challenge.ActorID, challenge.Name, auditChange{Kind: "actor.credential.added", Payload: credentialAdded{ActorID: challenge.ActorID, Credential: cred}}, now)
	if code != 200 {
		if code == 404 {
			refuseInvitation(w, s.auth)
		} else {
			http.Error(w, msg, code)
		}
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

// ---- single sign-on ----

func (s *server) oidcRedirectURI(r *http.Request) string {
	origin, _ := s.auth.expectedOrigin(r)
	return origin + "/api/oidc/callback"
}

func (s *server) beginOIDC(w http.ResponseWriter, r *http.Request, pending oidcPending) {
	verifier, challenge := oidc.PKCE()
	pending.Nonce, pending.Verifier = oidc.Random(), verifier
	// Bind this flow to this browser so a callback URL cannot be replayed into another one.
	binding := oidc.Random()
	pending.Binding = invitationID(binding)
	http.SetCookie(w, &http.Cookie{Name: "commonroom_oidc", Value: binding, Path: "/api/oidc", HttpOnly: true, Secure: s.auth.publicOrigin != "", SameSite: http.SameSiteLaxMode, MaxAge: int(oidcPendingTTL.Seconds())})
	state, ok := s.identityStore.putPending(pending, time.Now())
	if !ok {
		http.Error(w, "Too many sign-ins in progress", 429)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	target, err := s.identity.provider.AuthURL(ctx, s.oidcRedirectURI(r), state, pending.Nonce, challenge)
	if err != nil {
		http.Error(w, "Single sign-on is unavailable", 503)
		return
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (s *server) oidcLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	if err != nil || !s.oidcAvailable(st) {
		http.Error(w, "Single sign-on is not available", 403)
		return
	}
	if s.auth.throttled(time.Now()) {
		http.Error(w, "Too many attempts; try again in a minute", 429)
		return
	}
	s.beginOIDC(w, r, oidcPending{Purpose: "login"})
}

func (s *server) oidcLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	st, err := s.load(r.Context())
	actor, ok := member(st, requestActor(r))
	if err != nil || !ok || actor.Kind != "human" || !s.oidcAvailable(st) {
		http.Error(w, "Single sign-on is not available", 403)
		return
	}
	s.beginOIDC(w, r, oidcPending{Purpose: "link", ActorID: actor.ID})
}

func (s *server) oidcInviteStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	token, name := r.URL.Query().Get("token"), strings.TrimSpace(r.URL.Query().Get("name"))
	now := time.Now()
	if s.auth.throttled(now) {
		http.Error(w, "Too many attempts; try again in a minute", 429)
		return
	}
	st, err := s.load(r.Context())
	if err != nil {
		http.Error(w, "Unavailable", 503)
		return
	}
	if _, ok := openInvitation(st, token, now); !validTokenShape(token) || !ok || !s.identityFlag(st, flagInvitations) || !s.oidcAvailable(st) {
		refuseInvitation(w, s.auth)
		return
	}
	if !validPersonName(name) {
		http.Error(w, "Enter your name (1-100 characters)", 400)
		return
	}
	s.beginOIDC(w, r, oidcPending{Purpose: "invite", Token: token, Name: name})
}

func (s *server) oidcCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	now := time.Now()
	back := func(result string) { http.Redirect(w, r, "/#signin="+result, http.StatusSeeOther) }
	pending, ok := s.identityStore.takePending(r.URL.Query().Get("state"), now)
	cookie, cookieErr := r.Cookie("commonroom_oidc")
	bound := cookieErr == nil && ok && subtle.ConstantTimeCompare([]byte(invitationID(cookie.Value)), []byte(pending.Binding)) == 1
	http.SetCookie(w, &http.Cookie{Name: "commonroom_oidc", Value: "", Path: "/api/oidc", MaxAge: -1})
	code := r.URL.Query().Get("code")
	if !ok || !bound || code == "" || len(code) > 2048 || r.URL.Query().Get("error") != "" {
		s.auth.countFailure()
		back("failed")
		return
	}
	if s.auth.throttled(now) {
		back("failed")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	claims, err := s.identity.provider.Exchange(ctx, code, s.oidcRedirectURI(r), pending.Verifier, pending.Nonce)
	if err != nil {
		s.auth.countFailure()
		back("failed")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.load(r.Context())
	if err != nil || !s.oidcAvailable(st) {
		back("failed")
		return
	}
	link := room.OIDCLink{Issuer: claims.Issuer, Subject: claims.Subject, Email: claims.Email}
	startSession := func(actor room.Actor) {
		s.auth.mu.Lock()
		defer s.auth.mu.Unlock()
		if _, ok := s.auth.startSessionLocked(w, r, actor.ID, now); ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
	switch pending.Purpose {
	case "login":
		actor, found := findOIDCActor(st, link.Issuer, link.Subject)
		if !found || !actor.Enabled {
			s.auth.countFailure()
			back("unlinked")
			return
		}
		startSession(actor)
	case "link":
		// The state was created by an authenticated request and is bound to this browser;
		// the session cookie itself is not sent on a cross-site return (SameSite=Strict).
		actor, found := member(st, pending.ActorID)
		if !found || !actor.Enabled {
			back("failed")
			return
		}
		if other, taken := findOIDCActor(st, link.Issuer, link.Subject); taken && other.ID != actor.ID {
			back("taken")
			return
		}
		if err := s.append(context.WithValue(r.Context(), actorContextKey{}, actor.ID), "actor.oidc.linked", oidcLinked{ActorID: actor.ID, Link: link}); err != nil {
			back("failed")
			return
		}
		back("linked")
	case "invite":
		if _, taken := findOIDCActor(st, link.Issuer, link.Subject); taken {
			back("taken")
			return
		}
		inv, valid := openInvitation(st, pending.Token, now)
		if !valid {
			s.auth.countFailure()
			back("invalid")
			return
		}
		newID := id()
		actor, status, _ := s.redeemInvitationLocked(r.Context(), st, inv.ID, newID, pending.Name, auditChange{Kind: "actor.oidc.linked", Payload: oidcLinked{ActorID: newID, Link: link}}, now)
		if status != 200 {
			back("failed")
			return
		}
		startSession(actor)
	default:
		back("failed")
	}
}
