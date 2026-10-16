# cmd/server/task_dispatch_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/task_dispatch_test.go)

Automated regression checks for task_dispatch; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`dispatchFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler, chan struct{})`.
  Calls: `ctx.Done`, `ctx.Err`, `db.seed`, `http.NewServeMux`, `make`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`.
- **`dispatchRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, key, body string) *httptest.ResponseRecorder`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `r.AddCookie`, `r.Header.Set`, `strings.NewReader`.
- **`waitDispatch`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, s *server) room.State`.
  Calls: `context.Background`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `t.Fatal`, `t.Helper`, `time.Now`, `time.Now().Add`, `time.Now().Before`, `time.Sleep`.
- **`TestTaskDispatchAtomicReplayAndCompletion`** — Regression contract: TestTaskDispatchAtomicReplayAndCompletion.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `context.Background`, `db.snapshot`, `dispatchFixture`, `dispatchRequest`, `len`, `replay.Header`, `replay.Header().Get`, `s.load`, `t.Fatal`, `t.Parallel`, `w.Body.String`, `waitDispatch`.
- **`TestTaskDispatchRollbackNeverLaunches`** — Regression contract: TestTaskDispatchRollbackNeverLaunches.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `dispatchFixture`, `dispatchRequest`, `len`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestTaskDispatchRejections`** — Regression contract: TestTaskDispatchRejections.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `dispatchFixture`, `dispatchRequest`, `errors.New`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `r.AddCookie`, `r.Header.Set`, `rune`, `string`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestTaskFailureAndReconciliation`** — Regression contract: TestTaskFailureAndReconciliation.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `db.seed`, `dispatchFixture`, `dispatchRequest`, `errors.New`, `t.Fatal`, `t.Parallel`, `w.Body.String`, `waitDispatch`.
- **`TestTaskEditCannotForgeOrChangePendingDispatch`** — Regression contract: TestTaskEditCannotForgeOrChangePendingDispatch.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `dispatchFixture`, `json.Unmarshal`, `len`, `mockRequest`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestTaskResultWriteFailureLeavesIntentPendingAndStopsExecution`** — Regression contract: TestTaskResultWriteFailureLeavesIntentPendingAndStopsExecution.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `dispatchFixture`, `dispatchRequest`, `s.mu.Lock`, `s.mu.Unlock`, `t.Fatal`, `t.Parallel`, `w.Body.String`, `waitDispatch`.
- **`TestTaskDispatchFlagOffPreservesInflightResult`** — Regression contract: TestTaskDispatchFlagOffPreservesInflightResult.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `db.seed`, `dispatchFixture`, `dispatchRequest`, `s.mu.Lock`, `s.mu.Unlock`, `t.Fatal`, `t.Parallel`, `w.Body.String`, `waitDispatch`.
