# cmd/server/workspace_probe_physical.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_probe_physical.go)

Go implementation for workspace probe physical; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workspaceProbePhysicalEvidence`**: workspaceProbePhysicalEvidence is a worker-local metadata observer scoped to the actual creator ownership callback. It reads Docker and private lifecycle metadata only. It is not a FrozenLease, remote attestation or model authority. Trusted setup must supply an authenticated worker/approved daemon and a reviewed immutable helper image. A source fork substituting those loses these claims.
  Declaration: `struct { lifetime context.Context dir, root, worker, image string record workerJobRecord runner toolbroker.DockerRunner }`.
- **`newWorkspaceProbePhysicalEvidence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, worker, image string) (*workspaceProbePhysicalEvidence, error)`.
  Calls: `detachedProbeWorkerRecord`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `nilWorkerObservationDependency`, `probeHex`, `retainedObservationRecord`, `strings.HasPrefix`, `strings.TrimPrefix`, `workerAliasPattern.MatchString`.
- **Type `inspectionOnlyProbeJournal`**: Inspection never executes a probe; its journal refuses every effect. This prevents a metadata validator from becoming a create/start/remove capability.
  Declaration: `struct{}`.
- **`(inspectionOnlyProbeJournal).Begin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, toolbroker.WorkspaceProbeIntent) error`.
- **`(inspectionOnlyProbeJournal).Bound`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, toolbroker.WorkspaceProbeIntent, string) error`.
- **`(inspectionOnlyProbeJournal).Finish`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, toolbroker.WorkspaceProbeIntent, string, bool) error`.
- **`(*workspaceProbePhysicalEvidence).inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, i toolbroker.WorkspaceProbeIntent, check func(context.Context, *toolbroker.WorkspaceMountProbe) error) error`.
  Calls: `cancel`, `check`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `ctx.Err`, `e.lifetime.Err`, `readRetainedWorkspaceObservation`, `retainedObservationRecord`, `sameWorkerBinding`, `stop`, `toolbroker.NewWorkspaceMountProbe`, `validManagedJob`, `validWorkerProbeIntent`.
- **`(*workspaceProbePhysicalEvidence).zeroReferences`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, probe *toolbroker.WorkspaceMountProbe) error`.
  Calls: `verifyRetainedWorkspaceReferences`.
- **`(*workspaceProbePhysicalEvidence).ValidateWorkspaceProbe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r workerJobRecord, i toolbroker.WorkspaceProbeIntent, operation string) error`.
  Calls: `controllerProbeKey`, `e.inspect`, `e.zeroReferences`, `pairInventory`, `probeWorkerRecordDigest`, `recordedProbeID`, `validControllerProbeRecord`, `verifyRetainedWorkspaceReferences`.
- **`(*workspaceProbePhysicalEvidence).ValidateProbe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, reservation room.WorkspaceProbeReservation, event workerProbeEvent) error`.
  Calls: `e.inspect`, `e.zeroReferences`, `len`, `pairInventory`, `probe.InspectRecovery`, `probeHex`, `probeRecordName`, `readWorkerProbeRecords`, `validControllerProbeRecord`, `verifyRetainedWorkspaceReferences`.
