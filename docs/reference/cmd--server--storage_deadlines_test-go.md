# cmd/server/storage_deadlines_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/storage_deadlines_test.go)

Automated regression checks for storage_deadlines; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `stalledDatabase`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *mockDB kind string started chan struct{} once sync.Once }`.
- **`(*stalledDatabase).signal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `close`, `d.once.Do`.
- **`(*stalledDatabase).Query`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)`.
  Calls: `ctx.Done`, `ctx.Err`, `d.mockDB.Query`, `d.signal`.
- **`(*stalledDatabase).Begin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (pgx.Tx, error)`.
  Calls: `ctx.Done`, `ctx.Err`, `d.mockDB.Begin`, `d.signal`.
- **Type `stalledRows`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { pgx.Rows ctx context.Context }`.
- **`(stalledRows).Next`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() bool`.
  Calls: `r.ctx.Done`.
- **`(stalledRows).Err`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `r.ctx.Err`.
- **Type `stalledTransaction`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ pgx.Tx }`.
- **`(stalledTransaction).Exec`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)`.
  Calls: `ctx.Done`, `ctx.Err`, `tx.Tx.Exec`.
- **`TestStorageWorkHasFiniteDeadline`** — Regression contract: TestStorageWorkHasFiniteDeadline.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.snapshot`, `errors.Is`, `len`, `mockServer`, `s.append`, `s.load`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Since`.
- **`TestStorageStallReleasesControllerForEmergencyStop`** — Regression contract: TestStorageStallReleasesControllerForEmergencyStop.
  Signature: `func(t *testing.T)`.
  Calls: `agent.Err`, `cancel`, `context.Background`, `context.WithCancel`, `errors.Is`, `http.NewServeMux`, `make`, `mockRequest`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `t.Fatal`, `t.Parallel`, `time.After`, `w.Body.String`.
- **`TestIntegrityReadTimeoutStopsJobs`** — Regression contract: TestIntegrityReadTimeoutStopsJobs.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `httptest.NewRecorder`, `httptest.NewRequest`, `job.Err`, `mockServer`, `s.integrity`, `t.Fatal`, `t.Parallel`.
- **`TestStorageDeadlineCannotExtendCallerOrDisableCap`** — Regression contract: TestStorageDeadlineCannotExtendCallerOrDisableCap.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `ctx.Deadline`, `done`, `parent.Deadline`, `pd.Equal`, `s.storageContext`, `t.Fatal`, `t.Parallel`, `time.Until`.
