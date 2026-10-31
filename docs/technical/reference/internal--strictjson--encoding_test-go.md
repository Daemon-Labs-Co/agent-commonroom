# internal/strictjson/encoding_test.go

[Module index](internal.md) · [Open source](../../../internal/strictjson/encoding_test.go)

Automated regression checks for encoding; read test names for individual contracts.

Package: `strictjson`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestEncodingRefusesReplacement`** — Regression contract: TestEncodingRefusesReplacement.
  Signature: `func(t *testing.T)`.
  Calls: `ValidEncoding`, `t.Fatalf`, `t.Parallel`.
- **`TestEncodingPreservesValidUnicodeAndLiteralEscapes`** — Regression contract: TestEncodingPreservesValidUnicodeAndLiteralEscapes.
  Signature: `func(t *testing.T)`.
  Calls: `ValidEncoding`, `t.Fatalf`, `t.Parallel`.
- **`TestEncodingNoAllocation`** — Regression contract: TestEncodingNoAllocation.
  Signature: `func(t *testing.T)`.
  Calls: `ValidEncoding`, `panic`, `t.Fatalf`, `testing.AllocsPerRun`.
