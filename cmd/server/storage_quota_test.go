package main

import (
	"agentroom/internal/room"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
	"time"
)

type storageInventoryFunc func(context.Context, string) (managedStorageInventory, error)

func (f storageInventoryFunc) Inventory(c context.Context, s string) (managedStorageInventory, error) {
	return f(c, s)
}
func storageRecord(i int, bytes int64, category string) room.StorageReservation {
	j := quotaJob(1, true)
	return room.StorageReservation{Version: 1, ID: fmt.Sprintf("%032x", i), CompanyID: j.CompanyID, JobID: j.JobID, RoomID: j.RoomID, Worker: j.Worker, RepositoryDigest: j.RepositoryDigest, Category: category, Bytes: bytes}
}
func storageFixture(t *testing.T, initial ...room.StorageReservation) (*server, *mockDB, context.Context) {
	t.Helper()
	s, db, _ := mockServer(t)
	db.seed("worker.job.updated", quotaJob(1, true))
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	if initial == nil {
		initial = []room.StorageReservation{}
	}
	if err := s.initializeManagedStorage(ctx, "company", storageInventoryFunc(func(context.Context, string) (managedStorageInventory, error) {
		return managedStorageInventory{1, "company", strings.Repeat("d", 64), initial}, nil
	})); err != nil {
		t.Fatal(err)
	}
	return s, db, ctx
}
func TestStorageReservationExactBoundaryAndRetainedCategories(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"exact", "over", "retained", "paid-flags-off"} {
		t.Run(mode, func(t *testing.T) {
			used := managedStorageLimit - 1
			initial := []room.StorageReservation{storageRecord(10, used, "workspace")}
			if mode == "over" || mode == "paid-flags-off" {
				initial[0].Bytes++
			}
			if mode == "retained" {
				initial = []room.StorageReservation{storageRecord(10, 3_000_000_000, "workspace"), storageRecord(11, 1_000_000_000, "bundle"), storageRecord(12, 1_000_000_000, "artifact")}
			}
			s, db, ctx := storageFixture(t, initial...)
			if mode == "paid-flags-off" {
				db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
				db.seed("feature.flag.updated", map[string]any{"name": "agentBackgroundWork", "enabled": false})
			}
			before := len(db.snapshot())
			total, err := s.reserveManagedStorage(ctx, storageRecord(20, 1, "artifact"))
			if mode == "exact" {
				if err != nil || total != managedStorageLimit || len(db.snapshot()) != before+1 {
					t.Fatal("exact boundary failed", total, err)
				}
			} else if err != errStorageCapacity || len(db.snapshot()) != before {
				t.Fatal("capacity bypass", err)
			}
			// Reload through another controller to prove the charge is durable.
			latest, err := (&server{db: db}).load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			charged, err := managedStorageTotal(latest)
			if err != nil || charged != managedStorageLimit {
				t.Fatal("restart lost charge", charged, err)
			}
		})
	}
}
func TestStorageReservationsSurviveContainerShutdown(t *testing.T) {
	t.Parallel()
	s, db, ctx := storageFixture(t, storageRecord(10, managedStorageLimit, "workspace"))
	job := quotaJob(1, true)
	job.Status = "shutdown_confirmed"
	job.ShutdownConfirmed = true
	// Update the exact original journal binding, including its timestamps.
	st, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	job = st.ManagedJobs[job.JobID]
	job.Status = "shutdown_confirmed"
	job.DiagnosticCode = "shutdown_confirmed"
	job.ShutdownConfirmed = true
	db.seed("worker.job.updated", job)
	latest, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	total, err := managedStorageTotal(latest)
	if err != nil || total != managedStorageLimit {
		t.Fatal("shutdown refunded storage")
	}
	if latest.ManagedJobs[job.JobID].ShutdownConfirmed != true {
		t.Fatal("container reconciliation failed")
	}
}
func TestStorageUnknownInventoryAndInvalidAmountsRefuseWrites(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	db.seed("worker.job.updated", quotaJob(1, true))
	before := len(db.snapshot())
	if _, err := s.reserveManagedStorage(ctx, storageRecord(20, 1, "workspace")); err != errStorageUnknown {
		t.Fatal("unknown usage admitted", err)
	}
	if len(db.snapshot()) != before {
		t.Fatal("unknown inventory wrote reservation")
	}
	for _, amount := range []int64{-1, 0, math.MaxInt64} {
		r := storageRecord(20, amount, "workspace")
		st := room.State{Company: room.Company{ID: "company"}, ManagedStorage: &room.ManagedStorage{Version: 1, CompanyID: "company", InventoryDigest: strings.Repeat("d", 64), Reservations: map[string]room.StorageReservation{}}}
		if _, err := admitManagedStorage(st, r); err == nil {
			t.Fatal("invalid amount admitted")
		}
	}
	if err := s.initializeManagedStorage(ctx, "company", storageInventoryFunc(func(context.Context, string) (managedStorageInventory, error) {
		return managedStorageInventory{}, errors.New("unavailable meter")
	})); err != errStorageUnknown {
		t.Fatal("inventory failure admitted")
	}
	if len(db.snapshot()) != before {
		t.Fatal("failed inventory changed journal")
	}
}
func TestStorageLastByteIsSerializedAcrossControllers(t *testing.T) {
	t.Parallel()
	s, db, ctx := storageFixture(t, storageRecord(10, managedStorageLimit-1, "workspace"))
	other := &server{db: db}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i, c := range []*server{s, other} {
		wg.Add(1)
		go func(i int, c *server) {
			defer wg.Done()
			_, err := c.reserveManagedStorage(ctx, storageRecord(20+i, 1, "artifact"))
			results <- err
		}(i, c)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if err != errStorageCapacity && err != errAuditConflict {
			t.Fatal(err)
		}
	}
	st, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	total, err := managedStorageTotal(st)
	if success != 1 || err != nil || total != managedStorageLimit {
		t.Fatal("last byte overbooked", success, total, err)
	}
}
func TestStorageStaleRevisionCannotCommitLastByte(t *testing.T) {
	t.Parallel()
	s, db, ctx := storageFixture(t, storageRecord(10, managedStorageLimit-1, "workspace"))
	other := &server{db: db}
	st, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	first, second := storageRecord(20, 1, "artifact"), storageRecord(21, 1, "bundle")
	if _, err = admitManagedStorage(st, first); err != nil {
		t.Fatal(err)
	}
	if _, err = admitManagedStorage(st, second); err != nil {
		t.Fatal(err)
	}
	stale := context.WithValue(ctx, expectedEventCountKey{}, st.EventCount)
	if err = s.append(stale, "storage.reserved", first); err != nil {
		t.Fatal(err)
	}
	if err = other.append(stale, "storage.reserved", second); err != errAuditConflict {
		t.Fatal("stale reservation committed", err)
	}
}
func TestStoragePermissionAndBindingRefusals(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"no-actor", "agent", "disabled", "foreign-company", "foreign-room", "foreign-worker", "foreign-job", "foreign-repository", "duplicate", "negative", "unknown-category", "expired", "uncertain", "stopped", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			s, db, ctx := storageFixture(t, storageRecord(10, 1, "workspace"))
			r := storageRecord(20, 1, "artifact")
			switch mode {
			case "no-actor":
				ctx = t.Context()
			case "agent", "disabled":
				db.seed("actor.revised", room.Actor{ID: "caller", Kind: map[bool]string{true: "agent", false: "human"}[mode == "agent"], Enabled: mode != "disabled", RoomIDs: []string{r.RoomID}})
				ctx = context.WithValue(ctx, actorContextKey{}, "caller")
			case "foreign-company":
				r.CompanyID = "other"
			case "foreign-room":
				r.RoomID = "other"
			case "foreign-worker":
				r.Worker = "other"
			case "foreign-job":
				r.JobID = strings.Repeat("e", 32)
			case "foreign-repository":
				r.RepositoryDigest = strings.Repeat("e", 64)
			case "duplicate":
				r.ID = storageRecord(10, 1, "workspace").ID
			case "negative":
				r.Bytes = -1
			case "unknown-category":
				r.Category = "misc"
			case "expired", "uncertain":
				current, err := s.load(ctx)
				if err != nil {
					t.Fatal(err)
				}
				job := current.ManagedJobs[r.JobID]
				// Rebuild a separate valid fixture instead of mutating an immutable binding.
				fresh, newDB, _ := mockServer(t)
				if mode == "expired" {
					job.AllocatedAt = job.AllocatedAt.Add(-time.Hour)
					job.Deadline = job.Deadline.Add(-time.Hour)
				} else {
					job.Status = "shutdown_unconfirmed"
				}
				newDB.seed("worker.job.updated", job)
				newDB.seed("storage.inventory.recorded", managedStorageInventory{1, "company", strings.Repeat("d", 64), []room.StorageReservation{}})
				s, db = fresh, newDB
			case "stopped":
				s.stopAll = true
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			before := len(db.snapshot())
			if _, err := s.reserveManagedStorage(ctx, r); err == nil {
				t.Fatal("invalid caller/binding admitted")
			}
			if len(db.snapshot()) != before {
				t.Fatal("refusal appended event")
			}
		})
	}
}
func TestStorageProjectionRejectsCorruptOrRewrittenEvidence(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"duplicate", "wrong-company", "negative", "overflow", "rewrite", "uninitialized", "unknown-kind"} {
		t.Run(mode, func(t *testing.T) {
			s, _, ctx := storageFixture(t, storageRecord(10, 1, "workspace"))
			st, err := s.load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			kind := "storage.reserved"
			r := storageRecord(20, 1, "bundle")
			var payload any = r
			switch mode {
			case "duplicate":
				r.ID = storageRecord(10, 1, "workspace").ID
				payload = r
			case "wrong-company":
				r.CompanyID = "foreign"
				payload = r
			case "negative":
				r.Bytes = -1
				payload = r
			case "overflow":
				kind = "storage.inventory.recorded"
				st.ManagedStorage = nil
				payload = managedStorageInventory{1, "company", strings.Repeat("d", 64), []room.StorageReservation{storageRecord(10, math.MaxInt64, "workspace"), storageRecord(11, 1, "bundle")}}
			case "rewrite":
				kind = "storage.inventory.recorded"
				payload = managedStorageInventory{1, "company", strings.Repeat("d", 64), []room.StorageReservation{}}
			case "uninitialized":
				st.ManagedStorage = nil
			case "unknown-kind":
				kind = "storage.released"
			}
			raw, _ := json.Marshal(payload)
			if projectManagedStorage(&st, kind, raw) == nil {
				t.Fatal("corrupt storage evidence accepted")
			}
		})
	}
}
func TestStorageAppendFailureDoesNotGrantCapacity(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"begin", "lock", "canonical", "insert", "commit"} {
		t.Run(stage, func(t *testing.T) {
			s, db, ctx := storageFixture(t)
			db.failAt = stage
			if total, err := s.reserveManagedStorage(ctx, storageRecord(20, 1, "artifact")); err == nil || total != 0 {
				t.Fatal("failed durable reservation granted capacity")
			}
			db.failAt = ""
			st, err := s.load(ctx)
			if err != nil {
				t.Fatal(err)
			}
			total, err := managedStorageTotal(st)
			if err != nil || total != 0 {
				t.Fatal("failed reservation charged/released incorrectly")
			}
		})
	}
}

func TestStorageInventoryCannotBeResetOrProvidedWithoutOwner(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"no-owner", "reset", "foreign-company", "nil-inventory", "duplicate-inventory"} {
		t.Run(mode, func(t *testing.T) {
			s, db, _ := mockServer(t)
			ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
			company := "company"
			calls := 0
			if mode == "reset" {
				s, db, ctx = storageFixture(t, storageRecord(10, 1, "workspace"))
			}
			if mode == "no-owner" {
				ctx = t.Context()
			}
			if mode == "foreign-company" {
				company = "other"
			}
			before := len(db.snapshot())
			err := s.initializeManagedStorage(ctx, company, storageInventoryFunc(func(context.Context, string) (managedStorageInventory, error) {
				calls++
				records := []room.StorageReservation{}
				if mode == "nil-inventory" {
					records = nil
				}
				if mode == "duplicate-inventory" {
					records = []room.StorageReservation{storageRecord(10, 1, "workspace"), storageRecord(10, 1, "workspace")}
				}
				return managedStorageInventory{1, company, strings.Repeat("d", 64), records}, nil
			}))
			if err == nil || len(db.snapshot()) != before {
				t.Fatal("unsafe inventory accepted")
			}
			if (mode == "no-owner" || mode == "reset" || mode == "foreign-company") && calls != 0 {
				t.Fatal("unauthorized inventory adapter reached")
			}
		})
	}
}
func TestStorageEventsCannotBeInjectedThroughPublicEvents(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"storage.inventory.recorded", "storage.reserved", "storage.released"} {
		_, db, handler := mockServer(t)
		before := len(db.snapshot())
		response := mockRequest(handler, "POST", "/api/events", fmt.Sprintf(`{"kind":%q,"payload":{}}`, kind))
		if response.Code < 400 || len(db.snapshot()) != before {
			t.Fatal("public storage mutation allowed", kind, response.Code)
		}
	}
}

func TestStorageInitiatingHumanRetainsRoomAndJobPermissionChecks(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"authorized", "foreign-initiator", "no-room"} {
		t.Run(mode, func(t *testing.T) {
			s, db, ctx := storageFixture(t)
			actor := "human"
			rooms := []string{"room1"}
			if mode == "foreign-initiator" {
				actor = "another-human"
			}
			if mode == "no-room" {
				rooms = nil
			}
			db.seed("actor.revised", room.Actor{ID: actor, Kind: "human", Enabled: true, RoomIDs: rooms})
			ctx = context.WithValue(ctx, actorContextKey{}, actor)
			_, err := s.reserveManagedStorage(ctx, storageRecord(20, 1, "artifact"))
			if (err == nil) != (mode == "authorized") {
				t.Fatal("incorrect human permission result", err)
			}
		})
	}
}
func TestStorageOverLimitInventoryPreservesDataAndBlocksGrowth(t *testing.T) {
	t.Parallel()
	s, _, ctx := storageFixture(t, storageRecord(10, managedStorageLimit+1, "workspace"))
	st, err := s.load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	total, err := managedStorageTotal(st)
	if err != nil || total != managedStorageLimit+1 {
		t.Fatal("existing over-limit data erased")
	}
	if _, err = s.reserveManagedStorage(ctx, storageRecord(20, 1, "artifact")); err != errStorageCapacity {
		t.Fatal("over-limit account permitted growth", err)
	}
}
