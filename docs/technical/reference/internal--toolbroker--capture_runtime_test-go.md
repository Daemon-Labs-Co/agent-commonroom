# internal/toolbroker/capture_runtime_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/capture_runtime_test.go)

Automated regression checks for capture_runtime; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureRuntimeFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *captureAdmissionRunner mode string exists, begun, bound, started, finished, absent, removed bool reserves, seals, observations int process *mockCaptureProcess exitedChecks atomic.Int32 }`.
- **Type `mockCaptureProcess`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn f *captureRuntimeFixture done chan error closed atomic.Bool waited atomic.Bool earlyClose atomic.Bool }`.
- **`(*mockCaptureProcess).Wait`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `ctx.Done`, `ctx.Err`, `p.closed.Load`, `p.waited.Store`, `probeBool`.
- **`(*captureRuntimeFixture).BeginCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workspacecapture.Request, o CaptureObjectVolume, image string) error`.
  Calls: `ctx.Err`, `f.t.Fatal`.
- **`(*captureRuntimeFixture).BindCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workspacecapture.Request, id string) error`.
  Calls: `ctx.Err`, `f.t.Fatal`.
- **`(*captureRuntimeFixture).FinishCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workspacecapture.Request, id string, absent bool) error`.
  Calls: `ctx.Err`.
- **`(*captureRuntimeFixture).AuthorizeFrozenSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, b workspacesnapshot.Binding, p workspacemount.Proof) error`.
  Calls: `ctx.Err`, `f.t.Fatal`.
- **`(*captureRuntimeFixture).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i objectcustody.Intent) error`.
  Calls: `ctx.Err`.
- **`(*captureRuntimeFixture).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `ctx.Err`.
- **`(*captureRuntimeFixture).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Scope, objectcustody.Receipt) error`.
  Calls: `f.t.Fatal`.
- **`(*captureRuntimeFixture).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
  Calls: `f.t.Fatal`.
- **`(*captureRuntimeFixture).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `errors.New`, `f.captureAdmissionRunner.Run`, `f.exitedChecks.Add`, `f.p.CreateArgs`, `f.t.Fatal`, `io.WriteString`, `slices.Contains`, `slices.Equal`, `strings.Repeat`.
- **`(*captureRuntimeFixture).StartCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (CaptureProcess, error)`.
  Calls: `client.AuthorizeFrozenSource`, `client.Finish`, `client.Reserve`, `client.Seal`, `f.t.Fatal`, `helper.Close`, `make`, `net.Pipe`, `probeBool`, `strings.Repeat`, `workspacecapture.NewCaptureChannel`.
- **`newCaptureRuntimeFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mode string) (*CaptureRuntime, *captureRuntimeFixture)`.
  Calls: `NewCaptureRuntime`, `capturePlanFixture`, `t.Fatal`, `t.Helper`.
- **`TestCaptureRuntimeLifecycle`** — Regression contract: TestCaptureRuntimeLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `len`, `newCaptureRuntimeFixture`, `r.Run`, `t.Fatal`, `t.Parallel`.
- **`TestCaptureRuntimeFailureWithholdsReports`** — Regression contract: TestCaptureRuntimeFailureWithholdsReports.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `errors.Is`, `f.process.closed.Load`, `len`, `newCaptureRuntimeFixture`, `r.Run`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureRuntimeCanceledAndMissingDependencies`** — Regression contract: TestCaptureRuntimeCanceledAndMissingDependencies.
  Signature: `func(t *testing.T)`.
  Calls: `NewCaptureRuntime`, `cancel`, `context.Background`, `context.WithCancel`, `newCaptureRuntimeFixture`, `r.Run`, `t.Fatal`, `t.Parallel`.
- **`TestCaptureRuntimeFailedStreamIsClosed`** — Regression contract: TestCaptureRuntimeFailedStreamIsClosed.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `newCaptureRuntimeFixture`, `r.Run`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`(*mockCaptureProcess).Close`** — Closing the production attached stream cancels its command. Model that behavior deterministically: a close before Wait must not count as a successful exit.
  Signature: `func() error`.
  Calls: `p.Conn.Close`, `p.closed.Swap`, `p.earlyClose.Store`, `p.f.exitedChecks.Load`, `p.waited.Load`.
- **`TestCaptureRuntimeWaitsAndVerifiesExitBeforeClose`** — Regression contract: TestCaptureRuntimeWaitsAndVerifiesExitBeforeClose.
  Signature: `func(t *testing.T)`.
  Calls: `fixture.exitedChecks.Load`, `len`, `newCaptureRuntimeFixture`, `process.closed.Load`, `process.earlyClose.Load`, `process.waited.Load`, `runtime.Run`, `t.Context`, `t.Fatal`, `t.Parallel`.
