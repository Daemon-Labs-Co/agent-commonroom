# cmd/server/workspace_object_read_evidence.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_object_read_evidence.go)

Go implementation for workspace object read evidence; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadSessionWitness`**: objectReadSessionWitness MUST independently verify current operation-specific export/cleanup permission, creator ownership, approved worker/image/daemon, custody/writer exclusion, actual physical quota and account admission. Lifecycle checks also attest private persistence. Metadata is never authentication/proof. Callbacks must not reenter server/creator/admission locks or the waiting stream. There is no permissive default; a source fork can fake witnesses and void isolation.
  Declaration: `interface { ValidateObjectReadOperation(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error ValidateObjectReadLifecycle(context.Context, room.State, workerJobRecord, workerObjectReadRecord, workerProbeEvent) error }`.
- **Type `workspaceObjectReadEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { lifetime context.Context dir, worker string record workerJobRecord intent workerObjectReadIntent runner toolbroker.DockerRunner plan *toolbroker.ObjectReadContainerPlan witness objectReadSessionWitness }`.
- **`newWorkspaceObjectReadEvidence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, worker string, intent workerObjectReadIntent, witness objectReadSessionWitness) (*workspaceObjectReadEvidence, error)`.
  Calls: `detachedProbeWorkerRecord`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `nilWorkerObservationDependency`, `objectReadAuthorityBinding`, `toolbroker.NewObjectReadContainerPlan`, `workerAliasPattern.MatchString`.
- **`objectReadControllerEvents`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerObjectReadRecord) []workerProbeEvent`.
  Calls: `len`, `slices.Clone`.
- **`(*workspaceObjectReadEvidence).valid`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State) bool`.
  Calls: `ctx.Err`, `e.lifetime.Err`, `nilWorkerObservationDependency`, `sameWorkerBinding`, `validManagedJob`.
- **Type `objectReadLocalSnapshot`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Exists bool Record workerObjectReadRecord }`.
- **`(*workspaceObjectReadEvidence).localOperation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(event workerProbeEvent) (objectReadLocalSnapshot, error)`.
  Calls: `append`, `copyObjectReadRecord`, `len`, `objectReadControllerEvents`, `objectReadRecordName`, `objectReadRecordSettled`, `readWorkerObjectReadRecords`, `validWorkerObjectReadRecord`, `workerProbeStopped`.
- **`(*workspaceObjectReadEvidence).inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, event workerProbeEvent) error`.
  Calls: `e.plan.InspectRecorded`, `e.plan.VerifyAdmission`, `e.plan.VerifyNoReferences`.
- **`(*workspaceObjectReadEvidence).ValidateObjectReadOperation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r workerJobRecord, i workerObjectReadIntent, event workerProbeEvent) error`.
  Calls: `cancel`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `detachedProbeWorkerRecord`, `e.inspect`, `e.lifetime.Err`, `e.localOperation`, `e.valid`, `e.witness.ValidateObjectReadOperation`, `objectReadAuthorityBinding`, `objectReadAuthorityEvent`, `probeWorkerRecordDigest`, `readRetainedWorkspaceObservation`, `reflect.DeepEqual`, `stop`.
- **`(*workspaceObjectReadEvidence).localTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(candidate room.ObjectReadReservation, event workerProbeEvent) (workerObjectReadRecord, error)`.
  Calls: `copyObjectReadRecord`, `len`, `objectReadControllerEvents`, `objectReadRecordName`, `readWorkerObjectReadRecords`, `slices.Equal`.
- **`(*workspaceObjectReadEvidence).ValidateObjectReadTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, candidate room.ObjectReadReservation, event workerProbeEvent) error`.
  Calls: `cancel`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `copyObjectReadRecord`, `detachedProbeWorkerRecord`, `e.inspect`, `e.lifetime.Err`, `e.localTransition`, `e.valid`, `e.witness.ValidateObjectReadLifecycle`, `readRetainedWorkspaceObservation`, `reflect.DeepEqual`, `stop`, `validControllerObjectReadRecord`.
