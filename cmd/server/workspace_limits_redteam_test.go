package main

import (
	"agentroom/internal/room"
	"agentroom/internal/webauthn/webauthntest"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"testing"
	"time"
)

// RT-7: dailyTurns=0 is accepted as a valid limit (agents=0 means none), but the budget
// treats a zero limit as "unset" and silently grants the default 200 turns a day.
func TestRedteamZeroDailyTurnsLimitGrantsTheDefault(t *testing.T) {
	t.Parallel()
	limits, err := parseWorkspaceLimits([]byte(`{"dailyTurns":0}`))
	if err != nil {
		t.Skip("zero rejected at parse time, which would also be a fix:", err)
	}
	s := &server{dailyTurnLimit: 100}
	s.setLimits(&limits)
	if err := s.admitDailyTurns(room.State{}, time.Now(), 1); err == nil {
		t.Fatalf("workspace limited to 0 daily turns was admitted; budget limit=%d", s.dailyTurnBudget(room.State{}, time.Now()).Limit)
	}
}

// RT-2: the member cap is enforced at two creation paths that count differently. Invitation
// creation counts open invitations; direct member creation (/api/members) ignores them; and
// redeeming an invitation never re-checks. Together they exceed the operator's cap.
func TestRedteamMemberCapExceededByInvitationPlusDirectAdd(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	if w := f.call("PUT", "/platform/v1/workspaces/acme/limits", `{"members":1}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	owner := f.ownerSession("acme")
	token, w := f.invite(owner, "Developer") // 0 members + 0 open invitations < 1: allowed
	if w.Code != 200 {
		t.Fatal("first invitation", w.Code, w.Body.String())
	}
	if _, w := f.invite(owner, "QA"); w.Code != 409 {
		t.Fatal("second invitation must be refused at the cap", w.Code)
	}
	key, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pemKey := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
	body, _ := json.Marshal(map[string]any{"name": "Direct", "role": "Developer", "roomIds": []string{}, "publicKey": pemKey})
	direct := f.request("acme", "POST", "/api/members", string(body), &owner, nil)
	redeemed := f.redeemWithPasskey(token, "Invited", webauthntest.New())
	humans := 0
	for _, a := range f.state("acme").Actors {
		if a.Kind == "human" && a.ID != ownerID {
			humans++
		}
	}
	if humans > 1 {
		t.Fatalf("cap is 1 member besides the owner but %d exist (direct add %d, redeem %d)", humans, direct.Code, redeemed.Code)
	}
}
