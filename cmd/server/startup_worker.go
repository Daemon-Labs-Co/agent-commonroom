package main

import (
	"context"
	"log"
	"time"
)

// verifyWorkerAtStartup refuses fresh dispatch over unverified orphan work.
// It does not adopt, kill or replay jobs. Source forks CAN skip it, BUT risk
// concurrent stale writers and bypassing controller admission limits.
func (s *server) verifyWorkerAtStartup(parent context.Context, endpoint string) {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	probe := s.recoveryIdle
	if probe == nil {
		probe = remoteWorkerIdle
	}
	idle, err := probe(ctx, endpoint)
	if err == nil && idle {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopAll = true
	auditCtx, auditCancel := context.WithTimeout(parent, 5*time.Second)
	defer auditCancel()
	if s.append(auditCtx, "agents.stop", map[string]string{"scope": "application", "reason": "Startup worker state is active or unverified; inspect worker jobs and explicitly resume after confirmed shutdown"}) != nil {
		// Storage failures must survive restart via the integrity stop marker.
		s.tripIntegrityLocked()
	}
	log.Print("Managed execution stopped: startup worker state is active or unverified; owner investigation required")
}
