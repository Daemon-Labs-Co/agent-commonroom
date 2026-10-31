# cmd/server/task_context_policy_fence_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_context_policy_fence_test.go)

Automated regression checks for task_context_policy_fence; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestTypedTaskContextPolicyRevokesPendingAuthority`** — Regression contract: TestTypedTaskContextPolicyRevokesPendingAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `pendingResultState`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskDefinitionContextAllowed`.
- **`TestContextPolicyCycleWithholdsTypedResultAndNativeSession`** — Regression contract: TestContextPolicyCycleWithholdsTypedResultAndNativeSession.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `definedDispatchBody`, `dispatchRequest`, `len`, `s.recordAgentSession`, `strings.Contains`, `strings.Join`, `structuredTaskFixture`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`, `waitDispatch`.
- **`TestPersonalRoomContextPolicySurvivesReplay`** — Regression contract: TestPersonalRoomContextPolicySurvivesReplay.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `len`, `mockRequest`, `mockServer`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
