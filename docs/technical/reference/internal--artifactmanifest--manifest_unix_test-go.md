# internal/artifactmanifest/manifest_unix_test.go

[Module index](internal.md) · [Open source](../../../internal/artifactmanifest/manifest_unix_test.go)

Automated regression checks for manifest_unix; read test names for individual contracts.

Package: `artifactmanifest`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `journal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { sealed map[string]objectcustody.Receipt reserved map[string]objectcustody.Intent allowed bool reads int failAck bool }`.
- **`(*journal).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, i objectcustody.Intent) error`.
- **`(*journal).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, r objectcustody.Receipt) error`.
- **`(*journal).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, s objectcustody.Scope, r objectcustody.Receipt) error`.
- **`(*journal).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, r objectcustody.Receipt) error`.
- **`definition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(bundle bool) taskcontract.Definition`.
- **`fixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, d taskcontract.Definition, contents ...string) (*objectcustody.Store, *journal, objectcustody.ExecutionScope, []Claim, string)`.
  Calls: `append`, `bytes.NewBufferString`, `objectcustody.New`, `os.Chmod`, `os.OpenRoot`, `root.Close`, `store.Put`, `strings.Repeat`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`, `t.TempDir`, `taskcontract.Digest`, `taskcontract.Validate`.
- **`TestManifestCanonicalBundleAndActualRevalidation`** — Regression contract: TestManifestCanonicalBundleAndActualRevalidation.
  Signature: `func(t *testing.T)`.
  Calls: `Resolve`, `Revalidate`, `definition`, `fixture`, `int64`, `len`, `manifestDigest`, `reflect.DeepEqual`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestManifestPreflightRefusesEntireSetBeforeReads`** — Regression contract: TestManifestPreflightRefusesEntireSetBeforeReads.
  Signature: `func(t *testing.T)`.
  Calls: `Resolve`, `append`, `cancel`, `context.WithCancel`, `definition`, `errors.Is`, `fixture`, `reflect.DeepEqual`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `taskcontract.Digest`.
- **`TestRegisteredFormatsCheckActualSealedBytes`** — Regression contract: TestRegisteredFormatsCheckActualSealedBytes.
  Signature: `func(t *testing.T)`.
  Calls: `Resolve`, `definition`, `fixture`, `len`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestManifestIntegrityAndReleaseAcknowledgment`** — Regression contract: TestManifestIntegrityAndReleaseAcknowledgment.
  Signature: `func(t *testing.T)`.
  Calls: `Resolve`, `Revalidate`, `definition`, `delete`, `entries[0].Name`, `filepath.Join`, `fixture`, `len`, `os.ReadDir`, `os.WriteFile`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestManifestMetadataIsCanonicalShapeNotByteAuthority`** — Regression contract: TestManifestMetadataIsCanonicalShapeNotByteAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `Resolve`, `ValidateMetadata`, `append`, `definition`, `fixture`, `manifestDigest`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestManifestMalformedLaterReceiptRefusesBeforeAnyRead`** — Regression contract: TestManifestMalformedLaterReceiptRefusesBeforeAnyRead.
  Signature: `func(t *testing.T)`.
  Calls: `Resolve`, `definition`, `fixture`, `t.Context`, `t.Fatal`, `t.Parallel`.
