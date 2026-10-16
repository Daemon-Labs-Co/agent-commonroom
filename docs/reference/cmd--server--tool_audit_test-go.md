# cmd/server/tool_audit_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/tool_audit_test.go)

Automated regression checks for tool_audit; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`scopedAuditFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, *controllerToolAuditor, toolbroker.Intent)`.
  Calls: `context.Background`, `db.seed`, `mockServer`, `newControllerToolAuditor`, `now.Add`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Round`.
- **`TestControllerToolAuditPersistsScopeRealActorAndExactlyOneOutcome`** — Regression contract: TestControllerToolAuditPersistsScopeRealActorAndExactlyOneOutcome.
  Signature: `func(t *testing.T)`.
  Calls: `auditor.RecordIntent`, `auditor.RecordOutcome`, `context.Background`, `db.seed`, `db.snapshot`, `json.Unmarshal`, `s.load`, `scopedAuditFixture`, `strings.Contains`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `toolAuditKey`.
- **`TestControllerToolAuditRevalidatesBeforeEveryAcknowledgement`** — Regression contract: TestControllerToolAuditRevalidatesBeforeEveryAcknowledgement.
  Signature: `func(t *testing.T)`.
  Calls: `auditor.RecordIntent`, `auditor.RecordOutcome`, `context.Background`, `db.seed`, `db.snapshot`, `len`, `member`, `s.load`, `scopedAuditFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **Type `scopedAuditExecutor`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, toolbroker.Scope, toolbroker.Action, io.Writer) error`.
- **`(scopedAuditExecutor).Execute`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, action toolbroker.Action, out io.Writer) error`.
  Calls: `executor`.
- **`TestScopedAuditTransportUsesRealStoreBeforeMockToolAndHidesPayloads`** — Regression contract: TestScopedAuditTransportUsesRealStoreBeforeMockToolAndHidesPayloads.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `context.WithValue`, `db.snapshot`, `dispatcher.Dispatch`, `errors.New`, `io.WriteString`, `left.Close`, `make`, `net.Pipe`, `right.Close`, `s.load`, `s.serveToolAudit`, `scopedAuditExecutor`, `scopedAuditFixture`, `string`, `strings.Contains`, `t.Fatal`, `t.Run`, `t.Setenv`, `time.Now`, `toolAuditKey`, `toolbroker.New`, `toolbroker.NewAuditClient`, `toolbroker.NewDispatcher`.
- **`TestManagedJobRecordsRealInitiatorAndNeverPromotesLegacyJobsForTools`** — Regression contract: TestManagedJobRecordsRealInitiatorAndNeverPromotesLegacyJobsForTools.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `context.WithValue`, `db2.seed`, `fresh.load`, `fresh.managedJobContext`, `fresh.managedJobContext(ctx, room.Participant{ID: "agent", RoomID: "engineering", ConnectionID: "connection"}).Value`, `legacy.load`, `legacyDB.seed`, `mockServer`, `newControllerToolAuditor`, `registration.Register`, `s.load`, `scopedAuditFixture`, `t.Fatal`, `t.Parallel`.
- **`TestControllerToolAuditFailuresLeaveNoSuccessAndNoOwnerFallback`** — Regression contract: TestControllerToolAuditFailuresLeaveNoSuccessAndNoOwnerFallback.
  Signature: `func(t *testing.T)`.
  Calls: `auditor.RecordIntent`, `auditor.RecordOutcome`, `context.Background`, `db.snapshot`, `newControllerToolAuditor`, `s.toolAuditor`, `scopedAuditFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestToolAuditProjectionRejectsOrphanAndMutatedOutcome`** — Regression contract: TestToolAuditProjectionRejectsOrphanAndMutatedOutcome.
  Signature: `func(t *testing.T)`.
  Calls: `auditor.RecordIntent`, `context.Background`, `db.seed`, `s.load`, `scopedAuditFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `toolAuditKey`.
