# internal/toolbroker/audit_idle_test.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/audit_idle_test.go)

Automated regression checks for audit_idle; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `idleAuditMock`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { mockAudit idle func(context.Context, Scope, IdleStop) error }`.
- **`(idleAuditMock).RecordIdleStop`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, s Scope, i IdleStop) error`.
  Calls: `m.idle`.
- **`auditIdleFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() IdleStop`.
  Calls: `now.Add`, `time.Now`, `time.Now().UTC`.
- **`TestPrivateIdleFrameIsDurableAndTerminal`** — Regression contract: TestPrivateIdleFrameIsDurableAndTerminal.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `ServeAudit`, `auditIdleFixture`, `client.RecordIdleStop`, `client.RecordIntent`, `committed.Load`, `committed.Store`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `strings.Repeat`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **`TestIdleFrameRefusesOldPeerAndDowngradedAck`** — Regression contract: TestIdleFrameRefusesOldPeerAndDowngradedAck.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `ServeAudit`, `auditFrameDigest`, `auditIdleFixture`, `auditRead`, `auditWrite`, `bufio.NewReader`, `client.RecordIdleStop`, `close`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestIdleFrameCannotInterruptPendingTool`** — Regression contract: TestIdleFrameCannotInterruptPendingTool.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `ServeAudit`, `auditIdleFixture`, `client.RecordIdleStop`, `client.RecordIntent`, `idleCalls.Add`, `idleCalls.Load`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestAuditBudgetReservesOnlyOneLifecycleFrame`** — Regression contract: TestAuditBudgetReservesOnlyOneLifecycleFrame.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `ServeAudit`, `auditIdleFixture`, `client.RecordIdleStop`, `client.RecordIntent`, `client.RecordOutcome`, `fmt.Sprintf`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestIdleWireRejectsAmbiguityBeforeController`** — Regression contract: TestIdleWireRejectsAmbiguityBeforeController.
  Signature: `func(t *testing.T)`.
  Calls: `ServeAudit`, `append`, `auditIdleFixture`, `authority.scope.Deadline.Add`, `calls.Add`, `calls.Load`, `json.Marshal`, `left.Close`, `left.Write`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `string`, `strings.Repeat`, `strings.Replace`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestIdleTerminalServerRejectsFurtherTools`** — Regression contract: TestIdleTerminalServerRejectsFurtherTools.
  Signature: `func(t *testing.T)`.
  Calls: `ServeAudit`, `auditIdleFixture`, `auditRead`, `auditWrite`, `bufio.NewReader`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `tools.Add`, `tools.Load`.
- **`TestSessionIdleCannotRevokeActiveToolSlot`** — Regression contract: TestSessionIdleCannotRevokeActiveToolSlot.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `auditIdleFixture`, `authority.Reserve`, `left.Close`, `liveFixture`, `net.Pipe`, `right.Close`, `session.RecordIdleStop`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`.
- **`TestReservedIdleFrameCannotIncreaseToolBudget`** — Regression contract: TestReservedIdleFrameCannotIncreaseToolBudget.
  Signature: `func(t *testing.T)`.
  Calls: `NewAuditClient`, `ServeAudit`, `auditWrite`, `calls.Add`, `calls.Load`, `client.RecordIntent`, `client.RecordOutcome`, `fmt.Sprintf`, `int32`, `left.Close`, `liveFixture`, `make`, `net.Pipe`, `right.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
