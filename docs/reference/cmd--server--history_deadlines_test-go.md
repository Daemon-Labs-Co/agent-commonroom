# cmd/server/history_deadlines_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/history_deadlines_test.go)

Automated regression checks for history_deadlines; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `historyStallDatabase`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *mockDB phase string }`.
- **`(*historyStallDatabase).Query`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)`.
  Calls: `ctx.Done`, `ctx.Err`, `db.mockDB.Query`.
- **Type `partialHistoryRows`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { pgx.Rows ctx context.Context started bool }`.
- **`(*partialHistoryRows).Next`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() bool`.
  Calls: `rows.Rows.Next`, `rows.ctx.Done`.
- **`(*partialHistoryRows).Err`** — Deliberately omit a driver error to verify the handler checks context itself.
  Signature: `func() error`.
- **`TestHistoryStorageStallsNeverReturnPartialPages`** — Regression contract: TestHistoryStorageStallsNeverReturnPartialPages.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `httptest.NewRecorder`, `httptest.NewRequest`, `newMockDB`, `s.events`, `strings.Contains`, `strings.HasPrefix`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Since`, `w.Body.String`.
- **`TestHistorySettingReadPreservesCallerDeadline`** — Regression contract: TestHistorySettingReadPreservesCallerDeadline.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `newMockDB`, `s.historyPresentationEnabled`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Since`.
