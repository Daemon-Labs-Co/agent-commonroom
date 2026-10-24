package main

import (
	"agentroom/internal/room"
	"encoding/json"
	"strings"
)

// Historical replay still derives legacy identities/rooms for scope validation.
// Only unused implicit agent rooms are omitted from the resulting room catalog.
// Any historical reference conservatively retains the room, including settings,
// jobs, audit records and removed work; hiding a room must never erase history.
func rememberAgentRoomReferences(data []byte, referenced map[string]bool) {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return
	} // Replay reports malformed events.
	var visit func(any)
	visit = func(v any) {
		switch x := v.(type) {
		case string:
			if strings.HasPrefix(x, "personal-bot-") {
				referenced[x] = true
			}
		case []any:
			for _, item := range x {
				visit(item)
			}
		case map[string]any:
			for _, item := range x {
				visit(item)
			}
		}
	}
	visit(value)
}

func omitUnusedAgentRooms(st *room.State, referenced map[string]bool) {
	if st.FeatureFlags["automaticAgentRooms"] {
		return
	}
	kept := st.Rooms[:0]
	for _, r := range st.Rooms {
		actor, exists := member(*st, r.PersonalActorID)
		implicitAgent := exists && actor.Kind == "agent" && r.ID == "personal-"+actor.ID
		if implicitAgent && !referenced[r.ID] {
			continue
		}
		kept = append(kept, r)
	}
	st.Rooms = kept
}
