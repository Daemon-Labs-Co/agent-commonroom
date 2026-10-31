# cmd/server/task_value_inputs_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_value_inputs_test.go)

Automated regression checks for task_value_inputs; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`reviseValueTask`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, h http.Handler, task room.Task)`.
  Calls: `fmt.Sprintf`, `json.Marshal`, `mockRequest`, `t.Fatal`, `t.Helper`, `w.Body.String`.
- **`valueDependencyFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler, room.State, room.Task, taskcontract.Binding)`.
  Calls: `definedDispatchBody`, `dispatchRequest`, `mockRequest`, `reviseValueTask`, `s.load`, `structuredTaskFixture`, `syntheticDefinition`, `t.Context`, `t.Fatal`, `t.Helper`, `w.Body.String`, `waitDispatch`.
- **`valueDispatchBody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, target room.Task, b taskcontract.Binding) string`.
  Calls: `json.Marshal`, `string`, `t.Fatal`, `t.Helper`.
- **`TestVerifiedTaskValueDispatchDiscoveryAndImmutableProvenance`** — Regression contract: TestVerifiedTaskValueDispatchDiscoveryAndImmutableProvenance.
  Signature: `func(t *testing.T)`.
  Calls: `dispatchRequest`, `fmt.Sprintf`, `http.NewServeMux`, `json.Marshal`, `json.Unmarshal`, `len`, `mockRequest`, `mux.HandleFunc`, `reviseValueTask`, `s.actorGuard`, `s.auth.protect`, `s.load`, `strings.Contains`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `taskHasValidCurrentOutput`, `valueDependencyFixture`, `valueDispatchBody`, `w.Body.Bytes`, `w.Body.String`, `waitDispatch`.
- **`cloneValueState`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) room.State`.
  Calls: `append`, `json.Marshal`, `json.Unmarshal`.
- **`TestVerifiedTaskValueInputRefusalsAndLineageRevocation`** — Regression contract: TestVerifiedTaskValueInputRefusalsAndLineageRevocation.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `change`, `cloneValueState`, `json.RawMessage`, `prepareTaskExecution`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskExecutionInputsCurrent`, `tc.change`, `valueDependencyFixture`.
- **`TestTaskValueSourceChangeRevokesActualBrokerAndResult`** — Regression contract: TestTaskValueSourceChangeRevokesActualBrokerAndResult.
  Signature: `func(t *testing.T)`.
  Calls: `audit.RecordIntent`, `close`, `db.snapshot`, `dispatchRequest`, `fmt.Errorf`, `jc.Value`, `len`, `make`, `newControllerToolAuditor`, `reg.Deadline.UTC`, `reg.Register`, `reg.Update`, `reviseValueTask`, `s.managedExecutionJobContext`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `time.After`, `valueDependencyFixture`, `valueDispatchBody`, `w.Body.String`, `waitDispatch`.
- **`TestTaskValueLineageCyclesDepthBudgetsAndHistoricalEpoch`** — Regression contract: TestTaskValueLineageCyclesDepthBudgetsAndHistoricalEpoch.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `cloneValueState`, `fmt.Sprintf`, `prepareTaskExecution`, `t.Fatal`, `t.Parallel`, `taskExecutionInputsCurrent`, `taskValueLineageCurrent`, `taskcontract.ReturnedValues`, `valueDependencyFixture`.
- **`TestTaskValueSourceAPIFlagsPermissionsAndCompatibility`** — Regression contract: TestTaskValueSourceAPIFlagsPermissionsAndCompatibility.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `http.NewServeMux`, `mockRequest`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `strings.Contains`, `strings.Replace`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `valueDependencyFixture`, `w.Body.String`.
- **`TestDependencyFlagCycleCancelsOnlyDependentTurns`** — Regression contract: TestDependencyFlagCycleCancelsOnlyDependentTurns.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `ctx.Done`, `dispatchRequest`, `fmt.Sprintf`, `len`, `make`, `mockRequest`, `s.mu.Lock`, `s.mu.Unlock`, `t.Fatal`, `t.Parallel`, `time.After`, `valueDependencyFixture`, `valueDispatchBody`, `w.Body.String`, `waitDispatch`.
- **`TestTaskValueSourceDiscoveryPaginatesAcceptedValuesAndRetainsReviewIdentity`** — Regression contract: TestTaskValueSourceDiscoveryPaginatesAcceptedValuesAndRetainsReviewIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `db.seed`, `db.snapshot`, `delete`, `dispatchRequest`, `fmt.Sprintf`, `http.NewServeMux`, `json.Marshal`, `json.Unmarshal`, `len`, `mockRequest`, `mux.HandleFunc`, `resolveTaskValueInputs`, `reviseValueTask`, `s.actorGuard`, `s.auth.protect`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskcontract.ReturnedValues`, `valueDependencyFixture`, `w.Body.Bytes`, `w.Body.String`, `waitDispatch`.
