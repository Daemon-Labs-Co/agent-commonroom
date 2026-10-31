# cmd/server/task_definition_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_definition_test.go)

Automated regression checks for task_definition; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`syntheticDefinition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() taskcontract.Definition`.
- **`structuredTaskFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler, room.Task)`.
  Calls: `dispatchFixture`, `fmt.Sprintf`, `json.Marshal`, `mockRequest`, `s.load`, `syntheticDefinition`, `t.Context`, `t.Fatal`, `t.Helper`, `w.Body.String`.
- **`definedDispatchBody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(task room.Task) string`.
  Calls: `fmt.Sprintf`.
- **`TestStructuredTaskDispatchOutputsAndAuthority`** — Regression contract: TestStructuredTaskDispatchOutputsAndAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `audit.RecordIntent`, `audit.RecordOutcome`, `ctx.Deadline`, `definedDispatchBody`, `dispatchRequest`, `fmt.Sprintf`, `jc.Value`, `json.Marshal`, `len`, `make`, `mockRequest`, `newControllerToolAuditor`, `reg.Deadline.UTC`, `reg.Register`, `reg.Update`, `s.managedExecutionJobContext`, `strings.Contains`, `strings.Repeat`, `structuredTaskFixture`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.After`, `time.Until`, `w.Body.String`, `waitDispatch`.
- **`TestTaskDefinitionMutationPermissionsVersionsAndLegacy`** — Regression contract: TestTaskDefinitionMutationPermissionsVersionsAndLegacy.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `context.Background`, `db.seed`, `fmt.Sprintf`, `json.Marshal`, `mockRequest`, `s.load`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskcontract.CriteriaText`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestStructuredTaskFailureRefusalsAndFlagEpoch`** — Regression contract: TestStructuredTaskFailureRefusalsAndFlagEpoch.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `definedDispatchBody`, `dispatchRequest`, `fmt.Sprintf`, `json.Marshal`, `len`, `mockRequest`, `s.load`, `strings.Repeat`, `strings.Replace`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`, `waitDispatch`.
- **`TestStructuredDelegationKeepsExactParentInputs`** — Regression contract: TestStructuredDelegationKeepsExactParentInputs.
  Signature: `func(t *testing.T)`.
  Calls: `json.Marshal`, `json.RawMessage`, `json.Unmarshal`, `prepareTaskExecution`, `projectTaskExecution`, `syntheticDefinition`, `t.Fatal`, `t.Parallel`, `taskDefinitionDigest`, `taskcontract.Bind`.
- **`TestStructuredReviewerUsesValidatedWorkSnapshot`** — Regression contract: TestStructuredReviewerUsesValidatedWorkSnapshot.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `definedDispatchBody`, `dispatchRequest`, `fmt.Sprintf`, `jobCtx.Value`, `json.Marshal`, `len`, `make`, `mockRequest`, `reg.Register`, `reg.Update`, `s.load`, `s.managedExecutionJobContext`, `s.mu.Lock`, `s.mu.Unlock`, `strings.Contains`, `strings.Repeat`, `structuredTaskFixture`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `taskHasValidCurrentOutput`, `time.After`, `time.Now`, `time.Now().Add`, `time.Now().After`, `time.Sleep`, `w.Body.String`, `waitDispatch`.
