package main

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"strings"
	"time"

	"agentroom/internal/room"
)

const ownerID = "local-owner"

type actorContextKey struct{}

func requestActor(r *http.Request) string {
	if v, ok := r.Context().Value(actorContextKey{}).(string); ok {
		return v
	}
	// Absence of middleware identity grants no authority. In particular, a
	// handler added without the actor guard must never inherit owner privileges.
	return ""
}
func member(st room.State, id string) (room.Actor, bool) {
	for _, a := range st.Actors {
		if a.ID == id {
			return a, true
		}
	}
	return room.Actor{}, false
}
func canAccess(st room.State, actor, rid string) bool {
	a, ok := member(st, actor)
	if !ok || !a.Enabled {
		return false
	}
	if actor == ownerID || rid == "personal-"+actor {
		return true
	}
	for _, id := range a.RoomIDs {
		if id == rid {
			return true
		}
	}
	return false
}
func (s *server) actorGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api/login" || r.URL.Path == "/api/challenge" {
			next.ServeHTTP(w, r)
			return
		}
		session, ok := s.auth.session(r)
		if !ok {
			http.Error(w, "Sign in required", 401)
			return
		}
		actor := session.ActorID
		if actor == "" {
			http.Error(w, "Session identity missing; sign in again", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/api/agent-safety" && actor == ownerID {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), actorContextKey{}, actor)))
			return
		}
		st, e := s.load(r.Context())
		if e != nil {
			http.Error(w, "Account unavailable", 503)
			return
		}
		a, ok := member(st, actor)
		if !ok || !a.Enabled {
			http.Error(w, "Account disabled", 403)
			return
		}
		if a.Kind != "human" {
			http.Error(w, "Human login required; agent API credentials are unsupported", 403)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), actorContextKey{}, actor))
		if r.Method != "GET" && r.Method != "HEAD" && !s.admitActorMutation(actor, time.Now()) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Too many state-changing requests; wait before retrying", 429)
			return
		}
		path := r.URL.Path
		if actor != ownerID {
			if path == "/api/members" || path == "/api/owner-key/download" || path == "/api/integrity" || path == "/api/events" && r.Method == "GET" {
				http.Error(w, "Owner permission required", 403)
				return
			}
			if path == "/api/task-dispatch" || path == "/api/agents" || path == "/api/provider-auth" {
				http.Error(w, "Agent control currently requires the owner", 403)
				return
			}
		}
		if path == "/api/task-dispatch" || path == "/api/state" || path == "/api/events" || path == "/api/agents" {
			if !canAccess(st, actor, requestedRoom(r)) {
				http.Error(w, "Room access denied", 403)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func visibleState(st room.State, actor, rid string) room.State {
	st = scoped(st, rid)
	st.CurrentActorID = actor
	rooms := []room.Room{}
	for _, r := range st.Rooms {
		if canAccess(st, actor, r.ID) {
			rooms = append(rooms, r)
		}
	}
	st.Rooms = rooms
	actors := []room.Actor{}
	for _, a := range st.Actors {
		if actor == ownerID || a.ID == actor || canAccess(st, a.ID, rid) {
			a.PublicKey = ""
			actors = append(actors, a)
		}
	}
	st.Actors = actors
	if actor != ownerID {
		st.Organization = room.OrganizationChart{}
		st.WorkflowRuns = nil
		cs := []room.Connection{}
		for _, c := range st.Connections {
			for _, p := range st.Participants {
				if p.ConnectionID == c.ID {
					c.Endpoint = ""
					cs = append(cs, c)
					break
				}
			}
		}
		st.Connections = cs
	}
	return st
}
func projectActors(st *room.State) {
	if _, ok := member(*st, ownerID); !ok {
		st.Actors = append(st.Actors, room.Actor{ID: ownerID, Name: st.HumanName, Title: st.HumanTitle, Kind: "human", Role: "Owner", Enabled: true, Revision: 1})
	}
	for i := range st.Actors {
		if st.Actors[i].ID == ownerID {
			st.Actors[i].Name = st.HumanName
			st.Actors[i].Title = st.HumanTitle
			if st.Actors[i].Name == "" {
				st.Actors[i].Name = "Mark"
			}
		}
	}
	for _, c := range st.Connections {
		bid := "bot-" + c.ID
		if _, ok := member(*st, bid); !ok {
			st.Actors = append(st.Actors, room.Actor{ID: bid, Name: c.Name, Kind: "agent", Role: "Agent", Enabled: true, Revision: 1})
		} else {
			// Connection actors are derived; early projection must not freeze
			// their display names before a later connection rename is replayed.
			for i := range st.Actors {
				if st.Actors[i].ID == bid && st.Actors[i].Kind == "agent" {
					st.Actors[i].Name = c.Name
				}
			}
		}
	}
	for i := range st.Participants {
		st.Participants[i].ActorID = "bot-" + st.Participants[i].ConnectionID
	}
	for _, a := range st.Actors {
		rid := "personal-" + a.ID
		found := false
		for i, r := range st.Rooms {
			if r.ID == rid {
				found = true
				// Refresh only derived personal labels. Identity, access and
				// all other saved room fields remain unchanged.
				if r.PersonalActorID == a.ID {
					st.Rooms[i].Name = a.Name + " · Personal"
				}
				break
			}
		}
		if !found {
			present := a.Kind == "human"
			st.Rooms = append(st.Rooms, room.Room{ID: rid, Name: a.Name + " · Personal", CompanyID: st.Company.ID, PersonalActorID: a.ID, HumanPresent: &present})
		}
	}
}
func (s *server) members(w http.ResponseWriter, r *http.Request) {
	if requestActor(r) != ownerID {
		http.Error(w, "Owner permission required", 403)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.validCSRF(r) {
		http.Error(w, "Invalid request token", 403)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, e := s.load(r.Context())
	if e != nil {
		http.Error(w, "State unavailable", 503)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16384)
	var v room.Actor
	if json.NewDecoder(r.Body).Decode(&v) != nil || strings.TrimSpace(v.Name) == "" || len(v.Name) > 100 {
		http.Error(w, "Name required (maximum 100 characters)", 400)
		return
	}
	if v.Role != "Developer" && v.Role != "QA" {
		http.Error(w, "Choose Developer or QA", 400)
		return
	}
	for _, rid := range v.RoomIDs {
		found := false
		for _, room := range st.Rooms {
			if room.ID == rid && room.PersonalActorID == "" {
				found = true
			}
		}
		if !found {
			http.Error(w, "Only shared rooms can be assigned", 400)
			return
		}
	}
	if v.ID == "" {
		block, _ := pem.Decode([]byte(v.PublicKey))
		if block == nil {
			http.Error(w, "Public key required", 400)
			return
		}
		parsed, e := x509.ParsePKIXPublicKey(block.Bytes)
		key, ok := parsed.(*rsa.PublicKey)
		if e != nil || !ok || key.N.BitLen() < 3072 || key.N.BitLen() > 4096 || key.E != 65537 {
			http.Error(w, "RSA public key must be 3072–4096 bits", 400)
			return
		}
		if s.auth.publicKey != nil && key.N.Cmp(s.auth.publicKey.N) == 0 {
			http.Error(w, "Each account requires its own key", 400)
			return
		}
		for _, other := range st.Actors {
			block, _ := pem.Decode([]byte(other.PublicKey))
			if block == nil {
				continue
			}
			parsed, e := x509.ParsePKIXPublicKey(block.Bytes)
			if old, ok := parsed.(*rsa.PublicKey); e == nil && ok && old.N.Cmp(key.N) == 0 {
				http.Error(w, "Each account requires its own key", 400)
				return
			}
		}
		v.Image = ""
		v.ID = id()
		v.Kind = "human"
		v.Enabled = true
		v.Revision = 1
	} else {
		old, ok := member(st, v.ID)
		if !ok || old.ID == ownerID || old.Kind != "human" {
			http.Error(w, "Invited human not found", 404)
			return
		}
		if old.Revision != v.Revision {
			http.Error(w, "Member changed; reload", 409)
			return
		}
		v.Image = old.Image
		v.Kind = old.Kind
		v.PublicKey = old.PublicKey
		v.Revision++
	}
	v.Name = strings.TrimSpace(v.Name)
	v.Title = ""
	v.ChangedBy = ownerID
	if e = s.append(r.Context(), "actor.revised", v); e != nil {
		http.Error(w, "Invitation could not be saved", 500)
		return
	}
	v.PublicKey = ""
	reply(w, v)
}
func (s *server) verifyMemberLogin(ctx context.Context, actor, cid, sig string) bool {
	// Consume before database access so one invalid member attempt cannot reuse
	// a known challenge to force repeated full-state replays. Caller holds auth.mu.
	challenge, ok := s.auth.challenges[cid]
	delete(s.auth.challenges, cid)
	if !ok || !time.Now().Before(challenge.Expires) {
		return false
	}
	raw, e := base64.StdEncoding.DecodeString(sig)
	if e != nil || (len(raw) < 384 || len(raw) > 512) {
		return false
	}
	st, e := s.load(ctx)
	if e != nil {
		return false
	}
	a, ok := member(st, actor)
	if !ok || !a.Enabled || a.Kind != "human" || actor == ownerID {
		return false
	}
	block, _ := pem.Decode([]byte(a.PublicKey))
	if block == nil {
		return false
	}
	parsed, e := x509.ParsePKIXPublicKey(block.Bytes)
	key, ok := parsed.(*rsa.PublicKey)
	if e != nil || !ok {
		return false
	}
	digest := sha256.Sum256([]byte(challenge.Value))
	return rsa.VerifyPSS(key, crypto.SHA256, digest[:], raw, &rsa.PSSOptions{SaltLength: 32, Hash: crypto.SHA256}) == nil
}
