# cmd/server/main_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/main_test.go)

Automated regression checks for main; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestChecksumDetectsAlteration`** — Regression contract: TestChecksumDetectsAlteration.
  Signature: `func(t *testing.T)`.
  Calls: `checksum`, `t.Fatal`, `t.Parallel`.
