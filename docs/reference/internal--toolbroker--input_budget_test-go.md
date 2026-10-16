# internal/toolbroker/input_budget_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/input_budget_test.go)

Automated regression checks for input_budget; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestTypedBudgetRefusalDoesNotAllocate`** — Regression contract: TestTypedBudgetRefusalDoesNotAllocate.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `a.Reserve`, `fixture`, `strings.Repeat`, `t.Fatal`, `t.Fatalf`, `testing.AllocsPerRun`.
- **`TestActionBudgetCountsCombinedCommandInput`** — Regression contract: TestActionBudgetCountsCombinedCommandInput.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `digest`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validAction`.
- **`TestDispatcherBudgetRefusesBeforeDependencyCalls`** — Regression contract: TestDispatcherBudgetRefusesBeforeDependencyCalls.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `context.Background`, `d.Dispatch`, `dispatcher`, `fixture`, `mockExecutor`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **`TestEscapedPayloadStillHasWireLimit`** — Regression contract: TestEscapedPayloadStillHasWireLimit.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `a.Reserve`, `fixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validRequest`.
- **`TestTypedExecutorBudgetRefusesBeforeDocker`** — Regression contract: TestTypedExecutorBudgetRefusesBeforeDocker.
  Signature: `func(t *testing.T)`.
  Calls: `NewContainerExecutor`, `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `executorFixture`, `mockDocker`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Parallel`, `validToolImage`.
