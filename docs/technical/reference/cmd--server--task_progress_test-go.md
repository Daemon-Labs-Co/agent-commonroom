# cmd/server/task_progress_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_progress_test.go)

Automated regression checks for task_progress; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`progressHandler`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *server) http.Handler`.
  Calls: `http.NewServeMux`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`.
- **`readTaskProgress`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler) []taskProgressView`.
  Calls: `json.Unmarshal`, `mockRequest`, `t.Fatal`, `t.Helper`, `w.Body.Bytes`, `w.Body.String`.
- **`TestTaskProgressDescribesDurableLifecycleWithoutProbingWorkers`** — Regression contract: TestTaskProgressDescribesDurableLifecycleWithoutProbingWorkers.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `db.snapshot`, `definedDispatchBody`, `dispatchRequest`, `json.Marshal`, `len`, `make`, `mockRequest`, `progressHandler`, `readTaskProgress`, `reviseValueTask`, `s.load`, `string`, `strings.Contains`, `structuredTaskFixture`, `submitTaskAcceptance`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskAcceptanceBody`, `time.After`, `w.Body.String`, `waitDispatch`.
- **`TestTaskProgressRecordsStaleDependenciesAndDistinctAttempts`** — Regression contract: TestTaskProgressRecordsStaleDependenciesAndDistinctAttempts.
  Signature: `func(t *testing.T)`.
  Calls: `describeTaskProgress`, `dispatchRequest`, `json.Marshal`, `len`, `reviseValueTask`, `s.load`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `valueDependencyFixture`, `valueDispatchBody`, `w.Body.String`, `waitDispatch`.
- **`TestTaskProgressFiniteStatesLegacyDefinitionChangeAndFailedReview`** — Regression contract: TestTaskProgressFiniteStatesLegacyDefinitionChangeAndFailedReview.
  Signature: `func(t *testing.T)`.
  Calls: `describeTaskProgress`, `json.RawMessage`, `prepareTaskExecution`, `s.load`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskcontract.ReturnedValues`.
- **`TestTaskProgressScopePermissionsFlagsAndBoundedMetadata`** — Regression contract: TestTaskProgressScopePermissionsFlagsAndBoundedMetadata.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `fmt.Sprintf`, `json.Unmarshal`, `len`, `mockRequest`, `progressHandler`, `strings.Contains`, `structuredTaskFixture`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.Bytes`, `w.Body.String`.
