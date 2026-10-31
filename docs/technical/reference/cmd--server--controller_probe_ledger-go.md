# cmd/server/controller_probe_ledger.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_probe_ledger.go)

Go implementation for controller probe ledger; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerProbeEvidence`**: controllerProbeEvidence MUST authenticate the exact approved worker, creator ownership, reviewed image and physical volume/container observations for each event. Recovery requires fresh cleanup authority; absence must be independently verified before release. This cannot trust model/worker claims alone. No production transport/default exists. It must not re-enter server/admission locks. Source forks CAN falsify evidence and invalidate managed isolation guarantees.
  Declaration: `interface { ValidateProbe(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error }`.
- **Type `controllerProbeLedger`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server evidence controllerProbeEvidence }`.
- **Type `controllerProbeEvent`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' ActorID string 'json:"actorId"' CompanyID string 'json:"companyId"' RoomID string 'json:"roomId"' Worker string 'json:"worker"' Intent toolbroker.WorkspaceProbeIntent 'json:"intent"' Event workerProbeEvent 'json:"event"' }`.
- **`newControllerProbeLedger`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *server, e controllerProbeEvidence) (*controllerProbeLedger, error)`.
  Calls: `nilWorkerObservationDependency`.
- **`controllerProbeKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i toolbroker.WorkspaceProbeIntent) string`.
- **`releasedProbeEvents`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(events []workerProbeEvent) bool`.
- **`validControllerProbeRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r room.WorkspaceProbeReservation) bool`.
  Calls: `len`, `membershipPartPattern.MatchString`, `releasedProbeEvents`, `validWorkerProbeRecord`, `workerAliasPattern.MatchString`.
- **`controllerProbeSlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) (int, error)`.
  Calls: `controllerProbeKey`, `validControllerProbeRecord`, `validManagedJob`.
- **`admitControllerProbeQuota`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) error`.
  Calls: `controllerHelperSlots`, `managedContainerSlots`, `validManagedJob`.
- **`controllerProbeJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v controllerProbeEvent) (room.ManagedJob, bool)`.
  Calls: `validManagedJob`.
- **`validateControllerProbeEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v controllerProbeEvent) (room.WorkspaceProbeReservation, bool, error)`.
  Calls: `admitControllerProbeQuota`, `append`, `artifactRecordedActor`, `controllerProbeJob`, `controllerProbeKey`, `coordinationJobAllowed`, `delegatedJobAllowed`, `managedTaskCurrent`, `releasedProbeEvents`, `validControllerProbeRecord`, `validWorkerProbeIntent`.
- **`projectControllerProbe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `controllerProbeKey`, `json.Unmarshal`, `validateControllerProbeEvent`.
- **`(*controllerProbeLedger).current`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) (room.State, controllerProbeEvent, error)`.
  Calls: `artifactActor`, `coordinationJobAllowed`, `ctx.Value`, `delegatedJobAllowed`, `j.server.agentsBlockedLocked`, `j.server.load`, `managedTaskCurrent`.
- **`(*controllerProbeLedger).Acknowledge`** — Acknowledge is internal authenticated controller admission/audit, never an HTTP model event sink. The original principal MUST be supplied by trusted middleware or transport even for cleanup; absence never implies owner. Local worker journal and physical evidence remain mandatory. No host execution/fallback is introduced.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `cancel`, `context.WithValue`, `ctx.Err`, `errors.Is`, `j.current`, `j.evidence.ValidateProbe`, `nilWorkerObservationDependency`, `s.append`, `s.mu.Lock`, `s.mu.Unlock`, `s.storageContext`, `s.tripIntegrityLocked`, `validWorkerProbeIntent`, `validateControllerProbeEvent`.
