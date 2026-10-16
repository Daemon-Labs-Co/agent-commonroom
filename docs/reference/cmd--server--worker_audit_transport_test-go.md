# cmd/server/worker_audit_transport_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_audit_transport_test.go)

Automated regression checks for worker_audit_transport; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `transportTestAuditor`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ calls atomic.Int32 }`.
- **`(*transportTestAuditor).RecordIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, toolbroker.Intent) error`.
  Calls: `a.calls.Add`.
- **`(*transportTestAuditor).RecordOutcome`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, toolbroker.Intent, toolbroker.Outcome) error`.
  Calls: `a.calls.Add`.
- **`auditTransportFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (string, *workerPairSupervisor, toolbroker.Scope)`.
  Calls: `admitWorkerPair`, `os.MkdirTemp`, `os.RemoveAll`, `pairRequest`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`awaitAuditResult`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, ch <-chan error)`.
  Calls: `t.Fatal`, `t.Helper`, `time.After`.
- **`TestWorkerAuditTransportFramedRoundTripAndCancellation`** — Regression contract: TestWorkerAuditTransportFramedRoundTripAndCancellation.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `auditor.calls.Load`, `awaitAuditResult`, `cancel`, `client.RecordIntent`, `client.RecordOutcome`, `context.Background`, `context.WithTimeout`, `context.WithValue`, `controller.Close`, `link.Close`, `listener.Accept`, `listener.Close`, `make`, `net.Pipe`, `newWorkerAuditListener`, `relay.Close`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `toolbroker.NewAuditClient`, `toolbroker.ServeAudit`, `workerAttachAudit`.
- **`TestWorkerAuditSocketRefusesPublicDirectoryAndExistingPath`** — Regression contract: TestWorkerAuditSocketRefusesPublicDirectoryAndExistingPath.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `context.WithValue`, `filepath.Join`, `l.Close`, `newWorkerAuditListener`, `os.Chmod`, `os.ReadFile`, `os.Symlink`, `os.WriteFile`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerAuditSocket`.
- **`TestWorkerAuditScopeRequiresExactRecordAndCanonicalRequest`** — Regression contract: TestWorkerAuditScopeRequiresExactRecordAndCanonicalRequest.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `base64.RawURLEncoding.EncodeToString`, `changed.Deadline.Add`, `decodeWorkerAuditScope`, `json.Marshal`, `string`, `strings.Repeat`, `strings.Replace`, `t.Fatal`, `t.Parallel`, `validWorkerAuditScope`.
- **`TestWorkerAuditAttachRequiresHeldCreatorLease`** — Regression contract: TestWorkerAuditAttachRequiresHeldCreatorLease.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `input.Close`, `net.Pipe`, `output.Close`, `s.Close`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerAttachAudit`.
- **`TestWorkerAuditWrongHandshakeConsumesLink`** — Regression contract: TestWorkerAuditWrongHandshakeConsumesLink.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `cancel`, `conn.Close`, `context.WithTimeout`, `context.WithValue`, `ctx.Done`, `l.Accept`, `l.Close`, `make`, `net.Dial`, `newWorkerAuditListener`, `peer.Close`, `peer.Write`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerAuditSocket`.
- **`TestWorkerAuditRelayCancelsBlockedBothDirections`** — Regression contract: TestWorkerAuditRelayCancelsBlockedBothDirections.
  Signature: `func(t *testing.T)`.
  Calls: `awaitAuditResult`, `cancel`, `context.WithCancel`, `make`, `net.Pipe`, `other.Close`, `peer.Close`, `relayWorkerAudit`, `t.Context`, `t.Parallel`.
- **`TestWorkerAuditClosedSocketCannotBeRecreated`** — Regression contract: TestWorkerAuditClosedSocketCannotBeRecreated.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `l.Close`, `next.Close`, `s.OpenAuditListener`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerAuditListenerOriginalLifetimeCannotBeRevived`** — Regression contract: TestWorkerAuditListenerOriginalLifetimeCannotBeRevived.
  Signature: `func(t *testing.T)`.
  Calls: `auditScopeProof`, `auditTransportFixture`, `cancel`, `conn.Close`, `context.Background`, `context.WithCancel`, `context.WithValue`, `io.ReadFull`, `l.Accept`, `l.Close`, `make`, `net.Dial`, `newWorkerAuditListener`, `peer.Close`, `peer.Read`, `peer.SetReadDeadline`, `peer.Write`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `workerAuditSocket`.
- **`TestAuditCommandChild`** — Regression contract: TestAuditCommandChild.
  Signature: `func(t *testing.T)`.
  Calls: `io.Copy`, `os.Exit`, `os.Getenv`.
- **`TestCommandAuditLinkOwnsAndReapsDedicatedProcess`** — Regression contract: TestCommandAuditLinkOwnsAndReapsDedicatedProcess.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `awaitAuditResult`, `context.WithTimeout`, `exec.CommandContext`, `io.ReadFull`, `len`, `link.Close`, `link.Write`, `make`, `os.Environ`, `startCommandAuditLink`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerAuditListenerCloseInterruptsAcceptedRead`** — Regression contract: TestWorkerAuditListenerCloseInterruptsAcceptedRead.
  Signature: `func(t *testing.T)`.
  Calls: `auditScopeProof`, `auditTransportFixture`, `awaitAuditResult`, `conn.Close`, `conn.Read`, `context.WithValue`, `io.ReadFull`, `l.Accept`, `l.Close`, `make`, `net.Dial`, `newWorkerAuditListener`, `peer.Close`, `peer.Write`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerAuditSocket`.
- **`TestWorkerAuditListenerDeadlineBoundsUnattachedJob`** — Regression contract: TestWorkerAuditListenerDeadlineBoundsUnattachedJob.
  Signature: `func(t *testing.T)`.
  Calls: `auditTransportFixture`, `awaitAuditResult`, `conn.Close`, `context.Background`, `context.WithValue`, `l.Accept`, `l.Close`, `make`, `newWorkerAuditListener`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
