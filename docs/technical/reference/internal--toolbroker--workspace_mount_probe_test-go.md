# internal/toolbroker/workspace_mount_probe_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/workspace_mount_probe_test.go)

Automated regression checks for workspace_mount_probe; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `probeScenario`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T p *WorkspaceMountProbe c toolInspection v toolVolumeInspection mode string events []string exists, started, removed, finishedAbsent bool cancel context.CancelFunc }`.
- **`(*probeScenario).Begin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, WorkspaceProbeIntent) error`.
  Calls: `append`, `errors.New`.
- **`(*probeScenario).Bound`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, _ WorkspaceProbeIntent, id string) error`.
  Calls: `append`, `errors.New`, `f.t.Fatal`.
- **`(*probeScenario).Finish`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ WorkspaceProbeIntent, id string, absent bool) error`.
  Calls: `append`, `ctx.Err`, `errors.New`, `f.cancel`, `f.t.Fatal`.
- **`(*probeScenario).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `append`, `ctx.Err`, `errors.New`, `f.cancel`, `f.p.createArgs`, `f.t.Fatal`, `f.t.Fatalf`, `inspectJSON`, `json.Marshal`, `json.Unmarshal`, `len`, `out.Write`, `outError`, `probeBool`, `probeInt`, `slices.Contains`, `slices.Equal`, `string`, `strings.HasPrefix`, `strings.Repeat`, `strings.Replace`.
- **`probeBool`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v bool) *bool`.
- **`probeInt`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v int64) *int64`.
- **`newProbeScenario`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mode string) *probeScenario`.
  Calls: `NewWorkspaceMountProbe`, `append`, `p.labels`, `p.name`, `path.Join`, `probeBool`, `probeInt`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`TestWorkspaceMountProbeLifecycle`** — Regression contract: TestWorkspaceMountProbeLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `f.p.Run`, `len`, `newProbeScenario`, `slices.Contains`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `workspacemount.ValidProof`.
- **`TestProbeCreateIsReadOnlyNoCopy`** — Regression contract: TestProbeCreateIsReadOnlyNoCopy.
  Signature: `func(t *testing.T)`.
  Calls: `f.p.createArgs`, `newProbeScenario`, `slices.Contains`, `t.Fatalf`, `t.Parallel`.
- **`TestProbeRequiresTrustedDependencies`** — Regression contract: TestProbeRequiresTrustedDependencies.
  Signature: `func(t *testing.T)`.
  Calls: `NewWorkspaceMountProbe`, `newProbeScenario`, `t.Fatal`, `t.Parallel`.
- **`TestProbeRejectsInspectedPolicyDrift`** — Regression contract: TestProbeRejectsInspectedPolicyDrift.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `change`, `context.Background`, `f.p.Run`, `newProbeScenario`, `probeBool`, `probeInt`, `strings.Repeat`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`.
- **`outError`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(out io.Writer, s string) error`.
  Calls: `io.WriteString`.
- **Type `probeAdmissionDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { f *probeScenario extra, replace bool inspections int }`.
- **`(*probeAdmissionDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `d.f.Run`, `io.WriteString`, `strings.Repeat`.
- **`TestWorkspaceProbeAdmissionAllowsOnlyExactStoppedReadOnlyHelper`** — Regression contract: TestWorkspaceProbeAdmissionAllowsOnlyExactStoppedReadOnlyHelper.
  Signature: `func(t *testing.T)`.
  Calls: `NewWorkspaceMountProbe`, `VerifyRetainedWorkspace`, `newProbeScenario`, `p.VerifyAdmission`, `probeBool`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`.
