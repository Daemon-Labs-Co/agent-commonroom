# cmd/server/tool_audit.go

[Module index](cmd.md) · [Open source](../../cmd/server/tool_audit.go)

Go implementation for tool audit; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerToolAuditor`**: controllerToolAuditor owns the durable side of the supervisor's callback link. Construction is trusted controller work, not a model/plugin/public HTTP API. A source fork CAN substitute a no-op auditor BUT then tools are unaudited.
  Declaration: `struct { job room.ManagedJob server *server scope toolbroker.Scope actorID string actor room.Actor participantRevision int connection room.Connection companyID string repository func() (string, error) }`.
- **`(*server).serveToolAudit`** — serveToolAudit is the controller assembly point for a registered supervisor stream, not a public handler. SSH/peer authentication and engine/control-stream separation are required before calling it. An absent identity never means owner.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, channel io.ReadWriteCloser) error`.
  Calls: `channel.Close`, `s.toolAuditor`, `toolbroker.ServeAudit`.
- **`(*server).toolAuditor`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope toolbroker.Scope) (toolbroker.Auditor, error)`.
  Calls: `controllerDevelopmentRepository`, `ctx.Value`, `newControllerToolAuditor`, `os.Getenv`.
- **`newControllerToolAuditor`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, s *server, scope toolbroker.Scope, actor string, repository func() (string, error)) (*controllerToolAuditor, error)`.
  Calls: `a.check`, `s.mu.Lock`, `s.mu.Unlock`, `time.Now`, `toolbroker.ValidAuditScope`.
- **`(*controllerToolAuditor).check`** — check revalidates current membership, connection revision, repository policy, job lifetime and stop state. The initiating human never becomes the owner just because the selected provider credentials belong to the organization owner. Caller holds server.mu; persistence additionally checks the event sequence.
  Signature: `func(ctx context.Context, bound bool) (room.State, room.Actor, room.Participant, room.Connection, error)`.
  Calls: `a.repository`, `a.scope.Deadline.After`, `a.server.agentsBlockedLocked`, `a.server.load`, `canAccess`, `contextDestination`, `ctx.Err`, `job.Deadline.Equal`, `member`, `pairedLifecycleVersion`, `remoteWorkerEndpoint`, `time.Now`.
- **`(*controllerToolAuditor).RecordIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent toolbroker.Intent) error`.
  Calls: `a.check`, `a.persist`, `a.server.mu.Lock`, `a.server.mu.Unlock`, `time.Now`, `time.Now().UTC`, `toolAuditKey`, `toolbroker.ValidAuditIntent`.
- **`(*controllerToolAuditor).RecordOutcome`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent toolbroker.Intent, outcome toolbroker.Outcome) error`.
  Calls: `a.check`, `a.persist`, `a.server.mu.Lock`, `a.server.mu.Unlock`, `time.Now`, `time.Now().UTC`, `toolAuditIntent`, `toolAuditKey`, `toolbroker.ValidAuditOutcome`.
- **`(*controllerToolAuditor).persist`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, kind string, call room.ManagedToolCall) error`.
  Calls: `a.server.append`, `a.server.tripIntegrityLocked`, `context.WithValue`, `errors.Is`.
- **`(*controllerToolAuditor).RecordIdleStop`** — RecordIdleStop accepts only terminal supervisor lifecycle metadata. Current presentation/stop flags cannot erase evidence about an already admitted job. Tool/frame ordering is enforced by ServeAudit; this grants no executor access.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, stop toolbroker.IdleStop) error`.
  Calls: `a.server.recordIdleStop`, `time.Now`, `toolbroker.ValidAuditIdleStop`.
