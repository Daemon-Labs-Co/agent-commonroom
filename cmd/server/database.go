package main

import (
	"context"
	"github.com/jackc/pgx/v5"
)

// eventDatabase is the persistence boundary used by the application. Production
// uses pgxpool.Pool; tests supply a mock with no network or database connection.
type eventDatabase interface {
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
}
