package main

import (
	"context"
	"fmt"
	"testing"

	"agentroom/internal/room"
)

// Sizing benchmarks (no database): one more message arrives, then the state is needed. The
// full replay reads and decodes every event again; the cache reads one. Run with:
//
//	go test ./cmd/server -run '^$' -bench 'Replay' -benchtime 20x
func benchDB(b *testing.B, events int) *mockDB {
	b.Helper()
	db := &mockDB{t: &testing.T{}}
	for i := 0; i < events; i++ {
		raw := fmt.Sprintf(`{"id":"m%d","text":"a message of ordinary length for the benchmark","room":"engineering","author":"You","actorType":"human"}`, i)
		db.events = append(db.events, mockEvent{kind: "message.created", payload: raw, sequence: int64(i + 1), hash: fmt.Sprintf("h%d", i)})
	}
	return db
}

func BenchmarkReplayFull(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			s := &server{db: benchDB(b, n)}
			for i := 0; i < b.N; i++ {
				if _, err := s.loadAt(context.Background(), equivalenceNow); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkReplayCached(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			db := benchDB(b, n)
			s := &server{db: db}
			s.projection.enabled = true
			if _, err := s.loadCached(context.Background(), equivalenceNow); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				db.events = append(db.events, mockEvent{kind: "message.created", payload: `{"id":"x","text":"new","room":"engineering","author":"You","actorType":"human"}`, sequence: int64(len(db.events) + 1), hash: fmt.Sprintf("n%d", i)})
				if _, err := s.loadCached(context.Background(), equivalenceNow); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	_ = room.State{}
}

// The part of a cached load that still grows with the size of the state (not the history):
// copying the replayed state and deriving the views from it.
func BenchmarkSnapshot(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			db := benchDB(b, n)
			p := newProjection()
			for _, e := range db.events {
				if err := p.apply(e.kind, []byte(e.payload)); err != nil {
					b.Fatal(err)
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = p.snapshot(equivalenceNow)
			}
		})
	}
}

func BenchmarkSnapshotCloneOnly(b *testing.B) {
	for _, n := range []int{1000, 10000, 50000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			db := benchDB(b, n)
			p := newProjection()
			for _, e := range db.events {
				_ = p.apply(e.kind, []byte(e.payload))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = p.st.Clone()
			}
		})
	}
}
