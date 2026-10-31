# internal/projectquota/set_test.go

[Module index](internal.md) · [Open source](../../../internal/projectquota/set_test.go)

Automated regression checks for set; read test names for individual contracts.

Package: `projectquota`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`quotaSetFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) ([]Root, []*syntheticQuotaReader)`.
  Calls: `binary.LittleEndian.PutUint64`, `dir.Close`, `make`, `os.Open`, `quotaFixture`, `t.Cleanup`, `t.Fatal`, `t.Helper`, `t.TempDir`, `uint64`.
- **`TestInspectSetPinsAllRootsAndBoundsReservedCapacity`** — Regression contract: TestInspectSetPinsAllRootsAndBoundsReservedCapacity.
  Signature: `func(t *testing.T)`.
  Calls: `inspectSet`, `len`, `quotaSetFixture`, `roots[i].Directory.Stat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestInspectSetRefusesInvalidCompletePlanBeforeOpen`** — Regression contract: TestInspectSetRefusesInvalidCompletePlanBeforeOpen.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `inspectSet`, `len`, `quotaSetFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `uint64`.
- **`TestInspectSetFailureWithholdsPrefixAndJoinsReaders`** — Regression contract: TestInspectSetFailureWithholdsPrefixAndJoinsReaders.
  Signature: `func(t *testing.T)`.
  Calls: `binary.LittleEndian.PutUint16`, `binary.LittleEndian.PutUint64`, `cancel`, `context.WithCancel`, `inspectSet`, `quotaSetFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestInspectSetOwnsPolicySnapshotAndRefusesCanceledAdmission`** — Regression contract: TestInspectSetOwnsPolicySnapshotAndRefusesCanceledAdmission.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `inspectSet`, `len`, `quotaSetFixture`, `t.Context`, `t.Fatal`, `t.Parallel`.
