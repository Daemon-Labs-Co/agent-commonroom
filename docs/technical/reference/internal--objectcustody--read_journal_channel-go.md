# internal/objectcustody/read_journal_channel.go

[Module index](internal.md) · [Open source](../../../internal/objectcustody/read_journal_channel.go)

Go implementation for read journal channel; declarations and calls are listed below.

Package: `objectcustody`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ReadJournalChannel`**: ReadJournalChannel is the fixed isolated reader's read-only journal relay. The stream is already authenticated and supervisor-owned, pinned to one receipt, never model pipes. ServeReadChannel must own request/body framing on this same stream; methods run synchronously only during its trusted store Get. There is no database credential, permissive local acknowledgment, write or host fallback.
  Declaration: `struct { mu sync.Mutex stream io.ReadWriteCloser lifetime context.Context receipt Receipt phase int failed bool stop func() once sync.Once }`.
- **`NewReadJournalChannel`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, stream io.ReadWriteCloser, receipt Receipt, stop func()) (*ReadJournalChannel, error)`.
  Calls: `ctx.Err`, `readChannelAbsent`, `validReceipt`.
- **`(*ReadJournalChannel).Close`** — Close revokes local read authority and closes the owned stream exactly once. The stop callback must be local/nonblocking and must not reenter this journal.
  Signature: `func() error`.
  Calls: `c.once.Do`, `c.stop`, `c.stream.Close`.
- **`(*ReadJournalChannel).exchange`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r Receipt, operation, response string, phase int) error`.
  Calls: `c.Close`, `c.lifetime.Err`, `c.mu.Lock`, `c.mu.Unlock`, `cancel`, `child.Err`, `context.AfterFunc`, `context.WithTimeout`, `ctx.Err`, `fail`, `readReadChannelHeader`, `stopIO`, `stopLifetime`, `writeReadChannelHeader`.
- **`(*ReadJournalChannel).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope Scope, r Receipt) error`.
  Calls: `c.Close`, `c.exchange`.
- **`(*ReadJournalChannel).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r Receipt) error`.
  Calls: `c.exchange`.
- **`(*ReadJournalChannel).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, Intent) error`.
- **`(*ReadJournalChannel).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, Receipt) error`.
