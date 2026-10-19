package main

import (
	"agentroom/internal/room"
	"fmt"
	"strconv"
	"sync"
	"testing"
)

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
