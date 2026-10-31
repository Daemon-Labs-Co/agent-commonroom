# internal/toolbroker/container_inspection_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/container_inspection_test.go)

Automated regression checks for container_inspection; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`boolValue`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v bool) *bool`.
- **`intValue`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v int64) *int64`.
- **`inspectionFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*ToolContainerPlan, toolInspection, toolVolumeInspection, string, time.Time)`.
  Calls: `NewToolContainerPlan`, `append`, `boolValue`, `fixture`, `intValue`, `p.ContainerName`, `p.WorkspaceVolume`, `p.labels`, `path.Join`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `toolEnvironment`.
- **`inspectJSON`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, v any) []byte`.
  Calls: `json.Marshal`, `t.Fatal`, `t.Helper`.
- **`TestToolInspectionAcceptsDeclaredPolicy`** — Regression contract: TestToolInspectionAcceptsDeclaredPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `inspectJSON`, `inspectionFixture`, `len`, `p.VerifyInspection`, `p.VolumeCreateArgs`, `slices.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestToolInspectionRejectsContainerBoundaryChanges`** — Regression contract: TestToolInspectionRejectsContainerBoundaryChanges.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `boolValue`, `change`, `inspectJSON`, `inspectionFixture`, `intValue`, `json.RawMessage`, `p.VerifyInspection`, `strings.Repeat`, `t.Fatalf`, `t.Parallel`.
- **`TestToolInspectionRejectsRedirectedVolume`** — Regression contract: TestToolInspectionRejectsRedirectedVolume.
  Signature: `func(t *testing.T)`.
  Calls: `change`, `inspectJSON`, `inspectionFixture`, `p.VerifyInspection`, `t.Fatal`, `t.Parallel`.
- **`TestToolInspectionMalformedAndRevoked`** — Regression contract: TestToolInspectionMalformedAndRevoked.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `inspectJSON`, `inspectionFixture`, `p.VerifyInspection`, `p.VolumeCreateArgs`, `p.authority.Revoke`, `string`, `strings.Repeat`, `strings.Replace`, `t.Fatal`, `t.Parallel`.
- **`TestToolPlanMatchesInspectionEnvironment`** — Regression contract: TestToolPlanMatchesInspectionEnvironment.
  Signature: `func(t *testing.T)`.
  Calls: `inspectionFixture`, `len`, `p.RunArgs`, `p.labels`, `sameMap`, `slices.Contains`, `strings.Cut`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `toolEnvironment`.
