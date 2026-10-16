# internal/toolbroker/workspace_volume_cleanup_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/workspace_volume_cleanup_test.go)

Automated regression checks for workspace_volume_cleanup; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `cleanupVolumeDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T volume *toolVolumeInspection name string mode string inspectCalls, removed int }`.
- **`(*cleanupVolumeDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, _ []byte, out io.Writer) error`.
  Calls: `ctx.Err`, `errors.New`, `f.t.Fatal`, `f.t.Fatalf`, `inspectJSON`, `len`, `out.Write`, `strings.HasPrefix`, `strings.Repeat`, `time.Now`, `time.Now().Add`, `time.Now().Add(time.Hour).UTC`, `time.Now().Add(time.Hour).UTC().Format`.
- **`TestWorkspaceVolumeRemovalProof`** — Regression contract: TestWorkspaceVolumeRemovalProof.
  Signature: `func(t *testing.T)`.
  Calls: `RemoveWorkspaceVolume`, `path.Join`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `time.Now().Add(time.Hour).UTC`, `time.Now().Add(time.Hour).UTC().Format`, `time.Now().UTC`, `time.Now().UTC().Format`.
- **`TestWorkspaceVolumeCleanupRejectsInvalidBinding`** — Regression contract: TestWorkspaceVolumeCleanupRejectsInvalidBinding.
  Signature: `func(t *testing.T)`.
  Calls: `RemoveWorkspaceVolume`, `context.Background`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
