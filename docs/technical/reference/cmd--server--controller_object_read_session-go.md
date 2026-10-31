# cmd/server/controller_object_read_session.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_object_read_session.go)

Go implementation for controller object read session; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerObjectReadSessionEvidence`**: Evidence independently authenticates operation-specific export/cleanup authority, creator ownership, actual physical quotas and private persisted local history. It must not reenter controller/admission locks or the waiting authority stream. No permissive default or network authentication mechanism is supplied here.
  Declaration: `interface { controllerObjectReadEvidence ValidateObjectReadOperation(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error }`.
- **Type `controllerObjectReadSession`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server lifetime context.Context cancel context.CancelFunc actor, worker string record workerJobRecord intent workerObjectReadIntent evidence controllerObjectReadSessionEvidence ledger *controllerObjectReadLedger }`.
- **`newControllerObjectReadSession`** — Only authenticated private setup may supply principal, worker and creator. This pins existing authentication; decoding a record never authenticates a peer. Managed reads remain isolated, with no host/SSH fallback or security toggle.
  Signature: `func(ctx context.Context, s *server, worker string, r workerJobRecord, i workerObjectReadIntent, e controllerObjectReadSessionEvidence) (*controllerObjectReadSession, error)`.
  Calls: `cancel`, `context.WithCancel`, `ctx.Err`, `ctx.Value`, `detachedProbeWorkerRecord`, `membershipPartPattern.MatchString`, `newControllerObjectReadLedger`, `nilWorkerObservationDependency`, `objectReadAuthorityBinding`, `workerAliasPattern.MatchString`.
- **`objectReadUncertainFinish`** — Cancellation may still record finish uncertainty under current human authority. It never permits confirmed absence/release, a new operation or fresh recovery.
  Signature: `func(e workerProbeEvent) bool`.
- **`(*controllerObjectReadSession).callContext`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, e workerProbeEvent) (context.Context, context.CancelFunc, error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithCancel`, `context.WithValue`, `ctx.Err`, `g.lifetime.Err`, `objectReadUncertainFinish`, `stop`.
- **`(*controllerObjectReadSession).pinned`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, e workerProbeEvent) bool`.
  Calls: `ctx.Err`, `g.lifetime.Err`, `objectReadUncertainFinish`, `sameWorkerBinding`.
- **`(*controllerObjectReadSession).candidate`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, event workerProbeEvent) (room.State, room.ObjectReadReservation, error)`.
  Calls: `g.ledger.current`, `g.pinned`, `len`, `validateControllerObjectReadEvent`.
- **`(*controllerObjectReadSession).AuthorizeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord, i workerObjectReadIntent, event workerProbeEvent) error`.
  Calls: `cancel`, `captureAuthorityEvent`, `child.Err`, `detachedProbeWorkerRecord`, `g.callContext`, `g.candidate`, `g.evidence.ValidateObjectReadOperation`, `g.server.mu.Lock`, `g.server.mu.Unlock`, `g.server.storageContext`, `objectReadAuthorityBinding`, `probeWorkerRecordDigest`, `storageCancel`.
- **`objectReadAuthorityRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerObjectReadRecord, i workerObjectReadIntent, e workerProbeEvent) bool`.
  Calls: `captureAuthorityEvent`, `len`, `validWorkerObjectReadRecord`.
- **`(*controllerObjectReadSession).AcknowledgeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, local workerObjectReadRecord, event workerProbeEvent) error`.
  Calls: `cancel`, `copyObjectReadRecord`, `g.callContext`, `ledger.Acknowledge`, `newControllerObjectReadLedger`, `objectReadAuthorityRecord`.
- **Type `objectReadSessionAcknowledgment`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { session *controllerObjectReadSession local workerObjectReadRecord }`.
- **`(*objectReadSessionAcknowledgment).ValidateObjectReadTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.ObjectReadReservation, e workerProbeEvent) error`.
  Calls: `a.session.ValidateObjectReadTransition`, `slices.Equal`.
- **`(*controllerObjectReadSession).ValidateObjectReadTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.ObjectReadReservation, event workerProbeEvent) error`.
  Calls: `ctx.Err`, `g.evidence.ValidateObjectReadTransition`, `g.lifetime.Err`, `g.pinned`, `objectReadUncertainFinish`, `slices.Clone`.
- **`(*controllerObjectReadSession).StopObjectReadAuthority`** — Scoped revocation is nonblocking: it cannot be mistaken for cleanup evidence. Actual controller persistence faults independently trip the global integrity latch.
  Signature: `func()`.
  Calls: `g.cancel`.
- **`(*controllerObjectReadSession).ObjectReadAuthorityLifetime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() context.Context`.
