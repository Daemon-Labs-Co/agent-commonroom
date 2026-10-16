# cmd/server/workflows_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/workflows_test.go)

Automated regression checks for workflows; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestGraphRejectsInvalidConnections`** — Regression contract: TestGraphRejectsInvalidConnections.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `t.Fatal`, `t.Parallel`, `validateGraph`.
- **`TestWorkflowValidationBoundaries`** — Regression contract: TestWorkflowValidationBoundaries.
  Signature: `func(t *testing.T)`.
  Calls: `c.change`, `make`, `t.Fatal`, `t.Parallel`, `t.Run`, `valid`, `validPort`, `validateGraph`.
