package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"agentroom/internal/room"

	"github.com/jackc/pgx/v5"
)

type queryRecord struct {
	sql  string
	args []any
}

// recordingDB remembers every query so a test can say exactly what was read.
type recordingDB struct {
	eventDatabase
	mu      sync.Mutex
	queries []queryRecord
}

func (r *recordingDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	r.mu.Lock()
	r.queries = append(r.queries, queryRecord{sql, append([]any(nil), args...)})
	r.mu.Unlock()
	return r.eventDatabase.Query(ctx, sql, args...)
}

func (r *recordingDB) deltaArgs() []int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []int64{}
	for _, q := range r.queries {
		if q.sql == projectionDeltaSQL {
			out = append(out, q.args[0].(int64))
		}
	}
	return out
}

func cachedServer(db eventDatabase) *server {
	s := &server{db: db}
	s.projection.enabled = true
	return s
}

func seedMessages(db *mockDB, n int, prefix string) {
	for i := 0; i < n; i++ {
		db.seed("message.created", room.Message{ID: fmt.Sprintf("%s%d", prefix, i), Text: "hello", RoomID: "engineering", Author: "You", ActorType: "human"})
	}
}

func TestCacheReadsOnlyEventsItHasNotApplied(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	seedMessages(db, 5, "a")
	rec := &recordingDB{eventDatabase: db}
	s := cachedServer(rec)
	st, err := s.load(context.Background())
	if err != nil || len(st.Messages) != 5 {
		t.Fatal(err, len(st.Messages))
	}
	seedMessages(db, 3, "b")
	st, err = s.load(context.Background())
	if err != nil || len(st.Messages) != 8 || st.EventCount != 8 {
		t.Fatal(err, len(st.Messages), st.EventCount)
	}
	if _, err = s.load(context.Background()); err != nil {
		t.Fatal(err)
	}
	// First load from the beginning, then from the last applied sequence (5, then 8).
	if got := rec.deltaArgs(); !reflect.DeepEqual(got, []int64{0, 5, 8}) {
		t.Fatalf("delta reads started at %v", got)
	}
	for _, q := range rec.queries {
		if strings.Contains(q.sql, "ORDER BY sequence") && q.sql != projectionDeltaSQL {
			t.Fatalf("a full replay was read: %s", q.sql)
		}
	}
}

func TestCacheRebuildsWhenTheAnchorHashChanges(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	seedMessages(db, 4, "a")
	s := cachedServer(db)
	var reported atomic.Int32
	s.projection.onHistoryChanged = func() { reported.Add(1) }
	if _, err := s.load(context.Background()); err != nil {
		t.Fatal(err)
	}
	// An earlier record is rewritten in place (as direct database tampering would) and a
	// new event follows. The last applied row is no longer the one the cache applied.
	db.mu.Lock()
	db.events[3].hash = strings.Repeat("0", 64)
	db.mu.Unlock()
	seedMessages(db, 1, "b")
	st, err := s.load(context.Background())
	if err != nil || len(st.Messages) != 5 {
		t.Fatal(err, len(st.Messages))
	}
	if reported.Load() != 1 {
		t.Fatalf("history change reported %d times", reported.Load())
	}
}

func TestCacheRebuildsWhenHistoryShrinksOrDisappears(t *testing.T) {
	t.Parallel()
	for name, keep := range map[string]int{"purged to fewer events": 2, "purged completely": 0} {
		db := newMockDB(t)
		seedMessages(db, 5, "a")
		s := cachedServer(db)
		var reported atomic.Int32
		s.projection.onHistoryChanged = func() { reported.Add(1) }
		if _, err := s.load(context.Background()); err != nil {
			t.Fatal(name, err)
		}
		db.mu.Lock()
		db.events = db.events[:keep]
		db.mu.Unlock()
		st, err := s.load(context.Background())
		if err != nil || len(st.Messages) != keep || st.EventCount != int64(keep) || reported.Load() != 1 {
			t.Fatalf("%s: %v, %d messages, %d events, reported %d", name, err, len(st.Messages), st.EventCount, reported.Load())
		}
	}
}

func TestCacheNeverAdvancesPastAnInvalidEvent(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	seedMessages(db, 2, "a")
	s := cachedServer(db)
	if _, err := s.load(context.Background()); err != nil {
		t.Fatal(err)
	}
	db.seed("operation.receipt", map[string]any{"key": "short"}) // an event the replay must refuse
	seedMessages(db, 1, "b")
	_, cachedErr := s.load(context.Background())
	_, fullErr := (&server{db: db}).referenceLoad(context.Background(), equivalenceNow)
	if cachedErr == nil || fullErr == nil || cachedErr.Error() != fullErr.Error() {
		t.Fatalf("cached %v, full %v", cachedErr, fullErr)
	}
	// It keeps failing the same way (the refusal is not cached away) and holds no half-applied state.
	if _, again := s.load(context.Background()); again == nil || again.Error() != fullErr.Error() {
		t.Fatalf("second load: %v", again)
	}
	if s.projection.p != nil && s.projection.seq != 0 {
		t.Fatal("the cache kept a partially applied projection")
	}
}

func TestCacheQueryFailureKeepsTheCacheValid(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	seedMessages(db, 3, "a")
	s := cachedServer(db)
	if _, err := s.load(context.Background()); err != nil {
		t.Fatal(err)
	}
	db.failAt = "query"
	if _, err := s.load(context.Background()); err == nil {
		t.Fatal("a failed read was reported as success")
	}
	db.failAt = ""
	seedMessages(db, 1, "b")
	rec := &recordingDB{eventDatabase: db}
	s.db = rec
	st, err := s.load(context.Background())
	if err != nil || len(st.Messages) != 4 || !reflect.DeepEqual(rec.deltaArgs(), []int64{3}) {
		t.Fatal(err, len(st.Messages), rec.deltaArgs())
	}
}

func TestCachedStatesAreIndependentOfEachOther(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	seedMessages(db, 2, "a")
	s := cachedServer(db)
	first, _ := s.load(context.Background())
	first.Messages[0].Text = "changed by a caller"
	first.Messages = first.Messages[:0]
	second, _ := s.load(context.Background())
	if len(second.Messages) != 2 || second.Messages[0].Text != "hello" {
		t.Fatal("a caller's change reached a later state")
	}
}

func TestCacheUnderConcurrentLoadsAndAppends(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	seedMessages(db, 3, "a")
	s := cachedServer(db)
	var wg sync.WaitGroup
	errs := make(chan error, 64)
	for w := 0; w < 6; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				st, err := s.load(context.Background())
				if err != nil {
					errs <- err
					return
				}
				if int64(len(st.Messages)) != st.EventCount {
					errs <- errors.New("state is not a whole number of events")
					return
				}
			}
		}()
	}
	for i := 0; i < 20; i++ {
		seedMessages(db, 1, fmt.Sprintf("c%d-", i))
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	final, err := s.load(context.Background())
	if err != nil || len(final.Messages) != 23 {
		t.Fatal(err, len(final.Messages))
	}
}
