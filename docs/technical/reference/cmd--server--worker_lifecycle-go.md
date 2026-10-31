# cmd/server/worker_lifecycle.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_lifecycle.go)

Go implementation for worker lifecycle; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`workerContainerSlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(record workerJobRecord) int`.
- **Type `workerLocalJobContextKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **Type `workerLease`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ file *os.File }`.
- **`(*workerLease).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `int`, `l.file.Close`, `l.file.Fd`, `syscall.Flock`.
- **`workerTryLease`** — workerTryLease uses a persistent private inode; never delete lock files while running. Kernel ownership disappears on supervisor crash. A source fork CAN skip fencing, BUT allows delayed launches after cancellation and overlapping jobs.
  Signature: `func(dir, name string) (*workerLease, bool, error)`.
  Calls: `f.Close`, `f.Fd`, `filepath.Join`, `int`, `openPrivateRegular`, `syscall.Flock`.
- **`workerWaitLease`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir, name string) (*workerLease, error)`.
  Calls: `ctx.Done`, `ctx.Err`, `time.After`, `workerTryLease`.
- **`syncWorkerDirectory`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string) error`.
  Calls: `f.Close`, `f.Sync`, `os.Open`.
- **`workerMembership`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(key string) (string, string, string, string, error)`.
  Calls: `errors.New`, `hex.EncodeToString`, `len`, `membershipPartPattern.MatchString`, `sha256.Sum256`, `strings.SplitN`.
- **`newWorkerJobRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(request workerRunRequest) (workerJobRecord, error)`.
  Calls: `errors.New`, `now.Add`, `request.Deadline.After`, `request.Deadline.IsZero`, `time.Now`, `time.Now().UTC`, `workerMembership`.
- **`boundedWorkerRecords`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string) ([]workerJobRecord, error)`.
  Calls: `append`, `entry.Name`, `errors.New`, `f.Close`, `f.ReadDir`, `len`, `os.Open`, `readWorkerRecord`, `strings.HasSuffix`, `strings.TrimSuffix`, `validWorkerJobID`.
- **`workerJobSettled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v workerJobRecord) bool`.
  Calls: `pairedLifecycleVersion`, `validWorkerPair`, `workerTerminal`.
- **`workerAdmission`** — workerAdmission checks durable records and live containers under the shared admission lock. Scope locks serialize a membership across connections/providers.
  Signature: `func(records []workerJobRecord, containers []string, next workerJobRecord) error`.
  Calls: `errors.New`, `pairedLifecycleVersion`, `strings.HasPrefix`, `strings.TrimPrefix`, `validWorkerJobID`, `validWorkerPair`, `workerContainerSlots`, `workerJobSettled`.
- **Type `workerJobLeases`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ scope, job *workerLease }`.
- **`(*workerJobLeases).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `l.job.Close`, `l.scope.Close`.
- **`admitWorkerJob`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, request workerRunRequest, containers []string) (workerJobRecord, *workerJobLeases, error)`.
  Calls: `admitWorkerRecord`, `newWorkerJobRecord`.
- **`admitWorkerRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, record workerJobRecord, containers []string) (workerJobRecord, *workerJobLeases, error)`.
  Calls: `admission.Close`, `boundedWorkerRecords`, `errors.New`, `filepath.Join`, `leases.Close`, `os.IsNotExist`, `os.Lstat`, `scope.Close`, `validWorkerRecord`, `workerAdmission`, `workerContainerSlots`, `workerProbeCapacity`, `workerTryLease`, `workerWaitLease`, `writeWorkerRecord`.
- **`managedWorkerContainers`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) ([]string, error)`.
  Calls: `cmd.Run`, `errors.New`, `exec.CommandContext`, `len`, `output.String`, `strings.Fields`.
- **`stopWorkerContainer`** — stopWorkerContainer checks the exact name and managed label, removes only that container and confirms absence. Output/errors never contain provider data.
  Signature: `func(ctx context.Context, job string) error`.
  Calls: `cmd.Run`, `errors.New`, `exec.CommandContext`, `output.String`, `strings.TrimSpace`.
- **Type `workerFenceRequest`**: DiscardWorkspace is not human approval. Destructive callers must first obtain exact current human authorization, then bind the complete observed identity. The owner review endpoint requires a durable one-shot approval first. Ordinary reconciliation never sets these.
  Declaration: `struct { ExpectedWorkspace *toolbroker.WorkspaceVolumeIdentity 'json:"expectedWorkspace,omitempty"' Repository string 'json:"repository,omitempty"' DiscardWorkspace bool 'json:"discardWorkspace,omitempty"' Version int 'json:"version"' JobID string 'json:"jobId"' Provider string 'json:"provider"' Key string 'json:"key"' }`.
- **`(workerFenceRequest).record`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (workerJobRecord, error)`.
  Calls: `errors.New`, `newWorkerJobRecord`, `pairExecution`, `pairedLifecycleVersion`, `repositoryDigestPattern.MatchString`, `toolbroker.ValidWorkspaceVolume`, `validWorkerJobID`.
- **`fenceWorkerJob`** — fenceWorkerJob tombstones before launch and proves the supervisor lease is released before acknowledging shutdown. No automatic task replay is permitted.
  Signature: `func(ctx context.Context, dir string, request workerFenceRequest, stopContainer func(context.Context, string) error) (workerJobRecord, error)`.
  Calls: `fenceWorkerJobContainers`, `stopContainer`, `stopWorkerPair`.
- **`fenceWorkerJobContainers`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, request workerFenceRequest, stopContainers func(context.Context, workerJobRecord) (workerJobRecord, error)) (workerJobRecord, error)`.
  Calls: `context.WithValue`, `ctx.Done`, `ctx.Err`, `errors.New`, `file.Close`, `file.Sync`, `file.WriteString`, `filepath.Join`, `info.Mode`, `info.Mode().IsRegular`, `info.Mode().Perm`, `lease.Close`, `matchesWorkspaceDisposal`, `os.IsExist`, `os.IsNotExist`, `os.Lstat`, `os.OpenFile`, `readWorkerRecord`, `request.record`, `sameWorkerFenceBinding`, `sameWorkspaceEvidence`, `stopContainers`, `syncWorkerDirectory`, `time.After`, `time.Now`, `time.Now().UTC`, `validWorkerRecord`, `workerTerminal`, `workerTryLease`, `writeWorkerRecord`.
- **Type `workerJobView`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { workerJobRecord LeaseState string 'json:"leaseState"' }`.
- **Type `workerJobsSnapshot`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Jobs []workerJobView 'json:"jobs"' ManagedContainers int 'json:"managedContainers"' Unresolved int 'json:"unresolved"' Truncated bool 'json:"truncated"' }`.
- **`workerSnapshot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string, containers []string) (workerJobsSnapshot, error)`.
  Calls: `append`, `boundedWorkerRecords`, `lease.Close`, `len`, `pairedLifecycleVersion`, `records[i].UpdatedAt.After`, `sort.Slice`, `workerJobSettled`, `workerTryLease`.
- **`decodeFenceRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(encoded string) (workerFenceRequest, error)`.
  Calls: `base64.RawURLEncoding.DecodeString`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `errors.New`, `json.NewDecoder`, `len`, `new`, `request.record`, `string`, `strings.NewReader`, `validFenceWire`.
- **`workerRuntimeReady`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `cmd.Run`, `dockerSecurityReady`, `errors.New`, `exec.CommandContext`, `host.String`, `info.Bytes`, `os.Getenv`, `strings.HasPrefix`, `strings.TrimSpace`.
- **`validV2WorkerRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v workerJobRecord) bool`.
  Calls: `v.Deadline.After`, `v.IdleShutdownRequestedAt.Before`, `v.IdleShutdownRequestedAt.IsZero`, `v.StartedAt.IsZero`, `v.UpdatedAt.IsZero`, `validWorkerJobID`, `validWorkerPair`, `workerMembership`.
- **`sameWorkerFenceBinding`** — Pair cancellation is a separate versioned contract. Legacy requests may not reinterpret a pair as one container, including across mixed worker binaries.
  Signature: `func(record, expected workerJobRecord) bool`.
- **`matchesWorkspaceDisposal`** — Validate both before the cancellation tombstone and again under the creator lease. A stale identity may not cancel/delete a different retained workspace.
  Signature: `func(request workerFenceRequest, record workerJobRecord) bool`.
