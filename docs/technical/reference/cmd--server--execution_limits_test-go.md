# cmd/server/execution_limits_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/execution_limits_test.go)

Automated regression checks for execution_limits; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestAdmissionBoundsIncludeOtherRoomsAndRemovedActiveJobs`** — Regression contract: TestAdmissionBoundsIncludeOtherRoomsAndRemovedActiveJobs.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `delete`, `fmt.Sprint`, `mockServer`, `s.agentAdmission`, `t.Fatal`, `t.Parallel`.
- **`TestChatQuotaRejectsBeforeProviderOrHistoryWrite`** — Regression contract: TestChatQuotaRejectsBeforeProviderOrHistoryWrite.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `fmt.Sprint`, `http.HandlerFunc`, `len`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestWorkflowQuotaPreventsProviderDispatch`** — Regression contract: TestWorkflowQuotaPreventsProviderDispatch.
  Signature: `func(t *testing.T)`.
  Calls: `awaitExecution`, `executionServer`, `fmt.Sprint`, `startExecution`, `strings.Contains`, `t.Error`, `t.Fatal`, `t.Parallel`.
