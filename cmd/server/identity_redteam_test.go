package main

import (
	"agentroom/internal/webauthn/webauthntest"
	"testing"
)

// RT-9: the session table is capped at 32 per workspace, not per person. Any member can
// fill it by signing in repeatedly without presenting a cookie, and then nobody (including
// the owner) can start a new session for the 8 hours a session lives.
func TestRedteamMemberCanLockOwnerOutByFillingSessionTable(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	rt := f.runtime("acme")
	code, _, _ := rt.s.auth.setupSnapshot()
	owner := f.ownerSession("acme")
	token, w := f.invite(owner, "Developer")
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	device := webauthntest.New()
	if r := f.redeemWithPasskey(token, "Mallory", device); r.Code != 200 {
		t.Fatal("redeem", r.Code, r.Body.String())
	}
	for i := 0; i < 40; i++ {
		f.resetThrottle("acme")
		if r := f.signInWithPasskey(device); r.Code == 429 {
			break
		}
	}
	w = f.request("acme", "POST", "/api/login", map[string]string{"key": code}, nil, nil)
	if w.Code != 200 {
		t.Fatalf("owner could not sign in after a member filled the session table: %d %s", w.Code, w.Body.String())
	}
}

// RT-10: anonymous passkey sign-in options fill the 512-slot challenge table (5 minute
// lifetime) well inside the default 600 requests/minute, after which no one in the
// workspace can start a passkey sign-in or accept an invitation by passkey.
func TestRedteamAnonymousChallengeFloodBlocksPasskeySignIn(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	token, w := f.invite(owner, "Developer")
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	for i := 0; i < 520; i++ {
		f.do("POST", "/api/webauthn/login/options", `{}`, nil)
	}
	if r := f.do("POST", "/api/invitations/passkey/options", map[string]string{"token": token, "name": "Newcomer"}, nil); r.Code != 200 {
		t.Fatalf("a valid invitee was refused after an anonymous flood: %d %s", r.Code, r.Body.String())
	}
}
