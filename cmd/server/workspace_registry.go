package main

import (
	"context"
	"errors"
)

const (
	workspaceActive    = "active"
	workspaceSuspended = "suspended"
	workspaceDeleted   = "deleted"

	workspaceLookupSQL = "SELECT id,name,status FROM workspaces WHERE id=$1"
)

type workspaceRecord struct {
	ID, Name, Status string
}

// workspaceRegistry is the one cross-workspace lookup. Workspace servers never
// receive it; only the host that routes requests does.
type workspaceRegistry interface {
	Lookup(ctx context.Context, id string) (workspaceRecord, bool, error)
}

type databaseWorkspaceRegistry struct{ db catalogQuerier }

func (r databaseWorkspaceRegistry) Lookup(ctx context.Context, id string) (workspaceRecord, bool, error) {
	if !validHostedWorkspaceID(id) {
		return workspaceRecord{}, false, nil
	}
	rows, err := r.db.Query(ctx, workspaceLookupSQL, id)
	if err != nil {
		return workspaceRecord{}, false, errors.New("workspace registry unavailable")
	}
	defer rows.Close()
	if !rows.Next() {
		return workspaceRecord{}, false, rows.Err()
	}
	var rec workspaceRecord
	if rows.Scan(&rec.ID, &rec.Name, &rec.Status) != nil {
		return workspaceRecord{}, false, errors.New("workspace registry unavailable")
	}
	return rec, true, nil
}
