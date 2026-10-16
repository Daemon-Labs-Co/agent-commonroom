# cmd/server/development_snapshot_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/development_snapshot_test.go)

Automated regression checks for development_snapshot; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`snapshotDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) string`.
  Calls: `hex.EncodeToString`, `sha256.Sum256`.
- **`TestStagedSnapshotSurvivesSourceReplacement`** — Regression contract: TestStagedSnapshotSurvivesSourceReplacement.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Dir`, `filepath.Join`, `info.Mode`, `info.Mode().Perm`, `os.ReadFile`, `os.Remove`, `os.Stat`, `os.WriteFile`, `snapshotDigest`, `stageDevelopmentBundle`, `string`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestUnsafeSnapshotStagingFailsWithoutArtifacts`** — Regression contract: TestUnsafeSnapshotStagingFailsWithoutArtifacts.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `len`, `os.Chmod`, `os.ReadDir`, `os.Symlink`, `os.Truncate`, `os.WriteFile`, `snapshotDigest`, `stageDevelopmentBundle`, `syscall.Mkfifo`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
