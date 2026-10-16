# internal/toolbroker/executor_authorization_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/executor_authorization_test.go)

Automated regression checks for executor_authorization; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestExecutorRefusesUndeclaredOperationsBeforeDocker`** — Regression contract: TestExecutorRefusesUndeclaredOperationsBeforeDocker.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `NewToolContainerPlan`, `a.Reserve`, `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `liveFixture`, `mockDocker`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.Now`.
- **`TestExecutorPolicyGateAcceptsDeclaredRead`** — Regression contract: TestExecutorPolicyGateAcceptsDeclaredRead.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Equal`, `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `executorFixture`, `out.Write`, `output.Bytes`, `runtimeMock`, `syntheticReadResult`, `t.Fatal`, `t.Parallel`.
- **`TestExecutorPolicyGateScopeAndLifetime`** — Regression contract: TestExecutorPolicyGateScopeAndLifetime.
  Signature: `func(t *testing.T)`.
  Calls: `a.Revoke`, `a.permitsAction`, `fixture`, `now.Add`, `t.Fatal`, `t.Parallel`.
