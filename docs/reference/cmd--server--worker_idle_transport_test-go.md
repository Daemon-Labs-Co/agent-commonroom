# cmd/server/worker_idle_transport_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_idle_transport_test.go)

Automated regression checks for worker_idle_transport; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `idleTransportAuditor`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *controllerToolAuditor drop bool peer net.Conn }`.
- **`(idleTransportAuditor).RecordIdleStop`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, stop toolbroker.IdleStop) error`.
  Calls: `a.controllerToolAuditor.RecordIdleStop`, `a.peer.Close`.
- **`TestIdleAcknowledgmentAcrossPrivateWorkerTransport`** — Regression contract: TestIdleAcknowledgmentAcrossPrivateWorkerTransport.
  Signature: `func(t *testing.T)`.
  Calls: `awaitAuditResult`, `cancel`, `context.WithCancel`, `controller.Close`, `f.db.seed`, `f.s.load`, `f.worker.AcceptToolSession`, `f.worker.OpenAuditListener`, `f.worker.expireIdle`, `f.worker.mu.Lock`, `f.worker.mu.Unlock`, `len`, `listener.Close`, `make`, `net.Pipe`, `newControllerToolAuditor`, `newIdleNoticeFixture`, `relay.Close`, `session.Close`, `sessionIdleAnnouncer`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `toolbroker.ServeAudit`, `workerAttachAudit`.
