# internal/toolbroker/capture_plan_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/capture_plan_test.go)

Automated regression checks for capture_plan; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`capturePlanFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*CaptureContainerPlan, toolInspection, toolVolumeInspection, toolVolumeInspection)`.
  Calls: `NewCaptureContainerPlan`, `append`, `newProbeScenario`, `p.labels`, `p.name`, `p.objectLabels`, `path.Join`, `probeBool`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`TestCapturePlanFixedArgumentsAndPhases`** — Regression contract: TestCapturePlanFixedArgumentsAndPhases.
  Signature: `func(t *testing.T)`.
  Calls: `base64.RawURLEncoding.DecodeString`, `capturePlanFixture`, `inspectJSON`, `json.Unmarshal`, `len`, `p.CreateArgs`, `p.VerifyInspection`, `probeBool`, `slices.Contains`, `t.Fatal`, `t.Parallel`, `workspacecapture.RequestDigest`.
- **`TestCapturePlanRejectsPolicyDrift`** — Regression contract: TestCapturePlanRejectsPolicyDrift.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `capturePlanFixture`, `inspectJSON`, `mutate`, `p.VerifyInspection`, `probeBool`, `probeInt`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCapturePlanRefusesUntrustedInputs`** — Regression contract: TestCapturePlanRefusesUntrustedInputs.
  Signature: `func(t *testing.T)`.
  Calls: `NewCaptureContainerPlan`, `append`, `capturePlanFixture`, `change`, `inspectJSON`, `p.VerifyInspection`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **Type `captureAdmissionRunner`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T p *CaptureContainerPlan c toolInspection src, obj toolVolumeInspection mode string inspections int calls int cancel context.CancelFunc }`.
- **`(*captureAdmissionRunner).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `append`, `ctx.Deadline`, `ctx.Err`, `errors.New`, `f.cancel`, `f.p.name`, `f.t.Fatal`, `inspectJSON`, `len`, `out.Write`, `probeBool`, `slices.Contains`, `strings.Repeat`.
- **`TestCaptureAdmissionInventoriesAndRechecks`** — Regression contract: TestCaptureAdmissionInventoriesAndRechecks.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `capturePlanFixture`, `context.Background`, `context.WithCancel`, `p.name`, `plan.VerifyAdmission`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureAdmissionRefusesBeforeDaemon`** — Regression contract: TestCaptureAdmissionRefusesBeforeDaemon.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `capturePlanFixture`, `context.Background`, `context.WithCancel`, `p.VerifyAdmission`, `t.Fatal`, `t.Parallel`.
