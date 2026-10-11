package main

import (
	"agentroom/internal/room"
	"strings"
	"testing"
)

func TestRoomIsolation(t *testing.T) {
	t.Parallel()
	st := room.State{Connections: []room.Connection{{ID: "global"}}, Participants: []room.Participant{{ID: "a", RoomID: "one"}, {ID: "b", RoomID: "two"}}, Messages: []room.Message{{ID: "a", RoomID: "one"}, {ID: "b", RoomID: "two"}}, Memories: []room.Memory{{ID: "a", RoomID: "one", Scope: "Room"}, {ID: "b", RoomID: "two", Scope: "Room"}, {ID: "shared", RoomID: "two", Scope: "Application"}}, Tasks: []room.Task{{ID: "a", RoomID: "one"}, {ID: "b", RoomID: "two"}}, Evidence: []room.Evidence{{ID: "a", RoomID: "one"}, {ID: "b", RoomID: "two"}}}
	got := scoped(st, "one")
	if len(got.Participants) != 1 || got.Participants[0].ID != "a" || len(got.Messages) != 1 || len(got.Tasks) != 1 || len(got.Evidence) != 1 || len(got.Memories) != 2 || len(got.Connections) != 1 {
		t.Fatalf("Wrong room projection: %+v", got)
	}
	if len(st.Participants) != 2 {
		t.Fatal("Projection altered global state")
	}
}
func TestChallengeGuidance(t *testing.T) {
	t.Parallel()
	for n := 0; n <= 4; n++ {
		p := room.Participant{Challenge: &n}
		g := profileGuidance(p)
		if !strings.Contains(g, "no permissions or authority") {
			t.Fatal("Challenge must not grant authority")
		}
	}
	if !strings.Contains(profileGuidance(room.Participant{}), "Constructive") {
		t.Fatal("Default challenge must be constructive")
	}
}
