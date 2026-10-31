# cmd/workspace-bootstrap/main_test.go

[Module index](cmd.md) · [Open source](../../../cmd/workspace-bootstrap/main_test.go)

Automated regression checks for main; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestGitCommandHasFixedExecutableAndConfiguration`** — Regression contract: TestGitCommandHasFixedExecutableAndConfiguration.
  Signature: `func(t *testing.T)`.
  Calls: `(gitRunner{home: "/tmp/synthetic-home"}).command`, `context.Background`, `slices.Contains`, `slices.Equal`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestGitOutputBound`** — Regression contract: TestGitOutputBound.
  Signature: `func(t *testing.T)`.
  Calls: `make`, `t.Fatal`, `t.Parallel`, `w.Len`, `w.Write`.
