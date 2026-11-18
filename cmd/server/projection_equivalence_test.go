package main

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"agentroom/internal/room"
)

// The projection is compared against the frozen original replay (referenceLoad). The check
// is attached to every mockDB, so each of the existing tests also proves, on the events it
// produced, that the reducer-based projection and its snapshots give exactly the state and
// the errors the original did.

var equivalenceNow = time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)

func (d *mockDB) withEvents(events []mockEvent) *mockDB {
	return &mockDB{t: d.t, events: append([]mockEvent(nil), events...)}
}

func compareStates(want room.State, wantErr error, got room.State, gotErr error) string {
	if (wantErr == nil) != (gotErr == nil) || (wantErr != nil && wantErr.Error() != gotErr.Error()) {
		return fmt.Sprintf("errors differ: reference %v, projection %v", wantErr, gotErr)
	}
	if wantErr == nil && !reflect.DeepEqual(want, got) {
		return "states differ"
	}
	return ""
}

// checkProjectionEquivalence replays the database's events with the reference and with the
// projection, in full and as a sequence of snapshots taken while the replay continues.
func (d *mockDB) checkProjectionEquivalence() {
	events := d.snapshot()
	ctx := context.Background()
	whole := &server{db: d.withEvents(events)}
	want, wantErr := whole.referenceLoad(ctx, equivalenceNow)
	got, gotErr := whole.loadAt(ctx, equivalenceNow)
	if msg := compareStates(want, wantErr, got, gotErr); msg != "" {
		d.t.Errorf("projection differs from the reference replay over %d events: %s", len(events), msg)
		return
	}
	if len(events) > 3000 {
		return
	}
	// Snapshots at a handful of points, taken while the same projection keeps replaying.
	live := newProjection()
	cuts := map[int]bool{len(events): true}
	for i := 1; i <= 5; i++ {
		cuts[len(events)*i/6] = true
	}
	// The same cut points through the incremental cache: the database grows between loads.
	growing := d.withEvents(nil)
	cached := &server{db: growing}
	cached.projection.enabled = true
	defer func() {
		for _, k := range sortedCuts(cuts) {
			growing.mu.Lock()
			growing.events = append([]mockEvent(nil), events[:k]...)
			growing.mu.Unlock()
			prefix := &server{db: d.withEvents(events[:k])}
			want, wantErr := prefix.referenceLoad(ctx, equivalenceNow)
			got, gotErr := cached.loadCached(ctx, equivalenceNow)
			if msg := compareStates(want, wantErr, got, gotErr); msg != "" {
				d.t.Errorf("incremental cache after %d of %d events differs from the reference replay: %s", k, len(events), msg)
				return
			}
			if wantErr != nil {
				return
			}
		}
	}()
	for i, event := range events {
		err := live.apply(event.kind, []byte(event.payload))
		if err != nil {
			prefix := &server{db: d.withEvents(events[:i+1])}
			_, refErr := prefix.referenceLoad(ctx, equivalenceNow)
			if refErr == nil || refErr.Error() != err.Error() {
				d.t.Errorf("event %d (%s): projection error %v, reference %v", i+1, event.kind, err, refErr)
			}
			return
		}
		if !cuts[i+1] {
			continue
		}
		prefix := &server{db: d.withEvents(events[:i+1])}
		want, wantErr := prefix.referenceLoad(ctx, equivalenceNow)
		if msg := compareStates(want, wantErr, live.snapshot(equivalenceNow), nil); msg != "" {
			d.t.Errorf("snapshot after %d of %d events differs from the reference replay: %s", i+1, len(events), msg)
			return
		}
	}
}

// A snapshot must not disturb later replay, and must not share memory with the projection.
func TestSnapshotIsIndependentOfTheProjection(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	db.seed("participant.created", room.Participant{ID: "p1", Name: "Dev", RoomID: "engineering", ConnectionID: "c1"})
	db.seed("message.created", room.Message{ID: "m1", Text: "first", RoomID: "engineering", Author: "You", ActorType: "human"})
	live := newProjection()
	for _, e := range db.snapshot() {
		if err := live.apply(e.kind, []byte(e.payload)); err != nil {
			t.Fatal(err)
		}
	}
	first := live.snapshot(equivalenceNow)
	first.Messages[0].Text = "changed by a caller"
	first.Participants = nil
	second := live.snapshot(equivalenceNow)
	if second.Messages[0].Text != "first" || len(second.Participants) != 1 {
		t.Fatal("a caller's change reached the projection")
	}
}

func sortedCuts(cuts map[int]bool) []int {
	out := []int{}
	for k := range cuts {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}
