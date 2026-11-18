package main

// The cross-workspace attack suite (ADR 0004 stage 6). It builds two hosted workspaces that
// each hold recognisable data, then attacks every route one workspace serves with the other
// workspace's credentials and identifiers. Nothing here proves the database policy (that is
// scripts/check-workspace-isolation.sql); it proves the application layer, by construction and
// by exhaustion, never lets a request cross a workspace boundary.

import (
	"agentroom/internal/webauthn/webauthntest"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type isolationWorld struct {
	*idFixture
	owner   map[string]idSession
	member  map[string]idSession
	canary  map[string][]string
	roomID  map[string]string
	invite  map[string]string
	cred    map[string]string
	memberN map[string]string
}

func buildIsolationWorld(t *testing.T) *isolationWorld {
	t.Helper()
	f := newIdentityFixture(t, nil)
	f.create("beta")
	w := &isolationWorld{idFixture: f, owner: map[string]idSession{}, member: map[string]idSession{}, canary: map[string][]string{}, roomID: map[string]string{}, invite: map[string]string{}, cred: map[string]string{}, memberN: map[string]string{}}
	for _, ws := range []string{"acme", "beta"} {
		f.ws = ws
		owner := f.ownerSession(ws)
		w.owner[ws] = owner
		secret := "CANARY-" + strings.ToUpper(ws) + "-" + oidcRandom()
		w.canary[ws] = append(w.canary[ws], secret)
		// A room, a message and shared memory, each carrying the workspace's secret.
		resp := f.request(ws, "POST", "/api/events", map[string]any{"kind": "room.created", "payload": map[string]any{"name": "Room " + secret}}, &owner, nil)
		if resp.Code != 200 {
			t.Fatal("room", ws, resp.Code, resp.Body.String())
		}
		for _, r := range f.state(ws).Rooms {
			if strings.Contains(r.Name, secret) {
				w.roomID[ws] = r.ID
			}
		}
		if w.roomID[ws] == "" {
			t.Fatal("room not found for", ws)
		}
		if resp := f.request(ws, "POST", "/api/events?room="+w.roomID[ws], map[string]any{"kind": "message.created", "payload": map[string]any{"text": "message " + secret}}, &owner, nil); resp.Code != 200 {
			t.Fatal("message", ws, resp.Code, resp.Body.String())
		}
		// A member with a passkey, and an unused invitation.
		token, _ := f.invite(owner, "Developer")
		device := webauthntest.New()
		name := "Member " + secret
		redeem := f.request(ws, "POST", "/api/invitations/passkey/options", map[string]string{"token": token, "name": name}, nil, nil)
		opt := decodeOptions(t, redeem)
		_, att, cd := device.Create(webauthntest.Decode(opt.PublicKey.Challenge), ws+".rooms.example.com", "https://"+ws+".rooms.example.com")
		done := f.request(ws, "POST", "/api/invitations/passkey/complete", map[string]string{"challengeId": opt.ChallengeID, "attestationObject": att, "clientDataJSON": cd}, nil, nil)
		if done.Code != 200 {
			t.Fatal("member", ws, done.Code, done.Body.String())
		}
		w.member[ws] = sessionFrom(t, done)
		w.cred[ws] = device.CredentialID()
		f.ws = ws
		if _, resp := f.invite(owner, "QA"); resp.Code != 200 {
			t.Fatal(resp.Code)
		}
		for _, inv := range f.state(ws).Invitations {
			if inv.Status == invitationOpen {
				w.invite[ws] = inv.ID
			}
		}
		for _, a := range f.state(ws).Actors {
			if strings.Contains(a.Name, secret) {
				w.memberN[ws] = a.ID
			}
		}
		f.resetThrottle(ws)
	}
	f.ws = "acme"
	return w
}

func oidcRandom() string { return strings.ToUpper(b64url(randomChallenge()[:8])) }

// publicRoutes answer without a session. They may answer on any workspace, but never with
// another workspace's data.
var publicRoutes = map[string]bool{
	"/api/login": true, "/api/challenge": true, "/api/setup-status": true, "/api/auth/methods": true,
	"/api/webauthn/login/options": true, "/api/webauthn/login/verify": true,
	"/api/invitations/lookup": true, "/api/invitations/passkey/options": true, "/api/invitations/passkey/complete": true,
	"/api/invitations/oidc/start": true, "/api/oidc/login": true, "/api/oidc/callback": true,
	"/instance-label.js": true, "/": true, "/help/": true,
}

func TestEveryRouteRefusesAnotherWorkspacesCredentials(t *testing.T) {
	t.Parallel()
	w := buildIsolationWorld(t)
	routes := w.runtime("beta").routes
	if len(routes) < 60 {
		t.Fatal("the route list looks incomplete", len(routes))
	}
	secretsOfAcme := w.canary["acme"]
	leaks := func(body string) bool {
		for _, secret := range secretsOfAcme {
			if strings.Contains(body, secret) {
				return true
			}
		}
		return false // identifiers are chosen by the attacker here; only the secret content is a leak
	}
	// Control: the secrets really are visible to their own workspace, so the checks below could notice a leak.
	ownerAcme := w.owner["acme"]
	own := w.request("acme", "GET", "/api/state?room="+w.roomID["acme"], nil, &ownerAcme, nil)
	if own.Code != 200 || !leaks(own.Body.String()) {
		t.Fatal("the control request cannot see its own canaries", own.Code)
	}
	credentials := map[string]idSession{"acme owner": w.owner["acme"], "acme member": w.member["acme"]}
	attacks := 0
	for _, route := range routes {
		target := strings.TrimSuffix(route, "/")
		if strings.HasSuffix(route, "/") && route != "/" {
			target = route + "x"
		}
		for name, session := range credentials {
			for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
				for _, cookieName := range []string{hostedSessionCookie, "commonroom_session"} {
					sess := session
					resp := w.request("beta", method, target+"?room="+w.roomID["acme"], `{"kind":"message.created","payload":{"text":"x"},"id":"`+w.memberN["acme"]+`","credentialId":"`+w.cred["acme"]+`"}`, nil, func(r *http.Request) {
						r.AddCookie(&http.Cookie{Name: cookieName, Value: sess.cookie})
						r.Header.Set("X-CSRF-Token", sess.csrf)
					})
					attacks++
					if leaks(resp.Body.String()) {
						t.Fatal("LEAK: another workspace's data appeared", name, method, route, resp.Code)
					}
					if resp.Code >= 200 && resp.Code < 300 && !publicRoutes[route] && !strings.HasSuffix(route, ".js") {
						t.Fatal("another workspace's credentials were accepted:", name, method, route, cookieName, resp.Code, resp.Body.String())
					}
				}
			}
		}
	}
	if attacks < 600 {
		t.Fatal("too few attacks ran", attacks)
	}
	// And the other direction, through the same routes.
	for _, route := range w.runtime("acme").routes {
		if publicRoutes[route] || strings.HasSuffix(route, ".js") || strings.HasSuffix(route, "/") {
			continue
		}
		for _, method := range []string{"GET", "POST"} {
			sess := w.owner["beta"]
			resp := w.request("acme", method, route, `{}`, nil, func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: hostedSessionCookie, Value: sess.cookie})
				r.Header.Set("X-CSRF-Token", sess.csrf)
			})
			if resp.Code >= 200 && resp.Code < 300 {
				t.Fatal("beta's session was accepted by acme", method, route, resp.Code)
			}
			for _, secret := range w.canary["beta"] {
				if strings.Contains(resp.Body.String(), secret) {
					t.Fatal("LEAK from beta into acme", route)
				}
			}
		}
	}
}

func TestAnotherWorkspacesIdentifiersAreMeaninglessToAValidOwner(t *testing.T) {
	t.Parallel()
	w := buildIsolationWorld(t)
	beta := w.owner["beta"]
	before := len(w.dbs["beta"].snapshot())
	acmeRoom, acmeInvite, acmeCred, acmeMember := w.roomID["acme"], w.invite["acme"], w.cred["acme"], w.memberN["acme"]
	check := func(name string, resp *httptest.ResponseRecorder, wantRefusal bool) {
		t.Helper()
		for _, secret := range w.canary["acme"] {
			if strings.Contains(resp.Body.String(), secret) {
				t.Fatal("LEAK", name)
			}
		}
		if wantRefusal && resp.Code < 400 {
			t.Fatal("accepted another workspace's identifier:", name, resp.Code, resp.Body.String())
		}
	}
	do := func(method, path string, body any) *httptest.ResponseRecorder {
		return w.request("beta", method, path, body, &beta, nil)
	}
	// An unknown room falls back to the requester's own view; what matters is that none of acme's content is in it.
	view := do("GET", "/api/state?room="+acmeRoom, nil)
	check("read acme's room", view, false)
	if strings.Contains(view.Body.String(), "message CANARY-ACME") || strings.Contains(view.Body.String(), "Room CANARY-ACME") {
		t.Fatal("acme's room was readable")
	}
	check("write into acme's room", do("POST", "/api/events?room="+acmeRoom, map[string]any{"kind": "message.created", "payload": map[string]any{"text": "intrusion"}}), true)
	check("edit acme's room", do("POST", "/api/events", map[string]any{"kind": "room.updated", "payload": map[string]any{"id": acmeRoom, "name": "hijacked", "revision": 1}}), true)
	check("revoke acme's invitation", do("POST", "/api/invitations/revoke", map[string]string{"id": acmeInvite}), true)
	check("remove acme's passkey", do("POST", "/api/webauthn/remove", map[string]string{"credentialId": acmeCred}), true)
	check("edit acme's member", do("POST", "/api/members", map[string]any{"id": acmeMember, "name": "Taken", "role": "QA", "roomIds": []string{}, "enabled": true, "revision": 1}), true)
	check("agents in acme's room", do("POST", "/api/agents?room="+acmeRoom, map[string]any{"action": "broadcast", "text": "hi"}), true)
	check("tasks in acme's room", do("GET", "/api/task-progress?room="+acmeRoom, nil), false)
	check("workflow runs in acme's room", do("GET", "/api/workflow-runs?room="+acmeRoom, nil), false)
	check("integrity of beta only", do("GET", "/api/integrity", nil), false)
	// Beta's own view never includes acme's data, however it asks.
	for _, path := range []string{"/api/state", "/api/state?room=" + w.roomID["beta"], "/api/integrity", "/api/events"} {
		resp := do("GET", path, nil)
		check("beta's own "+path, resp, false)
	}
	if after := len(w.dbs["beta"].snapshot()); after != before {
		t.Fatal("refused requests still wrote to beta's history", before, after)
	}
	// Histories are physically separate: no workspace's events mention another's secrets.
	for ws, other := range map[string]string{"acme": "beta", "beta": "acme"} {
		for _, e := range w.dbs[ws].snapshot() {
			for _, secret := range w.canary[other] {
				if strings.Contains(e.payload, secret) {
					t.Fatal("an event in", ws, "mentions", other)
				}
			}
		}
	}
}

func TestTheSameIdentityCanBelongToTwoWorkspacesIndependently(t *testing.T) {
	t.Parallel()
	f := newSSOFixture(t)
	f.create("beta")
	ownerA := f.ownerSession("acme")
	ownerB := f.ownerSession("beta")
	f.idp.Subject = "shared-person"
	f.ws = "acme"
	token, _ := f.invite(ownerA, "Developer")
	authURL, binding := f.start("acme", "/api/invitations/oidc/start?token="+token+"&name=Sam", nil)
	if w := f.callback("acme", f.approve(authURL), binding, nil); sessionCookie(w) == "" {
		t.Fatal("could not join acme")
	}
	// Linked in acme only: beta has never heard of this identity.
	f.resetThrottle("beta")
	authURL, binding = f.start("beta", "/api/oidc/login", nil)
	if w := f.callback("beta", f.approve(authURL), binding, nil); sessionCookie(w) != "" || w.Header().Get("Location") != "/#signin=unlinked" {
		t.Fatal("an identity linked in one workspace signed in to another", w.Header().Get("Location"))
	}
	// And it can join beta separately, as a different account.
	f.ws = "beta"
	token, _ = f.invite(ownerB, "QA")
	authURL, binding = f.start("beta", "/api/invitations/oidc/start?token="+token+"&name=Sam", nil)
	w := f.callback("beta", f.approve(authURL), binding, nil)
	if sessionCookie(w) == "" {
		t.Fatal("the same person could not join a second workspace", w.Header().Get("Location"))
	}
	a, b := "", ""
	for _, actor := range f.state("acme").Actors {
		if actor.Name == "Sam" {
			a = actor.ID
		}
	}
	for _, actor := range f.state("beta").Actors {
		if actor.Name == "Sam" {
			b = actor.ID
		}
	}
	if a == "" || b == "" || a == b {
		t.Fatal("expected two independent accounts", a, b)
	}
	// A session from one is no session in the other.
	sess := idSession{cookie: sessionCookie(w)}
	if resp := f.request("acme", "GET", "/api/session", nil, &sess, nil); resp.Code == 200 {
		t.Fatal("a session crossed workspaces")
	}
}

func TestEveryWorkspaceRouteIsAccountedFor(t *testing.T) {
	t.Parallel()
	// A new route must be added knowingly: it is either public (listed) or behind a session. This
	// test fails when a route is added that answers an anonymous request with data, so the author
	// has to decide, and add it to publicRoutes on purpose.
	f := newIdentityFixture(t, nil)
	for _, route := range f.runtime("acme").routes {
		if publicRoutes[route] || strings.HasSuffix(route, ".js") || route == "/" || strings.HasSuffix(route, "/") {
			continue
		}
		for _, method := range []string{"GET", "POST"} {
			resp := f.request("acme", method, route, `{}`, nil, nil)
			if resp.Code >= 200 && resp.Code < 300 {
				t.Fatal("an anonymous request was answered with success:", method, route, resp.Code)
			}
		}
	}
	for _, route := range f.runtime("acme").routes {
		if publicRoutes[route] && !publicAuthPath(route) && route != "/instance-label.js" && route != "/" && route != "/help/" {
			t.Fatal("the test lists a public route that the server does not treat as public:", route)
		}
	}
	for path := range map[string]bool{"/api/login": true, "/api/webauthn/login/verify": true} {
		if !publicAuthPath(path) {
			t.Fatal(path)
		}
	}
	var seen = map[string]bool{}
	for _, route := range f.runtime("acme").routes {
		if seen[route] {
			t.Fatal("route registered twice:", route)
		}
		seen[route] = true
	}
}
