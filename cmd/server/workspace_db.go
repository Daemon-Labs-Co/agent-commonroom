package main

import (
	"context"
	"errors"
	"regexp"
	"sync"

	"github.com/jackc/pgx/v5"
)

// A workspace's history is isolated twice: application code only ever holds a
// database scoped to its own workspace, and PostgreSQL row-level security on
// events refuses any row whose workspace_id differs from the transaction's
// app.workspace_id setting. A query that forgets a filter therefore returns
// nothing from other workspaces instead of leaking them. See ADR 0004.
const (
	setWorkspaceSQL = "SELECT set_config('app.workspace_id', $1, true)"
	// One lock per workspace. Before the workspace migration the setting is unset,
	// every append hashes the same empty key and serialises exactly as before.
	appendLockSQL      = "SELECT pg_advisory_xact_lock(878701, hashtext(coalesce(current_setting('app.workspace_id', true), '')))"
	defaultWorkspaceID = "default"
)

var workspaceIDPattern = regexp.MustCompile(`^(default|[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)$`)

// validWorkspaceID admits only values that are safe as database setting values,
// directory names and DNS labels.
func validWorkspaceID(id string) bool { return workspaceIDPattern.MatchString(id) }

// scopedDatabase binds every transaction it opens to one workspace. It is the
// only way application code reaches events once workspace storage is in use.
type scopedDatabase struct {
	inner     eventDatabase
	workspace string
}

func newScopedDatabase(inner eventDatabase, workspace string) (*scopedDatabase, error) {
	if inner == nil || !validWorkspaceID(workspace) {
		return nil, errors.New("invalid workspace database scope")
	}
	return &scopedDatabase{inner: inner, workspace: workspace}, nil
}

// Begin starts a transaction whose first statement fixes the workspace. The
// setting is local to the transaction, so a pooled connection never carries one
// workspace's scope into another's work.
func (d *scopedDatabase) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := d.inner.Begin(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, setWorkspaceSQL, d.workspace); err != nil {
		_ = tx.Rollback(ctx)
		return nil, errors.New("workspace scope could not be applied")
	}
	return tx, nil
}

// Query runs one statement in a scoped transaction that ends when the rows are
// closed or exhausted.
func (d *scopedDatabase) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	return &scopedRows{Rows: rows, tx: tx, ctx: ctx}, nil
}

type scopedRows struct {
	pgx.Rows
	tx   pgx.Tx
	ctx  context.Context
	once sync.Once
}

func (r *scopedRows) finish() {
	r.once.Do(func() {
		r.Rows.Close()
		_ = r.tx.Rollback(context.WithoutCancel(r.ctx))
	})
}

func (r *scopedRows) Next() bool {
	if r.Rows.Next() {
		return true
	}
	r.finish()
	return false
}

func (r *scopedRows) Close() { r.finish() }
