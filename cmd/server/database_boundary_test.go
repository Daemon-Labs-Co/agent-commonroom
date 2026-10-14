package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRuntimeDatabaseRequiresAppendOnlyPrivileges(t *testing.T) {
	t.Parallel()
	for _, bad := range []int{-1, 0, 1, 2, 3, 4, 5} {
		t.Run(string(rune('a'+bad+1)), func(t *testing.T) {
			t.Parallel()
			db := newMockDB(t)
			db.runtimePrivileges = []any{false, true, true, true, false, false}
			if bad >= 0 {
				db.runtimePrivileges[bad] = !db.runtimePrivileges[bad].(bool)
			}
			err := validateRuntimeDatabase(context.Background(), db)
			if (bad >= 0) != (err != nil) {
				t.Fatal("runtime privilege boundary failed", bad, err)
			}
		})
	}
}

func TestDatabaseStartupTimeoutAndCredentialSafeFailure(t *testing.T) {
	t.Parallel()
	// nil catalog proves a failed connection never attempts privilege queries.
	err := verifyRuntimeDatabaseStartup(context.Background(), func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 10*time.Second {
			t.Fatal("startup deadline missing")
		}
		return errors.New("postgres://synthetic:synthetic-secret@example.invalid/db")
	}, nil)
	if err == nil || strings.Contains(err.Error(), "synthetic-secret") {
		t.Fatal("connection error leaked or accepted", err)
	}
	for _, phase := range []string{"ping", "query", "rows"} {
		t.Run(phase, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			db := &stalledDatabase{mockDB: newMockDB(t), kind: phase}
			db.runtimePrivileges = []any{false, true, true, true, false, false}
			ping := func(ctx context.Context) error {
				if phase == "ping" {
					<-ctx.Done()
					return ctx.Err()
				}
				return nil
			}
			start := time.Now()
			if err := verifyRuntimeDatabaseStartup(ctx, ping, db); err == nil {
				t.Fatal("stalled startup accepted")
			}
			if time.Since(start) > time.Second {
				t.Fatal("startup exceeded caller deadline")
			}
		})
	}
}

func TestDatabaseStartupAcceptsVerifiedLeastPrivileges(t *testing.T) {
	t.Parallel()
	db := newMockDB(t)
	db.runtimePrivileges = []any{false, true, true, true, false, false}
	if err := verifyRuntimeDatabaseStartup(context.Background(), func(context.Context) error { return nil }, db); err != nil {
		t.Fatal(err)
	}
	db.runtimePrivileges[0] = true
	if err := verifyRuntimeDatabaseStartup(context.Background(), func(context.Context) error { return nil }, db); err == nil {
		t.Fatal("overprivileged startup accepted")
	}
}

func TestRuntimeDatabasePinsSearchPathAndFailsOnUnknownPrivileges(t *testing.T) {
	t.Parallel()
	config, err := runtimeDatabaseConfig("postgres://localhost/example?search_path=untrusted,public")
	if err != nil || config.ConnConfig.RuntimeParams["search_path"] != "pg_catalog,public" {
		t.Fatal("search path not pinned")
	}
	db := newMockDB(t)
	db.failAt = "query"
	if validateRuntimeDatabase(context.Background(), db) == nil {
		t.Fatal("unknown privileges accepted")
	}
}
