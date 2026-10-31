# cmd/server/artifact_custody.go

[Module index](cmd.md) · [Open source](../../../cmd/server/artifact_custody.go)

Go implementation for artifact custody; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `artifactCustodyGuard`**: artifactCustodyGuard is trusted worker infrastructure, never a model/plugin. ValidateSource must attest a frozen isolated source, exact repository/revision, output membership and current export policy. EstablishQuota must enforce the exact durable reservation on the actual private object volume before writes. ValidateSeal binds actual stored content/size to that frozen source and checks the quota/root identity is still enforced. It cannot trust the model receipt. AuthorizeRelease checks current authenticated export policy and source identity, including completed work. No production implementation/default is supplied. A source fork CAN lie through these callbacks and invalidate the boundary.
  Declaration: `interface { ValidateSource(context.Context, room.State, objectcustody.Intent) error EstablishQuota(context.Context, room.StorageReservation) error ValidateSeal(context.Context, room.State, objectcustody.Receipt) error AuthorizeRelease(context.Context, room.State, objectcustody.Receipt) error }`.
- **Type `controllerArtifactJournal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server guard artifactCustodyGuard }`.
- **`newControllerArtifactJournal`** — Internal only: no HTTP/model constructor or optional security bypass. A missing guard refuses; the existing managed file/bundle dispatch barrier is unchanged.
  Signature: `func(s *server, guard artifactCustodyGuard) (*controllerArtifactJournal, error)`.
  Calls: `nilArtifactGuard`.
- **Type `artifactCustodyEvent`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' ActorID string 'json:"actorId"' Intent objectcustody.Intent 'json:"intent"' Reservation room.StorageReservation 'json:"reservation"' Receipt *objectcustody.Receipt 'json:"receipt,omitempty"' }`.
- **`artifactActor`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, ctx context.Context, rid string) bool`.
  Calls: `canAccess`, `ctx.Value`, `member`.
- **`artifactCaptureJobStatus`** — A fenced retained workspace may be captured while the original task attempt and deadline are still current. Status is not physical proof: the mandatory source guard must independently verify isolation, writer absence and quota. Confirmed shutdown/disposal and expired attempts grant no fresh source rights.
  Signature: `func(job room.ManagedJob) bool`.
- **`artifactSourceJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, intent objectcustody.Intent) (room.ManagedJob, bool)`.
  Calls: `artifactCaptureJobStatus`, `artifactRecordedSource`, `coordinationJobAllowed`, `delegatedJobAllowed`, `managedTaskCurrent`, `objectcustody.ValidIntent`, `time.Now`, `time.Now().Before`, `validManagedJob`.
- **`artifactRecordedSource`** — Historical identity checks have no clock, physical callback or current connection requirement. Retained completed objects remain interpretable.
  Signature: `func(st room.State, intent objectcustody.Intent) bool`.
  Calls: `taskcontract.Digest`, `validManagedJob`.
- **`(*controllerArtifactJournal).current`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent objectcustody.Intent, source bool) (room.State, error)`.
  Calls: `artifactActor`, `artifactSourceJob`, `ctx.Err`, `ctx.Value`, `j.guard.ValidateSource`, `s.agentsBlockedLocked`, `s.load`.
- **`(*controllerArtifactJournal).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent objectcustody.Intent) error`.
  Calls: `artifactSourceJob`, `cancel`, `ctx.Err`, `ctx.Value`, `j.append`, `j.current`, `j.guard.EstablishQuota`, `nilArtifactGuard`, `objectcustody.ValidIntent`, `s.mu.Lock`, `s.mu.Unlock`, `s.storageContext`, `validateArtifactTransition`.
- **`(*controllerArtifactJournal).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, receipt objectcustody.Receipt) error`.
  Calls: `j.record`.
- **`(*controllerArtifactJournal).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope objectcustody.Scope, receipt objectcustody.Receipt) error`.
  Calls: `j.record`.
- **`(*controllerArtifactJournal).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, receipt objectcustody.Receipt) error`.
  Calls: `j.record`.
- **`(*controllerArtifactJournal).record`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, kind string, receipt objectcustody.Receipt) error`.
  Calls: `cancel`, `ctx.Err`, `ctx.Value`, `j.append`, `j.current`, `j.guard.AuthorizeRelease`, `j.guard.ValidateSeal`, `nilArtifactGuard`, `objectcustody.ValidReceipt`, `s.mu.Lock`, `s.mu.Unlock`, `s.storageContext`, `validateArtifactTransition`.
- **`validateArtifactTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, kind string, v artifactCustodyEvent) error`.
  Calls: `admitManagedStorage`, `artifactCaptureJobStatus`, `artifactRecordedActor`, `artifactRecordedSource`, `membershipPartPattern.MatchString`, `objectcustody.ValidIntent`, `objectcustody.ValidReceipt`, `taskExecutionPendingCurrent`.
- **`projectArtifactCustody`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, kind string, raw []byte) error`.
  Calls: `json.Marshal`, `json.Unmarshal`, `projectManagedStorage`, `validateArtifactTransition`.
- **`nilArtifactGuard`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(g artifactCustodyGuard) bool`.
  Calls: `reflect.ValueOf`, `v.IsNil`, `v.Kind`.
- **`(*controllerArtifactJournal).append`** — Caller holds the server lock. A genuine persistence failure blocks later authority; an ordinary stale-controller conflict does not create an incident.
  Signature: `func(ctx context.Context, revision int64, kind string, event artifactCustodyEvent) error`.
  Calls: `context.WithValue`, `errors.Is`, `j.server.append`, `j.server.tripIntegrityLocked`.
- **`artifactRecordedActor`** — The owner is intrinsic when its default actor has not yet been projected at this event position. Explicit historical disabling still takes precedence.
  Signature: `func(st room.State, actor, rid string) bool`.
  Calls: `canAccess`, `member`.
