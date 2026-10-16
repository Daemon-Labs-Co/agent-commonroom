# tools/source-reference/main_test.go

[Module index](tools.md) · [Open source](../../tools/source-reference/main_test.go)

Automated regression checks for main; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestReferenceIncludesTypesMethodsAndCalls`** — Regression contract: TestReferenceIncludesTypesMethodsAndCalls.
  Signature: `func(t *testing.T)`.
  Calls: `describeGo`, `strings.Contains`, `t.Errorf`, `t.Fatal`.
- **`TestGenerateExcludesPrivateStateAndIsRepeatable`** — Regression contract: TestGenerateExcludesPrivateStateAndIsRepeatable.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.Dir`, `filepath.Join`, `generate`, `os.MkdirAll`, `os.ReadFile`, `os.WriteFile`, `string`, `strings.Contains`, `t.Errorf`, `t.Fatal`, `t.TempDir`.
