# cmd/server/worker_workspace_observation.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_workspace_observation.go)

Go implementation for worker workspace observation; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerWorkspaceObservationGuard`**: workerWorkspaceObservationGuard is trusted controller infrastructure. Each call must enforce current human/task/export authority and durably acknowledge bounded source observation before success. It is not a model callback or quota waiver. No production controller transport/guard is assembled here.
  Declaration: `interface { AuthorizeWorkspaceObservation(context.Context, workerJobRecord) error }`.
- **Type `workerWorkspaceObservation`**: workerWorkspaceObservation holds the actual creator lock during a trusted internal callback. It proves retained-volume admission, NOT a frozen read-only root. A future exporter must additionally attest an isolated read-only mount. Never use this alone as workspacesnapshot.FrozenLease. No public endpoint or host root exists. A fork bypassing the lock or mount contract invalidates claims.
  Declaration: `struct { mu sync.Mutex operations sync.WaitGroup active bool lifetime context.Context dir, dockerRoot string record workerJobRecord runner toolbroker.DockerRunner guard workerWorkspaceObservationGuard ownership *workerLease }`.
- **`(*workerWorkspaceObservation).beginOperation`** — Admission and teardown share mu: once active is cleared no Add can race the teardown Wait. The completion callback must run only after all cleanup joins.
  Signature: `func(ctx context.Context) (context.Context, func(), error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithCancel`, `ctx.Err`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `l.operations.Add`, `l.operations.Done`, `stop`.
- **`(*workerWorkspaceObservation).operationCurrent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) bool`.
  Calls: `ctx.Err`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`.
- **`(*workerWorkspaceObservation).Check`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `l.checkProbe`.
- **`(*workerWorkspaceObservation).checkProbe`** — Only the assembled probe may supply an exact stopped read-only helper here. Ordinary Check retains the stronger zero-reference contract.
  Signature: `func(ctx context.Context, probe *toolbroker.WorkspaceMountProbe, probeID string) error`.
  Calls: `l.checkWorkspace`.
- **`(*workerWorkspaceObservation).checkRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, probe *toolbroker.WorkspaceMountProbe, probeID string) error`.
  Calls: `l.checkWorkspace`.
- **`(*workerWorkspaceObservation).checkWorkspace`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, probe *toolbroker.WorkspaceMountProbe, probeID string, recovery bool) error`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithCancel`, `ctx.Err`, `errors.New`, `l.guard.AuthorizeWorkspaceObservation`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `readRetainedWorkspaceObservation`, `stop`, `verifyRetainedWorkspaceReferences`.
- **`readRetainedWorkspaceObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string, expected workerJobRecord) (workerJobRecord, error)`.
  Calls: `errors.New`, `filepath.Join`, `io.LimitReader`, `io.ReadAll`, `latch.Close`, `openPrivateRegular`, `readWorkerRecord`, `retainedObservationRecord`, `sameWorkerFenceBinding`, `string`.
- **`verifyRetainedWorkspaceReferences`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, current workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, probe *toolbroker.WorkspaceMountProbe, probeID string, recovery bool) error`.
  Calls: `ctx.Err`, `errors.New`, `pairContainerName`, `pairInventory`, `probe.InspectRecovery`, `probe.VerifyAdmission`, `strings.HasPrefix`, `toolbroker.VerifyRetainedWorkspace`.
- **`retainedObservationRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerJobRecord) bool`.
  Calls: `toolbroker.ValidWorkspaceVolume`, `validPairContainerID`, `validWorkerRecord`.
- **`withWorkerWorkspaceObservation`** — withWorkerWorkspaceObservation never stops a live creator or acquires a host mount. It admits only an already-reconciled retained workspace. The persistent cancellation latch prevents managed recreation; the exclusive creator lease excludes concurrent cleanup/disposal until the callback and final checks end. Callback failures/cancellation retain all objects and accounting; no cleanup success, export receipt or release authority is inferred from this return value.
  Signature: `func(ctx context.Context, dir string, expected workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, guard workerWorkspaceObservationGuard, use func(*workerWorkspaceObservation) error) error`.
  Calls: `errors.New`, `l.Check`, `use`, `withWorkerWorkspaceOwnership`.
- **`withWorkerWorkspaceOwnership`** — Internal ownership primitive: use MUST first validate observation or exact-ID recovery before acting. Final successful use always requires zero references.
  Signature: `func(ctx context.Context, dir string, expected workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, guard workerWorkspaceObservationGuard, use func(*workerWorkspaceObservation) error) error`.
  Calls: `context.WithCancel`, `ctx.Err`, `errors.New`, `expire`, `l.Check`, `l.mu.Lock`, `l.mu.Unlock`, `l.operations.Wait`, `nilWorkerObservationDependency`, `ownership.Close`, `retainedObservationRecord`, `use`, `workerTryLease`.
- **`nilWorkerObservationDependency`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v any) bool`.
  Calls: `r.IsNil`, `r.Kind`, `reflect.ValueOf`.
