# cmd/server/task_execution_configuration_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_execution_configuration_test.go)

Automated regression checks for task_execution_configuration; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestTypedExecutionConfigurationFencesStableProfileAndConnection`** — Regression contract: TestTypedExecutionConfigurationFencesStableProfileAndConnection.
  Signature: `func(t *testing.T)`.
  Calls: `pendingResultState`, `structuredTaskResult`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskDefinitionContextAllowed`, `taskExecutionPendingCurrent`, `validTaskExecutionConfiguration`.
- **`TestConfigurationChangesWithholdRealMockedTaskResult`** — Regression contract: TestConfigurationChangesWithholdRealMockedTaskResult.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `definedDispatchBody`, `dispatchRequest`, `len`, `s.load`, `strings.Contains`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`, `waitDispatch`.
- **`TestSuccessfulConfigurationReplayRejectsChangedCurrentConfig`** — Regression contract: TestSuccessfulConfigurationReplayRejectsChangedCurrentConfig.
  Signature: `func(t *testing.T)`.
  Calls: `json.Marshal`, `pendingResultState`, `projectTaskExecution`, `responseDigest`, `t.Fatal`, `t.Parallel`.
- **`TestCompletedTypedValuesRemainHistoryAfterConfigurationChanges`** — Regression contract: TestCompletedTypedValuesRemainHistoryAfterConfigurationChanges.
  Signature: `func(t *testing.T)`.
  Calls: `pendingResultState`, `t.Fatal`, `t.Parallel`, `taskHasValidCurrentOutput`, `taskcontract.ReturnedValues`.
- **`TestOwnerProfileSaveCancelsLocalTypedTurnAfterAcknowledgment`** — Regression contract: TestOwnerProfileSaveCancelsLocalTypedTurnAfterAcknowledgment.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `ctx.Done`, `definedDispatchBody`, `dispatchRequest`, `fmt.Sprintf`, `json.Marshal`, `len`, `make`, `mockRequest`, `s.load`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`, `w.Body.String`, `waitDispatch`.
- **`TestFailedProfileSaveDoesNotCancelAdmittedTurn`** — Regression contract: TestFailedProfileSaveDoesNotCancelAdmittedTurn.
  Signature: `func(t *testing.T)`.
  Calls: `ctx.Done`, `ctx.Err`, `definedDispatchBody`, `dispatchRequest`, `fmt.Sprintf`, `json.Marshal`, `make`, `mockRequest`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`, `turn.Err`, `w.Body.String`, `waitDispatch`.
- **`TestTypedSessionCallbackRechecksConfiguration`** — Regression contract: TestTypedSessionCallbackRechecksConfiguration.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `definedDispatchBody`, `dispatchRequest`, `fmt.Errorf`, `fmt.Sprint`, `make`, `s.load`, `s.recordAgentSession`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `w.Body.String`, `waitDispatch`.
