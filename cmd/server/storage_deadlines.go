package main

import (
	"context"
	"time"
)

const maximumStorageWait = 5 * time.Second

// storageContext bounds cooperative DB work, including work holding s.mu.
// The private shorter duration is for deterministic tests; it cannot disable
// the production cap or extend an earlier caller deadline. Source forks CAN
// remove it, BUT stalled storage can block cancellation and exhaust handlers.
func (s *server) storageContext(parent context.Context) (context.Context, context.CancelFunc) {
	timeout := s.storageTimeout
	if timeout <= 0 || timeout > maximumStorageWait {
		timeout = maximumStorageWait
	}
	return context.WithTimeout(parent, timeout)
}
