# internal/projectquota/unsupported_test.go

[Module index](internal.md) · [Open source](../../../internal/projectquota/unsupported_test.go)

Automated regression checks for unsupported; read test names for individual contracts.

Package: `projectquota`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestUnsupportedPlatformDoesNotSubstituteHostEvidence`** — Regression contract: TestUnsupportedPlatformDoesNotSubstituteHostEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `Inspect`, `context.Background`, `filepath.Clean`, `os.Open`, `quotaFixture`, `root.Close`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
