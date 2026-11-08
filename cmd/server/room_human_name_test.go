package main

import (
	"fmt"
	"net/http"
	"testing"
)

// saveRoomHumanName posts room.updated with the room's current revision, as the
// room profile form does, and returns the stored room name after the save.
func saveRoomHumanName(t *testing.T, s *server, h http.Handler, fields string) string {
	t.Helper()
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`{"kind":"room.updated","payload":{"id":"engineering","name":"Engineering","revision":%d,%s}}`, st.Rooms[0].Revision, fields)
	if w := mockRequest(h, "POST", "/api/events", body); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, err = s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return st.Rooms[0].HumanName
}

func TestRoomHumanNameFollowsProfileUnlessOverridden(t *testing.T) {
	t.Parallel()
	s, db, h := mockServer(t)
	db.seed("human.profile.updated", map[string]any{"name": "Original", "title": ""})
	if got := saveRoomHumanName(t, s, h, `"humanName":"Room Name"`); got != "Room Name" {
		t.Fatal("custom room name not stored", got)
	}
	if got := saveRoomHumanName(t, s, h, `"humanName":""`); got != "Room Name" {
		t.Fatal("empty name must keep the existing room name", got)
	}
	if got := saveRoomHumanName(t, s, h, `"humanName":"Room Name","useProfileName":true`); got != "" {
		t.Fatal("useProfileName must clear the room-specific name", got)
	}
	db.seed("human.profile.updated", map[string]any{"name": "Renamed", "title": "", "revision": 1})
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	owner, _ := member(st, ownerID)
	if owner.Name != "Renamed" || st.Rooms[0].HumanName != "" {
		t.Fatalf("room must follow the renamed profile: actor=%q room=%q", owner.Name, st.Rooms[0].HumanName)
	}
}
