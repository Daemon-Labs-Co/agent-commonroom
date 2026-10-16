# internal/toolbroker/encoding_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/encoding_test.go)

Automated regression checks for encoding; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWireEncodingRefusedBeforeReplacement`** — Regression contract: TestWireEncodingRefusedBeforeReplacement.
  Signature: `func(t *testing.T)`.
  Calls: `Decode`, `fixture`, `json.Marshal`, `string`, `strings.NewReader`, `strings.Replace`, `t.Fatal`, `t.Parallel`.
- **`TestTypedEncodingRefusedBeforeMarshal`** — Regression contract: TestTypedEncodingRefusedBeforeMarshal.
  Signature: `func(t *testing.T)`.
  Calls: `fixture`, `json.Marshal`, `json.Unmarshal`, `t.Fatal`, `t.Parallel`, `utf8.ValidString`, `validAction`, `validRequest`.
