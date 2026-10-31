# internal/objectread/reader.go

[Module index](internal.md) · [Open source](../../../internal/objectread/reader.go)

Go implementation for reader; declarations and calls are listed below.

Package: `objectread`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Request`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Receipt objectcustody.Receipt 'json:"receipt"' Challenge string 'json:"challenge"' }`.
- **`ValidRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r Request) bool`.
  Calls: `objectcustody.ValidReceipt`, `workspacemount.ValidChallenge`.
- **Type `containerView`**: Behavior contract; required method signatures are shown in the declaration.
  Declaration: `interface { OpenRoot() (*os.Root, error) Identity() (os.FileInfo, error) Observe(context.Context, string) (workspacemount.Proof, error) }`.
- **Type `fixedView`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(fixedView).OpenRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (*os.Root, error)`.
  Calls: `os.OpenRoot`.
- **`(fixedView).Identity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (os.FileInfo, error)`.
  Calls: `os.Lstat`.
- **`(fixedView).Observe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, c string) (workspacemount.Proof, error)`.
  Calls: `workspacemount.ObserveObjects`.
- **Type `readLease`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ctx context.Context view containerView root *os.Root challenge string proof workspacemount.Proof journal *objectcustody.ReadJournalChannel }`.
- **`(*readLease).verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `ctx.Err`, `fixed.IsDir`, `fixed.Mode`, `l.ctx.Err`, `l.root.Lstat`, `l.view.Identity`, `l.view.Observe`, `os.SameFile`, `workspacemount.ValidProof`.
- **`(*readLease).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, s objectcustody.Scope, r objectcustody.Receipt) error`.
  Calls: `l.journal.AuthorizeRead`, `l.verify`.
- **`(*readLease).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `l.journal.AcknowledgeRead`, `l.verify`.
- **`(*readLease).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Intent) error`.
- **`(*readLease).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
- **Type `leasedReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { lease *readLease store *objectcustody.Store }`.
- **`(leasedReader).Get`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, s objectcustody.Scope, receipt objectcustody.Receipt) ([]byte, error)`.
  Calls: `r.lease.verify`, `r.store.Get`.
- **`Serve`** — Serve reads one exact sealed receipt through current supervisor authorization and durable source acknowledgment. The supervisor MUST bind the actual reviewed running reader/volume, fence writers and enforce real quota/account/export policy. Local kernel reports are not independent authentication. No paths or credentials are accepted, and no bytes are returned on mount, policy or audit uncertainty.
  Signature: `func(ctx context.Context, q Request, stream io.ReadWriteCloser) error`.
  Calls: `serve`.
- **`serve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q Request, stream io.ReadWriteCloser, view containerView) error`.
  Calls: `ValidRequest`, `cancel`, `context.WithTimeout`, `ctx.Err`, `journal.Close`, `lease.verify`, `missing`, `objectcustody.New`, `objectcustody.NewReadJournalChannel`, `objectcustody.ServeReadChannel`, `root.Close`, `stream.Close`, `view.OpenRoot`.
- **`missing`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value any) bool`.
  Calls: `reflect.ValueOf`, `v.IsNil`, `v.Kind`.
