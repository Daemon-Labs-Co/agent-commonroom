# cmd/server/database_mock_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/database_mock_test.go)

Automated regression checks for database_mock; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `mockEvent`**: This mock accepts only the application's event-store operations. Fixtures and transaction writes exist solely in memory. Unexpected SQL fails the test.
  Declaration: `struct { kind, payload, previous, hash string id string sequence int64 at time.Time }`.
- **Type `mockDB`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T mu sync.Mutex events []mockEvent failAt string failInsertAt int streamErr error runtimePrivileges []any }`.
- **`newMockDB`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) *mockDB`.
- **`(*mockDB).snapshot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() []mockEvent`.
  Calls: `append`, `d.mu.Lock`, `d.mu.Unlock`.
- **`(*mockDB).seed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(kind string, value any)`.
  Calls: `d.seedRecord`.
- **`(*mockDB).seedLegacy`** — seedLegacy deliberately bypasses the format barrier for downgrade regressions.
  Signature: `func(kind string, value any)`.
  Calls: `d.seedRecord`.
- **`(*mockDB).seedRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(kind string, value any, legacy bool)`.
  Calls: `append`, `checksum`, `d.mu.Lock`, `d.mu.Unlock`, `d.t.Fatal`, `fmt.Sprintf`, `int64`, `json.Marshal`, `json.Unmarshal`, `len`, `protectAuditMetadata`, `string`, `time.Unix`, `time.Unix(int64(len(d.events)), 0).UTC`.
- **`(*mockDB).Begin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (pgx.Tx, error)`.
  Calls: `ctx.Err`, `d.mu.Lock`, `errors.New`.
- **`(*mockDB).Query`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)`.
  Calls: `append`, `at.Before`, `ctx.Err`, `d.snapshot`, `d.t.Errorf`, `errors.New`, `int64`, `json.Unmarshal`, `len`.
- **Type `mockTx`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { pgx.Tx db *mockDB pending []mockEvent done bool }`.
- **`(*mockTx).Exec`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)`.
  Calls: `append`, `checksum`, `ctx.Err`, `errors.New`, `len`, `pgconn.NewCommandTag`, `tx.db.t.Errorf`.
- **`(*mockTx).QueryRow`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, sql string, args ...any) pgx.Row`.
  Calls: `errors.New`, `int64`, `json.MarshalIndent`, `json.Unmarshal`, `len`, `string`, `tx.db.t.Errorf`.
- **`(*mockTx).Commit`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `append`, `ctx.Err`, `errors.New`, `tx.db.mu.Unlock`.
- **`(*mockTx).Rollback`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context) error`.
  Calls: `tx.db.mu.Unlock`.
- **Type `mockRow`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { values []any err error }`.
- **`(mockRow).Scan`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dest ...any) error`.
  Calls: `errors.New`, `fmt.Errorf`, `len`, `reflect.ValueOf`, `target.Elem`, `target.Elem().Set`, `target.Elem().Type`, `target.Kind`, `value.Type`, `value.Type().AssignableTo`.
- **Type `mockRows`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { pgx.Rows values [][]any index int err error closed, scanErr bool }`.
- **`(*mockRows).Next`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() bool`.
  Calls: `len`, `r.Close`.
- **`(*mockRows).Scan`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dest ...any) error`.
  Calls: `(mockRow{values: r.values[r.index-1]}).Scan`, `errors.New`.
- **`(*mockRows).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`(*mockRows).Err`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
- **`TestAppendAtomicity`** — Regression contract: TestAppendAtomicity.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.snapshot`, `len`, `make`, `newMockDB`, `s.append`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestConcurrentAppendChain`** — Regression contract: TestConcurrentAppendChain.
  Signature: `func(t *testing.T)`.
  Calls: `checksum`, `context.Background`, `db.snapshot`, `len`, `newMockDB`, `s.append`, `strings.Contains`, `t.Error`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
