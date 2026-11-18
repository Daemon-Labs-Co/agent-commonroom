package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Recording fakes: these tests never touch a database.
type scopeLog struct{ ops []string }

type scopeTx struct {
	pgx.Tx
	log       *scopeLog
	failScope bool
	failQuery bool
}

func (tx *scopeTx) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if sql == setWorkspaceSQL {
		tx.log.ops = append(tx.log.ops, "scope:"+args[0].(string))
		if tx.failScope {
			return pgconn.CommandTag{}, errors.New("secret detail: password=hunter2")
		}
		return pgconn.NewCommandTag("SELECT 1"), nil
	}
	tx.log.ops = append(tx.log.ops, "exec:"+sql)
	return pgconn.NewCommandTag("OK"), nil
}
func (tx *scopeTx) Query(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	tx.log.ops = append(tx.log.ops, "query:"+sql)
	if tx.failQuery {
		return nil, errors.New("query failed")
	}
	return &scopeRows{log: tx.log, left: 2}, nil
}
func (tx *scopeTx) Rollback(ctx context.Context) error {
	if ctx.Err() != nil {
		tx.log.ops = append(tx.log.ops, "rollback-with-cancelled-context")
	}
	tx.log.ops = append(tx.log.ops, "rollback")
	return nil
}

type scopeRows struct {
	pgx.Rows
	log  *scopeLog
	left int
}

func (r *scopeRows) Next() bool { r.left--; return r.left >= 0 }
func (r *scopeRows) Close()     { r.log.ops = append(r.log.ops, "rows-closed") }

type scopeInner struct {
	log       *scopeLog
	failScope bool
	failQuery bool
}

func (d scopeInner) Begin(context.Context) (pgx.Tx, error) {
	d.log.ops = append(d.log.ops, "begin")
	return &scopeTx{log: d.log, failScope: d.failScope, failQuery: d.failQuery}, nil
}
func (d scopeInner) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unscoped query reached the pool")
}

func TestScopedDatabaseSetsWorkspaceFirstInEveryTransaction(t *testing.T) {
	t.Parallel()
	log := &scopeLog{}
	db, err := newScopedDatabase(scopeInner{log: log}, "acme")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	tx.Exec(t.Context(), "SELECT 1")
	if got := log.ops[:3]; !reflect.DeepEqual(got, []string{"begin", "scope:acme", "exec:SELECT 1"}) {
		t.Fatal("workspace was not fixed before other statements", log.ops)
	}
}

func TestScopedQueryScopesThenEndsTransactionOnceWhenExhaustedOrClosed(t *testing.T) {
	t.Parallel()
	log := &scopeLog{}
	db, _ := newScopedDatabase(scopeInner{log: log}, "acme")
	rows, err := db.Query(t.Context(), "SELECT kind FROM events")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
	}
	rows.Close() // closing after exhaustion must not roll back twice
	rows.Close()
	want := []string{"begin", "scope:acme", "query:SELECT kind FROM events", "rows-closed", "rollback"}
	if !reflect.DeepEqual(log.ops, want) {
		t.Fatal(log.ops)
	}
	// Early close ends the transaction too, even when the request context was cancelled.
	log.ops = nil
	ctx, cancel := context.WithCancel(t.Context())
	rows, _ = db.Query(ctx, "SELECT 2")
	rows.Next()
	cancel()
	rows.Close()
	if log.ops[len(log.ops)-1] != "rollback" || containsOp(log.ops, "rollback-with-cancelled-context") {
		t.Fatal("transaction leaked or rolled back with a dead context", log.ops)
	}
}

func containsOp(ops []string, want string) bool {
	for _, op := range ops {
		if op == want {
			return true
		}
	}
	return false
}

func TestScopedDatabaseFailsClosed(t *testing.T) {
	t.Parallel()
	log := &scopeLog{}
	db, _ := newScopedDatabase(scopeInner{log: log, failScope: true}, "acme")
	if _, err := db.Query(t.Context(), "SELECT 1"); err == nil || strings.Contains(err.Error(), "hunter2") {
		t.Fatal("scope failure accepted or leaked detail:", err)
	}
	if containsOp(log.ops, "query:SELECT 1") || log.ops[len(log.ops)-1] != "rollback" {
		t.Fatal("a query ran without its workspace scope", log.ops)
	}
	log.ops = nil
	failing, _ := newScopedDatabase(scopeInner{log: log, failQuery: true}, "acme")
	if _, err := failing.Query(t.Context(), "SELECT 1"); err == nil || log.ops[len(log.ops)-1] != "rollback" {
		t.Fatal("failed query left its transaction open", log.ops)
	}
}

func TestScopedDatabasesNeverShareScope(t *testing.T) {
	t.Parallel()
	log := &scopeLog{}
	a, _ := newScopedDatabase(scopeInner{log: log}, "alpha")
	b, _ := newScopedDatabase(scopeInner{log: log}, "beta")
	a.Begin(t.Context())
	b.Begin(t.Context())
	a.Begin(t.Context())
	var scopes []string
	for _, op := range log.ops {
		if strings.HasPrefix(op, "scope:") {
			scopes = append(scopes, op)
		}
	}
	if !reflect.DeepEqual(scopes, []string{"scope:alpha", "scope:beta", "scope:alpha"}) {
		t.Fatal(scopes)
	}
}

func TestWorkspaceIDValidation(t *testing.T) {
	t.Parallel()
	for _, ok := range []string{"default", "a", "acme", "acme-co", "a1-b2", strings.Repeat("a", 63)} {
		if !validWorkspaceID(ok) {
			t.Fatal("rejected", ok)
		}
	}
	for _, bad := range []string{"", "A", "Acme", "-a", "a-", "a_b", "a.b", "../x", "a b", "a;b", "a'b", "é", strings.Repeat("a", 64), "default\n"} {
		if validWorkspaceID(bad) {
			t.Fatal("accepted", bad)
		}
		if _, err := newScopedDatabase(scopeInner{log: &scopeLog{}}, bad); err == nil {
			t.Fatal("scoped database accepted", bad)
		}
	}
	if _, err := newScopedDatabase(nil, "acme"); err == nil {
		t.Fatal("nil inner database accepted")
	}
}

func TestVerifyWorkspaceStorage(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name    string
		row     []any
		present bool
		fail    bool
	}{
		{"before the migration the previous layout is served", []any{false, false, false, false}, false, false},
		{"protected", []any{true, true, false, false}, true, false},
		{"row-level security off", []any{true, false, false, false}, true, true},
		{"role bypasses row-level security", []any{true, true, true, false}, true, true},
		{"role owns the table", []any{true, true, false, true}, true, true},
	} {
		db := newMockDB(t)
		db.workspaceStorage = c.row
		got, err := verifyWorkspaceStorage(t.Context(), db)
		if (err != nil) != c.fail || got.Present != c.present || got.Protected != (c.present && !c.fail) {
			t.Fatal(c.name, got, err)
		}
	}
}

func TestWorkspaceMigrationIsIdempotentAndNeverMutatesHistory(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../migrations/002.sql")
	if err != nil {
		t.Fatal(err)
	}
	var code []string
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			code = append(code, line)
		}
	}
	sql := strings.Join(code, "\n")
	for _, required := range []string{
		"ADD COLUMN IF NOT EXISTS workspace_id TEXT NOT NULL DEFAULT 'default'",
		"SET DEFAULT NULLIF(current_setting('app.workspace_id'), '')",
		"ADD CONSTRAINT events_workspace_id_format",
		"CREATE INDEX IF NOT EXISTS events_workspace_sequence ON events (workspace_id, sequence)",
		"ENABLE ROW LEVEL SECURITY",
		"DROP POLICY IF EXISTS workspace_isolation ON events",
		"USING (workspace_id = NULLIF(current_setting('app.workspace_id', true), ''))",
		"WITH CHECK (workspace_id = NULLIF(current_setting('app.workspace_id', true), ''))",
	} {
		if !strings.Contains(sql, required) {
			t.Fatal("migration missing:", required)
		}
	}
	for _, forbidden := range []string{"DELETE", "TRUNCATE", "UPDATE ", "DROP TABLE", "DROP TRIGGER", "FORCE ROW", "DISABLE TRIGGER", "DROP COLUMN"} {
		if strings.Contains(strings.ToUpper(sql), forbidden) {
			t.Fatal("migration must not contain", forbidden)
		}
	}
}

func TestApplyMigrationsOrderAndFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for name, body := range map[string]string{"002.sql": "two", "001.sql": "one", "010.sql": "ten", "notes.sql": "ignored", "1.sql": "ignored"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var ran []string
	if err := applyMigrations(t.Context(), dir, func(_ context.Context, sql string) error { ran = append(ran, sql); return nil }); err != nil || !reflect.DeepEqual(ran, []string{"one", "two", "ten"}) {
		t.Fatal(ran, err)
	}
	ran = nil
	err := applyMigrations(t.Context(), dir, func(_ context.Context, sql string) error {
		ran = append(ran, sql)
		return errors.New("postgres://user:hunter2@host/db")
	})
	if err == nil || strings.Contains(err.Error(), "hunter2") || len(ran) != 1 {
		t.Fatal("failure did not stop at the first file or leaked detail:", err, ran)
	}
	if err := applyMigrations(t.Context(), t.TempDir(), func(context.Context, string) error { return nil }); err == nil {
		t.Fatal("empty migration directory accepted")
	}
	// The repository ships both files, applied in order.
	var shipped []string
	if err := applyMigrations(t.Context(), "../../migrations", func(_ context.Context, sql string) error {
		shipped = append(shipped, strings.Fields(sql)[0]+" "+strings.Fields(sql)[1])
		return nil
	}); err != nil || len(shipped) < 2 || shipped[0] != "CREATE TABLE" {
		t.Fatal(shipped, err)
	}
}

func TestAppendLockIsPerWorkspace(t *testing.T) {
	t.Parallel()
	for _, part := range []string{"pg_advisory_xact_lock(878701,", "hashtext(", "current_setting('app.workspace_id', true)"} {
		if !strings.Contains(appendLockSQL, part) {
			t.Fatal("append lock is not keyed by workspace:", appendLockSQL)
		}
	}
}
