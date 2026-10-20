package main

import (
	"agentroom/internal/room"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Inventory failure must leave no authoritative ledger, including after replay.
// This exercises controller accounting only; no physical allocation is attempted.
func driveP0StorageMeteringUnknown(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	if scenario.Given["meter"] != "unavailable" {
		t.Fatal("unknown metering failure fixture")
	}
	s, db, _ := mockServer(t)
	db.seed("worker.job.updated", quotaJob(1, true))
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	before := db.snapshot()
	calls := 0
	err := s.initializeManagedStorage(ctx, "company", storageInventoryFunc(func(ctx context.Context, company string) (managedStorageInventory, error) {
		calls++
		if ctx.Err() != nil || company != "company" {
			t.Fatal("inventory invoked with invalid account/context")
		}
		// Even a structurally valid partial result is unusable when the trusted
		// source cannot attest completeness. Never reinterpret it as empty usage.
		return managedStorageInventory{1, company, strings.Repeat("d", 64), []room.StorageReservation{}}, errors.New("synthetic meter unavailable")
	}))
	if calls != 1 || !errors.Is(err, errStorageUnknown) {
		t.Fatal("meter failure was not propagated", calls, err)
	}
	out := pendingP0Observation{State: "block", Values: map[string]string{}}
	for _, controller := range []*server{s, {db: db}} {
		state, err := controller.load(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if state.ManagedStorage != nil {
			out.Effects = append(out.Effects, "inventory_recorded")
		}
		for i, category := range []string{"workspace", "bundle", "artifact"} {
			total, err := controller.reserveManagedStorage(ctx, storageRecord(20+i, 1, category))
			if err == nil {
				out.State = "allow"
				out.Effects = append(out.Effects, "storage_reserved")
			} else if !errors.Is(err, errStorageUnknown) || total != 0 {
				t.Fatal("reservation did not reach unknown-inventory gate", category, total, err)
			}
		}
	}
	after := db.snapshot()
	if len(after) < len(before) || !reflect.DeepEqual(before, after[:len(before)]) {
		out.Effects = append(out.Effects, "history_mutated")
	}
	for _, event := range after[min(len(before), len(after)):] {
		out.Effects = append(out.Effects, event.kind)
	}
	return out
}

func TestStorageMeteringFailureRequiresExplicitCompleteInventoryForRecovery(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	db.seed("worker.job.updated", quotaJob(1, true))
	ctx := context.WithValue(t.Context(), actorContextKey{}, ownerID)
	if err := s.initializeManagedStorage(ctx, "company", storageInventoryFunc(func(context.Context, string) (managedStorageInventory, error) {
		return managedStorageInventory{}, errors.New("synthetic unavailable meter")
	})); !errors.Is(err, errStorageUnknown) {
		t.Fatal("failed inventory was accepted", err)
	}
	restarted := &server{db: db}
	retained := storageRecord(10, managedStorageLimit-1, "workspace")
	if err := restarted.initializeManagedStorage(ctx, "company", storageInventoryFunc(func(context.Context, string) (managedStorageInventory, error) {
		return managedStorageInventory{1, "company", strings.Repeat("d", 64), []room.StorageReservation{retained}}, nil
	})); err != nil {
		t.Fatal("complete inventory could not recover", err)
	}
	if total, err := restarted.reserveManagedStorage(ctx, storageRecord(20, 1, "artifact")); err != nil || total != managedStorageLimit {
		t.Fatal("healthy metering control failed", total, err)
	}
	replayed, err := (&server{db: db}).load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if total, err := managedStorageTotal(replayed); err != nil || total != managedStorageLimit || replayed.ManagedStorage.Reservations[retained.ID] != retained {
		t.Fatal("recovery lost retained charges", total, err)
	}
	before := db.snapshot()
	if _, err := restarted.reserveManagedStorage(ctx, storageRecord(21, 1, "bundle")); !errors.Is(err, errStorageCapacity) || !reflect.DeepEqual(before, db.snapshot()) {
		t.Fatal("recovered inventory overbooked capacity", err)
	}
}

// Drives real controller reservation methods and journal replay with the in-memory
// database adapter. These cases establish accounting, not physical disk quotas.
// Observe committed events/results; never manufacture observations from Want.
func driveP0StorageReservation(t *testing.T, scenario pendingP0Case) pendingP0Observation {
	t.Helper()
	used := int64(0)
	if value, ok := scenario.Given["account_used_bytes"]; ok {
		var err error
		used, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
	}
	initial := []roomStorageFixture{}
	// Category-specific retained fixtures must all contribute to the same ledger.
	for i, category := range []string{"workspace", "bundle", "artifact"} {
		field := map[string]string{"workspace": "retained_workspace_bytes", "bundle": "staged_bundle_bytes", "artifact": "artifact_bytes"}[category]
		if value, ok := scenario.Given[field]; ok {
			amount, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			initial = append(initial, roomStorageFixture{10 + i, amount, category})
		}
	}
	if len(initial) == 0 && used > 0 {
		initial = append(initial, roomStorageFixture{10, used, "workspace"})
	}
	records := []room.StorageReservation{}
	for _, fixture := range initial {
		records = append(records, storageRecord(fixture.id, fixture.bytes, fixture.category))
	}
	s, db, ctx := storageFixture(t, records...)
	if scenario.ID == "storage-no-paid-bypass" {
		db.seed("feature.flag.updated", map[string]any{"name": "jobDiagnostics", "enabled": false})
		db.seed("feature.flag.updated", map[string]any{"name": "agentBackgroundWork", "enabled": false})
	}
	before := len(db.snapshot())
	out := pendingP0Observation{State: "deny", Values: map[string]string{}}
	if scenario.ID == "storage-concurrent-last-byte" {
		other := &server{db: db}
		results := make(chan error, 2)
		var wg sync.WaitGroup
		for i, controller := range []*server{s, other} {
			wg.Add(1)
			go func(i int, controller *server) {
				defer wg.Done()
				_, err := controller.reserveManagedStorage(ctx, storageRecord(20+i, 1, "artifact"))
				results <- err
			}(i, controller)
		}
		wg.Wait()
		close(results)
		success := 0
		for err := range results {
			if err == nil {
				success++
			}
		}
		latest, err := s.load(ctx)
		if err != nil {
			t.Fatal(err)
		}
		total, err := managedStorageTotal(latest)
		if err != nil {
			t.Fatal(err)
		}
		events := db.snapshot()[before:]
		reservations := 0
		for _, event := range events {
			if event.kind == "storage.reserved" {
				reservations++
			} else {
				out.Effects = append(out.Effects, "unexpected:"+event.kind)
			}
		}
		if reservations == 1 && success == 1 {
			out.State = "allow"
			out.Effects = append(out.Effects, "reservation_serialized")
		}
		out.Values["successful_reservations"] = strconv.Itoa(success)
		out.Values["used_plus_reserved_bytes"] = strconv.FormatInt(total, 10)
		return out
	}
	requested, err := strconv.ParseInt(scenario.Given["reserve_bytes"], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	total, err := s.reserveManagedStorage(ctx, storageRecord(20, requested, "artifact"))
	if err == nil {
		out.State = "allow"
		out.Values["used_plus_reserved_bytes"] = strconv.FormatInt(total, 10)
	}
	for _, event := range db.snapshot()[before:] {
		if event.kind == "storage.reserved" {
			out.Effects = append(out.Effects, "storage_reserved")
		} else {
			out.Effects = append(out.Effects, fmt.Sprintf("unexpected:%s", event.kind))
		}
	}
	return out
}

type roomStorageFixture struct {
	id       int
	bytes    int64
	category string
}
