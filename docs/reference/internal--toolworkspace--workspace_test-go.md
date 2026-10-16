# internal/toolworkspace/workspace_test.go

[Module index](internal.md) · [Open source](../../internal/toolworkspace/workspace_test.go)

Automated regression checks for workspace; read test names for individual contracts.

Package: `toolworkspace`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`fixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*Workspace, string)`.
  Calls: `New`, `os.OpenRoot`, `root.Close`, `t.Cleanup`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`TestBoundedReadAndAtomicWrite`** — Regression contract: TestBoundedReadAndAtomicWrite.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `e.Name`, `filepath.Join`, `fixture`, `newInfo.Mode`, `newInfo.Mode().Perm`, `old.Close`, `old.Stat`, `os.Open`, `os.ReadDir`, `os.SameFile`, `os.Stat`, `os.WriteFile`, `string`, `strings.HasPrefix`, `t.Fatal`, `t.Parallel`, `w.Read`, `w.Write`.
- **`TestRefusedPathsAndEscapingSymlinks`** — Regression contract: TestRefusedPathsAndEscapingSymlinks.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `errors.Is`, `filepath.Join`, `fixture`, `os.ReadFile`, `os.Stat`, `os.Symlink`, `os.WriteFile`, `string`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `w.Read`, `w.Write`.
- **`TestSpecialLinkedWritableAndOversizedFilesRefused`** — Regression contract: TestSpecialLinkedWritableAndOversizedFilesRefused.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `filepath.Join`, `fixture`, `os.Chmod`, `os.Link`, `os.WriteFile`, `strings.Repeat`, `syscall.Mkfifo`, `t.Fatal`, `t.Parallel`, `w.Read`, `w.Write`.
- **`TestCancellationAndUnconfirmedDurability`** — Regression contract: TestCancellationAndUnconfirmedDurability.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `cancel`, `context.Background`, `context.WithCancel`, `errors.New`, `filepath.Join`, `fixture`, `os.ReadFile`, `string`, `t.Fatal`, `t.Parallel`, `w.Read`, `w.Write`, `w.write`.
- **`TestParentReplacementDuringSyncIsNotAcknowledged`** — Regression contract: TestParentReplacementDuringSyncIsNotAcknowledged.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `errors.Is`, `filepath.Join`, `fixture`, `os.Mkdir`, `os.ReadFile`, `os.Rename`, `os.SameFile`, `os.Stat`, `parent.Stat`, `string`, `t.Error`, `t.Fatal`, `t.Parallel`, `w.write`.
