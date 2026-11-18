package main

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	workspaceActive    = "active"
	workspaceSuspended = "suspended"
	workspaceDeleted   = "deleted"

	workspaceLookupSQL    = "SELECT id,name,status,limits::text,created_at FROM workspaces WHERE id=$1"
	workspaceListSQL      = "SELECT id,name,status,limits::text,created_at FROM workspaces WHERE id>$1 ORDER BY id LIMIT $2"
	workspaceInsertSQL    = "INSERT INTO workspaces(id,name,limits) VALUES($1,$2,$3::jsonb) ON CONFLICT (id) DO NOTHING"
	workspaceStatusSQL    = "UPDATE workspaces SET status=$2, updated_at=clock_timestamp() WHERE id=$1 AND status<>'deleted' AND status<>$2"
	workspaceSetLimitsSQL = "UPDATE workspaces SET limits=$2::jsonb, updated_at=clock_timestamp() WHERE id=$1 AND status<>'deleted'"
)

type workspaceRecord struct {
	ID, Name, Status string
	Limits           workspaceLimits
	Created          time.Time
}

// workspaceRegistry is the one cross-workspace lookup. Workspace servers never
// receive it; only the host that routes requests and the platform API do.
type workspaceRegistry interface {
	Lookup(ctx context.Context, id string) (workspaceRecord, bool, error)
}

// workspaceStore adds the platform operator's writes. Deleted is terminal: a
// deleted workspace keeps its row as a tombstone and its id is never reused.
type workspaceStore interface {
	workspaceRegistry
	Create(ctx context.Context, id, name string, limits workspaceLimits) (created bool, err error)
	SetStatus(ctx context.Context, id, status string) (changed bool, err error)
	SetLimits(ctx context.Context, id string, limits workspaceLimits) (changed bool, err error)
	List(ctx context.Context, after string, limit int) ([]workspaceRecord, error)
}

var errRegistryUnavailable = errors.New("workspace registry unavailable")

// validWorkspaceName admits a visible name of 1-100 characters without controls.
func validWorkspaceName(name string) bool {
	if name != strings.TrimSpace(name) || name == "" || utf8.RuneCountInString(name) > 100 || !utf8.ValidString(name) {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == ' ' || r == ' ' {
			return false
		}
	}
	return true
}

type databaseWorkspaceRegistry struct {
	db eventDatabase
}

func scanWorkspace(scan func(...any) error) (workspaceRecord, error) {
	var rec workspaceRecord
	var limits string
	if scan(&rec.ID, &rec.Name, &rec.Status, &limits, &rec.Created) != nil {
		return workspaceRecord{}, errRegistryUnavailable
	}
	parsed, err := parseWorkspaceLimits([]byte(limits))
	if err != nil {
		return workspaceRecord{}, errRegistryUnavailable
	}
	rec.Limits = parsed
	return rec, nil
}

func (r databaseWorkspaceRegistry) Lookup(ctx context.Context, id string) (workspaceRecord, bool, error) {
	if !validHostedWorkspaceID(id) {
		return workspaceRecord{}, false, nil
	}
	rows, err := r.db.Query(ctx, workspaceLookupSQL, id)
	if err != nil {
		return workspaceRecord{}, false, errRegistryUnavailable
	}
	defer rows.Close()
	if !rows.Next() {
		if rows.Err() != nil {
			return workspaceRecord{}, false, errRegistryUnavailable
		}
		return workspaceRecord{}, false, nil
	}
	rec, err := scanWorkspace(rows.Scan)
	return rec, err == nil, err
}

func (r databaseWorkspaceRegistry) List(ctx context.Context, after string, limit int) ([]workspaceRecord, error) {
	if limit < 1 || limit > 100 || len(after) > 63 {
		return nil, errors.New("invalid list bounds")
	}
	rows, err := r.db.Query(ctx, workspaceListSQL, after, limit)
	if err != nil {
		return nil, errRegistryUnavailable
	}
	defer rows.Close()
	out := []workspaceRecord{}
	for rows.Next() {
		rec, err := scanWorkspace(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	if rows.Err() != nil {
		return nil, errRegistryUnavailable
	}
	return out, nil
}

// exec runs one write in its own transaction.
func (r databaseWorkspaceRegistry) exec(ctx context.Context, sql string, args ...any) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, errRegistryUnavailable
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return 0, errRegistryUnavailable
	}
	if tx.Commit(ctx) != nil {
		return 0, errRegistryUnavailable
	}
	return tag.RowsAffected(), nil
}

func (r databaseWorkspaceRegistry) Create(ctx context.Context, id, name string, limits workspaceLimits) (bool, error) {
	if !validHostedWorkspaceID(id) || !validWorkspaceName(name) {
		return false, errors.New("invalid workspace")
	}
	raw, _ := jsonMarshal(limits)
	n, err := r.exec(ctx, workspaceInsertSQL, id, name, string(raw))
	return n == 1, err
}

func (r databaseWorkspaceRegistry) SetStatus(ctx context.Context, id, status string) (bool, error) {
	if !validHostedWorkspaceID(id) || (status != workspaceActive && status != workspaceSuspended && status != workspaceDeleted) {
		return false, errors.New("invalid workspace status change")
	}
	n, err := r.exec(ctx, workspaceStatusSQL, id, status)
	return n == 1, err
}

func (r databaseWorkspaceRegistry) SetLimits(ctx context.Context, id string, limits workspaceLimits) (bool, error) {
	if !validHostedWorkspaceID(id) {
		return false, errors.New("invalid workspace")
	}
	raw, _ := jsonMarshal(limits)
	n, err := r.exec(ctx, workspaceSetLimitsSQL, id, string(raw))
	return n == 1, err
}
