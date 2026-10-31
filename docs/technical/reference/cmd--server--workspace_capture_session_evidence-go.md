# cmd/server/workspace_capture_session_evidence.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_capture_session_evidence.go)

Go implementation for workspace capture session evidence; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureSessionWitness`**: captureSessionWitness must independently verify operation-specific current creator/worker/account admission and physically enforced quotas. Lifecycle witness checks additionally verify persisted transitions. Neither can trust stream claims, reacquire server/admission locks or call the waiting stream. No default exists.
  Declaration: `interface { captureLifecycleWitness ValidateCaptureOperation(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error }`.
- **Type `workspaceCaptureSessionEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *workspaceCaptureLedgerEvidence operation captureSessionWitness }`.
- **`newWorkspaceCaptureSessionEvidence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, worker string, intent workerCaptureIntent, witness captureSessionWitness) (*workspaceCaptureSessionEvidence, error)`.
  Calls: `newWorkspaceCaptureLedgerEvidence`, `nilWorkerObservationDependency`.
- **Type `captureLocalOperationState`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { exists bool record workerCaptureRecord }`.
- **`(*workspaceCaptureSessionEvidence).localOperation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(event workerProbeEvent) (captureLocalOperationState, error)`.
  Calls: `append`, `captureControllerEvents`, `captureIntentEqual`, `captureRecordName`, `captureRecordSettled`, `copyCaptureRecord`, `len`, `readWorkerCaptureRecords`, `validWorkerCaptureRecord`, `workerProbeStopped`.
- **`(*workspaceCaptureSessionEvidence).ValidateCaptureOperation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r workerJobRecord, intent workerCaptureIntent, event workerProbeEvent) error`.
  Calls: `cancel`, `captureAuthorityEvent`, `captureIntentEqual`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `copyCaptureIntent`, `ctx.Err`, `detachedProbeWorkerRecord`, `e.inspectTransition`, `e.lifetime.Err`, `e.localOperation`, `e.operation.ValidateCaptureOperation`, `nilWorkerObservationDependency`, `probeWorkerRecordDigest`, `readRetainedWorkspaceObservation`, `reflect.DeepEqual`, `retainedObservationRecord`, `sameWorkerBinding`, `stop`, `validManagedJob`.
