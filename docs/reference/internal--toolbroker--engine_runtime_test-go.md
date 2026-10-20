# internal/toolbroker/engine_runtime_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/engine_runtime_test.go)

Automated regression checks for engine_runtime; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `engineDockerFake`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { runtime *EngineContainerRuntime calls [][]string existing, failedStart, running bool mutation func(map[string]any) imageMutation func(map[string]any) }`.
- **`(*engineDockerFake).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, _ []byte, out io.Writer) error`.
  Calls: `append`, `ctx.Err`, `engineInspectionFixture`, `errors.New`, `f.imageMutation`, `f.mutation`, `io.WriteString`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `strings.Repeat`.
- **`engineInspectionFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r *EngineContainerRuntime, running bool) map[string]any`.
  Calls: `append`, `engineEnvironment`, `r.labels`, `r.name`, `strings.Repeat`.
- **`engineFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*EngineContainerRuntime, *engineDockerFake)`.
  Calls: `NewEngineContainerRuntime`, `strings.Repeat`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Add`.
- **`TestEngineRuntimeCreatesFreshStoppedAndStartsByImmutableID`** — Regression contract: TestEngineRuntimeCreatesFreshStoppedAndStartsByImmutableID.
  Signature: `func(t *testing.T)`.
  Calls: `engineFixture`, `len`, `runtime.Create`, `runtime.Inspect`, `runtime.Start`, `runtime.VerifyCreated`, `strings.Contains`, `strings.Join`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestEngineRuntimeRejectsUnsafeReportedPolicyBeforeStart`** — Regression contract: TestEngineRuntimeRejectsUnsafeReportedPolicyBeforeStart.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `engineFixture`, `runtime.Create`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestEngineRuntimeImagePreflightRefusesImplicitVolumesAndCodeInjection`** — Regression contract: TestEngineRuntimeImagePreflightRefusesImplicitVolumesAndCodeInjection.
  Signature: `func(t *testing.T)`.
  Calls: `engineFixture`, `runtime.Create`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestEngineRuntimeCancelledLifetimeAndUncertainStartCannotRevive`** — Regression contract: TestEngineRuntimeCancelledLifetimeAndUncertainStartCannotRevive.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `engineFixture`, `len`, `runtime.Close`, `runtime.Create`, `runtime.Inspect`, `runtime.Start`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestEngineRuntimeRechecksPolicyImmediatelyBeforeAndAfterStart`** — Regression contract: TestEngineRuntimeRechecksPolicyImmediatelyBeforeAndAfterStart.
  Signature: `func(t *testing.T)`.
  Calls: `engineFixture`, `len`, `runtime.Create`, `runtime.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestEngineRuntimeOriginalCancellationPrecedesDerivedPropagation`** — Regression contract: TestEngineRuntimeOriginalCancellationPrecedesDerivedPropagation.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `engineFixture`, `len`, `runtime.Create`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestEngineRuntimeRefusesUnboundRepositoryDigestBeforeCreation`** — Regression contract: TestEngineRuntimeRefusesUnboundRepositoryDigestBeforeCreation.
  Signature: `func(t *testing.T)`.
  Calls: `delete`, `engineFixture`, `runtime.Create`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestEngineRuntimeDistinguishesManifestDigestFromLocalImageID`** — Regression contract: TestEngineRuntimeDistinguishesManifestDigestFromLocalImageID.
  Signature: `func(t *testing.T)`.
  Calls: `delete`, `engineFixture`, `runtime.Create`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
