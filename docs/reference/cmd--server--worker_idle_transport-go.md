# cmd/server/worker_idle_transport.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_idle_transport.go)

Go implementation for worker idle transport; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`sessionIdleAnnouncer`** — sessionIdleAnnouncer uses only the same private authenticated audit stream as the admitted tool session. Frames cannot select sockets/peers or carry actions. A source fork can self-acknowledge instead, invalidating durable audit trust.
  Signature: `func(session *toolbroker.ToolSession, job room.ManagedJob) workerIdleAnnounce`.
  Calls: `record.IdleShutdownRequestedAt.UTC`, `record.IdleShutdownRequestedAt.UTC().Round`, `record.StartedAt.UTC`, `record.StartedAt.UTC().Round`, `sameWorkerBinding`, `session.RecordIdleStop`, `session.Scope`, `validManagedJob`, `validWorkerAuditScope`, `validWorkerRecord`.
