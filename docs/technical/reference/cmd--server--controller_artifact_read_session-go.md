# cmd/server/controller_artifact_read_session.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_artifact_read_session.go)

Go implementation for controller artifact read session; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `artifactReleaseEvidence`**: artifactReleaseEvidence independently authenticates the approved isolated custody reader, actual sealed object/root identity, current export/retention policy and physical quota/account state. It must not trust peer metadata or reenter server locks/the waiting read stream. No permissive/default implementation exists.
  Declaration: `interface { AuthorizeRelease(context.Context, room.State, objectcustody.Receipt) error }`.
- **Type `controllerArtifactReadSession`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server lifetime context.Context cancel context.CancelFunc actor, worker string receipt objectcustody.Receipt evidence artifactReleaseEvidence journal *controllerArtifactJournal }`.
- **`newControllerArtifactReadSession`** — Trusted authenticated private setup supplies identity and exact receipt. This adapter creates no public endpoint or network authentication, never selects paths, and grants no reserve/seal/host execution authority. Canceled sessions cannot revive.
  Signature: `func(ctx context.Context, s *server, worker string, r objectcustody.Receipt, e artifactReleaseEvidence) (*controllerArtifactReadSession, error)`.
  Calls: `cancel`, `context.WithCancel`, `ctx.Err`, `ctx.Value`, `membershipPartPattern.MatchString`, `newControllerArtifactJournal`, `nilWorkerObservationDependency`, `objectcustody.ValidReceipt`, `workerAliasPattern.MatchString`.
- **`(*controllerArtifactReadSession).callContext`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (context.Context, context.CancelFunc, error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithCancel`, `context.WithValue`, `ctx.Err`, `g.lifetime.Err`, `stop`.
- **`(*controllerArtifactReadSession).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope objectcustody.Scope, r objectcustody.Receipt) error`.
  Calls: `cancel`, `g.callContext`, `g.journal.AuthorizeRead`.
- **`(*controllerArtifactReadSession).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `cancel`, `g.callContext`, `g.journal.AcknowledgeRead`.
- **`(*controllerArtifactReadSession).pinned`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r objectcustody.Receipt) bool`.
  Calls: `artifactRecordedSource`, `ctx.Err`, `ctx.Value`, `g.lifetime.Err`, `validManagedJob`.
- **`(*controllerArtifactReadSession).AuthorizeRelease`** — Called only inside the journal's locked policy snapshot. The post-evidence load checks current access and sealed/reservation linkage before either metadata approval or durable acknowledgment. A changed audit revision is a conflict, not a grant.
  Signature: `func(ctx context.Context, st room.State, r objectcustody.Receipt) error`.
  Calls: `ctx.Err`, `g.evidence.AuthorizeRelease`, `g.journal.current`, `g.lifetime.Err`, `g.pinned`, `nilWorkerObservationDependency`, `validateArtifactTransition`.
- **`(*controllerArtifactReadSession).ValidateSource`** — A read session cannot become a writer by being passed as a full custody guard.
  Signature: `func(context.Context, room.State, objectcustody.Intent) error`.
- **`(*controllerArtifactReadSession).EstablishQuota`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, room.StorageReservation) error`.
- **`(*controllerArtifactReadSession).ValidateSeal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, room.State, objectcustody.Receipt) error`.
- **`(*controllerArtifactReadSession).StopReadAuthority`** — Scoped revocation is deliberately nonblocking. Persistence failures still trip the existing global journal integrity latch. Transfer failures never refund data.
  Signature: `func()`.
  Calls: `g.cancel`.
