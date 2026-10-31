# internal/toolbroker/workspace_bootstrap_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/workspace_bootstrap_test.go)

Automated regression checks for workspace_bootstrap; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `bootstrapScenario`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *runtimeScenario mode string bootCalls int source string }`.
- **`(*bootstrapScenario).RunInput`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in io.Reader, out io.Writer) error`.
  Calls: `append`, `bytes.Repeat`, `bytes.Replace`, `errors.New`, `f.t.Fatal`, `io.ReadAll`, `json.Marshal`, `len`, `out.Write`, `slices.Equal`, `string`, `strings.Repeat`.
- **Type `trackedInput`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *strings.Reader closed atomic.Bool }`.
- **`(*trackedInput).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `r.closed.Store`.
- **`TestRuntimeBootstrapBindsSnapshotAndGatesTools`** — Regression contract: TestRuntimeBootstrapBindsSnapshotAndGatesTools.
  Signature: `func(t *testing.T)`.
  Calls: `again.closed.Load`, `cancel`, `f.plan.workspaceReady`, `input.closed.Load`, `len`, `r.Bootstrap`, `r.Create`, `r.Start`, `r.executor.Execute`, `runtimeSetup`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `workspacebootstrap.ValidReceipt`.
- **`TestRuntimeBootstrapRequiresRunningStreamingAdapter`** — Regression contract: TestRuntimeBootstrapRequiresRunningStreamingAdapter.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `input.closed.Load`, `r.Bootstrap`, `r.Create`, `r.Start`, `runtimeSetup`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **Type `blockedSnapshot`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { entered chan struct{} closed chan struct{} readOnce sync.Once closeOnce sync.Once }`.
- **`(*blockedSnapshot).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `close`, `r.readOnce.Do`.
- **`(*blockedSnapshot).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `close`, `r.closeOnce.Do`.
- **`TestRuntimeBootstrapCancellationClosesBlockedSnapshot`** — Regression contract: TestRuntimeBootstrapCancellationClosesBlockedSnapshot.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `cancelLifetime`, `context.WithCancel`, `f.plan.workspaceReady`, `make`, `r.Bootstrap`, `r.Create`, `r.Start`, `runtimeSetup`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestBootstrappedExecutorCannotReviveCancelledJob`** — Regression contract: TestBootstrappedExecutorCannotReviveCancelledJob.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `io.NopCloser`, `len`, `r.Bootstrap`, `r.Create`, `r.Start`, `r.executor.Execute`, `runtimeSetup`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestExecutorSuppressesSuccessWhenWorkspaceJobEndsDuringAction`** — Regression contract: TestExecutorSuppressesSuccessWhenWorkspaceJobEndsDuringAction.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `containerExecutor`, `context.WithCancel`, `context.WithDeadline`, `e.Execute`, `endJob`, `executorFixture`, `out.Write`, `output.Len`, `p.workspace.Store`, `runtimeMock`, `syntheticReadResult`, `t.Context`, `t.Fatal`, `t.Parallel`.
