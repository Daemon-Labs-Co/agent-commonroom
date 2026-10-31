# internal/workspacecapture/contract_test.go

[Module index](internal.md) · [Open source](../../../internal/workspacecapture/contract_test.go)

Automated regression checks for contract; read test names for individual contracts.

Package: `workspacecapture`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`captureDefinition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() taskcontract.Definition`.
- **`TestContractRequestRequiredOptionalAndCanonical`** — Regression contract: TestContractRequestRequiredOptionalAndCanonical.
  Signature: `func(t *testing.T)`.
  Calls: `ContractRequest`, `ValidateContractRequest`, `captureDefinition`, `captureFixture`, `len`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `taskcontract.Digest`.
- **`TestContractRequestRejectsOmissionsAndSubstitutions`** — Regression contract: TestContractRequestRejectsOmissionsAndSubstitutions.
  Signature: `func(t *testing.T)`.
  Calls: `ContractRequest`, `ValidateContractRequest`, `append`, `captureDefinition`, `captureFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskcontract.Digest`.
- **`TestContractRequestBundleMembersAndDuplicatePaths`** — Regression contract: TestContractRequestBundleMembersAndDuplicatePaths.
  Signature: `func(t *testing.T)`.
  Calls: `ContractRequest`, `append`, `captureDefinition`, `captureFixture`, `len`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `taskcontract.Digest`.
- **`TestContractCaptureToActualFormatManifest`** — Regression contract: TestContractCaptureToActualFormatManifest.
  Signature: `func(t *testing.T)`.
  Calls: `ContractRequest`, `append`, `artifactmanifest.Resolve`, `artifactmanifest.Revalidate`, `captureDefinition`, `captureFixture`, `captureToCustody`, `filepath.Join`, `files[0].Name`, `len`, `objectcustody.New`, `os.ReadDir`, `os.WriteFile`, `reflect.DeepEqual`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskcontract.Digest`, `transferRoot`.
- **`TestContractCapturedInvalidFormatRetainsCustody`** — Regression contract: TestContractCapturedInvalidFormatRetainsCustody.
  Signature: `func(t *testing.T)`.
  Calls: `ContractRequest`, `artifactmanifest.Resolve`, `captureDefinition`, `captureFixture`, `captureToCustody`, `len`, `objectcustody.New`, `os.ReadDir`, `t.Context`, `t.Fatal`, `t.Parallel`, `taskcontract.Digest`, `transferRoot`.
