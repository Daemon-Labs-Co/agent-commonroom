package main

import (
	"agentroom/internal/room"
	"agentroom/internal/webauthn/webauthntest"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type idSession struct{ cookie, csrf string }

type idFixture struct {
	*platformFixture
	ws string
}

func newIdentityFixture(t *testing.T, mutate func(f *platformFixture)) *idFixture {
	t.Helper()
	pf := newPlatformFixture(t)
	if mutate != nil {
		mutate(pf)
	}
	pf.create("acme")
	return &idFixture{platformFixture: pf, ws: "acme"}
}

func (f *idFixture) hostName(ws string) string { return ws + ".rooms.example.com" }

func (f *idFixture) runtime(ws string) *workspaceRuntime {
	f.t.Helper()
	rt, err := f.host.adminRuntime(f.t.Context(), ws)
	if err != nil {
		f.t.Fatal(err)
	}
	return rt
}

func (f *idFixture) request(ws, method, path string, body any, s *idSession, mutate func(*http.Request)) *httptest.ResponseRecorder {
	var raw string
	switch v := body.(type) {
	case nil:
	case string:
		raw = v
	default:
		b, _ := json.Marshal(v)
		raw = string(b)
	}
	h := f.hostName(ws)
	r := httptest.NewRequest(method, "https://"+h+path, strings.NewReader(raw))
	r.Host = h
	r.Header.Set("Origin", "https://"+h)
	r.Header.Set("Content-Type", "application/json")
	if s != nil {
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: s.cookie})
		r.Header.Set("X-CSRF-Token", s.csrf)
	}
	if mutate != nil {
		mutate(r)
	}
	w := httptest.NewRecorder()
	f.host.ServeHTTP(w, r)
	return w
}

func (f *idFixture) do(method, path string, body any, s *idSession) *httptest.ResponseRecorder {
	return f.request(f.ws, method, path, body, s, nil)
}

func sessionFrom(t *testing.T, w *httptest.ResponseRecorder) idSession {
	t.Helper()
	var out struct {
		CSRF string `json:"csrf"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	for _, c := range w.Result().Cookies() {
		if c.Name == "commonroom_session" && c.Value != "" {
			return idSession{cookie: c.Value, csrf: out.CSRF}
		}
	}
	t.Fatal("no session issued", w.Code, w.Body.String())
	return idSession{}
}

// ownerSession signs in with the setup code and marks setup finished, as claiming would.
func (f *idFixture) ownerSession(ws string) idSession {
	f.t.Helper()
	rt := f.runtime(ws)
	code, _, pending := rt.s.auth.setupSnapshot()
	if !pending {
		f.t.Fatal("setup already claimed")
	}
	w := f.request(ws, "POST", "/api/login", map[string]string{"key": code}, nil, nil)
	if w.Code != 200 {
		f.t.Fatal(w.Code, w.Body.String())
	}
	s := sessionFrom(f.t, w)
	rt.s.auth.mu.Lock()
	rt.s.auth.setupPending = false
	rt.s.auth.mu.Unlock()
	return s
}

func (f *idFixture) state(ws string) room.State {
	f.t.Helper()
	st, err := f.runtime(ws).s.load(f.t.Context())
	if err != nil {
		f.t.Fatal(err)
	}
	return st
}

func (f *idFixture) resetThrottle(ws string) {
	rt := f.runtime(ws)
	rt.s.auth.mu.Lock()
	rt.s.auth.attempts = 0
	rt.s.auth.mu.Unlock()
}

func (f *idFixture) origin() string { return "https://" + f.hostName(f.ws) }
func (f *idFixture) rp() string     { return f.hostName(f.ws) }

func (f *idFixture) invite(owner idSession, role string) (token string, resp *httptest.ResponseRecorder) {
	resp = f.do("POST", "/api/invitations", map[string]any{"role": role, "roomIds": []string{"engineering"}}, &owner)
	var out struct {
		Link string `json:"link"`
	}
	json.Unmarshal(resp.Body.Bytes(), &out)
	if i := strings.Index(out.Link, "#invite="); i >= 0 {
		token = out.Link[i+len("#invite="):]
	}
	return token, resp
}

type options struct {
	ChallengeID string `json:"challengeId"`
	PublicKey   struct {
		Challenge string `json:"challenge"`
		RPID      string `json:"rpId"`
		RP        struct {
			ID string `json:"id"`
		} `json:"rp"`
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		Exclude []struct {
			ID string `json:"id"`
		} `json:"excludeCredentials"`
	} `json:"publicKey"`
}

func decodeOptions(t *testing.T, w *httptest.ResponseRecorder) options {
	t.Helper()
	var o options
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &o) != nil || o.ChallengeID == "" {
		t.Fatal("options", w.Code, w.Body.String())
	}
	return o
}

// redeemWithPasskey runs the invitee's whole passkey flow and returns the final response.
func (f *idFixture) redeemWithPasskey(token, name string, a *webauthntest.Authenticator) *httptest.ResponseRecorder {
	f.t.Helper()
	opt := decodeOptions(f.t, f.do("POST", "/api/invitations/passkey/options", map[string]string{"token": token, "name": name}, nil))
	if opt.PublicKey.RP.ID != f.rp() {
		f.t.Fatal("relying party must be the workspace host", opt.PublicKey.RP.ID)
	}
	id, att, cd := a.Create(webauthntest.Decode(opt.PublicKey.Challenge), f.rp(), f.origin())
	_ = id
	return f.do("POST", "/api/invitations/passkey/complete", map[string]string{"challengeId": opt.ChallengeID, "attestationObject": att, "clientDataJSON": cd}, nil)
}

func (f *idFixture) signInWithPasskey(a *webauthntest.Authenticator) *httptest.ResponseRecorder {
	f.t.Helper()
	opt := decodeOptions(f.t, f.do("POST", "/api/webauthn/login/options", `{}`, nil))
	ad, cd, sig := a.Get(webauthntest.Decode(opt.PublicKey.Challenge), f.rp(), f.origin())
	return f.do("POST", "/api/webauthn/login/verify", map[string]string{"challengeId": opt.ChallengeID, "id": a.CredentialID(), "authenticatorData": ad, "clientDataJSON": cd, "signature": sig}, nil)
}

func TestInvitationLinkIsSingleUseAndStoredOnlyAsAHash(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	token, w := f.invite(owner, "Developer")
	if w.Code != 200 || len(token) != 43 {
		t.Fatal(w.Code, w.Body.String(), token)
	}
	var link struct{ Link string }
	json.Unmarshal(w.Body.Bytes(), &link)
	if !strings.HasPrefix(link.Link, "https://acme.rooms.example.com/#invite=") {
		t.Fatal("invitation link must point at the workspace:", link.Link)
	}
	for _, e := range f.dbs["acme"].snapshot() {
		if strings.Contains(e.payload, token) {
			t.Fatal("the invitation secret was written to the history")
		}
	}
	if w := f.do("POST", "/api/invitations/lookup", map[string]string{"token": token}, nil); w.Code != 200 || !strings.Contains(w.Body.String(), `"role":"Developer"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	member := webauthntest.New()
	w = f.redeemWithPasskey(token, "Dana", member)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	session := sessionFrom(t, w)
	st := f.state("acme")
	var dana room.Actor
	for _, a := range st.Actors {
		if a.Name == "Dana" {
			dana = a
		}
	}
	if dana.Kind != "human" || dana.Role != "Developer" || !dana.Enabled || len(dana.Credentials) != 1 || dana.Credentials[0].ID != member.CredentialID() {
		t.Fatal("account not created as invited", dana)
	}
	if len(st.Invitations) != 1 || st.Invitations[0].Status != invitationRedeemed || st.Invitations[0].RedeemedBy != dana.ID {
		t.Fatal("invitation was not marked redeemed", st.Invitations)
	}
	if w := f.do("GET", "/api/session", nil, &session); w.Code != 200 {
		t.Fatal("the new member has no working session", w.Code)
	}
	// The same link, however it is presented, is now worthless.
	f.resetThrottle("acme")
	for name, run := range map[string]func() int{
		"lookup": func() int {
			return f.do("POST", "/api/invitations/lookup", map[string]string{"token": token}, nil).Code
		},
		"options": func() int {
			return f.do("POST", "/api/invitations/passkey/options", map[string]string{"token": token, "name": "Eve"}, nil).Code
		},
	} {
		if code := run(); code != 404 {
			t.Fatal("a redeemed invitation was usable again:", name, code)
		}
	}
}

func TestInvitationRefusalsAreIdentical(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	revoked, _ := f.invite(owner, "QA")
	var id string
	for _, inv := range f.state("acme").Invitations {
		id = inv.ID
	}
	if w := f.do("POST", "/api/invitations/revoke", map[string]string{"id": id}, &owner); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	expiredToken, _ := f.invite(owner, "QA")
	for _, inv := range f.state("acme").Invitations {
		if inv.Status == invitationOpen {
			// Age the invitation past its expiry by appending a corrected record.
			rt := f.runtime("acme")
			inv.Expires = time.Now().Add(-time.Hour)
			inv.ID = invitationID("expired-" + expiredToken)
			_ = rt
		}
	}
	bodies := map[string]string{}
	for name, token := range map[string]string{"revoked": revoked, "unknown": strings.Repeat("a", 43), "short": "abc", "empty": "", "junk": strings.Repeat("!", 43)} {
		f.resetThrottle("acme")
		w := f.do("POST", "/api/invitations/lookup", map[string]string{"token": token}, nil)
		if w.Code != 404 {
			t.Fatal(name, w.Code, w.Body.String())
		}
		bodies[name] = w.Body.String()
	}
	for name, body := range bodies {
		if body != bodies["unknown"] {
			t.Fatal("refusals differ, so they reveal why:", name, body)
		}
	}
	// Failed attempts share the sign-in budget.
	f.resetThrottle("acme")
	var last int
	for i := 0; i < 7; i++ {
		last = f.do("POST", "/api/invitations/lookup", map[string]string{"token": strings.Repeat("b", 43)}, nil).Code
	}
	if last != 429 {
		t.Fatal("guessing invitation links was not rate limited", last)
	}
}

func TestInvitationConcurrentRedemptionCreatesExactlyOneMember(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	token, _ := f.invite(owner, "Developer")
	const attempts = 4
	codes := make([]int, attempts)
	prepared := make([]func() *httptest.ResponseRecorder, attempts)
	for i := 0; i < attempts; i++ {
		a := webauthntest.New()
		opt := decodeOptions(t, f.do("POST", "/api/invitations/passkey/options", map[string]string{"token": token, "name": "Racer"}, nil))
		_, att, cd := a.Create(webauthntest.Decode(opt.PublicKey.Challenge), f.rp(), f.origin())
		prepared[i] = func() *httptest.ResponseRecorder {
			return f.do("POST", "/api/invitations/passkey/complete", map[string]string{"challengeId": opt.ChallengeID, "attestationObject": att, "clientDataJSON": cd}, nil)
		}
	}
	var wg sync.WaitGroup
	for i := range prepared {
		wg.Add(1)
		go func() { defer wg.Done(); codes[i] = prepared[i]().Code }()
	}
	wg.Wait()
	ok := 0
	for _, c := range codes {
		if c == 200 {
			ok++
		}
	}
	humans := 0
	for _, a := range f.state("acme").Actors {
		if a.Kind == "human" && a.ID != ownerID {
			humans++
		}
	}
	if ok != 1 || humans != 1 {
		t.Fatal("one invitation must create exactly one member", codes, humans)
	}
}

func TestPasskeySignInEndToEndAndItsRefusals(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	token, _ := f.invite(owner, "QA")
	device := webauthntest.New()
	if w := f.redeemWithPasskey(token, "Quinn", device); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	signIn := f.signInWithPasskey(device)
	if signIn.Code != 200 {
		t.Fatal(signIn.Code, signIn.Body.String())
	}
	session := sessionFrom(t, signIn)
	var stored room.Credential
	for _, a := range f.state("acme").Actors {
		for _, c := range a.Credentials {
			stored = c
		}
	}
	if stored.SignCount != 1 {
		t.Fatal("the signature counter was not recorded", stored.SignCount)
	}
	if w := f.do("GET", "/api/session", nil, &session); w.Code != 200 {
		t.Fatal(w.Code)
	}
	f.resetThrottle("acme")
	// A challenge works once.
	opt := decodeOptions(t, f.do("POST", "/api/webauthn/login/options", `{}`, nil))
	ad, cd, sig := device.Get(webauthntest.Decode(opt.PublicKey.Challenge), f.rp(), f.origin())
	body := map[string]string{"challengeId": opt.ChallengeID, "id": device.CredentialID(), "authenticatorData": ad, "clientDataJSON": cd, "signature": sig}
	if w := f.do("POST", "/api/webauthn/login/verify", body, nil); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := f.do("POST", "/api/webauthn/login/verify", body, nil); w.Code != 401 {
		t.Fatal("an assertion was replayed", w.Code)
	}
	// One challenge cannot be spent twice even by a fresh, validly counted assertion.
	f.resetThrottle("acme")
	twice := decodeOptions(t, f.do("POST", "/api/webauthn/login/options", `{}`, nil))
	submit := func() int {
		ad, cd, sig := device.Get(webauthntest.Decode(twice.PublicKey.Challenge), f.rp(), f.origin())
		return f.do("POST", "/api/webauthn/login/verify", map[string]string{"challengeId": twice.ChallengeID, "id": device.CredentialID(), "authenticatorData": ad, "clientDataJSON": cd, "signature": sig}, nil).Code
	}
	if first, second := submit(), submit(); first != 200 || second != 401 {
		t.Fatal("a sign-in challenge was usable twice", first, second)
	}
	f.resetThrottle("acme")
	refusals := map[string]func() *httptest.ResponseRecorder{
		"unknown credential": func() *httptest.ResponseRecorder { return f.signInWithPasskey(webauthntest.New()) },
		"wrong origin": func() *httptest.ResponseRecorder {
			o := decodeOptions(t, f.do("POST", "/api/webauthn/login/options", `{}`, nil))
			ad, cd, sig := device.Get(webauthntest.Decode(o.PublicKey.Challenge), f.rp(), "https://evil.example.com")
			return f.do("POST", "/api/webauthn/login/verify", map[string]string{"challengeId": o.ChallengeID, "id": device.CredentialID(), "authenticatorData": ad, "clientDataJSON": cd, "signature": sig}, nil)
		},
		"wrong relying party": func() *httptest.ResponseRecorder {
			o := decodeOptions(t, f.do("POST", "/api/webauthn/login/options", `{}`, nil))
			ad, cd, sig := device.Get(webauthntest.Decode(o.PublicKey.Challenge), "beta.rooms.example.com", f.origin())
			return f.do("POST", "/api/webauthn/login/verify", map[string]string{"challengeId": o.ChallengeID, "id": device.CredentialID(), "authenticatorData": ad, "clientDataJSON": cd, "signature": sig}, nil)
		},
		"made-up challenge id": func() *httptest.ResponseRecorder {
			ad, cd, sig := device.Get(make([]byte, 32), f.rp(), f.origin())
			return f.do("POST", "/api/webauthn/login/verify", map[string]string{"challengeId": "nope", "id": device.CredentialID(), "authenticatorData": ad, "clientDataJSON": cd, "signature": sig}, nil)
		},
		"unverified user": func() *httptest.ResponseRecorder {
			weak := *device
			weak.Flags = 0x01 // presence only
			return f.signInWithPasskey(&weak)
		},
		"malformed fields": func() *httptest.ResponseRecorder {
			return f.do("POST", "/api/webauthn/login/verify", map[string]string{"challengeId": "x", "id": "!!", "authenticatorData": "!!", "clientDataJSON": "!!", "signature": "!!"}, nil)
		},
	}
	for name, run := range refusals {
		f.resetThrottle("acme")
		if w := run(); w.Code != 401 {
			t.Fatal("accepted:", name, w.Code, w.Body.String())
		}
	}
	// A challenge issued for one purpose cannot be used for another, and expires.
	f.resetThrottle("acme")
	reg := decodeOptions(t, f.do("POST", "/api/webauthn/register/options", `{}`, &session))
	ad2, cd2, sig2 := device.Get(webauthntest.Decode(reg.PublicKey.Challenge), f.rp(), f.origin())
	if w := f.do("POST", "/api/webauthn/login/verify", map[string]string{"challengeId": reg.ChallengeID, "id": device.CredentialID(), "authenticatorData": ad2, "clientDataJSON": cd2, "signature": sig2}, nil); w.Code != 401 {
		t.Fatal("a registration challenge was accepted for sign-in", w.Code)
	}
	opt = decodeOptions(t, f.do("POST", "/api/webauthn/login/options", `{}`, nil))
	rt := f.runtime("acme")
	rt.s.identityStore.mu.Lock()
	c := rt.s.identityStore.challenges[opt.ChallengeID]
	c.Expires = time.Now().Add(-time.Second)
	rt.s.identityStore.challenges[opt.ChallengeID] = c
	rt.s.identityStore.mu.Unlock()
	ad, cd, sig = device.Get(webauthntest.Decode(opt.PublicKey.Challenge), f.rp(), f.origin())
	if w := f.do("POST", "/api/webauthn/login/verify", map[string]string{"challengeId": opt.ChallengeID, "id": device.CredentialID(), "authenticatorData": ad, "clientDataJSON": cd, "signature": sig}, nil); w.Code != 401 {
		t.Fatal("an expired challenge was accepted", w.Code)
	}
}

func TestPasskeyRefusedForDisabledAccountsAndAnotherWorkspace(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	f.create("beta")
	owner := f.ownerSession("acme")
	token, _ := f.invite(owner, "Developer")
	device := webauthntest.New()
	w := f.redeemWithPasskey(token, "Dee", device)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	// The same authenticator cannot sign in to another workspace (its own host, relying
	// party and history) even though the credential id is the same.
	opt := decodeOptions(t, f.request("beta", "POST", "/api/webauthn/login/options", `{}`, nil, nil))
	ad, cd, sig := device.Get(webauthntest.Decode(opt.PublicKey.Challenge), "beta.rooms.example.com", "https://beta.rooms.example.com")
	if w := f.request("beta", "POST", "/api/webauthn/login/verify", map[string]string{"challengeId": opt.ChallengeID, "id": device.CredentialID(), "authenticatorData": ad, "clientDataJSON": cd, "signature": sig}, nil, nil); w.Code != 401 {
		t.Fatal("a credential from one workspace worked in another", w.Code)
	}
	// Disabling the member stops passkey sign-in.
	var dee room.Actor
	for _, a := range f.state("acme").Actors {
		if a.Name == "Dee" {
			dee = a
		}
	}
	dee.Enabled = false
	dee.Revision++
	if err := f.runtime("acme").s.append(t.Context(), "actor.revised", dee); err != nil {
		t.Fatal(err)
	}
	f.resetThrottle("acme")
	if w := f.signInWithPasskey(device); w.Code != 401 {
		t.Fatal("a disabled account signed in", w.Code)
	}
}

func TestSignedInPeopleManageTheirOwnPasskeys(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	token, _ := f.invite(owner, "Developer")
	first := webauthntest.New()
	member := sessionFrom(t, f.redeemWithPasskey(token, "Mo", first))
	register := func(s idSession, a *webauthntest.Authenticator, name string) *httptest.ResponseRecorder {
		opt := decodeOptions(t, f.do("POST", "/api/webauthn/register/options", `{}`, &s))
		_, att, cd := a.Create(webauthntest.Decode(opt.PublicKey.Challenge), f.rp(), f.origin())
		return f.do("POST", "/api/webauthn/register/verify", map[string]string{"challengeId": opt.ChallengeID, "attestationObject": att, "clientDataJSON": cd, "name": name}, &s)
	}
	second := webauthntest.New()
	if w := register(member, second, "Phone"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	// The same authenticator cannot be added twice.
	opt := decodeOptions(t, f.do("POST", "/api/webauthn/register/options", `{}`, &member))
	if len(opt.PublicKey.Exclude) != 2 {
		t.Fatal("existing passkeys must be excluded from registration", opt.PublicKey.Exclude)
	}
	_, att, cd := second.Create(webauthntest.Decode(opt.PublicKey.Challenge), f.rp(), f.origin())
	if w := f.do("POST", "/api/webauthn/register/verify", map[string]string{"challengeId": opt.ChallengeID, "attestationObject": att, "clientDataJSON": cd}, &member); w.Code != 409 {
		t.Fatal("a duplicate passkey was accepted", w.Code)
	}
	// A registration challenge belongs to the person who asked for it.
	opt = decodeOptions(t, f.do("POST", "/api/webauthn/register/options", `{}`, &member))
	_, att, cd = webauthntest.New().Create(webauthntest.Decode(opt.PublicKey.Challenge), f.rp(), f.origin())
	if w := f.do("POST", "/api/webauthn/register/verify", map[string]string{"challengeId": opt.ChallengeID, "attestationObject": att, "clientDataJSON": cd}, &owner); w.Code != 400 {
		t.Fatal("someone else completed another person's registration", w.Code)
	}
	if w := f.do("POST", "/api/webauthn/register/options", `{}`, nil); w.Code != 401 {
		t.Fatal("registration without a session", w.Code)
	}
	if w := f.request("acme", "POST", "/api/webauthn/register/options", `{}`, &idSession{cookie: member.cookie, csrf: "wrong"}, nil); w.Code != 403 {
		t.Fatal("registration without a CSRF token", w.Code)
	}
	// Removing a passkey keeps at least one way to sign in, and only your own can go.
	stateMember := func() room.Actor {
		for _, a := range f.state("acme").Actors {
			if a.Name == "Mo" {
				return a
			}
		}
		return room.Actor{}
	}
	creds := stateMember().Credentials
	if len(creds) != 2 {
		t.Fatal(creds)
	}
	remove := func(s idSession, id string) int {
		return f.do("POST", "/api/webauthn/remove", map[string]string{"credentialId": id}, &s).Code
	}
	if code := remove(owner, creds[0].ID); code != 404 {
		t.Fatal("removed someone else's passkey", code)
	}
	if code := remove(member, creds[0].ID); code != 200 {
		t.Fatal(code)
	}
	if code := remove(member, creds[1].ID); code != 409 {
		t.Fatal("removed the last way to sign in", code)
	}
	f.resetThrottle("acme")
	if w := f.signInWithPasskey(first); w.Code == 200 && stateMember().Credentials[0].ID == first.CredentialID() {
		t.Fatal("a removed passkey still works")
	}
}

func TestIdentityRecordsAreVisibleOnlyToTheirOwners(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	tokenA, _ := f.invite(owner, "Developer")
	tokenB, _ := f.invite(owner, "QA")
	a, b := webauthntest.New(), webauthntest.New()
	sessionA := sessionFrom(t, f.redeemWithPasskey(tokenA, "Ann", a))
	f.redeemWithPasskey(tokenB, "Bob", b)
	for name, check := range map[string]func(w *httptest.ResponseRecorder){
		"member state": func(w *httptest.ResponseRecorder) {
			var st struct {
				Actors      []room.Actor      `json:"actors"`
				Invitations []room.Invitation `json:"invitations"`
			}
			json.Unmarshal(w.Body.Bytes(), &st)
			if len(st.Invitations) != 0 {
				t.Fatal("a member saw invitations")
			}
			for _, actor := range st.Actors {
				if actor.Name == "Bob" && len(actor.Credentials) > 0 {
					t.Fatal("a member saw another member's passkeys")
				}
			}
		},
	} {
		w := f.do("GET", "/api/state", nil, &sessionA)
		if w.Code != 200 {
			t.Fatal(name, w.Code, w.Body.String())
		}
		check(w)
	}
	if w := f.do("POST", "/api/invitations", map[string]any{"role": "QA"}, &sessionA); w.Code != 403 {
		t.Fatal("a member created an invitation", w.Code)
	}
	if w := f.do("POST", "/api/invitations", map[string]any{"role": "QA"}, nil); w.Code != 401 {
		t.Fatal("an anonymous caller created an invitation", w.Code)
	}
}

func TestIdentityEventsCannotBeForgedThroughTheGenericEndpoint(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	for _, kind := range []string{"actor.credential.added", "actor.credential.used", "actor.credential.removed", "actor.oidc.linked", "actor.oidc.unlinked", "invitation.created", "invitation.redeemed", "invitation.revoked"} {
		w := f.do("POST", "/api/events", map[string]any{"kind": kind, "payload": map[string]any{"actorId": "local-owner"}}, &owner)
		if w.Code != 400 {
			t.Fatal("the events endpoint accepted", kind, w.Code)
		}
	}
}

func TestIdentityFlagsDefaultOnForHostedAndCanBeTurnedOff(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	methods := func() map[string]any {
		var m map[string]any
		json.Unmarshal(f.do("GET", "/api/auth/methods", nil, nil).Body.Bytes(), &m)
		return m
	}
	if m := methods(); m["passkeys"] != true || m["invitations"] != true || m["oidc"] != false {
		t.Fatal("hosted workspaces offer passkeys and invitations by default", m)
	}
	for _, flag := range []string{"passkeySignIn", "memberInvitations"} {
		if w := f.do("POST", "/api/events", map[string]any{"kind": "feature.flag.updated", "payload": map[string]any{"name": flag, "enabled": false}}, &owner); w.Code != 200 {
			t.Fatal(flag, w.Code, w.Body.String())
		}
	}
	if m := methods(); m["passkeys"] != false || m["invitations"] != false {
		t.Fatal("the flags did not turn the methods off", m)
	}
	if w := f.do("POST", "/api/webauthn/login/options", `{}`, nil); w.Code != 403 {
		t.Fatal("passkey sign-in worked while off", w.Code)
	}
	if w := f.do("POST", "/api/invitations", map[string]any{"role": "QA"}, &owner); w.Code != 403 {
		t.Fatal("invitations worked while off", w.Code)
	}
	if w := f.do("POST", "/api/invitations/lookup", map[string]string{"token": strings.Repeat("a", 43)}, nil); w.Code != 404 {
		t.Fatal(w.Code)
	}
	// A single-workspace installation keeps these off until its owner turns them on.
	s, _, _ := mockServer(t)
	st, _ := s.load(t.Context())
	if s.identityFlag(st, flagPasskeys) || s.identityFlag(st, flagInvitations) || s.identityFlag(st, flagOIDC) {
		t.Fatal("identity methods must default off outside hosted mode")
	}
}

func TestInvitationsRespectTheMemberLimit(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	if w := f.call("PUT", "/platform/v1/workspaces/acme/limits", `{"members":2}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	for i := 0; i < 2; i++ {
		if _, w := f.invite(owner, "QA"); w.Code != 200 {
			t.Fatal(i, w.Code, w.Body.String())
		}
	}
	if _, w := f.invite(owner, "QA"); w.Code != 409 || !strings.Contains(w.Body.String(), "limit of 2 members") {
		t.Fatal("unused invitations must count towards the member limit", w.Code, w.Body.String())
	}
}

func TestInvitationInputValidation(t *testing.T) {
	t.Parallel()
	f := newIdentityFixture(t, nil)
	owner := f.ownerSession("acme")
	for name, body := range map[string]any{
		"unknown role":     map[string]any{"role": "Owner"},
		"no role":          map[string]any{},
		"zero rooms valid": nil,
		"too long":         map[string]any{"role": "QA", "hours": 169},
		"negative":         map[string]any{"role": "QA", "hours": -1},
		"personal room":    map[string]any{"role": "QA", "roomIds": []string{"personal-local-owner"}},
		"unknown room":     map[string]any{"role": "QA", "roomIds": []string{"nope"}},
		"extra field":      map[string]any{"role": "QA", "admin": true},
	} {
		if body == nil {
			continue
		}
		if w := f.do("POST", "/api/invitations", body, &owner); w.Code != 400 {
			t.Fatal(name, w.Code, w.Body.String())
		}
	}
	if w := f.do("POST", "/api/invitations", map[string]any{"role": "QA", "roomIds": []string{"engineering"}, "hours": 1}, &owner); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	token, _ := f.invite(owner, "QA")
	f.resetThrottle("acme")
	for name, name2 := range map[string]string{"empty": "", "blank": "   ", "control": "a\nb", "long": strings.Repeat("n", 101)} {
		if w := f.do("POST", "/api/invitations/passkey/options", map[string]string{"token": token, "name": name2}, nil); w.Code != 400 {
			t.Fatal("accepted name", name, w.Code)
		}
	}
}
