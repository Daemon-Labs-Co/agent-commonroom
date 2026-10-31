# cmd/server/worker_probe_journal.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_probe_journal.go)

Go implementation for worker probe journal; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerProbeGuard`**: workerProbeGuard MUST enforce current human/task/export authority, actual creator ownership, approved daemon/image and full reference inventory. Acknowledge must durably audit each transition and reserve account-wide probe capacity on begin before returning. StopAuthority must revoke controller/worker execution on ambiguous persistence. No production transport/default exists. Source forks CAN substitute dishonest guards, BUT invalidate these guarantees.
  Declaration: `interface { Authorize(context.Context, toolbroker.WorkspaceProbeIntent) error Acknowledge(context.Context, toolbroker.WorkspaceProbeIntent, workerProbeEvent) error StopAuthority() }`.
- **Type `workerProbeEvent`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `room.WorkspaceProbeEvent`.
- **Type `workerProbeRecord`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' Intent toolbroker.WorkspaceProbeIntent 'json:"intent"' Events []workerProbeEvent 'json:"events"' }`.
- **Type `workerProbeJournal`**: workerProbeJournal persists bounded lifecycle history under the shared worker admission lock. The caller MUST retain creator ownership through probe cleanup. This is internal infrastructure, not an exporter, public endpoint or FrozenLease.
  Declaration: `struct { dir string guard workerProbeGuard persist func(string, []byte, func(string) error) error }`.
- **`newWorkerProbeJournal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string, guard workerProbeGuard) (*workerProbeJournal, error)`.
  Calls: `filepath.Clean`, `filepath.IsAbs`, `info.IsDir`, `info.Mode`, `info.Mode().Perm`, `info.Sys`, `nilWorkerObservationDependency`, `os.Geteuid`, `os.Lstat`, `uint32`.
- **`probeHex`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(s string, n int) bool`.
  Calls: `hex.DecodeString`, `hex.EncodeToString`, `len`.
- **`validWorkerProbeIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i toolbroker.WorkspaceProbeIntent) bool`.
  Calls: `probeHex`, `strings.HasPrefix`, `strings.TrimPrefix`, `toolbroker.ValidWorkspaceVolume`, `validWorkerJobID`, `workspacemount.ValidChallenge`.
- **`probeRecordName`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(i toolbroker.WorkspaceProbeIntent) string`.
- **`validWorkerProbeRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerProbeRecord) bool`.
  Calls: `len`, `validWorkerHelperEvents`, `validWorkerProbeIntent`.
- **`validWorkerHelperEvents`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(events []workerProbeEvent) bool`.
  Calls: `len`, `probeHex`.
- **`workerProbeStopped`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string) bool`.
  Calls: `filepath.Join`, `os.IsNotExist`, `os.Lstat`.
- **`readWorkerProbeRecords`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string) (map[string]workerProbeRecord, error)`.
  Calls: `bytes.Equal`, `entry.Name`, `f.Close`, `f.ReadDir`, `filepath.Join`, `json.Marshal`, `json.Unmarshal`, `len`, `os.Open`, `probeRecordName`, `readPrivateBounded`, `strings.HasPrefix`, `validWorkerProbeRecord`.
- **`workerProbeCapacity`** — All callers hold .admission.lease. Missing/invalid evidence never frees a slot. The per-worker ceiling is conservative: retained creator roles also stay charged.
  Signature: `func(dir string, jobs []workerJobRecord, next int) error`.
  Calls: `readWorkerProbeRecords`, `workerCaptureSlots`, `workerContainerSlots`, `workerJobSettled`, `workerObjectReadSlots`, `workerProbeSettled`, `workerProbeStopped`.
- **`(*workerProbeJournal).fail`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `filepath.Join`, `j.guard.StopAuthority`, `replacePrivateDurably`.
- **`probeRecordMatchesWorker`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(dir string, i toolbroker.WorkspaceProbeIntent) bool`.
  Calls: `readWorkerRecord`, `retainedObservationRecord`.
- **`(*workerProbeJournal).transition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `append`, `boundedWorkerRecords`, `ctx.Err`, `filepath.Join`, `j.fail`, `j.guard.Acknowledge`, `j.guard.Authorize`, `j.persist`, `json.Marshal`, `len`, `lock.Close`, `nilWorkerObservationDependency`, `probeRecordMatchesWorker`, `probeRecordName`, `readWorkerProbeRecords`, `validWorkerProbeIntent`, `validWorkerProbeRecord`, `workerProbeCapacity`, `workerProbeSettled`, `workerProbeStopped`, `workerWaitLease`.
- **`(*workerProbeJournal).Begin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error`.
  Calls: `j.transition`.
- **`(*workerProbeJournal).Bound`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string) error`.
  Calls: `j.transition`.
- **`(*workerProbeJournal).Finish`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string, absent bool) error`.
  Calls: `j.transition`.
- **`workerProbeSettled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerProbeRecord) bool`.
  Calls: `len`.
