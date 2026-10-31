# cmd/object-read/main_test.go

[Module index](cmd.md) · [Open source](../../../cmd/object-read/main_test.go)

Automated regression checks for main; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`testRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() objectread.Request`.
  Calls: `strings.Repeat`.
- **Type `testStream`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { bytes.Buffer closed bool }`.
- **`(*testStream).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
- **`TestReaderCommandFiniteCanonicalRequest`** — Regression contract: TestReaderCommandFiniteCanonicalRequest.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `base64.RawURLEncoding.EncodeToString`, `decodeRequest`, `json.Marshal`, `run`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `testRequest`.
