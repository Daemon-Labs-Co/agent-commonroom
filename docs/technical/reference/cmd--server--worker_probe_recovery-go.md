# cmd/server/worker_probe_recovery.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_probe_recovery.go)

Go implementation for worker probe recovery; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerProbeRecoveryGuard`**: Recovery requires separate current authenticated cleanup authority and durable recovery audit, never inference from an old export/probe approval. Controller transport/implementation remains mandatory and is not supplied by models.
  Declaration: `interface { workerWorkspaceObservationGuard AuthorizeRecovery(context.Context, toolbroker.WorkspaceProbeIntent, string) error Acknowledge(context.Context, toolbroker.WorkspaceProbeIntent, workerProbeEvent) error StopAuthority() }`.
- **Type `guardedProbeRecovery`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { absentRequired bool active bool intent toolbroker.WorkspaceProbeIntent id string lease *workerWorkspaceObservation probe *toolbroker.WorkspaceMountProbe guard workerProbeRecoveryGuard }`.
- **`(*guardedProbeRecovery).Authorize`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error`.
  Calls: `g.guard.AuthorizeRecovery`, `g.lease.Check`, `g.lease.checkRecovery`.
- **`(*guardedProbeRecovery).Acknowledge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `g.guard.Acknowledge`.
- **`(*guardedProbeRecovery).StopAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `g.guard.StopAuthority`.
- **`recordedProbeID`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerProbeRecord) string`.
- **`(*workerProbeJournal).recoveryTransition`** — recoveryTransition preserves monotonic history. Explicit retries may re-ack an existing recovery intent/outcome; they never delete evidence or repeat launch.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string, complete bool) error`.
  Calls: `append`, `ctx.Err`, `filepath.Join`, `j.fail`, `j.guard.Acknowledge`, `j.guard.Authorize`, `j.persist`, `json.Marshal`, `len`, `lock.Close`, `nilWorkerObservationDependency`, `persist`, `probeHex`, `probeRecordMatchesWorker`, `probeRecordName`, `readWorkerProbeRecords`, `recordedProbeID`, `validWorkerProbeIntent`, `validWorkerProbeRecord`, `workerProbeSettled`, `workerProbeStopped`, `workerWaitLease`.
- **`recoverRetainedWorkspaceProbe`** — recoverRetainedWorkspaceProbe reconciles only a durable exact helper identity while holding creator ownership. It never adopts unknown create replies, starts a helper, clears integrity stops, disposes volumes or returns source bytes/proof/FrozenLease. Source forks can falsify trusted guards/runners and invalidate these guarantees. No host/uncontained execution fallback exists.
  Signature: `func(ctx context.Context, dir string, record workerJobRecord, intent toolbroker.WorkspaceProbeIntent, dockerRoot string, runner toolbroker.DockerRunner, guard workerProbeRecoveryGuard) error`.
  Calls: `cancel`, `context.WithTimeout`, `ctx.Err`, `journal.recoveryTransition`, `lease.checkRecovery`, `newWorkerProbeJournal`, `nilWorkerObservationDependency`, `probe.RemoveRecorded`, `probeHex`, `probeRecordName`, `readWorkerProbeRecords`, `recordedProbeID`, `toolbroker.NewWorkspaceMountProbe`, `validWorkerProbeIntent`, `withWorkerWorkspaceOwnership`, `workerProbeSettled`.
