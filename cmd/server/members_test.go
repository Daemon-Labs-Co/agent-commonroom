package main

import (
	"agentroom/internal/room"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"

	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPersonalRoomsAndVisibility(t *testing.T) {
	t.Parallel()
	st := room.State{Company: room.Company{ID: "company"}, Rooms: []room.Room{{ID: "dev"}, {ID: "qa"}}, Actors: []room.Actor{{ID: "one", Name: "Darrell", Role: "Developer", Kind: "human", Enabled: true, RoomIDs: []string{"dev"}}, {ID: "two", Name: "Darrell", Role: "QA", Kind: "human", Enabled: true, RoomIDs: []string{"qa"}}}, Connections: []room.Connection{{ID: "cli", Name: "Bot"}}}
	projectActors(&st)
	if !canAccess(st, "one", "personal-one") || canAccess(st, "one", "personal-two") || canAccess(st, "one", "qa") || !canAccess(st, ownerID, "personal-two") {
		t.Fatal("room authorization incorrect")
	}
	if !hasRoom(st, "personal-bot-cli") {
		t.Fatal("bot personal room missing")
	}
	visible := visibleState(st, "one", "dev")
	if len(visible.Rooms) != 2 {
		t.Fatalf("member sees %d rooms", len(visible.Rooms))
	}
	for _, a := range visible.Actors {
		if a.PublicKey != "" {
			t.Fatal("credential metadata exposed")
		}
	}
	st.Actors[0].Enabled = false
	if canAccess(st, "one", "personal-one") {
		t.Fatal("disabled account admitted")
	}
}
func TestInvitationAuthorizationWithMockDB(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newMockDB(t)
	a, _ := newOwnerAuth(strings.Repeat("a", 64), "127.0.0.1:8787")
	s := &server{db: db, auth: a, active: map[string]bool{}, cancels: map[string]context.CancelFunc{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", s.login)
	mux.HandleFunc("/api/members", s.members)
	mux.HandleFunc("/api/events", s.events)
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/session", s.session)
	h := a.protect(s.actorGuard(mux))
	a.sessions["owner"] = ownerSession{ActorID: ownerID, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	call := func(token, path string, payload any) *httptest.ResponseRecorder {
		var body string
		if payload != nil {
			raw, _ := json.Marshal(payload)
			body = string(raw)
		}
		method := "GET"
		if payload != nil {
			method = "POST"
		}
		r := httptest.NewRequest(method, "http://127.0.0.1:8787"+path, strings.NewReader(body))
		r.AddCookie(&http.Cookie{Name: "commonroom_session", Value: token})
		r.Header.Set("X-CSRF-Token", "csrf")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	key, e := rsa.GenerateKey(rand.Reader, 3072)
	if e != nil {
		t.Fatal(e)
	}
	der, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pub := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
	invited := room.Actor{Name: "Darrell", Role: "Developer", PublicKey: pub, RoomIDs: []string{"engineering"}}
	w := call("owner", "/api/members", invited)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var first room.Actor
	json.Unmarshal(w.Body.Bytes(), &first)
	secondKey, e := rsa.GenerateKey(rand.Reader, 3072)
	if e != nil {
		t.Fatal(e)
	}
	secondDER, _ := x509.MarshalPKIXPublicKey(&secondKey.PublicKey)
	invited.PublicKey = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: secondDER}))
	invited.Role = "QA"
	invited.RoomIDs = nil
	w = call("owner", "/api/members", invited)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var second room.Actor
	json.Unmarshal(w.Body.Bytes(), &second)
	if first.ID == second.ID {
		t.Fatal("duplicate names share identity")
	}
	a.challenges["invited-login"] = ownerChallenge{Value: "invitation login", Expires: time.Now().Add(time.Minute)}
	digest := sha256.Sum256([]byte("invitation login"))
	signature, e := rsa.SignPSS(rand.Reader, key, crypto.SHA256, digest[:], &rsa.PSSOptions{SaltLength: 32, Hash: crypto.SHA256})
	if e != nil {
		t.Fatal(e)
	}
	w = call("", "/api/login", map[string]string{"actorId": first.ID, "challengeId": "invited-login", "signature": base64.StdEncoding.EncodeToString(signature)})
	if w.Code != 200 {
		t.Fatal("invited key login failed", w.Body.String())
	}
	a.sessions["member"] = ownerSession{ActorID: first.ID, CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	if w = call("member", "/api/state?room=personal-"+first.ID, nil); w.Code != 200 {
		t.Fatal("own room denied", w.Body.String())
	}
	if w = call("member", "/api/state?room=personal-"+second.ID, nil); w.Code != 403 {
		t.Fatal("cross-room access allowed")
	}
	if w = call("member", "/api/members", room.Actor{ID: first.ID, Name: "Darrell", Role: "QA", Revision: 1}); w.Code != 403 {
		t.Fatal("self role change allowed")
	}
	for _, kind := range []string{"company.updated", "room.updated", "participant.updated", "human.profile.updated"} {
		if w = call("member", "/api/events", map[string]any{"kind": kind, "payload": map[string]string{"name": "attack"}}); w.Code != 403 {
			t.Fatalf("member allowed %s: %d", kind, w.Code)
		}
	}
	if w = call("member", "/api/events", map[string]any{"kind": "actor.name.updated", "payload": map[string]string{"id": second.ID, "name": "Darrell Developer"}}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	st, e := s.load(ctx)
	if e != nil {
		t.Fatal(e)
	}
	one, _ := member(st, first.ID)
	two, _ := member(st, second.ID)
	if one.Name != "Darrell Developer" || two.Name != "Darrell" {
		t.Fatal("name update impersonates other member")
	}
	first.Enabled = false
	first.Role = "Developer"
	first.Name = "Darrell Developer"
	first.Revision = 1
	w = call("owner", "/api/members", first)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w = call("member", "/api/state?room=personal-"+first.ID, nil); w.Code != 403 {
		t.Fatal("disabled session still works")
	}
	st, e = s.load(ctx)
	if e != nil || !hasRoom(st, "personal-"+first.ID) {
		t.Fatal("disabling deleted personal room")
	}
	if w = call("owner", "/api/state?room=personal-"+first.ID, nil); w.Code != 200 {
		t.Fatal("owner administration denied")
	}
	var count int
	for _, event := range db.snapshot() {
		if event.kind == "room.owner_access" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owner access not audited")
	}
}

func TestPersonalLabelRefreshPreservesRoomSettingsAndActorAuthority(t *testing.T) {
	t.Parallel()
	present := false
	actor := room.Actor{ID: "member", Name: "Current", Kind: "human", Role: "QA", Enabled: false, Revision: 7, RoomIDs: []string{"engineering"}}
	bot := room.Actor{ID: "bot-connection", Name: "Old connection", Kind: "agent", Role: "Agent", Enabled: false, Revision: 9}
	st := room.State{Company: room.Company{ID: "company"}, Actors: []room.Actor{actor, bot}, Connections: []room.Connection{{ID: "connection", Name: "Current connection"}}, Rooms: []room.Room{
		{ID: "engineering", Name: "Custom shared name", Description: "Shared settings"},
		{ID: "personal-member", PersonalActorID: "member", Name: "Earlier · Personal", Description: "Private settings", HumanPresent: &present},
	}}
	before := st.Rooms[1]
	before.Name = "Current · Personal"
	projectActors(&st)
	projectActors(&st)
	if !reflect.DeepEqual(st.Rooms[1], before) || st.Rooms[0].Name != "Custom shared name" {
		t.Fatal("refresh rewrote room settings")
	}
	actual, _ := member(st, "member")
	if !reflect.DeepEqual(actual, actor) {
		t.Fatal("refresh changed human authority")
	}
	actual, _ = member(st, "bot-connection")
	bot.Name = "Current connection"
	if !reflect.DeepEqual(actual, bot) {
		t.Fatal("connection rename changed agent authority")
	}
	if canAccess(st, "member", "personal-member") {
		t.Fatal("label refresh enabled disabled actor")
	}
}
