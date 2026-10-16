# cmd/server/agent_status_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/agent_status_test.go)

Automated regression checks for agent_status; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestDeadlineFailurePreservesCauseInsteadOfKilledProcess`** — Regression contract: TestDeadlineFailurePreservesCauseInsteadOfKilledProcess.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithDeadline`, `db.seed`, `errors.New`, `len`, `mockServer`, `s.finishAgentTurn`, `s.load`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestHistoricalKilledProcessDoesNotInventTimeout`** — Regression contract: TestHistoricalKilledProcessDoesNotInventTimeout.
  Signature: `func(t *testing.T)`.
  Calls: `safeAgentFailure`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestLongRunningAgeIsNotAClaimOfWorkerHealth`** — Regression contract: TestLongRunningAgeIsNotAClaimOfWorkerHealth.
  Signature: `func(t *testing.T)`.
  Calls: `now.Add`, `projectAgentElapsed`, `projectAgentStatusDetails`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `time.Date`, `time.Duration`.
- **`TestBlockedStatusUsesSafeFailureAndRecordedActivity`** — Regression contract: TestBlockedStatusUsesSafeFailureAndRecordedActivity.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `mockServer`, `s.load`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestAgentStatusHistoricalFallbackScopeAndSafetyPrecedence`** — Regression contract: TestAgentStatusHistoricalFallbackScopeAndSafetyPrecedence.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `len`, `mockServer`, `s.load`, `scoped`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestAgentFailureCategoriesNeverReturnRawDiagnostics`** — Regression contract: TestAgentFailureCategoriesNeverReturnRawDiagnostics.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `safeAgentFailure`, `statusActivity`, `strings.Contains`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
