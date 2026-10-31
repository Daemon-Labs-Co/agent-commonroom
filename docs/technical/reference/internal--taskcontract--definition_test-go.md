# internal/taskcontract/definition_test.go

[Module index](internal.md) · [Open source](../../../internal/taskcontract/definition_test.go)

Automated regression checks for definition; read test names for individual contracts.

Package: `taskcontract`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`sample`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() Definition`.
- **`TestFiniteContractsAndCanonicalBindings`** — Regression contract: TestFiniteContractsAndCanonicalBindings.
  Signature: `func(t *testing.T)`.
  Calls: `Allows`, `Bind`, `Decode`, `Digest`, `ReturnedValues`, `json.Marshal`, `json.RawMessage`, `json.Unmarshal`, `len`, `sample`, `string`, `t.Fatal`, `t.Parallel`.
- **`TestDefinitionValidationAdversaries`** — Regression contract: TestDefinitionValidationAdversaries.
  Signature: `func(t *testing.T)`.
  Calls: `Decode`, `Validate`, `append`, `json.Marshal`, `sample`, `string`, `strings.Replace`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestInputAndOutputRefusals`** — Regression contract: TestInputAndOutputRefusals.
  Signature: `func(t *testing.T)`.
  Calls: `Bind`, `ReturnedValues`, `json.RawMessage`, `sample`, `t.Fatal`, `t.Parallel`.
- **`TestFileAndBundleContracts`** — Regression contract: TestFileAndBundleContracts.
  Signature: `func(t *testing.T)`.
  Calls: `Executable`, `ValidData`, `Validate`, `sample`, `t.Fatal`, `t.Parallel`.
- **`TestRegisteredValueSchemas`** — Regression contract: TestRegisteredValueSchemas.
  Signature: `func(t *testing.T)`.
  Calls: `ValidValue`, `t.Fatal`, `t.Parallel`.
- **`TestConcreteObjectAndArrayContracts`** — Regression contract: TestConcreteObjectAndArrayContracts.
  Signature: `func(t *testing.T)`.
  Calls: `Bind`, `ValidContractValue`, `ValidData`, `sample`, `t.Fatal`, `t.Parallel`.
- **`TestUnresolvedReturnedValueReferenceIsNeverAManualBinding`** — Regression contract: TestUnresolvedReturnedValueReferenceIsNeverAManualBinding.
  Signature: `func(t *testing.T)`.
  Calls: `Bind`, `json.RawMessage`, `sample`, `t.Fatal`, `t.Parallel`.
