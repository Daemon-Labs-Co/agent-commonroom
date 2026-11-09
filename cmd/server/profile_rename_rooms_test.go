package main

import (
	"agentroom/internal/room"
	"testing"
)

func roomHumanNames(t *testing.T, s *server) map[string]string {
	t.Helper()
	st, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, r := range st.Rooms {
		out[r.ID] = r.HumanName
	}
	return out
}

func TestProfileRenameReleasesEveryRoomName(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("room.created", room.Room{ID: "a", Name: "A", HumanName: "Mark"})
	db.seed("room.created", room.Room{ID: "b", Name: "B", HumanName: "Board Chair"})
	db.seed("human.profile.updated", map[string]any{"name": "Mark Dhas", "title": ""})
	for id, got := range roomHumanNames(t, s) {
		if got != "" {
			t.Fatalf("room %s kept a room-only name after a profile rename: %q", id, got)
		}
	}
}

func TestFirstProfileNameReleasesRoomOnlyNames(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("room.created", room.Room{ID: "a", Name: "A", HumanName: "Mark"})
	db.seed("human.profile.updated", map[string]any{"name": "Mark", "title": ""})
	if got := roomHumanNames(t, s)["a"]; got != "" {
		t.Fatalf("first profile name must be what the room shows: %q", got)
	}
}

func TestProfileRenameToSameNameChangesNothing(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("human.profile.updated", map[string]any{"name": "Same", "title": ""})
	db.seed("room.created", room.Room{ID: "kept", Name: "Kept", HumanName: "Same"})
	db.seed("human.profile.updated", map[string]any{"name": "Same", "title": "New title"})
	if got := roomHumanNames(t, s)["kept"]; got != "Same" {
		t.Fatalf("a title-only change must not clear room names: %q", got)
	}
}
