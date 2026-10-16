# internal/toolbroker/audit_stream_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/audit_stream_test.go)

Automated regression checks for audit_stream; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestAuditStreamCommitsBeforeToolAndResult`** — Regression contract: TestAuditStreamCommitsBeforeToolAndResult.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `ServeAudit`, `cancel`, `context.Background`, `context.WithCancel`, `dispatch.Dispatch`, `dispatcher`, `errors.New`, `io.WriteString`, `left.Close`, `liveFixture`, `make`, `mockExecutor`, `net.Pipe`, `right.Close`, `stage.CompareAndSwap`, `stage.Load`, `string`, `t.Cleanup`, `t.Fatal`, `t.Parallel`.
- **`TestAuditStreamPersistenceFailureNeverRunsTool`** — Regression contract: TestAuditStreamPersistenceFailureNeverRunsTool.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `ServeAudit`, `authority.Reserve`, `context.Background`, `dispatch.Dispatch`, `dispatcher`, `executed.Load`, `executed.Store`, `left.Close`, `liveFixture`, `make`, `mockExecutor`, `net.Pipe`, `right.Close`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`.
- **`TestAuditClientRejectsStaleMalformedAndPartialAcknowledgements`** — Regression contract: TestAuditClientRejectsStaleMalformedAndPartialAcknowledgements.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `append`, `auditFrameDigest`, `auditRead`, `bufio.NewReaderSize`, `client.RecordIntent`, `close`, `context.Background`, `json.Marshal`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `right.Write`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **Type `auditWriteStarted`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn started chan struct{} once sync.Once }`.
- **`(*auditWriteStarted).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) (int, error)`.
  Calls: `channel.Conn.Write`, `channel.once.Do`, `close`.
- **`TestAuditCancellationInterruptsBlockedReadAndWrite`** — Regression contract: TestAuditCancellationInterruptsBlockedReadAndWrite.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `auditRead`, `bufio.NewReaderSize`, `cancel`, `client.RecordIntent`, `close`, `context.Background`, `context.WithCancel`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
- **`TestAuditServerRejectsForgedScopeAndReorderedFramesBeforePersistence`** — Regression contract: TestAuditServerRejectsForgedScopeAndReorderedFramesBeforePersistence.
  Signature: `func(t *testing.T)`.
  Calls: `ServeAudit`, `append`, `context.Background`, `json.Marshal`, `left.Close`, `left.Write`, `liveFixture`, `make`, `net.Pipe`, `persisted.Load`, `persisted.Store`, `right.Close`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAuditAckCannotBeReplayedOnNewLinkForSameJobAndCall`** — Regression contract: TestAuditAckCannotBeReplayedOnNewLinkForSameJobAndCall.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `auditFrameDigest`, `auditRead`, `auditWrite`, `bufio.NewReaderSize`, `client.RecordIntent`, `close`, `context.Background`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestAuditWireRejectsAmbiguousEncodingAndMissingPrimitiveFields`** — Regression contract: TestAuditWireRejectsAmbiguousEncodingAndMissingPrimitiveFields.
  Signature: `func(t *testing.T)`.
  Calls: `auditDecode`, `t.Fatal`, `t.Parallel`.
