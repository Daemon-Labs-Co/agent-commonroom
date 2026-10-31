# cmd/server/worker_object_read_journal.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_object_read_journal.go)

Go implementation for worker object read journal; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerObjectReadIntent`**: workerObjectReadIntent contains finite immutable metadata, never execution rights. This local adapter requires a retained-workspace creator lease. It grants no host read path, public download or isolation toggle. Source forks can fake guards and invalidate the mandatory containment/physical-quota contract. Account-wide audit, actual sealed-receipt custody and export authority MUST be independently checked.
  Declaration: `toolbroker.ObjectReadIntent`.
- **Type `workerObjectReadRecord`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Intent workerObjectReadIntent 'json:"intent"' Events []workerProbeEvent 'json:"events"' }`.
- **Type `workerObjectReadJournalGuard`**: workerObjectReadJournalGuard MUST authenticate current operation-specific reads or cleanup authority, actual creator/quota/image/volume evidence, and account-wide capacity. AcknowledgeObjectRead must durably mirror the exact local transition before returning; failed persistence must stop controller/worker authority. An exact already-begun recovery may be acknowledged again after fresh current authorization. It resumes the same logical cleanup, never creates a new reader. It must not reenter creator/admission locks. Finishing uncertainty must be acknowledged even when read authority was canceled; it grants no export rights. No permissive/default implementation exists.
  Declaration: `interface { AuthorizeObjectRead(context.Context, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error AcknowledgeObjectRead(context.Context, workerObjectReadRecord, workerProbeEvent) error StopObjectReadAuthority() }`.
- **Type `workerObjectReadJournal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { owner *workerWorkspaceObservation dir string lifetime context.Context record workerJobRecord intent workerObjectReadIntent guard workerObjectReadJournalGuard persist func(string, []byte, func(string) error) error }`.
- **`objectReadRecordName`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i workerObjectReadIntent) string`.
- **`validWorkerObjectReadRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerObjectReadRecord) bool`.
  Calls: `toolbroker.ValidObjectReadMetadata`, `validWorkerHelperEvents`.
- **`readWorkerObjectReadRecords`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string) (map[string]workerObjectReadRecord, error)`.
  Calls: `bytes.Equal`, `entry.Name`, `f.Close`, `f.ReadDir`, `filepath.Join`, `json.Marshal`, `json.Unmarshal`, `len`, `objectReadRecordName`, `os.Open`, `readPrivateBounded`, `strings.HasPrefix`, `validWorkerObjectReadRecord`.
- **`objectReadRecordSettled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerObjectReadRecord) bool`.
  Calls: `len`.
- **`workerObjectReadSlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string) (int, error)`.
  Calls: `objectReadRecordSettled`, `readWorkerObjectReadRecords`.
- **`newWorkerObjectReadJournal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, q objectread.Request, image string, objects toolbroker.CaptureObjectVolume, g workerObjectReadJournalGuard) (*workerObjectReadJournal, error)`.
  Calls: `detachedProbeWorkerRecord`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `nilWorkerObservationDependency`, `retainedObservationRecord`, `toolbroker.NewObjectReadContainerPlan`.
- **`(*workerObjectReadJournal).fail`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `filepath.Join`, `j.guard.StopObjectReadAuthority`, `replacePrivateDurably`.
- **`copyObjectReadRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerObjectReadRecord) workerObjectReadRecord`.
  Calls: `slices.Clone`.
- **`objectReadIntentEqual`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a, b workerObjectReadIntent) bool`.
- **`(*workerObjectReadJournal).matches`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(q objectread.Request) bool`.
  Calls: `objectReadIntentEqual`.
- **`(*workerObjectReadJournal).transition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, event workerProbeEvent) error`.
  Calls: `append`, `boundedWorkerRecords`, `cancel`, `captureRecordSettled`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `copyObjectReadRecord`, `ctx.Err`, `detachedProbeWorkerRecord`, `filepath.Join`, `j.fail`, `j.guard.AcknowledgeObjectRead`, `j.guard.AuthorizeObjectRead`, `j.lifetime.Err`, `j.owner.mu.Lock`, `j.owner.mu.Unlock`, `j.persist`, `json.Marshal`, `len`, `lock.Close`, `nilWorkerObservationDependency`, `objectReadIntentEqual`, `objectReadRecordName`, `objectReadRecordSettled`, `readRetainedWorkspaceObservation`, `readWorkerCaptureRecords`, `readWorkerObjectReadRecords`, `stop`, `validWorkerObjectReadRecord`, `workerProbeCapacity`, `workerProbeStopped`, `workerWaitLease`.
- **`(*workerObjectReadJournal).BeginObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, o toolbroker.CaptureObjectVolume, image string) error`.
  Calls: `j.matches`, `j.transition`.
- **`(*workerObjectReadJournal).BindObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, id string) error`.
  Calls: `j.matches`, `j.transition`.
- **`(*workerObjectReadJournal).FinishObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, id string, absent bool) error`.
  Calls: `j.matches`, `j.transition`.
- **`(*workerObjectReadJournal).BeginObjectReadRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, o toolbroker.CaptureObjectVolume, image, id string) error`.
  Calls: `j.matches`, `j.transition`.
- **`(*workerObjectReadJournal).FinishObjectReadRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, id string, absent bool) error`.
  Calls: `j.matches`, `j.transition`.
