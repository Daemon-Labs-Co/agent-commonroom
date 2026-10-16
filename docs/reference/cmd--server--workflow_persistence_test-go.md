# cmd/server/workflow_persistence_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/workflow_persistence_test.go)

Automated regression checks for workflow_persistence; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWorkflowTransitionRollbackStopsExecution`** — Regression contract: TestWorkflowTransitionRollbackStopsExecution.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `len`, `mockServer`, `s.mu.Lock`, `s.mu.Unlock`, `s.persistWorkflowTransitionLocked`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkflowTransitionCommitsOutcomeTogether`** — Regression contract: TestWorkflowTransitionCommitsOutcomeTogether.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `len`, `mockServer`, `s.mu.Lock`, `s.mu.Unlock`, `s.persistWorkflowTransitionLocked`, `t.Fatal`, `t.Parallel`.
- **`TestWorkflowOutcomeFailurePreventsNextAgentDispatch`** — Regression contract: TestWorkflowOutcomeFailurePreventsNextAgentDispatch.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `executionServer`, `len`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `startExecution`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `time.Now().Before`, `time.Sleep`.
