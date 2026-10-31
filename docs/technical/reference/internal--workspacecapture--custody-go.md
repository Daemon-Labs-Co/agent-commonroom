# internal/workspacecapture/custody.go

[Module index](internal.md) · [Open source](../../../internal/workspacecapture/custody.go)

Go implementation for custody; declarations and calls are listed below.

Package: `workspacecapture`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Request`**: Request is supervisor-owned finite output metadata, never a model-selected source root or proof of authorization. Capture computes its actual revision.
  Declaration: `struct { Binding workspacesnapshot.Binding 'json:"binding"' Members []workspacesnapshot.Member 'json:"members"' Challenge string 'json:"challenge"' }`.
- **Type `CapturedObjects`**: CapturedObjects is returned only after every declared selected member is sealed and the complete source is still current. It is not human acceptance, public release authority or an atomic storage transaction: failures retain partial data.
  Declaration: `struct { Snapshot workspacesnapshot.Snapshot Receipts []objectcustody.Receipt }`.
- **`CaptureToCustody`** — CaptureToCustody transfers a fixed isolated source into an independently private quota-backed object root. Journal MUST durably reserve and physically enforce exact object quota before writes; it must independently validate source/seal and current authority. Neither this adapter nor directory permissions supply quota. There is no host source path, release operation, automatic retry/refund/deletion or permissive journal. The caller owns the object root throughout the call.
  Signature: `func(ctx context.Context, request Request, authority Authority, objectRoot *os.Root, journal objectcustody.Journal) (CapturedObjects, error)`.
  Calls: `captureToCustody`.
- **`captureToCustody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, view containerView, request Request, authority Authority, objectRoot *os.Root, journal objectcustody.Journal) (CapturedObjects, error)`.
  Calls: `append`, `cancel`, `context.WithTimeout`, `ctx.Err`, `errors.Join`, `len`, `make`, `missingJournal`, `objectcustody.New`, `reader.Close`, `source.Current`, `source.Open`, `source.Snapshot`, `source.VerifyReceipt`, `store.Put`, `withFrozenSource`.
- **`missingJournal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(j objectcustody.Journal) bool`.
  Calls: `isNilDependency`.
- **Type `capturedSourceJournal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { source *workspacesnapshot.Source journal objectcustody.Journal }`.
- **`(*capturedSourceJournal).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent objectcustody.Intent) error`.
  Calls: `j.journal.Reserve`, `j.source.Current`, `j.source.Snapshot`, `objectcustody.ValidIntent`.
- **`(*capturedSourceJournal).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, receipt objectcustody.Receipt) error`.
  Calls: `j.journal.Seal`, `j.source.VerifyReceipt`.
- **`(*capturedSourceJournal).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Scope, objectcustody.Receipt) error`.
- **`(*capturedSourceJournal).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
