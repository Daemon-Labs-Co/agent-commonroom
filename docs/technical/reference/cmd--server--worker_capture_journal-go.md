# cmd/server/worker_capture_journal.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_journal.go)

Go implementation for worker capture journal; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerCaptureIntent`**: workerCaptureIntent contains finite immutable metadata, never execution rights.
  Declaration: `toolbroker.CaptureIntent`.
- **Type `workerCaptureRecord`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Intent workerCaptureIntent 'json:"intent"' Events []workerProbeEvent 'json:"events"' }`.
- **Type `workerCaptureJournalGuard`**: workerCaptureJournalGuard MUST authenticate current operation-specific source or cleanup authority, actual creator/quota/image/volume evidence, and account-wide capacity. AcknowledgeCapture must durably mirror the exact local transition before returning; failed persistence must stop controller/worker authority. It must not reenter admission locks. No permissive/default implementation exists.
  Declaration: `interface { AuthorizeCapture(context.Context, workerJobRecord, workerCaptureIntent, workerProbeEvent) error AcknowledgeCapture(context.Context, workerCaptureRecord, workerProbeEvent) error StopCaptureAuthority() }`.
- **Type `workerCaptureJournal`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { dir string lifetime context.Context record workerJobRecord intent workerCaptureIntent guard workerCaptureJournalGuard persist func(string, []byte, func(string) error) error }`.
- **`captureRecordName`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i workerCaptureIntent) string`.
- **`validWorkerCaptureRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerCaptureRecord) bool`.
  Calls: `toolbroker.ValidCaptureMetadata`, `validWorkerHelperEvents`.
- **`readWorkerCaptureRecords`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string) (map[string]workerCaptureRecord, error)`.
  Calls: `bytes.Equal`, `captureRecordName`, `entry.Name`, `f.Close`, `f.ReadDir`, `filepath.Join`, `json.Marshal`, `json.Unmarshal`, `len`, `os.Open`, `readPrivateBounded`, `strings.HasPrefix`, `validWorkerCaptureRecord`.
- **`captureRecordSettled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerCaptureRecord) bool`.
  Calls: `len`.
- **`workerCaptureSlots`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string) (int, error)`.
  Calls: `captureRecordSettled`, `readWorkerCaptureRecords`.
- **`newWorkerCaptureJournal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(l *workerWorkspaceObservation, q workspacecapture.Request, image string, objects toolbroker.CaptureObjectVolume, g workerCaptureJournalGuard) (*workerCaptureJournal, error)`.
  Calls: `copyCaptureRequest`, `detachedProbeWorkerRecord`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `nilWorkerObservationDependency`, `retainedObservationRecord`, `toolbroker.NewCaptureContainerPlan`.
- **`(*workerCaptureJournal).fail`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `filepath.Join`, `j.guard.StopCaptureAuthority`, `replacePrivateDurably`.
- **`copyCaptureIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i workerCaptureIntent) workerCaptureIntent`.
  Calls: `copyCaptureRequest`.
- **`copyCaptureRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerCaptureRecord) workerCaptureRecord`.
  Calls: `copyCaptureIntent`, `slices.Clone`.
- **`captureIntentEqual`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a, b workerCaptureIntent) bool`.
  Calls: `bytes.Equal`, `json.Marshal`.
- **`(*workerCaptureJournal).matches`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(q workspacecapture.Request) bool`.
  Calls: `captureIntentEqual`.
- **`(*workerCaptureJournal).transition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, event workerProbeEvent) error`.
  Calls: `append`, `boundedWorkerRecords`, `cancel`, `captureIntentEqual`, `captureRecordName`, `captureRecordSettled`, `child.Err`, `context.AfterFunc`, `context.WithCancel`, `copyCaptureIntent`, `copyCaptureRecord`, `ctx.Err`, `detachedProbeWorkerRecord`, `filepath.Join`, `j.fail`, `j.guard.AcknowledgeCapture`, `j.guard.AuthorizeCapture`, `j.lifetime.Err`, `j.persist`, `json.Marshal`, `len`, `lock.Close`, `nilWorkerObservationDependency`, `objectReadRecordSettled`, `readRetainedWorkspaceObservation`, `readWorkerCaptureRecords`, `readWorkerObjectReadRecords`, `stop`, `validWorkerCaptureRecord`, `workerProbeCapacity`, `workerProbeStopped`, `workerWaitLease`.
- **`(*workerCaptureJournal).BeginCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q workspacecapture.Request, o toolbroker.CaptureObjectVolume, image string) error`.
  Calls: `j.matches`, `j.transition`.
- **`(*workerCaptureJournal).BindCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q workspacecapture.Request, id string) error`.
  Calls: `j.matches`, `j.transition`.
- **`(*workerCaptureJournal).FinishCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q workspacecapture.Request, id string, absent bool) error`.
  Calls: `j.matches`, `j.transition`.
- **`(*workerCaptureJournal).BeginCaptureRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q workspacecapture.Request, o toolbroker.CaptureObjectVolume, image, id string) error`.
  Calls: `j.matches`, `j.transition`.
- **`(*workerCaptureJournal).FinishCaptureRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q workspacecapture.Request, id string, absent bool) error`.
  Calls: `j.matches`, `j.transition`.
