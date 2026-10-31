# cmd/server/database_boundary.go

[Module index](cmd.md) · [Open source](../../../cmd/server/database_boundary.go)

Go implementation for database boundary; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `catalogQuerier`**: Behavior contract; required method signatures are shown in the declaration.
  Declaration: `interface { Query(context.Context, string, ...any) (pgx.Rows, error) }`.
- **`verifyRuntimeDatabaseStartup`** — verifyRuntimeDatabaseStartup bounds connection and privilege verification as one operation. Never log the raw connection error: drivers/backends may include credentials or untrusted server detail. Unknown privileges refuse startup.
  Signature: `func(ctx context.Context, ping func(context.Context) error, db catalogQuerier) error`.
  Calls: `cancel`, `context.WithTimeout`, `ctx.Err`, `errors.New`, `ping`, `validateRuntimeDatabase`.
- **`runtimeDatabaseConfig`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(url string) (*pgxpool.Config, error)`.
  Calls: `errors.New`, `pgxpool.ParseConfig`.
- **`validateRuntimeDatabase`** — validateRuntimeDatabase prevents the serving process from using a migration owner or an account that can mutate audit rows/protections. This has no off switch: a source fork CAN bypass it, BUT undermines append-only guarantees.
  Signature: `func(ctx context.Context, db catalogQuerier) error`.
  Calls: `db.Query`, `errors.New`, `rows.Close`, `rows.Err`, `rows.Next`, `rows.Scan`.
- **`runDatabaseMigration`** — runDatabaseMigration uses a distinct administrator-supplied URL. It never falls back to runtime credentials, and emits no credential-bearing errors.
  Signature: `func(ctx context.Context) error`.
  Calls: `db.Close`, `db.Exec`, `errors.New`, `os.Getenv`, `os.ReadFile`, `pgxpool.New`, `string`.
