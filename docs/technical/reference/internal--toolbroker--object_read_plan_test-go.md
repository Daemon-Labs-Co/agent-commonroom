# internal/toolbroker/object_read_plan_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/object_read_plan_test.go)

Automated regression checks for object_read_plan; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`objectReadPlanFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*ObjectReadContainerPlan, toolInspection, toolVolumeInspection)`.
  Calls: `NewObjectReadContainerPlan`, `capturePlanFixture`, `p.labels`, `p.name`, `probeBool`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`TestObjectReadPlanFixedPolicyAndPhases`** — Regression contract: TestObjectReadPlanFixedPolicyAndPhases.
  Signature: `func(t *testing.T)`.
  Calls: `base64.RawURLEncoding.DecodeString`, `inspectJSON`, `json.Unmarshal`, `len`, `objectReadPlanFixture`, `p.CreateArgs`, `p.VerifyInspection`, `probeBool`, `slices.Contains`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestObjectReadPlanRejectsPolicyDrift`** — Regression contract: TestObjectReadPlanRejectsPolicyDrift.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `change`, `inspectJSON`, `objectReadPlanFixture`, `p.VerifyInspection`, `probeBool`, `probeInt`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **Type `objectReadPlanRunner`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T p *ObjectReadContainerPlan c toolInspection v toolVolumeInspection mode string inspects int calls int }`.
- **`(*objectReadPlanRunner).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `ctx.Err`, `f.t.Error`, `inspectJSON`, `io.WriteString`, `len`, `out.Write`, `strings.Repeat`.
- **`TestObjectReadAdmissionIndependentPasses`** — Regression contract: TestObjectReadAdmissionIndependentPasses.
  Signature: `func(t *testing.T)`.
  Calls: `objectReadPlanFixture`, `p.VerifyAdmission`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObjectReadMetadataRefusesUnboundedOrUnpinnedPlan`** — Regression contract: TestObjectReadMetadataRefusesUnboundedOrUnpinnedPlan.
  Signature: `func(t *testing.T)`.
  Calls: `NewObjectReadContainerPlan`, `objectReadPlanFixture`, `t.Fatal`, `t.Parallel`.
