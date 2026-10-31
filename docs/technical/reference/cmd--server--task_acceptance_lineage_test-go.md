# cmd/server/task_acceptance_lineage_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_acceptance_lineage_test.go)

Automated regression checks for task_acceptance_lineage; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`acceptanceLineageState`** — Synthetic projected records deliberately include a forged cycle/overlong graph: even invalid historical state must refuse within one traversal budget. No DB, provider, filesystem or live identity participates.
  Signature: `func(t *testing.T, n int, cycle bool) room.State`.
  Calls: `append`, `fmt.Sprintf`, `syntheticDefinition`, `t.Fatal`, `t.Helper`, `taskAcceptanceContentSnapshot`, `taskcontract.Bind`, `taskcontract.Digest`, `taskcontract.ReturnedValues`.
- **`TestAcceptanceUsesOneBoundedLineageWalk`** — Regression contract: TestAcceptanceUsesOneBoundedLineageWalk.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceLineageState`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskAcceptanceSnapshot`, `taskHasCurrentAcceptance`, `taskHasRecordedAcceptance`, `taskHasValidCurrentOutput`, `taskValueLineageCurrent`.
- **`TestAcceptanceRevokesStaleLineageWithoutErasingRecordedDecision`** — Regression contract: TestAcceptanceRevokesStaleLineageWithoutErasingRecordedDecision.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceLineageState`, `t.Fatal`, `t.Parallel`, `taskHasCurrentAcceptance`, `taskHasRecordedAcceptance`.
- **`TestAcceptanceSharedBudgetBoundsRepeatedDAGBranches`** — Regression contract: TestAcceptanceSharedBudgetBoundsRepeatedDAGBranches.
  Signature: `func(t *testing.T)`.
  Calls: `acceptanceLineageState`, `append`, `fmt.Sprintf`, `t.Fatal`, `t.Parallel`, `taskAcceptanceContentSnapshot`, `taskHasCurrentAcceptance`, `taskHasRecordedAcceptance`, `taskValueLineageCurrent`, `taskcontract.Digest`, `taskcontract.Validate`.
