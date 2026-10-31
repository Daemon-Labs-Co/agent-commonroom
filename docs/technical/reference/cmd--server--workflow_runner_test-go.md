# cmd/server/workflow_runner_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workflow_runner_test.go)

Automated regression checks for workflow_runner; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`executionGraph`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() room.Workflow`.
  Calls: `append`, `float64`.
- **`executionServer`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler)`.
  Calls: `db.seed`, `executionGraph`, `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`.
- **`startExecution`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler) room.WorkflowRun`.
  Calls: `json.Unmarshal`, `mockRequest`, `res.Body.Bytes`, `res.Body.String`, `t.Fatal`, `t.Fatalf`, `t.Helper`.
- **`awaitExecution`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, s *server, status string) room.WorkflowRun`.
  Calls: `context.Background`, `len`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `t.Fatal`, `t.Fatalf`, `t.Helper`, `time.Now`, `time.Now().Add`, `time.Now().Before`, `time.Sleep`.
- **`gateExecution`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler, run room.WorkflowRun, action, port string)`.
  Calls: `json.Marshal`, `mockRequest`, `string`, `t.Fatalf`, `t.Helper`, `w.Body.String`.
- **`TestWorkflowExecutesReviewReworkAndHumanAcceptance`** — Regression contract: TestWorkflowExecutesReviewReworkAndHumanAcceptance.
  Signature: `func(t *testing.T)`.
  Calls: `awaitExecution`, `developments.Add`, `developments.Load`, `executionServer`, `gateExecution`, `len`, `reviews.Add`, `startExecution`, `strings.Contains`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **`TestWorkflowRunnableRejectsUnsupportedAndUnsafeGraphs`** — Regression contract: TestWorkflowRunnableRejectsUnsupportedAndUnsafeGraphs.
  Signature: `func(t *testing.T)`.
  Calls: `c.change`, `context.Background`, `executionGraph`, `executionServer`, `len`, `s.load`, `t.Fatal`, `t.Parallel`, `t.Run`, `validateRunnable`.
- **`TestWorkflowBoundariesAndExplicitRecovery`** — Regression contract: TestWorkflowBoundariesAndExplicitRecovery.
  Signature: `func(t *testing.T)`.
  Calls: `awaitExecution`, `context.Background`, `db.seed`, `errors.New`, `executionGraph`, `executionServer`, `gateExecution`, `len`, `mockRequest`, `s.load`, `startExecution`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestWorkflowAuthStaleAndTamperedHistory`** — Regression contract: TestWorkflowAuthStaleAndTamperedHistory.
  Signature: `func(t *testing.T)`.
  Calls: `awaitExecution`, `context.Background`, `db.mu.Lock`, `db.mu.Unlock`, `db.seed`, `executionServer`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Marshal`, `len`, `mockRequest`, `r.AddCookie`, `s.load`, `startExecution`, `string`, `strings.NewReader`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`.
- **`TestWorkflowCutoffsDiscardLateOutputs`** — Regression contract: TestWorkflowCutoffsDiscardLateOutputs.
  Signature: `func(t *testing.T)`.
  Calls: `awaitExecution`, `close`, `context.Background`, `db.seed`, `executionServer`, `gateExecution`, `make`, `mockRequest`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `startExecution`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `w.Body.String`.
- **`TestWorkflowWaitDecisionAndPreservedDisabledInspection`** — Regression contract: TestWorkflowWaitDecisionAndPreservedDisabledInspection.
  Signature: `func(t *testing.T)`.
  Calls: `awaitExecution`, `context.Background`, `db.seed`, `executionServer`, `gateExecution`, `len`, `mockRequest`, `s.load`, `strings.Contains`, `t.Error`, `t.Fatal`, `t.Parallel`, `visibleState`, `workflowExecutionEnabled`, `wres.Body.String`.
- **`TestWorkflowRejectsForgedRunEventsAndDuplicateStarts`** — Regression contract: TestWorkflowRejectsForgedRunEventsAndDuplicateStarts.
  Signature: `func(t *testing.T)`.
  Calls: `awaitExecution`, `executionServer`, `gateExecution`, `mockRequest`, `startExecution`, `t.Fatal`, `t.Parallel`.
