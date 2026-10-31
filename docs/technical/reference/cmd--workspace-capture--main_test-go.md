# cmd/workspace-capture/main_test.go

[Module index](cmd.md) · [Open source](../../../cmd/workspace-capture/main_test.go)

Automated regression checks for main; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`helperRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() workspacecapture.Request`.
  Calls: `strings.Repeat`.
- **`encodedRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workspacecapture.Request) string`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `json.Marshal`.
- **Type `helperStream`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { bytes.Buffer closed bool }`.
- **`(*helperStream).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
- **Type `helperAuthority`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(helperAuthority).AuthorizeFrozenSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ workspacesnapshot.Binding, _ workspacemount.Proof) error`.
  Calls: `ctx.Err`.
- **Type `helperJournal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { reserved []objectcustody.Intent sealed []objectcustody.Receipt }`.
- **`(*helperJournal).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i objectcustody.Intent) error`.
  Calls: `append`, `ctx.Err`.
- **`(*helperJournal).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `append`, `ctx.Err`.
- **`(*helperJournal).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Scope, objectcustody.Receipt) error`.
  Calls: `errors.New`.
- **`(*helperJournal).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
  Calls: `errors.New`.
- **`TestHelperRejectsArgumentsBeforeOpeningRoots`** — Regression contract: TestHelperRejectsArgumentsBeforeOpeningRoots.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `base64.RawURLEncoding.EncodeToString`, `encodedRequest`, `helperRequest`, `json.Marshal`, `len`, `run`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestHelperSupervisedCustodyAndFinish`** — Regression contract: TestHelperSupervisedCustodyAndFinish.
  Signature: `func(t *testing.T)`.
  Calls: `authority.AuthorizeFrozenSource`, `encodedRequest`, `filepath.Join`, `helperRequest`, `len`, `make`, `net.Pipe`, `objectcustody.New`, `os.Chmod`, `os.OpenRoot`, `os.ReadFile`, `run`, `store.Put`, `string`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `workspacecapture.ServeCaptureChannel`.
- **`TestHelperRootFailureNeverCaptures`** — Regression contract: TestHelperRootFailureNeverCaptures.
  Signature: `func(t *testing.T)`.
  Calls: `encodedRequest`, `helperRequest`, `run`, `t.Context`, `t.Fatal`, `t.Parallel`.
