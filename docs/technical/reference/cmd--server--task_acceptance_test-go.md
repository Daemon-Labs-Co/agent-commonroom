# cmd/server/task_acceptance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_acceptance_test.go)

Automated regression checks for task_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`acceptanceFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler, room.Task)`.
  Calls: `append`, `definedDispatchBody`, `dispatchRequest`, `http.NewServeMux`, `mockRequest`, `mux.HandleFunc`, `reviseValueTask`, `s.actorGuard`, `s.auth.protect`, `s.load`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Helper`, `w.Body.String`, `waitDispatch`.
- **`readTaskAcceptance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler) map[string]json.RawMessage`.
  Calls: `json.Unmarshal`, `mockRequest`, `t.Fatal`, `t.Helper`, `w.Body.Bytes`, `w.Body.String`.
- **`taskAcceptanceBody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler, action string) taskAcceptanceRequest`.
  Calls: `json.Unmarshal`, `readTaskAcceptance`, `t.Helper`.
- **`submitTaskAcceptance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(h http.Handler, key string, body taskAcceptanceRequest) *httptest.ResponseRecorder`.
  Calls: `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Marshal`, `r.AddCookie`, `r.Header.Set`, `string`, `strings.NewReader`.
- **`TestTaskCriterionAcceptanceIsAtomicExplicitAndReplayable`** — Regression contract: TestTaskCriterionAcceptanceIsAtomicExplicitAndReplayable.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `acceptanceFixture`, `db.snapshot`, `fmt.Sprintf`, `json.Marshal`, `latestTaskAcceptance`, `len`, `mockRequest`, `readTaskAcceptance`, `s.load`, `string`, `submitTaskAcceptance`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskAcceptanceBody`, `taskHasCurrentAcceptance`, `w.Body.String`, `w.Header`, `w.Header().Get`.
- **`TestTaskCriterionReviewRefusals`** — Regression contract: TestTaskCriterionReviewRefusals.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceFixture`, `append`, `db.seed`, `db.snapshot`, `len`, `strings.Repeat`, `submitTaskAcceptance`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskAcceptanceBody`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestCriterionReviewFlagsStalenessHistoryAndFiniteSchemaFailure`** — Regression contract: TestCriterionReviewFlagsStalenessHistoryAndFiniteSchemaFailure.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceFixture`, `append`, `assessTaskCriteria`, `db.seed`, `fmt.Sprintf`, `mockRequest`, `readTaskAcceptance`, `reviseValueTask`, `s.load`, `string`, `strings.Contains`, `submitTaskAcceptance`, `syntheticDefinition`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskAcceptanceBody`, `taskHasCurrentAcceptance`, `taskReviewAllPassed`, `w.Body.String`.
- **`TestCriterionReviewHistoryPaginationShowsOriginalReviewedAttempts`** — Regression contract: TestCriterionReviewHistoryPaginationShowsOriginalReviewedAttempts.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceFixture`, `definedDispatchBody`, `dispatchRequest`, `fmt.Sprintf`, `json.Unmarshal`, `len`, `mockRequest`, `readTaskAcceptance`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `string`, `submitTaskAcceptance`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskAcceptanceBody`, `w.Body.Bytes`, `w.Body.String`, `waitDispatch`.
- **`TestCriterionReviewReplayRejectsForgedDecisionsAndGenericInjection`** — Regression contract: TestCriterionReviewReplayRejectsForgedDecisionsAndGenericInjection.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceFixture`, `db.snapshot`, `fmt.Sprintf`, `json.Marshal`, `json.Unmarshal`, `len`, `mockRequest`, `projectTaskAcceptanceReview`, `s.load`, `strings.Repeat`, `submitTaskAcceptance`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskAcceptanceBody`, `w.Body.String`.
- **`TestCriterionPolicyCannotReviveDependentExecutionAuthority`** — Regression contract: TestCriterionPolicyCannotReviveDependentExecutionAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `prepareTaskExecution`, `t.Fatal`, `t.Parallel`, `taskExecutionInputsCurrent`, `valueDependencyFixture`.
- **`TestCriterionReviewPolicyMakesRoomAndSourceReadinessStale`** — Regression contract: TestCriterionReviewPolicyMakesRoomAndSourceReadinessStale.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceFixture`, `db.seed`, `http.NewServeMux`, `mux.HandleFunc`, `readRoomReview`, `roomReviewBody`, `s.actorGuard`, `s.auth.protect`, `s.load`, `string`, `submitRoomReview`, `submitTaskAcceptance`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskAcceptanceBody`, `taskValueSource`, `w.Body.String`.
- **`TestCriterionReviewCannotAcceptOmittedSchemaOutput`** — Regression contract: TestCriterionReviewCannotAcceptOmittedSchemaOutput.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceFixture`, `append`, `definedDispatchBody`, `dispatchRequest`, `readTaskAcceptance`, `reviseValueTask`, `s.load`, `string`, `strings.Contains`, `submitTaskAcceptance`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskAcceptanceBody`, `w.Body.String`, `waitDispatch`.
