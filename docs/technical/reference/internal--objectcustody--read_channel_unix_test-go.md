# internal/objectcustody/read_channel_unix_test.go

[Module index](internal.md) · [Open source](../../../internal/objectcustody/read_channel_unix_test.go)

Automated regression checks for read_channel_unix; read test names for individual contracts.

Package: `objectcustody`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `channelReleaseGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { receipt Receipt deny, failAck bool authorize, ack, stopped atomic.Int32 }`.
- **`(*channelReleaseGuard).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, s Scope, r Receipt) error`.
  Calls: `ctx.Err`, `g.authorize.Add`.
- **`(*channelReleaseGuard).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r Receipt) error`.
  Calls: `ctx.Err`, `g.ack.Add`.
- **`(*channelReleaseGuard).StopReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `g.stopped.Add`.
- **`TestReadChannelActualSealedObject`** — Regression contract: TestReadChannelActualSealedObject.
  Signature: `func(t *testing.T)`.
  Calls: `ReadChannel`, `ServeReadChannel`, `bytes.Equal`, `bytes.NewReader`, `custodyFixture`, `g.ack.Load`, `g.authorize.Load`, `g.stopped.Load`, `j.mu.Lock`, `j.mu.Unlock`, `len`, `make`, `net.Pipe`, `rune`, `store.Put`, `string`, `strings.Join`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestReadChannelRefusalsReturnNoBytes`** — Regression contract: TestReadChannelRefusalsReturnNoBytes.
  Signature: `func(t *testing.T)`.
  Calls: `ReadChannel`, `ServeReadChannel`, `custodyFixture`, `g.ack.Load`, `g.stopped.Load`, `len`, `make`, `net.Pipe`, `store.Put`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestReadChannelRejectsHostileBodyAndHeaders`** — Regression contract: TestReadChannelRejectsHostileBodyAndHeaders.
  Signature: `func(t *testing.T)`.
  Calls: `ReadChannel`, `append`, `binary.BigEndian.PutUint32`, `bytes.NewReader`, `cancel`, `close`, `context.WithCancel`, `custodyFixture`, `g.ack.Load`, `g.stopped.Load`, `len`, `make`, `net.Pipe`, `readReadChannelHeader`, `right.Close`, `right.Read`, `right.Write`, `store.Put`, `strings.Repeat`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `uint32`, `writeReadChannelBytes`, `writeReadChannelHeader`.
- **Type `channelCountingReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls atomic.Int32 raw []byte }`.
- **`(*channelCountingReader).Get`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, Scope, Receipt) ([]byte, error)`.
  Calls: `r.calls.Add`.
- **`TestServeReadChannelRejectsSubstitutionBeforeStore`** — Regression contract: TestServeReadChannelRejectsSubstitutionBeforeStore.
  Signature: `func(t *testing.T)`.
  Calls: `ServeReadChannel`, `custodyFixture`, `left.Close`, `left.Read`, `make`, `net.Pipe`, `reader.calls.Load`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `writeReadChannelHeader`.
- **`TestReadChannelMissingAuthorityAndBounds`** — Regression contract: TestReadChannelMissingAuthorityAndBounds.
  Signature: `func(t *testing.T)`.
  Calls: `ReadChannel`, `ServeReadChannel`, `custodyFixture`, `left.Close`, `net.Pipe`, `right.Close`, `store.Put`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`.
