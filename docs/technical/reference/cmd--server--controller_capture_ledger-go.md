# cmd/server/controller_capture_ledger.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_capture_ledger.go)

Go implementation for controller capture ledger; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerCaptureLedgerEvidence`**: controllerCaptureLedgerEvidence MUST authenticate the exact approved worker, creator ownership, reviewed image and physical volume/container observations for each event. Recovery requires fresh cleanup authority; absence must be independently verified before release. This cannot trust model/worker claims alone. No production transport/default exists. It must not re-enter server/admission locks. Source forks CAN falsify evidence and invalidate managed isolation guarantees.
  Declaration: `interface { ValidateCaptureTransition(context.Context, room.State, room.WorkspaceCaptureReservation, workerProbeEvent) error }`.
- **Type `controllerCaptureLedger`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server evidence controllerCaptureLedgerEvidence }`.
- **Type `controllerCaptureEvent`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' ActorID string 'json:"actorId"' CompanyID string 'json:"companyId"' RoomID string 'json:"roomId"' Worker string 'json:"worker"' Intent workerCaptureIntent 'json:"intent"' Event workerProbeEvent 'json:"event"' }`.
- **`newControllerCaptureLedger`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s *server, e controllerCaptureLedgerEvidence) (*controllerCaptureLedger, error)`.
  Calls: `nilWorkerObservationDependency`.
- **`controllerCaptureKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i workerCaptureIntent) string`.
- **`validControllerCaptureRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r room.WorkspaceCaptureReservation) bool`.
  Calls: `len`, `membershipPartPattern.MatchString`, `releasedProbeEvents`, `validWorkerCaptureRecord`, `workerAliasPattern.MatchString`.
- **`controllerCaptureSlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) (int, error)`.
  Calls: `captureDefinitionBound`, `controllerCaptureKey`, `validControllerCaptureRecord`, `validManagedJob`.
- **`admitControllerCaptureQuota`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) error`.
  Calls: `admitControllerProbeQuota`.
- **`controllerCaptureJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v controllerCaptureEvent) (room.ManagedJob, bool)`.
  Calls: `captureDefinitionBound`, `validManagedJob`.
- **`validateControllerCaptureEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v controllerCaptureEvent) (room.WorkspaceCaptureReservation, bool, error)`.
  Calls: `admitControllerCaptureQuota`, `append`, `artifactCaptureJobStatus`, `artifactRecordedActor`, `captureIntentEqual`, `controllerCaptureJob`, `controllerCaptureKey`, `coordinationJobAllowed`, `delegatedJobAllowed`, `managedTaskCurrent`, `releasedProbeEvents`, `toolbroker.ValidCaptureMetadata`, `validControllerCaptureRecord`.
- **`projectControllerCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `controllerCaptureKey`, `json.Unmarshal`, `validateControllerCaptureEvent`.
- **`(*controllerCaptureLedger).current`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i workerCaptureIntent, e workerProbeEvent) (room.State, controllerCaptureEvent, error)`.
  Calls: `artifactActor`, `artifactCaptureJobStatus`, `coordinationJobAllowed`, `ctx.Value`, `delegatedJobAllowed`, `j.server.agentsBlockedLocked`, `j.server.load`, `managedTaskCurrent`, `time.Now`, `time.Now().Before`.
- **`(*controllerCaptureLedger).Acknowledge`** — Acknowledge is internal authenticated controller admission/audit, never an HTTP model event sink. The original principal MUST be supplied by trusted middleware or transport even for cleanup; absence never implies owner. Local worker journal and physical evidence remain mandatory. No host execution/fallback is introduced.
  Signature: `func(ctx context.Context, i workerCaptureIntent, e workerProbeEvent) error`.
  Calls: `cancel`, `context.WithValue`, `copyCaptureIntent`, `copyControllerCaptureReservation`, `ctx.Err`, `errors.Is`, `j.current`, `j.evidence.ValidateCaptureTransition`, `nilWorkerObservationDependency`, `s.append`, `s.mu.Lock`, `s.mu.Unlock`, `s.storageContext`, `s.tripIntegrityLocked`, `toolbroker.ValidCaptureMetadata`, `validateControllerCaptureEvent`.
- **`controllerHelperSlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) (int, error)`.
  Calls: `controllerCaptureSlots`, `controllerObjectReadSlots`, `controllerProbeSlots`.
- **`copyControllerCaptureReservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r room.WorkspaceCaptureReservation) room.WorkspaceCaptureReservation`.
  Calls: `copyCaptureIntent`, `slices.Clone`.
