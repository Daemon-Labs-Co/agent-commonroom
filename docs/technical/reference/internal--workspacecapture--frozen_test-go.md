# internal/workspacecapture/frozen_test.go

[Module index](internal.md) · [Open source](../../../internal/workspacecapture/frozen_test.go)

Automated regression checks for frozen; read test names for individual contracts.

Package: `workspacecapture`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `fixtureContainerView`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { dir string proof workspacemount.Proof opened bool }`.
- **`(*fixtureContainerView).OpenRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (*os.Root, error)`.
  Calls: `os.OpenRoot`.
- **`(*fixtureContainerView).RootIdentity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (os.FileInfo, error)`.
  Calls: `os.Lstat`.
- **`(*fixtureContainerView).Observe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, challenge string) (workspacemount.Proof, error)`.
  Calls: `ctx.Err`.
- **Type `fixtureAuthority`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls int check func(context.Context, workspacesnapshot.Binding, workspacemount.Proof) error }`.
- **`(*fixtureAuthority).AuthorizeFrozenSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, b workspacesnapshot.Binding, p workspacemount.Proof) error`.
  Calls: `a.check`, `ctx.Err`.
- **`captureFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*fixtureContainerView, workspacesnapshot.Binding, []workspacesnapshot.Member)`.
  Calls: `filepath.Join`, `os.Chmod`, `os.WriteFile`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `t.TempDir`, `time.Unix`, `time.Unix(123, 0).UTC`, `time.Unix(123, 0).UTC().Format`.
- **`capturedScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(source *workspacesnapshot.Source) objectcustody.Scope`.
  Calls: `source.Snapshot`.
- **`TestFrozenCaptureActualBytesAndScopeExpiry`** — Regression contract: TestFrozenCaptureActualBytesAndScopeExpiry.
  Signature: `func(t *testing.T)`.
  Calls: `captureFixture`, `capturedScope`, `escaped.Current`, `escaped.Open`, `fresh.Close`, `fresh.Read`, `io.ReadAll`, `make`, `reader.Close`, `reader.Read`, `source.Open`, `source.Snapshot`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `withFrozenSource`.
- **`TestFrozenCaptureRefusesViewAndAuthorityChanges`** — Regression contract: TestFrozenCaptureRefusesViewAndAuthorityChanges.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `captureFixture`, `context.WithCancel`, `errors.New`, `filepath.Join`, `os.Symlink`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `withFrozenSource`.
- **`TestFrozenCaptureRevocationBeforeRead`** — Regression contract: TestFrozenCaptureRevocationBeforeRead.
  Signature: `func(t *testing.T)`.
  Calls: `captureFixture`, `capturedScope`, `ctx.Deadline`, `errors.Is`, `errors.New`, `make`, `reader.Close`, `reader.Read`, `source.Open`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Until`, `withFrozenSource`.
- **`TestFrozenCaptureMissingAuthorityDoesNotOpenRoot`** — Regression contract: TestFrozenCaptureMissingAuthorityDoesNotOpenRoot.
  Signature: `func(t *testing.T)`.
  Calls: `WithFrozenSource`, `captureFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `withFrozenSource`.
- **`TestFrozenCaptureCallbackFailureAndPanicExpireLease`** — Regression contract: TestFrozenCaptureCallbackFailureAndPanicExpireLease.
  Signature: `func(t *testing.T)`.
  Calls: `captureFixture`, `errors.New`, `fmt.Sprint`, `panic`, `recover`, `source.Current`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `withFrozenSource`.
