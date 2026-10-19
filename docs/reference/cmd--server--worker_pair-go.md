# cmd/server/worker_pair.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_pair.go)

Go implementation for worker pair; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`pairedLifecycleVersion`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(version int) bool`.
- **`pairExecution`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(version int) string`.
- **Type `workerPairRecord`**: workerPairRecord is additive, explicitly versioned lifecycle evidence, not an execution capability. Both roles share one job lease and the old logical-job admission slot. Actual engine/tool resource policy still needs launch verification. No provider, token, command or host path belongs here. A source fork CAN skip creation journalling, BUT invalidates crash recovery.
  Declaration: `struct { Version int 'json:"version"' Repository string 'json:"repository"' Engine workerPairContainer 'json:"engine"' Tool workerPairContainer 'json:"tool"' Workspace *workerPairWorkspace 'json:"workspace,omitempty"' }`.
- **Type `workerPairWorkspace`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { CreationStarted bool 'json:"creationStarted"' Identity toolbroker.WorkspaceVolumeIdentity 'json:"identity"' Removed bool 'json:"removed"' }`.
- **Type `workerPairContainer`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { CreationStarted bool 'json:"creationStarted"' ID string 'json:"id,omitempty"' Absent bool 'json:"absent"' }`.
- **`pairContainerName`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(job, role string) string`.
- **`validPairContainerID`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(id string) bool`.
  Calls: `len`, `repositoryDigestPattern.MatchString`.
- **`validWorkerPair`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v workerJobRecord) bool`.
  Calls: `pairExecution`, `pairRoleSettled`, `pairedLifecycleVersion`, `repositoryDigestPattern.MatchString`, `toolbroker.ValidWorkspaceVolume`, `validPairContainerID`.
- **`pairRoleSettled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c workerPairContainer) bool`.
- **Type `workerPairSupervisor`**: workerPairSupervisor holds existing job/membership leases. The future launcher owns create, inspect, journal and start ordering through its internal adapter. No production launch adapter or public entry point is wired. Its caller must authenticate the controller and validate image/mount policy.
  Declaration: `struct { ctx context.Context cancel context.CancelFunc startup sync.Mutex turn sync.Mutex reconcile sync.Mutex mu sync.Mutex dir string record workerJobRecord leases *workerJobLeases closed bool assembled bool auditListener *workerAuditListener toolSession *toolbroker.ToolSession codexTurnStarted bool engineCredential *workerEngineCredential codexRunStarted bool }`.
- **`admitWorkerPair`** — admitWorkerPair is an internal worker seam, not a new run-v2 variant or public endpoint. The legacy development barrier remains enforced at all entry points.
  Signature: `func(ctx context.Context, dir string, request workerRunRequest, containers []string) (*workerPairSupervisor, error)`.
  Calls: `admitWorkerRecord`, `context.WithDeadline`, `ctx.Err`, `errors.New`, `newWorkerJobRecord`, `now.Add`, `pairExecution`, `pairedLifecycleVersion`, `repositoryDigestPattern.MatchString`, `request.Deadline.After`, `request.Deadline.Sub`, `request.IssuedAt.After`, `request.IssuedAt.IsZero`, `time.Now`, `validWorkerJobID`.
- **`(*workerPairSupervisor).Close`** — Close cancels first, then retains ownership until the active model turn and in-flight adapter work have returned. Cancellation alone cannot prove a create/start operation has ended. Every concurrent Close waits too; none can report released ownership early.
  Signature: `func()`.
  Calls: `s.cancel`, `s.leases.Close`, `s.mu.Lock`, `s.mu.Unlock`, `s.startup.Lock`, `s.startup.Unlock`, `s.turn.Lock`, `s.turn.Unlock`.
- **`(*workerPairSupervisor).pairUpdate`** — pairUpdate is under the owning lease, and refuses any post-cancellation create or bind. Intent is durable before Docker create, immutable ID before start.
  Signature: `func(ctx context.Context, role, id string, begin bool) error`.
  Calls: `errors.New`, `s.available`, `s.cancel`, `s.leases.Close`, `s.mu.Lock`, `s.mu.Unlock`, `time.Now`, `time.Now().UTC`, `validPairContainerID`, `validWorkerPair`, `writeWorkerRecord`.
- **`(*workerPairSupervisor).BeginCreate`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, role string) error`.
  Calls: `s.pairUpdate`.
- **`(*workerPairSupervisor).BindCreated`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, role, id string) error`.
  Calls: `s.pairUpdate`.
- **Type `workerPairLaunchAdapter`**: workerPairLaunchAdapter is trusted worker policy, never model-supplied launch arguments. Create MUST create without starting, Verify MUST inspect reviewed image, mounts, credentials and aggregate resource policy, and Start MUST use the immutable ID and respect cancellation. Bootstrap must return a job-bound receipt before engine startup. No production adapter is wired yet. A source fork CAN violate this contract, BUT loses container isolation proof.
  Declaration: `interface { Create(context.Context, workerPairIdentity) (string, error) Verify(context.Context, workerPairIdentity, string) error Inspect(context.Context, string) ([]byte, error) Start(context.Context, string) error Bootstrap(context.Context, string) (workspacebootstrap.Receipt, error) }`.
- **Type `workerPairIdentity`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ Job, Role, Repository string }`.
- **Type `workerWorkspaceLaunchAdapter`**: Workspace allocation is a version-4-only contract. Version-3 readers cannot ignore workspace evidence and falsely confirm container-only shutdown.
  Declaration: `interface { workerPairLaunchAdapter WorkspaceIntent(context.Context) (toolbroker.WorkspaceVolumeIdentity, error) PrepareVolume(context.Context, toolbroker.WorkspaceVolumeIdentity) (toolbroker.WorkspaceVolumeIdentity, error) }`.
- **`(*workerPairSupervisor).Start`** — Start owns the lifecycle ordering behind one interface. Both identities are durable before either process starts. Errors retain the journal for explicit fenced recovery; startup never retries uncertain creation or starts. This internal seam neither authenticates callers nor enables development.
  Signature: `func(ctx context.Context, adapter workerPairLaunchAdapter) error`.
  Calls: `adapter.Bootstrap`, `adapter.Create`, `adapter.Inspect`, `adapter.Start`, `adapter.Verify`, `cancel`, `context.AfterFunc`, `context.WithDeadline`, `errors.New`, `json.Unmarshal`, `len`, `s.BeginCreate`, `s.BindCreated`, `s.available`, `s.cancel`, `s.leases.Close`, `s.mu.Lock`, `s.mu.Unlock`, `s.prepareWorkspace`, `s.startup.Lock`, `s.startup.Unlock`, `stop`, `time.Now`, `time.Now().UTC`, `toolbroker.VerifyPairResources`, `validPairContainerID`, `workspacebootstrap.ValidReceipt`, `writeWorkerRecord`.
- **`(*workerPairSupervisor).available`** — available is called with mu held so Close and state publication serialize.
  Signature: `func(ctx context.Context) error`.
  Calls: `ctx.Err`, `errors.New`, `filepath.Join`, `os.IsNotExist`, `os.Lstat`, `s.ctx.Err`, `s.record.Deadline.After`, `time.Now`.
- **`pairCapture`** — pairCapture never interprets shell input, inherits credentials or returns raw diagnostics. The production runner revalidates the approved rootless daemon.
  Signature: `func(ctx context.Context, runner toolbroker.DockerRunner, args []string) (string, error)`.
  Calls: `cancel`, `child.Err`, `context.WithCancel`, `ctx.Err`, `errors.New`, `out.data.String`, `runner.Run`, `strings.TrimSpace`.
- **Type `workerPairOutput`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { data strings.Builder cancel context.CancelFunc overflow bool }`.
- **`(*workerPairOutput).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `len`, `w.cancel`, `w.data.Len`, `w.data.Write`.
- **`pairInventory`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, runner toolbroker.DockerRunner, filter string) (string, error)`.
  Calls: `errors.New`, `pairCapture`, `validPairContainerID`.
- **`inspectPairIdentity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, runner toolbroker.DockerRunner, v workerJobRecord, role, id string) error`.
  Calls: `errors.New`, `pairCapture`, `pairContainerName`.
- **`stopWorkerPair`** — stopWorkerPair attempts both roles even if one fails. Docker errors/EOF are never absence proof. Missing IDs from a create/journal crash are discovered by fixed name, inspected for exact role/job/digest, then removed by immutable ID. Callers must also wait for the supervisor lease; this function alone cannot prove a still-running creator won't launch later. It grants no restart authority.
  Signature: `func(ctx context.Context, v workerJobRecord, runner toolbroker.DockerRunner) (workerJobRecord, error)`.
  Calls: `errors.New`, `inspectPairIdentity`, `pairCapture`, `pairContainerName`, `pairInventory`, `stopWorkerWorkspace`, `toolbroker.StopContainer`, `validWorkerJobID`, `validWorkerPair`.
- **Type `workerFenceAck`**: Pair acknowledgements bind execution kind and repository as well as scope. Version-2 replies cannot settle a version-3 cancellation request.
  Declaration: `struct { DiscardedWorkspace *toolbroker.WorkspaceVolumeIdentity 'json:"discardedWorkspace,omitempty"' Version int 'json:"version"' JobID string 'json:"jobId"' Scope string 'json:"scope"' Stopped bool 'json:"stopped"' CancellationRequested bool 'json:"cancellationRequested"' Execution string 'json:"execution,omitempty"' Repository string 'json:"repository,omitempty"' WorkspaceRemoved bool 'json:"workspaceRemoved,omitempty"' }`.
- **`workerFenceAcknowledgement`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(request workerFenceRequest, record workerJobRecord) workerFenceAck`.
  Calls: `matchesWorkspaceDisposal`.
- **`validWorkerFenceAck`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(request workerFenceRequest, expected workerJobRecord, ack workerFenceAck) bool`.
  Calls: `pairExecution`, `pairedLifecycleVersion`, `request.record`.
- **`validWorkerRecord`** — Versions 3/4 cannot be interpreted as a version-2 single job; version 4 additionally prevents old container-only readers ignoring workspace cleanup. Reuse common scalar validation only after checking the actual pair evidence.
  Signature: `func(v workerJobRecord) bool`.
  Calls: `pairedLifecycleVersion`, `validV2WorkerRecord`, `validWorkerPair`.
