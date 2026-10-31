# cmd/server/database.go

[Module index](cmd.md) · [Open source](../../../cmd/server/database.go)

Go implementation for database; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `eventDatabase`**: eventDatabase is the persistence boundary used by the application. Production uses pgxpool.Pool; tests supply a mock with no network or database connection.
  Declaration: `interface { Begin(context.Context) (pgx.Tx, error) Query(context.Context, string, ...any) (pgx.Rows, error) }`.
