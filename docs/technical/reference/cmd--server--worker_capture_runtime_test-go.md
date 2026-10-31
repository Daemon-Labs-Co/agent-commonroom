# cmd/server/worker_capture_runtime_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_runtime_test.go)

Automated regression checks for worker_capture_runtime; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ownedCaptureFixture`**: This fixture drives the real plan/runtime/protocol/local lifecycle journal. All daemon, frozen-source, custody and controller evidence remains synthetic.
  Declaration: `struct { base *physicalCaptureFixture guard *captureJournalGuard phase string exists bool mode string starts, reserves, seals int onStart func(context.Context) onInspect func(context.Context) }`.
- **`(*ownedCaptureFixture).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `b.Bytes`, `errors.New`, `f.base.Run`, `f.base.probe.checkLock`, `f.base.probe.t.Fatal`, `f.onInspect`, `io.WriteString`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `json.Unmarshal`, `len`, `plan.CreateArgs`, `strings.Join`, `toolbroker.NewCaptureContainerPlan`.
- **`(*ownedCaptureFixture).StartCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (toolbroker.CaptureProcess, error)`.
  Calls: `captureRecordName`, `client.AuthorizeFrozenSource`, `client.Finish`, `client.Reserve`, `client.Seal`, `close`, `f.base.probe.t.Fatal`, `f.onStart`, `helper.Close`, `len`, `make`, `net.Pipe`, `readWorkerCaptureRecords`, `strings.Repeat`, `workspacecapture.NewCaptureChannel`.
- **Type `ownedCaptureProcess`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn f *ownedCaptureFixture done chan error joined chan struct{} }`.
- **`(*ownedCaptureProcess).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `p.Conn.Close`.
- **`(*ownedCaptureProcess).Wait`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `ctx.Done`, `ctx.Err`.
- **`(*ownedCaptureFixture).AuthorizeFrozenSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ workspacesnapshot.Binding, _ workspacemount.Proof) error`.
  Calls: `ctx.Err`.
- **`(*ownedCaptureFixture).Reserve`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ objectcustody.Intent) error`.
  Calls: `ctx.Err`.
- **`(*ownedCaptureFixture).Seal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ objectcustody.Receipt) error`.
  Calls: `ctx.Err`.
- **`(*ownedCaptureFixture).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Scope, objectcustody.Receipt) error`.
- **`(*ownedCaptureFixture).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
- **`ownedCaptureSetup`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*ownedCaptureFixture, *workerWorkspaceObservation, *workerCaptureJournal)`.
  Calls: `captureJournalFixtureScope`, `t.Helper`.
- **`TestOwnedCaptureAssemblesLifecycleAndRetainsUncertainty`** — Regression contract: TestOwnedCaptureAssemblesLifecycleAndRetainsUncertainty.
  Signature: `func(t *testing.T)`.
  Calls: `capture.Run`, `captureRecordName`, `captureRecordSettled`, `copyCaptureRequest`, `len`, `newOwnedWorkerCapture`, `ownedCaptureSetup`, `readWorkerCaptureRecords`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestOwnedCaptureRefusesMissingOrExpiredSetup`** — Regression contract: TestOwnedCaptureRefusesMissingOrExpiredSetup.
  Signature: `func(t *testing.T)`.
  Calls: `c.Run`, `cancel`, `context.WithCancel`, `copyCaptureRequest`, `l.mu.Lock`, `l.mu.Unlock`, `len`, `newOwnedWorkerCapture`, `ownedCaptureSetup`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **Type `blockingCaptureGuard`**: Returning from the owner callback cannot release the creator lock while a canceled helper operation is still joining its trusted dependencies.
  Declaration: `struct{ entered, canceled, release chan struct{} }`.
- **`(*blockingCaptureGuard).AuthorizeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ workerJobRecord, _ workerCaptureIntent, _ workerProbeEvent) error`.
  Calls: `close`, `ctx.Done`, `ctx.Err`.
- **`(*blockingCaptureGuard).AcknowledgeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, workerCaptureRecord, workerProbeEvent) error`.
- **`(*blockingCaptureGuard).StopCaptureAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`(*blockingCaptureGuard).CaptureAuthorityLifetime`** — This explicitly synthetic guard tests owner cancellation independently of transport loss; production must supply an authenticated cancelable lifetime.
  Signature: `func() context.Context`.
  Calls: `context.Background`.
- **`TestOwnedCaptureTeardownJoinsBeforeCreatorUnlock`** — Regression contract: TestOwnedCaptureTeardownJoinsBeforeCreatorUnlock.
  Signature: `func(t *testing.T)`.
  Calls: `checkOwnedCaptureJoin`, `t.Parallel`, `t.Run`.
- **`checkOwnedCaptureJoin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, recovery bool)`.
  Calls: `beginCaptureJournal`, `c.Run`, `close`, `errors.New`, `j.BindCapture`, `j.FinishCapture`, `l.ownership.Close`, `lease.Close`, `make`, `newOwnedWorkerCapture`, `ownedCaptureSetup`, `recoverRetainedWorkerCapture`, `t.Context`, `t.Fatal`, `t.Helper`, `time.After`, `withWorkerWorkspaceOwnership`, `workerTryLease`.
