# cmd/server/worker_pair_start_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_pair_start_test.go)

Automated regression checks for worker_pair_start; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `pairLaunchFake`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T supervisor *workerPairSupervisor calls []string fail string cancelAfterCreate context.CancelFunc }`.
- **`(*pairLaunchFake).Create`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, identity workerPairIdentity) (string, error)`.
  Calls: `append`, `errors.New`, `f.cancelAfterCreate`, `f.t.Fatal`, `readWorkerRecord`, `strings.Repeat`.
- **`(*pairLaunchFake).Verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, identity workerPairIdentity, id string) error`.
  Calls: `append`, `errors.New`.
- **`(*pairLaunchFake).Inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) ([]byte, error)`.
  Calls: `append`, `errors.New`, `fmt.Sprintf`.
- **`(*pairLaunchFake).Start`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `append`, `errors.New`, `f.t.Fatal`, `readWorkerRecord`.
- **`(*pairLaunchFake).Bootstrap`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (workspacebootstrap.Receipt, error)`.
  Calls: `append`, `errors.New`, `strings.Repeat`.
- **`startPairRequest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() workerRunRequest`.
  Calls: `pairRequest`.
- **`(*pairLaunchFake).WorkspaceIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (toolbroker.WorkspaceVolumeIdentity, error)`.
  Calls: `append`, `strings.Repeat`.
- **`(*pairLaunchFake).PrepareVolume`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, intent toolbroker.WorkspaceVolumeIdentity) (toolbroker.WorkspaceVolumeIdentity, error)`.
  Calls: `append`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Format`.
- **`TestWorkerPairStartOwnsDurableOrdering`** — Regression contract: TestWorkerPairStartOwnsDurableOrdering.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `reflect.DeepEqual`, `s.Close`, `s.Start`, `startPairRequest`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.TempDir`.
- **`TestWorkerPairStartRetainsUncertainFailure`** — Regression contract: TestWorkerPairStartRetainsUncertainFailure.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `len`, `readWorkerRecord`, `s.Close`, `s.Start`, `slices.Contains`, `startPairRequest`, `strings.HasPrefix`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestWorkerPairStartStopsOnCancellation`** — Regression contract: TestWorkerPairStartStopsOnCancellation.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `cancel`, `context.WithCancel`, `filepath.Join`, `os.WriteFile`, `s.Close`, `s.Start`, `startPairRequest`, `strings.HasPrefix`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **Type `pairBlockingLaunch`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *pairLaunchFake phase string entered chan struct{} release chan struct{} }`.
- **`(*pairBlockingLaunch).Create`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, identity workerPairIdentity) (string, error)`.
  Calls: `close`, `f.pairLaunchFake.Create`.
- **`(*pairBlockingLaunch).Start`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `close`, `f.pairLaunchFake.Start`.
- **`TestWorkerPairCloseRetainsLeaseUntilStartupReturns`** — Regression contract: TestWorkerPairCloseRetainsLeaseUntilStartupReturns.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `close`, `lease.Close`, `make`, `s.Close`, `s.Start`, `s.ctx.Done`, `startPairRequest`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workerTryLease`.
- **`(*pairLaunchFake).Scope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() toolbroker.Scope`.
  Calls: `r.Deadline.UTC`, `r.Deadline.UTC().Round`.
