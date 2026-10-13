package main

import (
	"time"
)

type actorRequestBudget struct {
	Tokens  float64
	Updated time.Time
}

// admitActorMutation is a mandatory single-controller budget: twenty-request
// bursts, refilling at one request/second. Emergency stop bypasses this policy
// through actorGuard's existing authenticated-owner safety path.
func (s *server) admitActorMutation(actor string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.requestBudgets == nil {
		s.requestBudgets = map[string]actorRequestBudget{}
	}
	b, exists := s.requestBudgets[actor]
	if !exists {
		if len(s.requestBudgets) >= 4096 {
			for id, old := range s.requestBudgets {
				if now.Sub(old.Updated) >= time.Minute {
					delete(s.requestBudgets, id)
				}
			}
			if len(s.requestBudgets) >= 4096 {
				return false
			}
		}
		b = actorRequestBudget{Tokens: 20, Updated: now}
	}
	if elapsed := now.Sub(b.Updated).Seconds(); elapsed > 0 {
		b.Tokens += elapsed
		if b.Tokens > 20 {
			b.Tokens = 20
		}
		b.Updated = now
	}
	allowed := b.Tokens >= 1
	if allowed {
		b.Tokens--
	}
	s.requestBudgets[actor] = b
	return allowed
}
