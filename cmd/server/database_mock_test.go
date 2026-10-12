package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// This mock accepts only the application's event-store operations. Fixtures and
// transaction writes exist solely in memory. Unexpected SQL fails the test.
type mockEvent struct{ kind, payload, previous, hash string }
type mockDB struct {
	t         *testing.T
	mu        sync.Mutex
	events    []mockEvent
	failAt    string
	streamErr error
}

func newMockDB(t *testing.T) *mockDB { return &mockDB{t: t} }
func (d *mockDB) snapshot() []mockEvent {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]mockEvent(nil), d.events...)
}
func (d *mockDB) seed(kind string, value any) {
	raw, err := json.Marshal(value)
	if err != nil {
		d.t.Fatal(err)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	previous := ""
	if len(d.events) > 0 {
		previous = d.events[len(d.events)-1].hash
	}
	d.events = append(d.events, mockEvent{kind, string(raw), previous, checksum(previous, kind, string(raw))})
}
func (d *mockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if d.failAt == "begin" {
		return nil, errors.New("mock begin failure")
	}
	d.mu.Lock()
	return &mockTx{db: d}, nil
}
func (d *mockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if d.failAt == "query" {
		return nil, errors.New("mock query failure")
	}
	events := d.snapshot()
	values := [][]any{}
	switch sql {
	case "SELECT kind,payload FROM events ORDER BY sequence":
		for _, e := range events {
			values = append(values, []any{e.kind, []byte(e.payload)})
		}
	case "SELECT kind,payload::text,previous_hash,hash FROM events ORDER BY sequence":
		for _, e := range events {
			values = append(values, []any{e.kind, e.payload, e.previous, e.hash})
		}
	case "SELECT sequence,kind,committed_at,hash,payload FROM events ORDER BY sequence DESC LIMIT 200":
		for i := len(events) - 1; i >= 0 && len(values) < 200; i-- {
			e := events[i]
			values = append(values, []any{int64(i + 1), e.kind, time.Unix(int64(i), 0).UTC(), e.hash, []byte(e.payload)})
		}
	case auditPageSQL:
		for i := len(events) - 1; i >= 0 && len(values) < 101; i-- {
			at := time.Unix(int64(i), 0).UTC()
			if args[0] != nil && int64(i+1) >= args[0].(int64) {
				continue
			}
			if args[1] != nil && at.Before(args[1].(time.Time)) {
				continue
			}
			if args[2] != nil && !at.Before(args[2].(time.Time)) {
				continue
			}
			e := events[i]
			values = append(values, []any{int64(i + 1), e.kind, at, e.hash, []byte(e.payload)})
		}
	default:
		d.t.Errorf("unexpected mock query: %s", sql)
		return nil, errors.New("unexpected SQL")
	}
	return &mockRows{values: values, err: d.streamErr, scanErr: d.failAt == "scan"}, nil
}

type mockTx struct {
	pgx.Tx
	db      *mockDB
	pending *mockEvent
	done    bool
}

func (tx *mockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if err := ctx.Err(); err != nil {
		return pgconn.CommandTag{}, err
	}
	switch sql {
	case "SELECT pg_advisory_xact_lock(878701)":
		if tx.db.failAt == "lock" {
			return pgconn.CommandTag{}, errors.New("mock lock failure")
		}
	case "INSERT INTO events(id,kind,payload,previous_hash,hash) VALUES($1,$2,$3::jsonb,$4,$5)":
		if tx.db.failAt == "insert" {
			return pgconn.CommandTag{}, errors.New("mock insert failure")
		}
		if len(args) != 5 {
			return pgconn.CommandTag{}, errors.New("wrong insert arguments")
		}
		e := mockEvent{args[1].(string), args[2].(string), args[3].(string), args[4].(string)}
		previous := ""
		if len(tx.db.events) > 0 {
			previous = tx.db.events[len(tx.db.events)-1].hash
		}
		if e.previous != previous || e.hash != checksum(previous, e.kind, e.payload) {
			return pgconn.CommandTag{}, errors.New("invalid inserted chain")
		}
		tx.pending = &e
	default:
		tx.db.t.Errorf("unexpected mock exec: %s", sql)
		return pgconn.CommandTag{}, errors.New("unexpected SQL")
	}
	return pgconn.NewCommandTag("SELECT 1"), nil
}
func (tx *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	switch sql {
	case "SELECT hash FROM events ORDER BY sequence DESC LIMIT 1":
		if tx.db.failAt == "previous" {
			return mockRow{err: errors.New("mock previous failure")}
		}
		if len(tx.db.events) == 0 {
			return mockRow{err: pgx.ErrNoRows}
		}
		return mockRow{values: []any{tx.db.events[len(tx.db.events)-1].hash}}
	case "SELECT $1::jsonb::text":
		if tx.db.failAt == "canonical" {
			return mockRow{err: errors.New("mock canonical failure")}
		}
		var value any
		if err := json.Unmarshal([]byte(args[0].(string)), &value); err != nil {
			return mockRow{err: err}
		}
		// Deliberately return differently spaced JSON to check that append hashes the
		// database result, rather than the original marshalled input.
		raw, err := json.MarshalIndent(value, "", " ")
		return mockRow{values: []any{string(raw)}, err: err}
	default:
		tx.db.t.Errorf("unexpected mock row query: %s", sql)
		return mockRow{err: errors.New("unexpected SQL")}
	}
}
func (tx *mockTx) Commit(ctx context.Context) error {
	if tx.done {
		return pgx.ErrTxClosed
	}
	tx.done = true
	defer tx.db.mu.Unlock()
	if tx.db.failAt == "commit" {
		return errors.New("mock commit failure")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if tx.pending != nil {
		tx.db.events = append(tx.db.events, *tx.pending)
	}
	return nil
}
func (tx *mockTx) Rollback(context.Context) error {
	if tx.done {
		return pgx.ErrTxClosed
	}
	tx.done = true
	tx.db.mu.Unlock()
	return nil
}

type mockRow struct {
	values []any
	err    error
}

func (r mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return errors.New("mock scan column mismatch")
	}
	for i, v := range r.values {
		target := reflect.ValueOf(dest[i])
		value := reflect.ValueOf(v)
		if target.Kind() != reflect.Pointer || !value.Type().AssignableTo(target.Elem().Type()) {
			return fmt.Errorf("mock scan type mismatch %T", dest[i])
		}
		target.Elem().Set(value)
	}
	return nil
}

type mockRows struct {
	pgx.Rows
	values          [][]any
	index           int
	err             error
	closed, scanErr bool
}

func (r *mockRows) Next() bool {
	if r.closed {
		return false
	}
	if r.index >= len(r.values) {
		r.Close()
		return false
	}
	r.index++
	return true
}
func (r *mockRows) Scan(dest ...any) error {
	if r.scanErr {
		return errors.New("mock scan failure")
	}
	return (mockRow{values: r.values[r.index-1]}).Scan(dest...)
}
func (r *mockRows) Close()     { r.closed = true }
func (r *mockRows) Err() error { return r.err }

func TestAppendAtomicity(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"begin", "lock", "previous", "canonical", "insert", "commit"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			db := newMockDB(t)
			db.failAt = stage
			s := server{db: db}
			if s.append(context.Background(), "message.created", map[string]string{"author": "Human"}) == nil {
				t.Fatal("failure lost")
			}
			if len(db.snapshot()) != 0 {
				t.Fatal("failed transaction persisted an event")
			}
		})
	}
	t.Run("marshal", func(t *testing.T) {
		t.Parallel()
		db := newMockDB(t)
		s := server{db: db}
		if s.append(context.Background(), "bad", make(chan int)) == nil {
			t.Fatal("invalid JSON accepted")
		}
		if len(db.snapshot()) != 0 {
			t.Fatal("event persisted")
		}
	})
}
func TestConcurrentAppendChain(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	s := server{db: db}
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := s.append(context.Background(), "message.created", map[string]int{"turn": i}); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	events := db.snapshot()
	if len(events) != 24 {
		t.Fatalf("lost writes: %d", len(events))
	}
	previous := ""
	for _, e := range events {
		if e.previous != previous || e.hash != checksum(previous, e.kind, e.payload) {
			t.Fatal("broken chain")
		}
		if !strings.Contains(e.payload, "\n") {
			t.Fatal("database canonical result not used")
		}
		previous = e.hash
	}
}
