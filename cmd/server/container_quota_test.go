package main

import (
	"agentroom/internal/room"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func quotaJob(i int, pair bool) room.ManagedJob {
	now := time.Now().UTC()
	job := room.ManagedJob{Version: 2, JobID: fmt.Sprintf("%032x", i), CompanyID: "company", InitiatorID: "human", RoomID: fmt.Sprintf("room%d", i), ParticipantID: fmt.Sprintf("agent%d", i), ConnectionID: "connection", Worker: fmt.Sprintf("worker%d", i), Provider: "codex", Status: "allocated", AllocatedAt: now, UpdatedAt: now, Deadline: now.Add(time.Minute)}
	if pair {
		job.Version = workerWorkspaceLifecycleVersion
		job.Execution = workerWorkspaceExecution
		job.RepositoryDigest = strings.Repeat("b", 64)
	}
	return job
}
func TestManagedContainerQuotaCountsPhysicalReservationsAcrossWorkers(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name              string
		pairs, singles    int
		nextPair, blocked bool
	}{
		{"second pair fits", 1, 0, true, false},
		{"third pair refused", 2, 0, true, true},
		{"fifth single refused", 0, 4, false, true},
		{"mixed exact limit", 1, 1, false, false},
		{"mixed overflow", 1, 1, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := room.State{Company: room.Company{ID: "company"}, ManagedJobs: map[string]room.ManagedJob{}}
			for i := 1; i <= tc.pairs+tc.singles; i++ {
				job := quotaJob(i, i <= tc.pairs)
				st.ManagedJobs[job.JobID] = job
			}
			if blocked := admitManagedContainerQuota(st, quotaJob(20, tc.nextPair)) != nil; blocked != tc.blocked {
				t.Fatalf("blocked %v, want %v", blocked, tc.blocked)
			}
		})
	}
}
func TestManagedContainerQuotaRetainsExpiredAndUnconfirmedReservations(t *testing.T) {
	t.Parallel()
	st := room.State{Company: room.Company{ID: "company"}, ManagedJobs: map[string]room.ManagedJob{}}
	for i := 1; i <= 2; i++ {
		job := quotaJob(i, true)
		job.AllocatedAt = job.AllocatedAt.Add(-time.Hour)
		job.Deadline = job.Deadline.Add(-time.Hour)
		job.Status = "shutdown_unconfirmed"
		st.ManagedJobs[job.JobID] = job
	}
	next := quotaJob(20, false)
	if admitManagedContainerQuota(st, next) == nil {
		t.Fatal("deadline released unconfirmed capacity")
	}
	job := st.ManagedJobs[fmt.Sprintf("%032x", 1)]
	job.Status = "shutdown_confirmed"
	job.ShutdownConfirmed = true
	st.ManagedJobs[job.JobID] = job
	if err := admitManagedContainerQuota(st, next); err != nil {
		t.Fatal("shutdown evidence did not release capacity", err)
	}
	job.ShutdownConfirmed = false // A forged terminal label is not evidence.
	st.ManagedJobs[job.JobID] = job
	if admitManagedContainerQuota(st, next) == nil {
		t.Fatal("invalid evidence released capacity")
	}
}
func TestManagedContainerQuotaConservativelyCountsLegacyAttribution(t *testing.T) {
	t.Parallel()
	st := room.State{Company: room.Company{ID: "company"}, ManagedJobs: map[string]room.ManagedJob{}}
	for i := 1; i <= 4; i++ {
		job := quotaJob(i, false)
		job.CompanyID = ""
		job.InitiatorID = ""
		st.ManagedJobs[job.JobID] = job
	}
	if admitManagedContainerQuota(st, quotaJob(20, false)) == nil {
		t.Fatal("legacy attribution bypassed quota")
	}
}
func TestAccountContainerQuotaPersistsAcrossRestartAndDisabledFlags(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	registrations := []managedJobRegistration{}
	for i := 0; i < 5; i++ {
		p := room.Participant{ID: fmt.Sprintf("quota-agent%d", i), RoomID: "engineering", ConnectionID: "connection"}
		db.seed("participant.created", p)
		registrations = append(registrations, s.managedJobContext(t.Context(), p).Value(managedJobContextKey{}).(managedJobRegistration))
	}
	for i := 0; i < 4; i++ {
		if err := registrations[i].Register(fmt.Sprintf("worker%d", i), "codex"); err != nil {
			t.Fatal(err)
		}
	}
	db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
	db.seed("feature.flag.updated", map[string]any{"name": "agentBackgroundWork", "enabled": false})
	if registrations[4].Register("another-worker", "codex") == nil {
		t.Fatal("fifth container admitted")
	}
	restarted := &server{db: db}
	st, err := restarted.load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if admitManagedContainerQuota(st, quotaJob(20, false)) == nil {
		t.Fatal("restart cleared durable reservations")
	}
	if err = registrations[0].Update("shutdown_confirmed"); err != nil {
		t.Fatal(err)
	}
	if err = registrations[4].Register("another-worker", "codex"); err != nil {
		t.Fatal("proof did not permit new allocation", err)
	}
}
func TestLastAccountContainerSlotIsAtomicAcrossRegistrations(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	registrations := []managedJobRegistration{}
	for i := 0; i < 5; i++ {
		p := room.Participant{ID: fmt.Sprintf("atomic-agent%d", i), RoomID: "engineering", ConnectionID: "connection"}
		db.seed("participant.created", p)
		registrations = append(registrations, s.managedJobContext(t.Context(), p).Value(managedJobContextKey{}).(managedJobRegistration))
	}
	for i := 0; i < 3; i++ {
		if err := registrations[i].Register("worker", "codex"); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 3; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results <- registrations[i].Register(fmt.Sprintf("worker%d", i), "codex")
		}(i)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	st, err := s.load(t.Context())
	if err != nil || success != 1 || len(st.ManagedJobs) != 4 {
		t.Fatal("last slot not serialized", success, len(st.ManagedJobs), err)
	}
}

func TestAccountQuotaStaleControllersCannotBothCommitLastSlot(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	for i := 1; i <= 3; i++ {
		db.seed("worker.job.updated", quotaJob(i, false))
	}
	other := &server{db: db}
	firstState, err := s.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	secondState, err := other.load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	first, second := quotaJob(10, false), quotaJob(11, false)
	if admitManagedContainerQuota(firstState, first) != nil || admitManagedContainerQuota(secondState, second) != nil {
		t.Fatal("last slot unexpectedly unavailable")
	}
	ctx := context.WithValue(t.Context(), expectedEventCountKey{}, firstState.EventCount)
	if err = s.append(ctx, "worker.job.updated", first); err != nil {
		t.Fatal(err)
	}
	ctx = context.WithValue(t.Context(), expectedEventCountKey{}, secondState.EventCount)
	if err = other.append(ctx, "worker.job.updated", second); err != errAuditConflict {
		t.Fatal("stale controller allocation committed", err)
	}
	latest, err := s.load(t.Context())
	if err != nil || len(latest.ManagedJobs) != 4 {
		t.Fatal("quota exceeded", len(latest.ManagedJobs), err)
	}
	if admitManagedContainerQuota(latest, second) == nil {
		t.Fatal("refreshed controller ignored committed reservation")
	}
}

func TestManagedContainerQuotaRejectsUnknownIdentityAndMalformedEvidence(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"missing company", "foreign incoming", "duplicate incoming", "malformed prior", "wrong map key", "confirmed incoming"} {
		t.Run(mode, func(t *testing.T) {
			st := room.State{Company: room.Company{ID: "company"}, ManagedJobs: map[string]room.ManagedJob{}}
			next := quotaJob(20, false)
			switch mode {
			case "missing company":
				st.Company.ID = ""
			case "foreign incoming":
				next.CompanyID = "other"
			case "duplicate incoming":
				st.ManagedJobs[next.JobID] = next
			case "malformed prior":
				prior := quotaJob(1, true)
				prior.Version = 99
				st.ManagedJobs[prior.JobID] = prior
			case "wrong map key":
				st.ManagedJobs["wrong"] = quotaJob(1, false)
			case "confirmed incoming":
				next.ShutdownConfirmed = true
				next.Status = "shutdown_confirmed"
			}
			if admitManagedContainerQuota(st, next) == nil {
				t.Fatal("invalid quota evidence admitted", mode)
			}
		})
	}
}
