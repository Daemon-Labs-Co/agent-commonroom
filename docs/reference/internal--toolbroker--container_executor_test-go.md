# internal/toolbroker/container_executor_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/container_executor_test.go)

Automated regression checks for container_executor; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `mockDocker`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, []string, []byte, io.Writer) error`.
- **`(mockDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, a []string, input []byte, w io.Writer) error`.
  Calls: `m`.
- **`executorFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*ToolContainerPlan, toolInspection, []byte, []byte, string, Scope, Action)`.
  Calls: `NewToolContainerPlan`, `context.Background`, `inspectJSON`, `inspectionFixture`, `liveFixture`, `p.workspace.Store`, `t.Fatal`, `t.Helper`, `time.Now`.
- **`runtimeMock`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, c, volume []byte, exec func(context.Context, []string, []byte, io.Writer) error, cleanup func([]string, io.Writer) error) DockerRunner`.
  Calls: `cleanup`, `errors.New`, `exec`, `mockDocker`, `out.Write`, `t.Error`, `t.Errorf`, `t.Helper`.
- **`containerExecutor`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, p *ToolContainerPlan, r DockerRunner, root string) *ContainerExecutor`.
  Calls: `NewContainerExecutor`, `t.Fatal`, `t.Helper`.
- **`TestContainerExecutorFixedHelperAndImmutableIdentity`** — Regression contract: TestContainerExecutorFixedHelperAndImmutableIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Contains`, `bytes.Equal`, `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `executorFixture`, `json.Unmarshal`, `out.Write`, `output.Bytes`, `runtimeMock`, `slices.Equal`, `syntheticReadResult`, `t.Fatal`, `t.Parallel`.
- **`TestContainerExecutorRejectsScopeInputAndContext`** — Regression contract: TestContainerExecutorRejectsScopeInputAndContext.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `executorFixture`, `mockDocker`, `scope.Deadline.Add`, `stop`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **`TestContainerExecutorRefusesUnsafeOrOversizedInspection`** — Regression contract: TestContainerExecutorRefusesUnsafeOrOversizedInspection.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `errors.New`, `executorFixture`, `inspectJSON`, `mockDocker`, `out.Len`, `p.RunArgs`, `runtimeMock`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `time.Now`.
- **`TestContainerExecutorFailureRequiresAbsenceProof`** — Regression contract: TestContainerExecutorFailureRequiresAbsenceProof.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `errors.New`, `executorFixture`, `io.WriteString`, `output.Len`, `runtimeMock`, `slices.Contains`, `slices.Equal`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestContainerExecutorCanceledWorkUsesIndependentCleanup`** — Regression contract: TestContainerExecutorCanceledWorkUsesIndependentCleanup.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `executorFixture`, `io.WriteString`, `mockDocker`, `runCtx.Deadline`, `runCtx.Done`, `runCtx.Err`, `runner.Run`, `runtimeMock`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.Until`.
- **`TestContainerExecutorRevocationCancelsRunner`** — Regression contract: TestContainerExecutorRevocationCancelsRunner.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `close`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `ctx.Done`, `ctx.Err`, `e.Execute`, `executorFixture`, `io.WriteString`, `make`, `p.authority.Revoke`, `runtimeMock`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestDispatcherUsesContainerExecutorAfterAudit`** — Regression contract: TestDispatcherUsesContainerExecutorAfterAudit.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `NewToolContainerPlan`, `bytes.Equal`, `containerExecutor`, `context.Background`, `d.Dispatch`, `dispatcher`, `executorFixture`, `out.Write`, `p.workspace.Store`, `runtimeMock`, `syntheticReadResult`, `t.Error`, `t.Fatal`, `t.Parallel`, `time.Now`.
- **`TestExecutorPinsOriginalBootstrapIdentity`** — Regression contract: TestExecutorPinsOriginalBootstrapIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `containerExecutor`, `context.WithDeadline`, `e.Execute`, `executorFixture`, `inspectJSON`, `len`, `mockDocker`, `out.Write`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
