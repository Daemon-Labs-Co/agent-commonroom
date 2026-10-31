# internal/objectcustody/store_unix_test.go

[Module index](internal.md) · [Open source](../../../internal/objectcustody/store_unix_test.go)

Automated regression checks for store_unix; read test names for individual contracts.

Package: `objectcustody`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `custodyJournal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { mu sync.Mutex capacity, total int64 reserved map[string]Intent sealed map[string]Receipt events []string allowed bool failReserve, failSeal, failRead, failAck bool afterReserve func() afterAuthorize func() }`.
- **`(*custodyJournal).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, intent Intent) error`.
  Calls: `append`, `errors.New`, `j.afterReserve`, `j.mu.Lock`, `j.mu.Unlock`.
- **`(*custodyJournal).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, receipt Receipt) error`.
  Calls: `append`, `errors.New`, `j.mu.Lock`, `j.mu.Unlock`.
- **`(*custodyJournal).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, scope Scope, receipt Receipt) error`.
  Calls: `append`, `errors.New`, `j.afterAuthorize`, `j.mu.Lock`, `j.mu.Unlock`.
- **`(*custodyJournal).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, receipt Receipt) error`.
  Calls: `append`, `errors.New`, `j.mu.Lock`, `j.mu.Unlock`.
- **`syntheticScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() Scope`.
  Calls: `strings.Repeat`.
- **`custodyFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*Store, *custodyJournal, string)`.
  Calls: `New`, `os.Chmod`, `os.OpenRoot`, `root.Close`, `t.Cleanup`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`TestCustodyActualBytesRequireAccountingSealAndReleaseAcknowledgment`** — Regression contract: TestCustodyActualBytesRequireAccountingSealAndReleaseAcknowledgment.
  Signature: `func(t *testing.T)`.
  Calls: `bytes.Equal`, `bytes.NewReader`, `custodyFixture`, `filepath.Join`, `int64`, `len`, `objectName`, `os.ReadFile`, `store.Get`, `store.Put`, `strings.Join`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `validReceipt`.
- **Type `readCounter`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ calls int }`.
- **`(*readCounter).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
- **`TestCustodyRefusesBeforeReadingOrAllocatingWithoutAdmission`** — Regression contract: TestCustodyRefusesBeforeReadingOrAllocatingWithoutAdmission.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `custodyFixture`, `errors.Is`, `int64`, `len`, `os.ReadDir`, `store.Put`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCustodyMaximumReservationRemainsChargedAtExactBoundary`** — Regression contract: TestCustodyMaximumReservationRemainsChargedAtExactBoundary.
  Signature: `func(t *testing.T)`.
  Calls: `custodyFixture`, `errors.Is`, `store.Put`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCustodyUncertainWritesRetainBytesAndChargeWithoutReceipt`** — Regression contract: TestCustodyUncertainWritesRetainBytesAndChargeWithoutReceipt.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `custodyFixture`, `errors.New`, `int64`, `len`, `os.ReadDir`, `store.Put`, `store.root.Close`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **Type `errorReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(errorReader).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `copy`, `errors.New`.
- **Type `changingReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { source *strings.Reader change func() }`.
- **`(*changingReader).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `r.change`, `r.source.Read`.
- **`TestCustodyUnsealedOrForeignClaimsCannotReleaseBytes`** — Regression contract: TestCustodyUnsealedOrForeignClaimsCannotReleaseBytes.
  Signature: `func(t *testing.T)`.
  Calls: `custodyFixture`, `delete`, `store.Get`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCustodyRefusesSymlinkHardlinkPermissionsAndTamperedActualBytes`** — Regression contract: TestCustodyRefusesSymlinkHardlinkPermissionsAndTamperedActualBytes.
  Signature: `func(t *testing.T)`.
  Calls: `custodyFixture`, `filepath.Join`, `objectName`, `os.Chmod`, `os.Link`, `os.ReadFile`, `os.Remove`, `os.Symlink`, `os.WriteFile`, `store.Get`, `store.Put`, `string`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **Type `emptyReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ reads int }`.
- **`(*emptyReader).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func([]byte) (int, error)`.
- **`TestCustodyEmptyContentAndNonProgressingStreamsAreBounded`** — Regression contract: TestCustodyEmptyContentAndNonProgressingStreamsAreBounded.
  Signature: `func(t *testing.T)`.
  Calls: `custodyFixture`, `errors.Is`, `len`, `store.Get`, `store.Put`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCustodyParallelObjectsNeverOverwriteOrLoseCharges`** — Regression contract: TestCustodyParallelObjectsNeverOverwriteOrLoseCharges.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `custodyFixture`, `len`, `make`, `store.Get`, `store.Put`, `string`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestCustodyRequiresPrivateOwnedRootAndJournal`** — Regression contract: TestCustodyRequiresPrivateOwnedRootAndJournal.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `custodyFixture`, `errors.Is`, `os.Chmod`, `t.Fatal`, `t.Parallel`.
- **`TestCustodyTypedNilDependenciesAndPreexistingObjectsRefuseSafely`** — Regression contract: TestCustodyTypedNilDependenciesAndPreexistingObjectsRefuseSafely.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `custodyFixture`, `entries[0].Name`, `errors.Is`, `filepath.Join`, `len`, `objectName`, `os.ReadDir`, `os.ReadFile`, `os.WriteFile`, `store.Put`, `string`, `syntheticScope`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **Type `zeroReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(zeroReader).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `clear`, `len`.
- **`TestCustodyExactMaximumActualBytesAndOverBoundStream`** — Regression contract: TestCustodyExactMaximumActualBytesAndOverBoundStream.
  Signature: `func(t *testing.T)`.
  Calls: `custodyFixture`, `entry.Info`, `errors.Is`, `info.Size`, `int64`, `io.LimitReader`, `len`, `os.ReadDir`, `store.Get`, `store.Put`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCustodySealsActualReadBackBytesRatherThanWriteBuffers`** — Regression contract: TestCustodySealsActualReadBackBytesRatherThanWriteBuffers.
  Signature: `func(t *testing.T)`.
  Calls: `custodyFixture`, `errors.Is`, `file.Sync`, `file.WriteAt`, `len`, `store.Put`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCustodyLogicalMemberNamesCannotSelectFilesystemPaths`** — Regression contract: TestCustodyLogicalMemberNamesCannotSelectFilesystemPaths.
  Signature: `func(t *testing.T)`.
  Calls: `custodyFixture`, `len`, `store.Get`, `store.Put`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCustodyPutCloseUncertaintyWithholdsReceiptWithoutRetry`** — Regression contract: TestCustodyPutCloseUncertaintyWithholdsReceiptWithoutRetry.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `custodyFixture`, `errors.Is`, `errors.New`, `f.Close`, `f.Stat`, `info.IsDir`, `len`, `os.ReadDir`, `store.Put`, `strings.Contains`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCustodyGetCloseUncertaintyWithholdsAuthorizedBytes`** — Regression contract: TestCustodyGetCloseUncertaintyWithholdsAuthorizedBytes.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `custodyFixture`, `errors.Is`, `errors.New`, `f.Close`, `f.Stat`, `filepath.Join`, `info.IsDir`, `len`, `objectName`, `os.ReadFile`, `store.Get`, `store.Put`, `string`, `strings.Contains`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
