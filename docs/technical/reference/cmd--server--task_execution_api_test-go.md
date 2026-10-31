# cmd/server/task_execution_api_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_execution_api_test.go)

Automated regression checks for task_execution_api; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`executionHistoryHandler`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *server) http.Handler`.
  Calls: `http.NewServeMux`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`.
- **`TestTaskExecutionHistoryPaginationAndScopedAccess`** — Regression contract: TestTaskExecutionHistoryPaginationAndScopedAccess.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `definedDispatchBody`, `dispatchRequest`, `executionHistoryHandler`, `fmt.Sprintf`, `json.Unmarshal`, `len`, `mockRequest`, `strings.Contains`, `structuredTaskFixture`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.Bytes`, `w.Body.String`, `waitDispatch`.
- **`TestTaskExecutionPersonalRoomReadsRetainOwnerAudit`** — Regression contract: TestTaskExecutionPersonalRoomReadsRetainOwnerAudit.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `executionHistoryHandler`, `len`, `mockRequest`, `mockServer`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestTaskContractPersistenceFailuresDoNotInventExecutionsOrOutputs`** — Regression contract: TestTaskContractPersistenceFailuresDoNotInventExecutionsOrOutputs.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `definedDispatchBody`, `dispatchRequest`, `len`, `make`, `structuredTaskFixture`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`, `waitDispatch`.
- **`TestTaskDefinitionFlagDoesNotPoisonLegacyJobEpoch`** — Regression contract: TestTaskDefinitionFlagDoesNotPoisonLegacyJobEpoch.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `dispatchFixture`, `dispatchRequest`, `jobCtx.Value`, `make`, `reg.Register`, `reg.Update`, `s.managedJobContext`, `t.Fatal`, `t.Parallel`, `time.After`, `w.Body.String`, `waitDispatch`.
