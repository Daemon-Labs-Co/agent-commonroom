package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type historyStallDatabase struct {
	*mockDB
	phase string
}

func (db *historyStallDatabase) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if db.phase == "flag" || db.phase == "query" && sql != historyFlagSQL {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	rows, err := db.mockDB.Query(ctx, sql, args...)
	if err == nil && db.phase == "rows" && sql != historyFlagSQL {
		return &partialHistoryRows{Rows: rows, ctx: ctx}, nil
	}
	return rows, err
}

type partialHistoryRows struct {
	pgx.Rows
	ctx     context.Context
	started bool
}

func (rows *partialHistoryRows) Next() bool {
	if !rows.started {
		rows.started = true
		return rows.Rows.Next()
	}
	<-rows.ctx.Done()
	return false
}

// Deliberately omit a driver error to verify the handler checks context itself.
func (*partialHistoryRows) Err() error { return nil }

func TestHistoryStorageStallsNeverReturnPartialPages(t *testing.T) {
	t.Parallel()
	for _, phase := range []string{"flag", "query", "rows"} {
		for _, path := range []string{"/api/events", "/api/events?page=true"} {
			t.Run(phase+path, func(t *testing.T) {
				t.Parallel()
				db := &historyStallDatabase{mockDB: newMockDB(t), phase: phase}
				db.seed("message.created", map[string]string{"roomId": "engineering", "text": "synthetic historical body"})
				s := server{db: db, storageTimeout: 20 * time.Millisecond}
				w := httptest.NewRecorder()
				started := time.Now()
				s.events(w, httptest.NewRequest(http.MethodGet, path, nil))
				if w.Code != http.StatusInternalServerError {
					t.Fatal("stalled history accepted", w.Code, w.Body.String())
				}
				if time.Since(started) > time.Second {
					t.Fatal("history deadline not honored")
				}
				if strings.Contains(w.Body.String(), "synthetic historical body") || strings.HasPrefix(w.Body.String(), "[") || strings.HasPrefix(w.Body.String(), `{"events":`) {
					t.Fatal("partial history leaked")
				}
			})
		}
	}
}

func TestHistorySettingReadPreservesCallerDeadline(t *testing.T) {
	t.Parallel()
	db := &historyStallDatabase{mockDB: newMockDB(t), phase: "flag"}
	s := server{db: db}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	if enabled, err := s.historyPresentationEnabled(ctx); err == nil || enabled {
		t.Fatal("timed-out setting treated as valid", enabled, err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("caller deadline extended")
	}
}
