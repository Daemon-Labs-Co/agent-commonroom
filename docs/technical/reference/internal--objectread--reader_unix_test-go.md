# internal/objectread/reader_unix_test.go

[Module index](internal.md) · [Open source](../../../internal/objectread/reader_unix_test.go)

Automated regression checks for reader_unix; read test names for individual contracts.

Package: `objectread`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `seedJournal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(seedJournal).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Intent) error`.
- **`(seedJournal).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
- **`(seedJournal).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Scope, objectcustody.Receipt) error`.
- **`(seedJournal).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
- **Type `syntheticView`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { mu sync.Mutex dir, identity string proof workspacemount.Proof calls int }`.
- **`(*syntheticView).OpenRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (*os.Root, error)`.
  Calls: `os.OpenRoot`.
- **`(*syntheticView).Identity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (os.FileInfo, error)`.
  Calls: `os.Lstat`, `v.mu.Lock`, `v.mu.Unlock`.
- **`(*syntheticView).Observe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ string) (workspacemount.Proof, error)`.
  Calls: `ctx.Err`, `v.mu.Lock`, `v.mu.Unlock`.
- **Type `releaseGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { auth, acks, stops atomic.Int32 deny bool onAuth, onAck func() }`.
- **`(*releaseGuard).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ objectcustody.Scope, _ objectcustody.Receipt) error`.
  Calls: `ctx.Err`, `g.auth.Add`, `g.onAuth`.
- **`(*releaseGuard).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ objectcustody.Receipt) error`.
  Calls: `ctx.Err`, `g.acks.Add`, `g.onAck`.
- **`(*releaseGuard).StopReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `g.stops.Add`.
- **`readerFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (Request, *syntheticView)`.
  Calls: `objectcustody.New`, `os.Chmod`, `os.OpenRoot`, `root.Close`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `t.TempDir`.
- **`TestObjectReaderFixedLeaseAndControllerRelay`** — Regression contract: TestObjectReaderFixedLeaseAndControllerRelay.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `g.acks.Load`, `g.auth.Load`, `g.stops.Load`, `make`, `net.Pipe`, `objectcustody.ReadChannel`, `readerFixture`, `serve`, `string`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `view.mu.Lock`, `view.mu.Unlock`.
- **`TestObjectReaderMetadataAndUnsupportedView`** — Regression contract: TestObjectReaderMetadataAndUnsupportedView.
  Signature: `func(t *testing.T)`.
  Calls: `ValidRequest`, `readerFixture`, `serve`, `t.Fatal`.
