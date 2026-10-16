# internal/toolbroker/workspace_volume_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/workspace_volume_test.go)

Automated regression checks for workspace_volume; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWorkspaceVolumePrepareAndIdentity`** — Regression contract: TestWorkspaceVolumePrepareAndIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `ValidWorkspaceVolume`, `cancel`, `f.plan.WorkspaceIntent`, `len`, `r.Create`, `r.PrepareVolume`, `r.Start`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `time.Now`, `unpreparedRuntimeSetup`.
- **`TestWorkspaceVolumeRefusesAdoptionAndUncertainty`** — Regression contract: TestWorkspaceVolumeRefusesAdoptionAndUncertainty.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `f.plan.WorkspaceIntent`, `len`, `r.Create`, `r.PrepareVolume`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `unpreparedRuntimeSetup`.
- **`TestWorkspaceVolumeIntentCannotBeChangedOrRevived`** — Regression contract: TestWorkspaceVolumeIntentCannotBeChangedOrRevived.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `f.plan.WorkspaceIntent`, `len`, `r.PrepareVolume`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Format`, `unpreparedRuntimeSetup`.
