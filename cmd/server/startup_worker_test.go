package main

import (
	"context"
	"testing"
)

func TestStartupWorkerRequiresVerifiedIdleWithoutLaunchingOrReplaying(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"idle", "busy", "unknown", "audit-unavailable"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s, db, _ := mockServer(t)
			s.recoveryIdle = func(context.Context, string) (bool, error) {
				if mode == "unknown" {
					return false, context.DeadlineExceeded
				}
				return mode == "idle", nil
			}
			if mode == "audit-unavailable" {
				db.failAt = "begin"
			}
			s.verifyWorkerAtStartup(context.Background(), "ssh://synthetic-worker")
			if s.stopAll != (mode != "idle") {
				t.Fatal("unverified worker admitted", mode)
			}
			if mode == "idle" && len(db.snapshot()) != 0 {
				t.Fatal("idle check created work")
			}
			if mode == "audit-unavailable" && !s.integrityFault {
				t.Fatal("storage outage stop not latched")
			}
		})
	}
}
