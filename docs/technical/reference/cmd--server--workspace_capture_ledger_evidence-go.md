# cmd/server/workspace_capture_ledger_evidence.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_capture_ledger_evidence.go)

Go implementation for workspace capture ledger evidence; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureLifecycleWitness`**: captureLifecycleWitness MUST independently authenticate operation-specific current cleanup/source authority, creator ownership, approved worker/image/volumes, account capacity/storage accounting and real physical quotas. It must not reenter server/ admission locks or the same control stream. No permissive/default witness exists.
  Declaration: `interface { ValidateCaptureLifecycle(context.Context, room.State, workerJobRecord, workerCaptureRecord, workerProbeEvent) error }`.
- **Type `workspaceCaptureLedgerEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { lifetime context.Context dir, worker string record workerJobRecord intent workerCaptureIntent runner toolbroker.DockerRunner plan *toolbroker.CaptureContainerPlan witness captureLifecycleWitness }`.
- **`newWorkspaceCaptureLedgerEvidence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, worker string, intent workerCaptureIntent, witness captureLifecycleWitness) (*workspaceCaptureLedgerEvidence, error)`.
  Calls: `copyCaptureIntent`, `detachedProbeWorkerRecord`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `nilWorkerObservationDependency`, `retainedObservationRecord`, `toolbroker.NewCaptureContainerPlan`, `workerAliasPattern.MatchString`.
- **`captureControllerEvents`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(local workerCaptureRecord) []workerProbeEvent`.
  Calls: `len`, `slices.Clone`.
- **`(*workspaceCaptureLedgerEvidence).local`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(candidate room.WorkspaceCaptureReservation, event workerProbeEvent) (workerCaptureRecord, error)`.
  Calls: `captureControllerEvents`, `captureIntentEqual`, `captureRecordName`, `len`, `readWorkerCaptureRecords`, `reflect.DeepEqual`.
- **`(*workspaceCaptureLedgerEvidence).ValidateCaptureTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, candidate room.WorkspaceCaptureReservation, event workerProbeEvent) error`.
  Calls: `cancel`, `captureIntentEqual`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `copyCaptureRecord`, `ctx.Err`, `detachedProbeWorkerRecord`, `e.inspectTransition`, `e.lifetime.Err`, `e.local`, `e.witness.ValidateCaptureLifecycle`, `nilWorkerObservationDependency`, `readRetainedWorkspaceObservation`, `reflect.DeepEqual`, `sameWorkerBinding`, `stop`, `validControllerCaptureRecord`, `validManagedJob`.
- **`(*workspaceCaptureLedgerEvidence).inspectTransition`** — inspectTransition observes exact known identities only; it never performs lifecycle effects.
  Signature: `func(ctx context.Context, event workerProbeEvent) error`.
  Calls: `e.plan.InspectRecorded`, `e.plan.VerifyAdmission`, `e.plan.VerifyNoReferences`.
