package main

import (
	"agentroom/internal/room"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func samplePhoto(t *testing.T, w, h int) string {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b.Bytes())
}
func TestProfilePhotoValidation(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"https://example.com/image.jpg", "data:image/svg+xml;base64,PHN2Zz4=", "data:image/jpeg;base64,broken", samplePhoto(t, 257, 1), "data:image/png;base64," + strings.Repeat("A", 44000)} {
		if _, err := normalizeProfileImage(value); err == nil {
			t.Fatal("unsafe image accepted")
		}
	}
	got, err := normalizeProfileImage(samplePhoto(t, 160, 160))
	if err != nil || !strings.HasPrefix(got, "data:image/jpeg;base64,") {
		t.Fatal("valid image normalization failed", err)
	}
	if got, err := normalizeProfileImage(""); got != "" || err != nil {
		t.Fatal("removal failed")
	}
}
func TestHumanPhotoSelfOnlyAndFlag(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("actor.revised", room.Actor{ID: "dev", Name: "Developer", Kind: "human", Role: "Developer", Enabled: true, RoomIDs: []string{"engineering"}})
	s.auth.sessions["owner"] = ownerSession{ActorID: "dev", CSRF: "csrf", Expires: time.Now().Add(time.Hour)}
	body, _ := json.Marshal(map[string]any{"kind": "actor.image.updated", "payload": map[string]string{"id": ownerID, "image": samplePhoto(t, 32, 32), "role": "Owner"}})
	w := mockRequest(h, "POST", "/api/events", string(body))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err := s.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	dev, _ := member(st, "dev")
	owner, _ := member(st, ownerID)
	if dev.Image == "" || owner.Image != "" || dev.Role != "Developer" {
		t.Fatal("photo changed another identity or role")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "humanProfileImages", "enabled": false})
	if w = mockRequest(h, "POST", "/api/events", string(body)); w.Code != 403 {
		t.Fatal("disabled upload accepted", w.Code)
	}
	if w = mockRequest(h, "POST", "/api/events", `{"kind":"actor.image.updated","payload":{"image":""}}`); w.Code != 200 {
		t.Fatal("disabled removal denied", w.Code)
	}
	if w = mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"humanProfileImages","enabled":true}}`); w.Code != 403 {
		t.Fatal("member changed flag")
	}
}
func TestOwnerFlagsPersistWithoutResumingAgents(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	s.safetyUIDefault = true
	db.seed("agents.stop", map[string]string{"scope": "application"})
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "agent-safety.js"), []byte("/* safety */"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []bool{false, true} {
		body, _ := json.Marshal(map[string]any{"kind": "feature.flag.updated", "payload": map[string]any{"name": "agentSafetyControls", "enabled": enabled}})
		if w := mockRequest(h, "POST", "/api/events", string(body)); w.Code != 200 {
			t.Fatal(w.Body.String())
		}
		st, err := s.load(context.Background())
		if err != nil || !st.AgentsStopped || s.agentSafetyControlsEnabled(st) != enabled {
			t.Fatal("flag lost or cleared stop")
		}
		fresh := server{db: db, safetyUIDefault: !enabled}
		w := httptest.NewRecorder()
		fresh.agentSafetyScript(root).ServeHTTP(w, httptest.NewRequest("GET", "/agent-safety.js", nil))
		if enabled && w.Code != 200 || !enabled && w.Code != 204 {
			t.Fatal("stored flag did not gate script", w.Code)
		}
	}
	for _, payload := range []string{`{"name":"unknown","enabled":false}`, `{"name":"humanProfileImages"}`} {
		if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":`+payload+`}`); w.Code != 400 {
			t.Fatal("invalid flag accepted")
		}
	}
	if w := mockRequest(h, "POST", "/api/events", `{"kind":"feature.flag.updated","payload":{"name":"humanProfileImages","enabled":false}}`); w.Code != 200 {
		t.Fatal("owner photo flag failed")
	}
	st, _ := s.load(context.Background())
	if humanProfileImagesEnabled(st) {
		t.Fatal("photo flag not replayed")
	}
}
