# cmd/server/managed_jobs.go

[Module index](cmd.md) · [Open source](../../cmd/server/managed_jobs.go)

Go implementation for managed jobs; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `managedJobContextKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **Type `managedJobRegistration`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int Repository string JobID string IssuedAt time.Time Deadline time.Time Register func(string, string) error Update func(string) error }`.
- **`(*server).managedJobContext`** — managedJobContext records intent before SSH dispatch, independent of optional diagnostics. Audit failure prevents launch or latches execution off. It never stores prompts, provider output, command arguments or credentials.
  Signature: `func(ctx context.Context, p room.Participant) context.Context`.
  Calls: `s.managedExecutionJobContext`.
- **`(*server).managedPairJobContext`** — managedPairJobContext journals identity only. It does not enable launch or bypass the mandatory development barrier. Repository policy is rechecked at registration; absent human provenance cannot become owner authority.
  Signature: `func(ctx context.Context, p room.Participant, repository string) context.Context`.
  Calls: `controllerDevelopmentRepository`, `os.Getenv`, `s.managedExecutionJobContext`.
- **`(*server).managedWorkspaceJobContext`** — managedWorkspaceJobContext records the version-4 workspace cleanup contract. It creates no process and cannot enable the blocked development launch path.
  Signature: `func(ctx context.Context, p room.Participant, repository string) context.Context`.
  Calls: `controllerDevelopmentRepository`, `os.Getenv`, `s.managedExecutionJobContextVersion`.
- **`(*server).managedExecutionJobContext`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, p room.Participant, repository string, policy func() (string, error)) context.Context`.
  Calls: `s.managedExecutionJobContextVersion`.
- **`(*server).managedExecutionJobContextVersion`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, p room.Participant, repository string, policy func() (string, error), version int) context.Context`.
  Calls: `admitManagedContainerQuota`, `canAccess`, `cancel`, `connectionTestingEnabled`, `context.Background`, `context.WithTimeout`, `context.WithValue`, `ctx.Deadline`, `ctx.Err`, `ctx.Value`, `errors.New`, `id`, `job.AllocatedAt.Add`, `job.Deadline.IsZero`, `member`, `pairExecution`, `policy`, `remoteWorkerEndpoint`, `repositoryDigestPattern.MatchString`, `s.admitDailyTurns`, `s.agentsBlockedLocked`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.recordManagedJobLocked`, `s.tripIntegrityLocked`, `time.Now`, `time.Now().UTC`, `validManagedJob`, `workerAliasPattern.MatchString`.
- **`(*server).recordManagedJobLocked`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(jobID, status string) error`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `delete`, `errors.New`, `log.Printf`, `s.append`, `s.load`, `s.tripIntegrityLocked`, `stop`, `time.Now`, `time.Now().UTC`, `withInitiatingActor`.
- **`validManagedJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v room.ManagedJob) bool`.
  Calls: `membershipPartPattern.MatchString`, `pairExecution`, `repositoryDigestPattern.MatchString`, `v.AllocatedAt.IsZero`, `v.Deadline.After`, `v.Deadline.IsZero`, `v.Deadline.Sub`, `v.UpdatedAt.IsZero`, `validWorkerJobID`, `workerAliasPattern.MatchString`.
- **`jobsDebugEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) bool`.
- **`managedJobFence`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(job room.ManagedJob) workerFenceRequest`.
  Calls: `strings.Join`.
- **`hasUnconfirmedManagedJobs`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) bool`.
- **Type `managedWorkerTransport`**: managedWorkerTransport is the internal transport seam. Production SSH and synthetic tests vary here; durable lifecycle ordering stays in this module. No adapter may bypass isolated-worker admission or enable paired startup.
  Declaration: `struct { command func(context.Context, string, string, string) *exec.Cmd fence func(context.Context, string, workerFenceRequest) error monitor func(context.Context, <-chan struct{}, string, string, string) }`.
- **`dispatchManagedWorker`** — dispatchManagedWorker owns intent, submission and terminal shutdown evidence. Callers supply a validated worker binding and invoke the returned cleanup after execution, including command-start failure. Cleanup outlives turn cancellation but has its own bounded lifetime and is safe to call repeatedly.
  Signature: `func(ctx context.Context, alias string, request workerRunRequest, transport managedWorkerTransport) (*exec.Cmd, func(), error)`.
  Calls: `cancel`, `cleanup`, `close`, `context.Background`, `context.WithTimeout`, `ctx.Err`, `ctx.Value`, `emitBackground`, `encodeWorkerRequest`, `errors.New`, `fmt.Fprintln`, `make`, `once.Do`, `registration.Register`, `registration.Update`, `transport.command`, `transport.fence`, `transport.monitor`, `validWorkerJobID`.
- **`(*server).reconcileManagedJob`** — reconcileManagedJob keeps transport proof and its durable lifecycle outcome together. Authorization and approved-worker checks precede this seam; worker evidence remains independently verified by the transport adapter.
  Signature: `func(ctx context.Context, job room.ManagedJob) (shutdownErr, auditErr error)`.
  Calls: `fence`, `managedJobFence`, `s.mu.Lock`, `s.mu.Unlock`, `s.recordManagedJobLocked`.
