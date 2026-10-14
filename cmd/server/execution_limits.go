package main

import (
	"agentroom/internal/room"
	"errors"
)

const maxActiveAgentTurns = 8
const maxRoomAgentTurns = 4

// agentAdmission requires s.mu and the complete state, never a room projection.
// This mandatory resource boundary covers chat and workflow model turns on a
// single controller. It is not a durable queue or a multi-controller quota.
func (s *server) agentAdmission(st room.State, rid string, requested int) error {
	global := 0
	for _, active := range s.active {
		if active {
			global++
		}
	}
	if global+requested > maxActiveAgentTurns {
		return errors.New("Agent concurrency limit reached; wait for active turns to finish")
	}
	local := 0
	for _, p := range st.Participants {
		if p.RoomID == rid && s.active[p.ID] {
			local++
		}
	}
	for pid, roomID := range s.verificationRooms {
		if roomID == rid && s.active[pid] {
			local++
		}
	}
	if local+requested > maxRoomAgentTurns {
		return errors.New("Room agent concurrency limit reached; request a smaller group or wait for active turns to finish")
	}
	return nil
}
