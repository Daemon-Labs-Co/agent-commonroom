# internal/toolbroker/container_plan_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/container_plan_test.go)

Automated regression checks for container_plan; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestToolPlanFixedCredentialFreeBoundary`** — Regression contract: TestToolPlanFixedCredentialFreeBoundary.
  Signature: `func(t *testing.T)`.
  Calls: `NewToolContainerPlan`, `fixture`, `len`, `p.ContainerName`, `p.RemoveArgs`, `p.RunArgs`, `p.WorkspaceVolume`, `slices.Contains`, `slices.Equal`, `strings.Contains`, `strings.Repeat`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestToolPlanRefusesUnsafeImagesAndScope`** — Regression contract: TestToolPlanRefusesUnsafeImagesAndScope.
  Signature: `func(t *testing.T)`.
  Calls: `NewToolContainerPlan`, `fixture`, `strings.Repeat`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestToolPlanOriginalLifetimeAndJobVolumeIsolation`** — Regression contract: TestToolPlanOriginalLifetimeAndJobVolumeIsolation.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `NewToolContainerPlan`, `fixture`, `now.Add`, `other.ContainerName`, `other.WorkspaceVolume`, `p.ContainerName`, `p.RunArgs`, `p.WorkspaceVolume`, `r.Scope.Deadline.Add`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestToolPlanRevocation`** — Regression contract: TestToolPlanRevocation.
  Signature: `func(t *testing.T)`.
  Calls: `NewToolContainerPlan`, `a.Revoke`, `fixture`, `p.RunArgs`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
