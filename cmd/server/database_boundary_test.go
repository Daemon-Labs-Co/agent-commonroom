package main

import (
	"context"
	"testing"
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
