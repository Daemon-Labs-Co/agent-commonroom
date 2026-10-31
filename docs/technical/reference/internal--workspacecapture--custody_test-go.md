# internal/workspacecapture/custody_test.go

[Module index](internal.md) · [Open source](../../../internal/workspacecapture/custody_test.go)

Automated regression checks for custody; read test names for individual contracts.

Package: `workspacecapture`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `transferJournal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { intents []objectcustody.Intent receipts []objectcustody.Receipt charged int64 reserve func(objectcustody.Intent) error seal func(objectcustody.Receipt) error reads int }`.
- **`(*transferJournal).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i objectcustody.Intent) error`.
  Calls: `append`, `ctx.Deadline`, `ctx.Err`, `errors.New`, `j.reserve`, `time.Until`.
- **`(*transferJournal).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `append`, `ctx.Err`, `j.seal`.
- **`(*transferJournal).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Scope, objectcustody.Receipt) error`.
- **`(*transferJournal).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
- **`transferRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*os.Root, string)`.
  Calls: `os.Chmod`, `os.OpenRoot`, `root.Close`, `t.Cleanup`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`TestCaptureToCustodyActualMemberAndReceipt`** — Regression contract: TestCaptureToCustodyActualMemberAndReceipt.
  Signature: `func(t *testing.T)`.
  Calls: `captureFixture`, `captureToCustody`, `filepath.Join`, `hex.EncodeToString`, `int64`, `len`, `os.ReadDir`, `os.ReadFile`, `sha256.Sum256`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `transferRoot`.
- **`TestCaptureToCustodyFailureRetainsChargesAndWithholdsResult`** — Regression contract: TestCaptureToCustodyFailureRetainsChargesAndWithholdsResult.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `captureFixture`, `captureToCustody`, `errors.Is`, `errors.New`, `filepath.Join`, `len`, `os.ReadDir`, `os.WriteFile`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `transferRoot`.
- **`TestCapturedSourceJournalCannotExpandScopeOrReleaseBytes`** — Regression contract: TestCapturedSourceJournalCannotExpandScopeOrReleaseBytes.
  Signature: `func(t *testing.T)`.
  Calls: `captureFixture`, `capturedScope`, `j.AcknowledgeRead`, `j.AuthorizeRead`, `j.Reserve`, `j.Seal`, `len`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `withFrozenSource`.
- **`TestCaptureToCustodyMissingJournalRefusesBeforeSource`** — Regression contract: TestCaptureToCustodyMissingJournalRefusesBeforeSource.
  Signature: `func(t *testing.T)`.
  Calls: `captureFixture`, `captureToCustody`, `t.Context`, `t.Fatal`, `t.Parallel`, `transferRoot`.
