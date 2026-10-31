# cmd/server/worker_object_read_runtime_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_object_read_runtime_test.go)

Automated regression checks for worker_object_read_runtime; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `assembledObjectReadFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *physicalObjectReadFixture t *testing.T planArgs []string mode string process *assembledObjectReadProcess release *controllerArtifactReadSession onStart func() }`.
- **`(*assembledObjectReadFixture).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `errors.New`, `f.physicalObjectReadFixture.Run`, `f.process.closed.Load`, `f.t.Fatal`, `io.WriteString`, `len`, `objectReadRecordName`, `readWorkerObjectReadRecords`, `slices.Equal`.
- **`(*assembledObjectReadFixture).StartObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (toolbroker.ObjectReadProcess, error)`.
  Calls: `f.onStart`, `f.t.Fatal`, `make`, `net.Pipe`, `objectReadRecordName`, `objectcustody.ServeReadChannel`, `readWorkerObjectReadRecords`.
- **Type `syntheticAssembledObjectSource`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(syntheticAssembledObjectSource).Get`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ objectcustody.Scope, _ objectcustody.Receipt) ([]byte, error)`.
  Calls: `ctx.Err`.
- **Type `assembledObjectReadProcess`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn f *assembledObjectReadFixture done chan error waited, closed atomic.Bool }`.
- **`(*assembledObjectReadProcess).Wait`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `ctx.Err`, `errors.Join`, `errors.New`, `p.waited.CompareAndSwap`.
- **`(*assembledObjectReadProcess).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `errors.New`, `p.Conn.Close`, `p.closed.CompareAndSwap`, `p.waited.Load`.
- **`TestWorkerObjectReadRuntimeControllerBridge`** — Regression contract: TestWorkerObjectReadRuntimeControllerBridge.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `c.ObjectReadAuthorityLifetime`, `c.ObjectReadAuthorityLifetime().Done`, `c.StopObjectReadAuthority`, `close`, `context.WithCancel`, `controllerObjectReadSlots`, `creator.Done`, `detachedProbeWorkerRecord`, `errors.New`, `f.process.closed.Load`, `filepath.Join`, `l.mu.Lock`, `l.mu.Unlock`, `lease.Close`, `left.Close`, `make`, `net.Pipe`, `newControllerArtifactReadSession`, `newObjectReadWorkerService`, `newSyntheticReadListener`, `newWorkerObjectReadRuntime`, `newWorkspaceObjectReadEvidence`, `objectReadProfileFixture`, `objectReadSessionFixture`, `os.Chmod`, `os.WriteFile`, `plan.CreateArgs`, `release.lifetime.Err`, `right.Close`, `runtime.Run`, `service.Close`, `service.Connect`, `service.ListenAndServe`, `stopCreator`, `string`, `strings.Repeat`, `syntheticReadDialer`, `syntheticReadListenerFactory`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `toolbroker.NewObjectReadContainerPlan`, `workerObjectReadSlots`, `workerTryLease`, `writeWorkerRecord`.
- **`TestWorkerObjectReadRuntimeRequiresBoundary`** — Regression contract: TestWorkerObjectReadRuntimeRequiresBoundary.
  Signature: `func(t *testing.T)`.
  Calls: `detachedProbeWorkerRecord`, `lease.Close`, `newControllerArtifactReadSession`, `newWorkerObjectReadRuntime`, `objectReadSessionFixture`, `os.Chmod`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workerTryLease`.
- **`TestWorkerObjectReadRecoveryControllerBridge`** — Regression contract: TestWorkerObjectReadRecoveryControllerBridge.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `c.StopObjectReadAuthority`, `close`, `controllerObjectReadSlots`, `db.seed`, `detachedProbeWorkerRecord`, `filepath.Join`, `journal.BeginObjectRead`, `journal.BindObjectRead`, `journal.FinishObjectRead`, `lease.Close`, `make`, `net.Pipe`, `newControllerObjectReadSession`, `newObjectReadAuthorityClient`, `newWorkerObjectReadJournal`, `newWorkspaceObjectReadEvidence`, `objectReadSessionFixture`, `open`, `os.Chmod`, `os.WriteFile`, `recoverRetainedWorkerObjectRead`, `serveObjectReadAuthority`, `strings.Repeat`, `t.Cleanup`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workerObjectReadSlots`, `workerTryLease`, `writeWorkerRecord`.
