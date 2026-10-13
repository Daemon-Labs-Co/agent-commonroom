package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMutationBudgetIsConcurrentAndActorScoped(t *testing.T) {
	t.Parallel()
	s, _, _ := mockServer(t)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var allowed atomic.Int32
	var jobs sync.WaitGroup
	for i := 0; i < 100; i++ {
		jobs.Add(1)
		go func() {
			defer jobs.Done()
			if s.admitActorMutation("one", now) {
				allowed.Add(1)
			}
		}()
	}
	jobs.Wait()
	if allowed.Load() != 20 || !s.admitActorMutation("other", now) || !s.admitActorMutation("one", now.Add(time.Second)) || s.admitActorMutation("one", now.Add(time.Second)) {
		t.Fatal("mutation budget failed", allowed.Load())
	}
}

func TestMutationBudgetCapsTrackingAndEmergencyStopBypassesIt(t *testing.T) {
	t.Parallel()
	s, db, h := safetyHarness(t)
	now := time.Now()
	for i := 0; i < 4096; i++ {
		if !s.admitActorMutation(fmt.Sprint(i), now) {
			t.Fatal("capacity filled prematurely")
		}
	}
	if s.admitActorMutation("excess", now) {
		t.Fatal("unbounded identity tracking")
	}
	if !s.admitActorMutation("after expiry", now.Add(2*time.Minute)) {
		t.Fatal("expired budget not reclaimed")
	}
	for i := 0; i < 20; i++ {
		s.admitActorMutation(ownerID, now)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.cancels["p"] = cancel
	w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"application"}`)
	if w.Code != 200 || ctx.Err() == nil {
		t.Fatal("rate limit blocked emergency stop", w.Code, w.Body.String())
	}
	if len(db.snapshot()) == 0 {
		t.Fatal("stop not audited")
	}
}
