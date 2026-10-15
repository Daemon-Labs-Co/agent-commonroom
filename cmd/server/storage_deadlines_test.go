package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type stalledDatabase struct {
	*mockDB
	kind    string
	started chan struct{}
	once    sync.Once
}

func (d *stalledDatabase) signal() {
	if d.started != nil {
		d.once.Do(func() { close(d.started) })
	}
}
func (d *stalledDatabase) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if d.kind == "query" {
		d.signal()
		<-ctx.Done()
		return nil, ctx.Err()
	}
	rows, err := d.mockDB.Query(ctx, sql, args...)
	if err == nil && d.kind == "rows" {
		return stalledRows{Rows: rows, ctx: ctx}, nil
	}
	return rows, err
}
func (d *stalledDatabase) Begin(ctx context.Context) (pgx.Tx, error) {
	if d.kind == "begin" {
		d.signal()
		<-ctx.Done()
		return nil, ctx.Err()
	}
	tx, err := d.mockDB.Begin(ctx)
	if err == nil && d.kind == "insert" {
		return stalledTransaction{Tx: tx}, nil
	}
	return tx, err
}

type stalledRows struct {
	pgx.Rows
	ctx context.Context
}

func (r stalledRows) Next() bool { <-r.ctx.Done(); return false }
func (r stalledRows) Err() error { return r.ctx.Err() }

type stalledTransaction struct{ pgx.Tx }

func (tx stalledTransaction) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if sql != "SELECT pg_advisory_xact_lock(878701)" {
		<-ctx.Done()
		return pgconn.CommandTag{}, ctx.Err()
	}
	return tx.Tx.Exec(ctx, sql, args...)
}
func TestStorageWorkHasFiniteDeadline(t *testing.T) {
	for _, kind := range []string{"query", "rows", "begin", "insert"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			s, db, _ := mockServer(t)
			s.storageTimeout = 20 * time.Millisecond
			s.db = &stalledDatabase{mockDB: db, kind: kind}
			var err error
			start := time.Now()
			if kind == "query" || kind == "rows" {
				_, err = s.load(context.Background())
			} else {
				err = s.append(context.Background(), "company.updated", map[string]string{"name": "Synthetic"})
			}
			if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
				t.Fatal("storage wait unbounded", kind, err)
			}
			if len(db.snapshot()) != 0 {
				t.Fatal("timeout committed partial data")
			}
		})
	}
}
func TestStorageStallReleasesControllerForEmergencyStop(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	started := make(chan struct{})
	s.db = &stalledDatabase{mockDB: db, kind: "query", started: started}
	s.storageTimeout = 30 * time.Millisecond
	agent, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.cancels["synthetic"] = cancel
	finished := make(chan error, 1)
	go func() { s.mu.Lock(); _, err := s.load(context.Background()); s.mu.Unlock(); finished <- err }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("stall never started")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/agent-safety", s.agentSafety)
	h := s.auth.protect(s.actorGuard(mux))
	w := mockRequest(h, "POST", "/api/agent-safety", `{"action":"stop","scope":"application","reason":"Synthetic storage outage"}`)
	if w.Code != 200 || agent.Err() != context.Canceled || !s.stopAll {
		t.Fatal("storage stall blocked stop", w.Code, w.Body.String())
	}
	if err := <-finished; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}
func TestIntegrityReadTimeoutStopsJobs(t *testing.T) {
	t.Parallel()
	s, db, _ := mockServer(t)
	s.db = &stalledDatabase{mockDB: db, kind: "query"}
	s.storageTimeout = 20 * time.Millisecond
	job, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.cancels["synthetic"] = cancel
	w := httptest.NewRecorder()
	s.integrity(w, httptest.NewRequest("GET", "/api/integrity", nil))
	if w.Code != 503 || !s.integrityFault || job.Err() != context.Canceled {
		t.Fatal("integrity deadline failed open", w.Code)
	}
}
func TestStorageDeadlineCannotExtendCallerOrDisableCap(t *testing.T) {
	t.Parallel()
	s := server{storageTimeout: time.Hour}
	parent, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	ctx, done := s.storageContext(parent)
	defer done()
	pd, _ := parent.Deadline()
	cd, _ := ctx.Deadline()
	if !pd.Equal(cd) {
		t.Fatal("extended caller deadline")
	}
	for _, value := range []time.Duration{0, -time.Second, time.Hour} {
		s.storageTimeout = value
		ctx, cancel := s.storageContext(context.Background())
		deadline, _ := ctx.Deadline()
		if time.Until(deadline) > maximumStorageWait || time.Until(deadline) < 4*time.Second {
			t.Fatal("cap bypass", value)
		}
		cancel()
	}
}
