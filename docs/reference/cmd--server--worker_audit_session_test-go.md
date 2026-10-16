# cmd/server/worker_audit_session_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_audit_session_test.go)

Automated regression checks for worker_audit_session; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWatchedWorkerAuditLinkDetectsIdleEOF`** — Regression contract: TestWatchedWorkerAuditLinkDetectsIdleEOF.
  Signature: `func(t *testing.T)`.
  Calls: `awaitAuditResult`, `cancel`, `closedAuditDone`, `context.WithCancel`, `ctx.Done`, `link.Close`, `link.Write`, `net.Pipe`, `newWatchedWorkerAuditLink`, `peer.Close`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`closedAuditDone`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(done <-chan struct{}) <-chan error`.
  Calls: `make`.
- **`TestWatchedWorkerAuditLinkPreservesFramedAcknowledgements`** — Regression contract: TestWatchedWorkerAuditLinkPreservesFramedAcknowledgements.
  Signature: `func(t *testing.T)`.
  Calls: `auditor.calls.Load`, `awaitAuditResult`, `cancel`, `client.RecordIntent`, `client.RecordOutcome`, `context.WithCancel`, `ctx.Done`, `failures.Add`, `failures.Load`, `link.Close`, `make`, `net.Pipe`, `newWatchedWorkerAuditLink`, `peer.Close`, `sampleToolScope`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`, `toolbroker.NewAuditClient`, `toolbroker.ServeAudit`.
- **`TestWatchedWorkerAuditLinkCancellationJoinsBlockedForwarding`** — Regression contract: TestWatchedWorkerAuditLinkCancellationJoinsBlockedForwarding.
  Signature: `func(t *testing.T)`.
  Calls: `awaitAuditResult`, `cancel`, `context.WithCancel`, `link.Close`, `make`, `net.Pipe`, `newWatchedWorkerAuditLink`, `peer.Close`, `peer.Write`, `t.Context`, `t.Parallel`.
- **`TestSupervisorAcceptToolSessionRevokesOnAuthenticatedLinkLoss`** — Regression contract: TestSupervisorAcceptToolSessionRevokesOnAuthenticatedLinkLoss.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `awaitAuditResult`, `context.Background`, `controller.Close`, `lease.Close`, `listener.Close`, `make`, `net.Pipe`, `relay.Close`, `s.AcceptToolSession`, `s.OpenAuditListener`, `s.ctx.Done`, `session.Close`, `session.Runtime.WorkspaceIntent`, `session.Scope`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`, `workerAttachAudit`, `workerTryLease`.
- **`TestSupervisorAcceptToolSessionRefusesForeignListenerWithoutConsumingInput`** — Regression contract: TestSupervisorAcceptToolSessionRefusesForeignListenerWithoutConsumingInput.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `context.WithValue`, `foreign.Close`, `newWorkerAuditListener`, `s.AcceptToolSession`, `s.ctx.Err`, `session.Close`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestSupervisorAcceptToolSessionCancellationInterruptsHandshake`** — Regression contract: TestSupervisorAcceptToolSessionCancellationInterruptsHandshake.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `awaitAuditResult`, `cancel`, `context.WithCancel`, `listener.Close`, `make`, `s.AcceptToolSession`, `s.OpenAuditListener`, `session.Close`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestSupervisorFailedToolSessionConstructionRevokesLifetime`** — Regression contract: TestSupervisorFailedToolSessionConstructionRevokesLifetime.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `awaitAuditResult`, `controller.Close`, `listener.Close`, `make`, `net.Pipe`, `relay.Close`, `s.AcceptToolSession`, `s.OpenAuditListener`, `s.ctx.Done`, `session.Close`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`, `workerAttachAudit`.
- **Type `auditSessionBlockedDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { entered chan struct{} once sync.Once }`.
- **`(*auditSessionBlockedDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ []string, _ []byte, _ io.Writer) error`.
  Calls: `close`, `ctx.Done`, `ctx.Err`, `d.once.Do`.
- **`TestAuditLinkLossInterruptsRealToolRuntimeBeforeEngineLaunch`** — Regression contract: TestAuditLinkLossInterruptsRealToolRuntimeBeforeEngineLaunch.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `awaitAuditResult`, `controller.Close`, `lease.Close`, `len`, `listener.Close`, `make`, `net.Pipe`, `readWorkerRecord`, `relay.Close`, `s.AcceptToolSession`, `s.AssemblePair`, `s.OpenAuditListener`, `s.Start`, `session.Close`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`, `workerAttachAudit`, `workerTryLease`.
