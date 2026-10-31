# cmd/server/controller_object_read_ledger.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_object_read_ledger.go)

Go implementation for controller object read ledger; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerObjectReadEvidence`**: controllerObjectReadEvidence MUST independently authenticate the approved worker, exact private local journal/creator ownership, reviewed image, custody volume, physical quota and operation-specific export or cleanup authority. Absence is independently observed, never accepted from model/worker metadata alone. It must not reenter server/admission locks. No permissive default or deployment handshake exists. Source forks can falsify these checks and invalidate mandatory isolation.
  Declaration: `interface { ValidateObjectReadTransition(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error }`.
- **Type `controllerObjectReadLedger`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server evidence controllerObjectReadEvidence }`.
- **Type `controllerObjectReadEvent`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' ActorID string 'json:"actorId"' CompanyID string 'json:"companyId"' RoomID string 'json:"roomId"' Worker string 'json:"worker"' Intent workerObjectReadIntent 'json:"intent"' Event workerProbeEvent 'json:"event"' }`.
- **`newControllerObjectReadLedger`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *server, e controllerObjectReadEvidence) (*controllerObjectReadLedger, error)`.
  Calls: `nilWorkerObservationDependency`.
- **`controllerObjectReadKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i workerObjectReadIntent) string`.
- **`validControllerObjectReadRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r room.ObjectReadReservation) bool`.
  Calls: `len`, `membershipPartPattern.MatchString`, `releasedProbeEvents`, `validWorkerObjectReadRecord`, `workerAliasPattern.MatchString`.
- **`controllerObjectReadJob`** — Historical identity checks do not depend on current task state or editing flags. Cleanup must remain interpretable after task completion and access revocation.
  Signature: `func(st room.State, r room.ObjectReadReservation) bool`.
  Calls: `artifactRecordedSource`, `validManagedJob`.
- **`controllerObjectReadSlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) (int, error)`.
  Calls: `controllerObjectReadJob`, `controllerObjectReadKey`, `len`, `validControllerObjectReadRecord`.
- **`validateControllerObjectReadEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v controllerObjectReadEvent) (room.ObjectReadReservation, bool, error)`.
  Calls: `admitControllerProbeQuota`, `append`, `artifactRecordedActor`, `controllerObjectReadJob`, `controllerObjectReadKey`, `len`, `releasedProbeEvents`, `slices.Clone`, `toolbroker.ValidObjectReadMetadata`, `validControllerObjectReadRecord`, `validateArtifactTransition`.
- **`projectControllerObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `controllerObjectReadKey`, `json.Unmarshal`, `validateControllerObjectReadEvent`.
- **`(*controllerObjectReadLedger).current`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i workerObjectReadIntent, e workerProbeEvent) (room.State, controllerObjectReadEvent, error)`.
  Calls: `artifactActor`, `ctx.Value`, `j.server.agentsBlockedLocked`, `j.server.load`.
- **`(*controllerObjectReadLedger).Acknowledge`** — Acknowledge is private authenticated controller infrastructure, never a public event sink. Caller identity must come from trusted middleware/transport; missing identity never implies owner. This ledger cannot launch helpers or release bytes.
  Signature: `func(ctx context.Context, i workerObjectReadIntent, e workerProbeEvent) error`.
  Calls: `cancel`, `context.WithValue`, `ctx.Err`, `errors.Is`, `j.current`, `j.evidence.ValidateObjectReadTransition`, `nilWorkerObservationDependency`, `s.append`, `s.mu.Lock`, `s.mu.Unlock`, `s.storageContext`, `s.tripIntegrityLocked`, `slices.Clone`, `toolbroker.ValidObjectReadMetadata`, `validateControllerObjectReadEvent`.
