# cmd/server/artifact_source_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_source_test.go)

Automated regression checks for artifact_source; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticFrozenArtifactLease`**: This lease is explicitly synthetic. Actual worker/volume fencing and read-only mount attestation remain required before production assembly can use Capture.
  Declaration: `struct{ deny bool }`.
- **`(*syntheticFrozenArtifactLease).VerifyFrozen`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, *os.Root, workspacesnapshot.Binding) error`.
  Calls: `errors.New`.
- **Type `artifactStreamProbe`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { io.ReadCloser read func() }`.
- **`(artifactStreamProbe).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(b []byte) (int, error)`.
  Calls: `p.ReadCloser.Read`, `p.read`.
- **`TestActualSourceSnapshotFeedsDurableCustodyAndRefusesChangedSource`** — Regression contract: TestActualSourceSnapshotFeedsDurableCustodyAndRefusesChangedSource.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `artifactmanifest.Resolve`, `artifactmanifest.Revalidate`, `errors.New`, `filepath.Join`, `int64`, `len`, `newControllerArtifactJournal`, `os.Chmod`, `os.OpenRoot`, `os.ReadDir`, `os.WriteFile`, `root.Close`, `source.Current`, `source.Open`, `source.Snapshot`, `source.VerifyReceipt`, `store.Get`, `store.Put`, `stream.Close`, `string`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `time.Unix`, `time.Unix(123, 0).UTC`, `time.Unix(123, 0).UTC().Format`, `workspacesnapshot.Capture`.
- **`TestSourceChangeDuringCopyCannotSealReceipt`** — Regression contract: TestSourceChangeDuringCopyCannotSealReceipt.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `artifactTotal`, `filepath.Join`, `len`, `newControllerArtifactJournal`, `os.Chmod`, `os.OpenRoot`, `os.ReadDir`, `os.WriteFile`, `root.Close`, `source.Current`, `source.Open`, `source.Snapshot`, `source.VerifyReceipt`, `store.Put`, `stream.Close`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `time.Unix`, `time.Unix(123, 0).UTC`, `time.Unix(123, 0).UTC().Format`, `workspacesnapshot.Capture`.
- **Type `mutatingArtifactSource`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { io.ReadCloser mutate func() }`.
- **`(*mutatingArtifactSource).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `r.ReadCloser.Read`, `r.mutate`.
