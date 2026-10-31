# internal/workspacesnapshot/snapshot_unix_test.go

[Module index](internal.md) · [Open source](../../../internal/workspacesnapshot/snapshot_unix_test.go)

Automated regression checks for snapshot_unix; read test names for individual contracts.

Package: `workspacesnapshot`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticLease`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls atomic.Int32 check func(int32) error }`.
- **`(*syntheticLease).VerifyFrozen`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(_ context.Context, _ *os.Root, _ Binding) error`.
  Calls: `l.calls.Add`, `l.check`.
- **`binding`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() Binding`.
  Calls: `strings.Repeat`, `time.Unix`, `time.Unix(123, 0).UTC`, `time.Unix(123, 0).UTC().Format`.
- **`fixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*os.Root, string, Binding, []Member)`.
  Calls: `binding`, `filepath.Join`, `os.Chmod`, `os.Mkdir`, `os.OpenRoot`, `os.WriteFile`, `root.Close`, `t.Cleanup`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`capture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*Source, string, *syntheticLease)`.
  Calls: `Capture`, `fixture`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`scope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *Source) objectcustody.Scope`.
  Calls: `s.Snapshot`.
- **`receipt`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *Source) objectcustody.Receipt`.
  Calls: `hex.EncodeToString`, `int64`, `len`, `scope`, `sha256.Sum256`, `strings.Repeat`.
- **`TestCaptureActualSourceRevisionAndImmutableResults`** — Regression contract: TestCaptureActualSourceRevisionAndImmutableResults.
  Signature: `func(t *testing.T)`.
  Calls: `capture`, `filepath.Join`, `int64`, `io.ReadAll`, `lease.calls.Load`, `len`, `make`, `os.WriteFile`, `r.Close`, `r.Read`, `receipt`, `s.Current`, `s.Open`, `s.Snapshot`, `s.VerifyReceipt`, `scope`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCaptureCannotAcceptCallerRevisionOrUnsafeSelection`** — Regression contract: TestCaptureCannotAcceptCallerRevisionOrUnsafeSelection.
  Signature: `func(t *testing.T)`.
  Calls: `Capture`, `append`, `errors.New`, `fixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureRefusesUnsafeTreeEntries`** — Regression contract: TestCaptureRefusesUnsafeTreeEntries.
  Signature: `func(t *testing.T)`.
  Calls: `Capture`, `errors.Is`, `f.Close`, `f.Truncate`, `filepath.Join`, `fixture`, `os.Chmod`, `os.Create`, `os.Link`, `os.Mkdir`, `os.MkdirAll`, `os.Symlink`, `os.WriteFile`, `safeName`, `string`, `strings.Repeat`, `syscall.Mkfifo`, `t.Context`, `t.Fatal`, `t.Log`, `t.Parallel`, `t.Run`.
- **`TestCaptureExcludesGitInternalsAndBindsSelectionSeparately`** — Regression contract: TestCaptureExcludesGitInternalsAndBindsSelectionSeparately.
  Signature: `func(t *testing.T)`.
  Calls: `Capture`, `append`, `filepath.Join`, `first.Snapshot`, `fixture`, `os.Mkdir`, `os.WriteFile`, `second.Snapshot`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCaptureMutationAndRevocationWithholdSource`** — Regression contract: TestCaptureMutationAndRevocationWithholdSource.
  Signature: `func(t *testing.T)`.
  Calls: `Capture`, `capture`, `errors.New`, `filepath.Join`, `fixture`, `make`, `os.WriteFile`, `r.Close`, `r.Read`, `rune`, `s.Open`, `scope`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestSourceRejectsForgedReceiptScopeAndLimit`** — Regression contract: TestSourceRejectsForgedReceiptScopeAndLimit.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `capture`, `context.WithCancel`, `make`, `r.Close`, `r.Read`, `reader.Close`, `receipt`, `s.Open`, `s.VerifyReceipt`, `scope`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureEmptyMemberAndDeclaredByteBoundary`** — Regression contract: TestCaptureEmptyMemberAndDeclaredByteBoundary.
  Signature: `func(t *testing.T)`.
  Calls: `Capture`, `filepath.Join`, `fixture`, `io.ReadAll`, `os.WriteFile`, `reader.Close`, `s.Open`, `scope`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
