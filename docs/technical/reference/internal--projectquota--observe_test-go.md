# internal/projectquota/observe_test.go

[Module index](internal.md) · [Open source](../../../internal/projectquota/observe_test.go)

Automated regression checks for observe; read test names for individual contracts.

Package: `projectquota`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticQuotaReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { root rootObservation status, quota []byte calls []string change func(string, int) closes int fail string }`.
- **`(*syntheticQuotaReader).step`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(name string) error`.
  Calls: `append`, `len`, `r.change`.
- **`(*syntheticQuotaReader).Root`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context) (rootObservation, error)`.
  Calls: `r.step`.
- **`(*syntheticQuotaReader).Status`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context) ([]byte, error)`.
  Calls: `append`, `r.step`.
- **`(*syntheticQuotaReader).Quota`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, project uint32) ([]byte, error)`.
  Calls: `append`, `r.step`.
- **`(*syntheticQuotaReader).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
- **`quotaFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (Expected, *syntheticQuotaReader)`.
  Calls: `binary.LittleEndian.PutUint16`, `binary.LittleEndian.PutUint32`, `binary.LittleEndian.PutUint64`, `make`.
- **`TestObserveEnforcedProjectQuota`** — Regression contract: TestObserveEnforcedProjectQuota.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `inspectOwned`, `len`, `quotaFixture`, `t.Fatal`, `t.Parallel`.
- **`TestObserveQuotaRefusals`** — Regression contract: TestObserveQuotaRefusals.
  Signature: `func(t *testing.T)`.
  Calls: `binary.LittleEndian.PutUint16`, `binary.LittleEndian.PutUint32`, `binary.LittleEndian.PutUint64`, `context.Background`, `inspectOwned`, `quotaFixture`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObserveQuotaChangedOrCanceled`** — Regression contract: TestObserveQuotaChangedOrCanceled.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `inspectOwned`, `quotaFixture`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestInvalidExpectedRefusesBeforeKernelReads`** — Regression contract: TestInvalidExpectedRefusesBeforeKernelReads.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `inspectOwned`, `len`, `quotaFixture`, `t.Fatal`, `t.Parallel`.
