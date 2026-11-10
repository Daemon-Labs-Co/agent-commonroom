package main

import (
	"agentroom/internal/room"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type extensionClient struct {
	id   string
	key  ed25519.PrivateKey
	h    http.Handler
	skew time.Duration
}

func (c extensionClient) do(method, target, body string) *httptest.ResponseRecorder {
	stamp := strconv.FormatInt(time.Now().Add(c.skew).Unix(), 10)
	r := httptest.NewRequest(method, "http://127.0.0.1:8787"+target, strings.NewReader(body))
	r.Header.Set("X-Extension-Id", c.id)
	r.Header.Set("X-Extension-Timestamp", stamp)
	r.Header.Set("X-Extension-Signature", base64.StdEncoding.EncodeToString(ed25519.Sign(c.key, extensionSigningInput(c.id, stamp, method, target, []byte(body)))))
	w := httptest.NewRecorder()
	c.h.ServeHTTP(w, r)
	return w
}

func extensionFixture(t *testing.T, enabled bool) (*server, *mockDB, extensionClient, string) {
	t.Helper()
	s, db, _ := mockServer(t)
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "extensions.json")
	registry, _ := json.Marshal([]map[string]any{{"id": "planner", "name": "Planner", "publicKey": base64.StdEncoding.EncodeToString(public), "rooms": []string{"engineering"}}})
	if err = os.WriteFile(path, registry, 0o600); err != nil {
		t.Fatal(err)
	}
	s.extensionRegistry = path
	if enabled {
		db.seed("feature.flag.updated", map[string]any{"name": "extensions", "enabled": true})
	}
	db.seed("message.created", room.Message{ID: "h1", RoomID: "engineering", ActorID: ownerID, ActorType: "human", Author: "Owner", Text: "We need login"})
	mux := http.NewServeMux()
	mux.HandleFunc("/ext/v1/", s.extensionAPI)
	mux.HandleFunc("/api/events", s.events)
	return s, db, extensionClient{id: "planner", key: private, h: s.auth.protect(s.actorGuard(mux))}, path
}

func TestExtensionReadsGrantedRoomAndPostsAttributedNotice(t *testing.T) {
	t.Parallel()
	s, db, client, _ := extensionFixture(t, true)
	db.seed("feature.flag.updated", map[string]any{"name": "governingSuggestions", "enabled": true})
	w := client.do("GET", "/ext/v1/rooms/engineering/messages", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "We need login") || strings.Contains(w.Body.String(), "actorId") {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = client.do("GET", "/ext/v1/rooms/engineering/messages?after=h1", ""); w.Code != 200 || strings.Contains(w.Body.String(), "We need login") {
		t.Fatal("after cursor ignored", w.Body.String())
	}
	w = client.do("POST", "/ext/v1/rooms/engineering/notices", `{"text":"Plan ready.\n[PROPOSE TASK] Add login :: Users can sign in"}`)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	last := st.Messages[len(st.Messages)-1]
	if last.ActorType != "extension" || last.ActorID != "extension:planner" || last.Author != "Planner" {
		t.Fatalf("notice not attributed to the extension: %+v", last)
	}
	if len(st.TaskSuggestions) != 1 || st.TaskSuggestions[0].Author != "Planner" || len(st.TaskProposals) != 0 || len(st.Tasks) != 0 {
		t.Fatal("an extension notice must yield a suggestion and create nothing else")
	}
	if len(st.Objections) != 0 || last.Stance != "" || last.HumanRequest != "" {
		t.Fatal("extension text was treated as an agent marker")
	}
}

func TestExtensionRefusals(t *testing.T) {
	t.Parallel()
	s, db, client, path := extensionFixture(t, true)
	before := len(db.snapshot())
	db.seed("room.created", room.Room{ID: "private", Name: "Private"})
	cases := []struct {
		name string
		run  func() *httptest.ResponseRecorder
		want int
	}{
		{"ungranted room", func() *httptest.ResponseRecorder { return client.do("GET", "/ext/v1/rooms/private/messages", "") }, 403},
		{"unknown room", func() *httptest.ResponseRecorder { return client.do("GET", "/ext/v1/rooms/nowhere/messages", "") }, 403},
		{"no other resource", func() *httptest.ResponseRecorder { return client.do("POST", "/ext/v1/rooms/engineering/tasks", `{}`) }, 404},
		{"empty notice", func() *httptest.ResponseRecorder {
			return client.do("POST", "/ext/v1/rooms/engineering/notices", `{"text":" "}`)
		}, 400},
		{"unknown field", func() *httptest.ResponseRecorder {
			return client.do("POST", "/ext/v1/rooms/engineering/notices", `{"text":"x","actorType":"human"}`)
		}, 400},
		{"stale timestamp", func() *httptest.ResponseRecorder {
			old := client
			old.skew = -10 * time.Minute
			return old.do("GET", "/ext/v1/rooms/engineering/messages", "")
		}, 403},
		{"wrong key", func() *httptest.ResponseRecorder {
			_, other, _ := ed25519.GenerateKey(rand.Reader)
			forged := client
			forged.key = other
			return forged.do("GET", "/ext/v1/rooms/engineering/messages", "")
		}, 403},
		{"unknown extension", func() *httptest.ResponseRecorder {
			stranger := client
			stranger.id = "stranger"
			return stranger.do("GET", "/ext/v1/rooms/engineering/messages", "")
		}, 403},
	}
	for _, c := range cases {
		if w := c.run(); w.Code != c.want {
			t.Fatalf("%s: got %d want %d (%s)", c.name, w.Code, c.want, w.Body.String())
		}
	}
	// A signature is bound to its body and cannot be replayed.
	stamp := strconv.FormatInt(time.Now().Unix(), 10)
	target := "/ext/v1/rooms/engineering/notices"
	signature := base64.StdEncoding.EncodeToString(ed25519.Sign(client.key, extensionSigningInput("planner", stamp, "POST", target, []byte(`{"text":"signed"}`))))
	send := func(body string) int {
		r := httptest.NewRequest("POST", "http://127.0.0.1:8787"+target, strings.NewReader(body))
		r.Header.Set("X-Extension-Id", "planner")
		r.Header.Set("X-Extension-Timestamp", stamp)
		r.Header.Set("X-Extension-Signature", signature)
		w := httptest.NewRecorder()
		client.h.ServeHTTP(w, r)
		return w.Code
	}
	if send(`{"text":"tampered"}`) != 403 || send(`{"text":"signed"}`) != 201 || send(`{"text":"signed"}`) != 429 {
		t.Fatal("signature not bound to the body, or a replay was accepted")
	}
	// The extension has no session: human endpoints stay closed to it.
	if w := client.do("POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"extensions","enabled":true}}`); w.Code != 401 {
		t.Fatal("extension reached a human endpoint", w.Code)
	}
	// Removing the registry entry removes access immediately.
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if w := client.do("GET", "/ext/v1/rooms/engineering/messages", ""); w.Code != 403 {
		t.Fatal("removed extension still has access", w.Code)
	}
	if got := len(db.snapshot()); got != before+1+2 {
		t.Fatalf("refused requests wrote history: %d events beyond the one accepted notice", got-before-3)
	}
	_ = s
}

func TestExtensionsDisabledAndRegistrySafety(t *testing.T) {
	t.Parallel()
	_, db, client, path := extensionFixture(t, false)
	if w := client.do("GET", "/ext/v1/rooms/engineering/messages", ""); w.Code != 403 || !strings.Contains(w.Body.String(), "disabled") {
		t.Fatal("disabled extensions answered", w.Code)
	}
	db.seed("feature.flag.updated", map[string]any{"name": "extensions", "enabled": true})
	if err := os.Chmod(path, 0o666); err != nil {
		t.Fatal(err)
	}
	if w := client.do("GET", "/ext/v1/rooms/engineering/messages", ""); w.Code != 403 {
		t.Fatal("a registry writable by others was trusted", w.Code)
	}
	if _, err := loadExtensionRegistry(""); err == nil {
		t.Fatal("missing registry path accepted")
	}
	bad := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(bad, []byte(`[{"id":"Bad Id","name":"x","publicKey":"AA==","rooms":[]}]`), 0o600)
	if _, err := loadExtensionRegistry(bad); err == nil {
		t.Fatal("invalid entry accepted")
	}
	rounds := db
	_ = rounds
	limiter := extensionLimiter{}
	now := time.Now()
	for i := 0; i < extensionRequestsPerMin; i++ {
		if err := limiter.admit("planner", strconv.Itoa(i), now); err != nil {
			t.Fatal("limit reached early", i)
		}
	}
	if limiter.admit("planner", "over", now) == nil || limiter.admit("other", "fresh", now) != nil {
		t.Fatal("per-extension rate limit not enforced or not isolated")
	}
	if limiter.admit("planner", "later", now.Add(61*time.Second)) != nil {
		t.Fatal("rate window did not move on")
	}
}
