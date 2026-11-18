package main

import (
	"agentroom/internal/room"
	"errors"
	"sync"
	"time"
)

const maxActiveAgentTurns = 8
const maxRoomAgentTurns = 4

// agentAdmission requires s.mu and the complete state, never a room projection.
// This mandatory resource boundary covers chat and workflow model turns on a
// single controller. It is not a durable queue or a multi-controller quota.
func (s *server) agentAdmission(st room.State, rid string, requested int) error {
	if s.hostedExecutionBlocked {
		return errors.New(hostedExecutionMessage)
	}
	if s.pool != nil {
		cap := 0
		if s.limits != nil {
			cap = s.limits.ConcurrentTurns
		}
		if msg := s.pool.refusal(s.workspaceID, requested, cap); msg != "" {
			return errors.New(msg)
		}
	}
	if err := s.admitDailyTurns(st, time.Now(), requested); err != nil {
		return err
	}
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

// setActive and clearActive are the only way a turn is marked running or finished. They
// keep the installation-wide count in step with this workspace's own map. Callers hold s.mu.
func (s *server) setActive(id string) {
	if !s.active[id] && s.pool != nil {
		s.pool.add(s.workspaceID, 1)
	}
	s.active[id] = true
}

func (s *server) clearActive(id string) {
	if s.active[id] && s.pool != nil {
		s.pool.add(s.workspaceID, -1)
	}
	delete(s.active, id)
}

// executionPool counts running agent turns across every workspace in an installation so
// that one workspace cannot use all the capacity. Limits protect resources; they are not
// a security boundary, and the check and the later start are not one atomic step.
type executionPool struct {
	mu    sync.Mutex
	total int
	per   map[string]int
	max   int
}

func newExecutionPool(max int) *executionPool {
	return &executionPool{per: map[string]int{}, max: max}
}

func (p *executionPool) add(workspace string, n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total += n
	p.per[workspace] += n
	if p.per[workspace] <= 0 {
		delete(p.per, workspace)
	}
	if p.total < 0 {
		p.total = 0
	}
}

// refusal explains why requested more turns cannot start, or returns "". It names no other
// workspace and gives no count of other workspaces' work.
func (p *executionPool) refusal(workspace string, requested, perWorkspace int) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if perWorkspace > 0 && p.per[workspace]+requested > perWorkspace {
		return "This workspace is already running its limit of " + itoa(perWorkspace) + " agent turns; wait for one to finish"
	}
	if p.total+requested > p.max {
		return "The installation is running as many agent turns as it can; try again shortly"
	}
	return ""
}

func (p *executionPool) running(workspace string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.per[workspace]
}
