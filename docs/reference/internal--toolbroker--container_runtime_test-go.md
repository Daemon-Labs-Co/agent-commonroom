# internal/toolbroker/container_runtime_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/container_runtime_test.go)

Automated regression checks for container_runtime; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `runtimeScenario`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T plan *ToolContainerPlan container toolInspection volume toolVolumeInspection mode string calls []string creates, starts int volumeExists bool volumeCreates int }`.
- **`(*runtimeScenario).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `append`, `boolValue`, `errors.New`, `f.plan.CreateArgs`, `f.t.Fatal`, `f.t.Fatalf`, `inspectJSON`, `len`, `out.Write`, `slices.Equal`, `strings.Cut`, `strings.Repeat`, `time.Now`, `time.Now().Add`, `time.Now().Add(time.Hour).UTC`, `time.Now().Add(time.Hour).UTC().Format`, `time.Now().UTC`, `time.Now().UTC().Format`.
- **`unpreparedRuntimeSetup`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mode string) (*ToolContainerRuntime, *runtimeScenario, context.CancelFunc)`.
  Calls: `NewToolContainerRuntime`, `boolValue`, `cancel`, `context.WithCancel`, `inspectionFixture`, `liveFixture`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`runtimeSetup`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mode string) (*ToolContainerRuntime, *runtimeScenario, context.CancelFunc)`.
  Calls: `cancel`, `f.plan.WorkspaceIntent`, `r.PrepareVolume`, `t.Context`, `t.Fatal`, `t.Helper`, `time.Now`, `unpreparedRuntimeSetup`.
- **`TestToolRuntimeCreateInspectStart`** — Regression contract: TestToolRuntimeCreateInspectStart.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `len`, `r.Create`, `r.Inspect`, `r.Start`, `runtimeSetup`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestToolRuntimeRefusesUncertainCreation`** — Regression contract: TestToolRuntimeRefusesUncertainCreation.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `len`, `r.Create`, `r.Start`, `runtimeSetup`, `slices.Contains`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestToolRuntimeRevalidatesBeforeStart`** — Regression contract: TestToolRuntimeRevalidatesBeforeStart.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `len`, `r.Create`, `r.Start`, `runtimeSetup`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestToolRuntimeJobLifetimeAndIdentity`** — Regression contract: TestToolRuntimeJobLifetimeAndIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `f.plan.authority.Revoke`, `len`, `r.Create`, `r.Start`, `runtimeSetup`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestToolCreateArgsAndCreatedState`** — Regression contract: TestToolCreateArgsAndCreatedState.
  Signature: `func(t *testing.T)`.
  Calls: `boolValue`, `inspectJSON`, `inspectionFixture`, `p.CreateArgs`, `p.VerifyCreatedInspection`, `p.VerifyInspection`, `p.authority.Revoke`, `slices.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestToolRuntimeOriginalCancellationPrecedesDerivedPropagation`** — Regression contract: TestToolRuntimeOriginalCancellationPrecedesDerivedPropagation.
  Signature: `func(t *testing.T)`.
  Calls: `NewToolContainerRuntime`, `base.Run`, `cancel`, `context.Background`, `context.WithCancel`, `context.WithDeadline`, `done`, `executorFixture`, `mockDocker`, `out.Write`, `plan.workspace.Store`, `runtime.WorkspaceIntent`, `runtime.executor.Execute`, `runtimeMock`, `syntheticReadResult`, `t.Context`, `t.Fatal`, `t.Parallel`.
