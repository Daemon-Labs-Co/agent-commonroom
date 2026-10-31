# cmd/server/task_result_binding_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_result_binding_test.go)

Automated regression checks for task_result_binding; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`pendingResultState`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mode string) (room.State, room.TaskExecution, room.Participant, context.Context)`.
  Calls: `bindTaskExecutionConfiguration`, `context.Background`, `context.WithValue`, `syntheticDefinition`, `t.Fatal`, `t.Helper`, `taskcontract.Bind`, `taskcontract.Digest`, `time.Now`, `time.Now().UTC`.
- **`TestTypedResultContextRequiresExactExecutionRole`** — Regression contract: TestTypedResultContextRequiresExactExecutionRole.
  Signature: `func(t *testing.T)`.
  Calls: `pendingResultState`, `strings.Repeat`, `structuredTaskResult`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskDefinitionContextAllowed`.
- **`TestSuccessfulExecutionReplayRequiresCompletedTaskBinding`** — Regression contract: TestSuccessfulExecutionReplayRequiresCompletedTaskBinding.
  Signature: `func(t *testing.T)`.
  Calls: `e.StartedAt.Add`, `json.Marshal`, `pendingResultState`, `projectTaskExecution`, `responseDigest`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskcontract.ReturnedValues`.
- **`TestChangedTaskRoleWithholdsActualTurnResult`** — Regression contract: TestChangedTaskRoleWithholdsActualTurnResult.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `definedDispatchBody`, `dispatchRequest`, `len`, `s.load`, `strings.Contains`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`, `waitDispatch`.
