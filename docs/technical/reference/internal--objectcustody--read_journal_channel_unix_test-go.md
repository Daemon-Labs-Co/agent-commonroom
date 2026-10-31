# internal/objectcustody/read_journal_channel_unix_test.go

[Module index](internal.md) · [Open source](../../../internal/objectcustody/read_journal_channel_unix_test.go)

Automated regression checks for read_journal_channel_unix; read test names for individual contracts.

Package: `objectcustody`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestReadJournalChannelActualStoreRelay`** — Regression contract: TestReadJournalChannelActualStoreRelay.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `NewReadJournalChannel`, `ReadChannel`, `ServeReadChannel`, `cancel`, `context.WithCancel`, `custodyFixture`, `g.ack.Load`, `g.authorize.Load`, `g.stopped.Load`, `len`, `make`, `net.Pipe`, `os.OpenRoot`, `relay.Close`, `relay.Reserve`, `relay.Seal`, `root.Close`, `stopped.Add`, `stopped.Load`, `store.Put`, `string`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestReadChannelRejectsRelayPhaseSubstitution`** — Regression contract: TestReadChannelRejectsRelayPhaseSubstitution.
  Signature: `func(t *testing.T)`.
  Calls: `ReadChannel`, `close`, `custodyFixture`, `g.ack.Load`, `g.stopped.Load`, `int32`, `make`, `net.Pipe`, `readReadChannelHeader`, `right.Close`, `send`, `store.Put`, `strings.NewReader`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `writeReadChannelHeader`.
- **`TestReadJournalChannelRejectsWrongAckAndDeadLifetime`** — Regression contract: TestReadJournalChannelRejectsWrongAckAndDeadLifetime.
  Signature: `func(t *testing.T)`.
  Calls: `NewReadJournalChannel`, `cancel`, `close`, `context.WithCancel`, `custodyFixture`, `make`, `net.Pipe`, `readReadChannelHeader`, `relay.AcknowledgeRead`, `relay.AuthorizeRead`, `right.Close`, `stopped.Add`, `stopped.Load`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `syntheticScope`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `writeReadChannelHeader`.
