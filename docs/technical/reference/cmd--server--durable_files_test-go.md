# cmd/server/durable_files_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/durable_files_test.go)

Automated regression checks for durable_files; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestPrivateReplacementRequiresDirectorySync`** — Regression contract: TestPrivateReplacementRequiresDirectorySync.
  Signature: `func(t *testing.T)`.
  Calls: `errors.Is`, `errors.New`, `filepath.Join`, `len`, `os.ReadDir`, `os.WriteFile`, `readPrivateBounded`, `replacePrivateDurably`, `string`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestFailedReplacementDoesNotAcknowledgeDirectorySync`** — Regression contract: TestFailedReplacementDoesNotAcknowledgeDirectorySync.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Join`, `len`, `os.Mkdir`, `os.ReadDir`, `replacePrivateDurably`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
