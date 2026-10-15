package main

import (
	"agentroom/internal/room"
	"context"
	"sync"
	"testing"
	"time"
)

func TestDailyTurnBudgetUTCWindowAndUniqueJobs(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 13, 20, 0, 0, 0, time.FixedZone("Toronto", -4*3600))
	midnight := now.UTC()
	s := server{dailyTurnLimit: 2}
	st := room.State{ManagedJobs: map[string]room.ManagedJob{
		"prior":     {AllocatedAt: midnight.Add(-time.Second)},
		"first":     {AllocatedAt: midnight, ShutdownConfirmed: true},
		"cancelled": {AllocatedAt: midnight.Add(time.Hour), Status: "shutdown_confirmed"},
		"tomorrow":  {AllocatedAt: midnight.Add(24 * time.Hour)},
	}}
	b := s.dailyTurnBudget(st, now)
	if b.Used != 2 || b.Remaining != 0 || !b.ResetsAt.Equal(midnight.Add(24*time.Hour)) {
		t.Fatal(b)
	}
	if s.admitDailyTurns(st, now, 1) == nil {
		t.Fatal("exhausted budget admitted work")
	}
	delete(st.ManagedJobs, "cancelled")
	if s.admitDailyTurns(st, now, 2) == nil || s.admitDailyTurns(st, now, 1) != nil {
		t.Fatal("batch budget not checked")
	}
	if next := s.dailyTurnBudget(st, midnight.Add(24*time.Hour)); next.Used != 1 {
		t.Fatal("daily rollover failed", next)
	}
}
func TestDailyTurnLimitConfiguration(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"0", "-1", "100001", "unlimited", " 200", "200 ", "1.5"} {
		if _, err := parseDailyTurnLimit(raw); err == nil {
			t.Fatal("unsafe config", raw)
		}
	}
	for raw, want := range map[string]int{"": 200, "1": 1, "100000": 100000} {
		if got, err := parseDailyTurnLimit(raw); err != nil || got != want {
			t.Fatal(raw, got, err)
		}
	}
}
func TestDailyTurnBudgetPersistsAcrossControllerRestart(t *testing.T) {
	t.Parallel()
	s, db, _, reg := syntheticManagedJob(t)
	s.dailyTurnLimit = 1
	if err := reg.Update("shutdown_confirmed"); err != nil {
		t.Fatal(err)
	}
	// New controller instance reads the same mock event journal, not a token counter.
	restarted := &server{db: db, dailyTurnLimit: 1}
	st, err := restarted.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if restarted.admitDailyTurns(st, time.Now(), 1) == nil {
		t.Fatal("restart reset budget")
	}
	db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
	st, err = restarted.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if restarted.admitDailyTurns(st, time.Now(), 1) == nil {
		t.Fatal("cosmetic flag disabled enforcement")
	}
}
func TestLastDailySlotIsAtomicAcrossRegistrations(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	s.dailyTurnLimit = 1
	registrations := []managedJobRegistration{}
	for _, pid := range []string{"one", "two"} {
		p := room.Participant{ID: pid, RoomID: "engineering", ConnectionID: "connection"}
		db.seed("participant.created", p)
		registrations = append(registrations, s.managedJobContext(context.Background(), p).Value(managedJobContextKey{}).(managedJobRegistration))
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, reg := range registrations {
		wg.Add(1)
		go func(reg managedJobRegistration) { defer wg.Done(); results <- reg.Register("worker", "codex") }(reg)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	st, err := s.load(context.Background())
	if err != nil || success != 1 || len(st.ManagedJobs) != 1 {
		t.Fatal("daily slot duplicated", success, err)
	}
}
func TestFailedAllocationDoesNotConsumeBudget(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	s.dailyTurnLimit = 1
	p := room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection"}
	db.seed("participant.created", p)
	reg := s.managedJobContext(context.Background(), p).Value(managedJobContextKey{}).(managedJobRegistration)
	db.failAt = "commit"
	if reg.Register("worker", "codex") == nil {
		t.Fatal("failed allocation accepted")
	}
	db.failAt = ""
	st, err := s.load(context.Background())
	if err != nil || s.dailyTurnBudget(st, time.Now()).Used != 0 || !s.integrityFault {
		t.Fatal("rollback consumed quota or lost stop", err)
	}
}

func TestConnectionVerificationCannotBypassDailyBudget(t *testing.T) {
	t.Parallel()
	s, db, c := verificationFixture(t)
	s.dailyTurnLimit = 1
	p := room.Participant{ID: "ordinary", RoomID: "engineering", ConnectionID: c.ID}
	db.seed("participant.created", p)
	reg := s.managedJobContext(context.Background(), p).Value(managedJobContextKey{}).(managedJobRegistration)
	if err := reg.Register("worker", "codex"); err != nil {
		t.Fatal(err)
	}
	if err := reg.Update("shutdown_confirmed"); err != nil {
		t.Fatal(err)
	}
	s.verificationTurn = func(context.Context) (string, error) {
		t.Error("exhausted verification launched provider")
		return "", nil
	}
	if _, err := s.managedVerification(context.Background(), c, "engineering"); err == nil {
		t.Fatal("verification bypassed daily budget")
	}
}
