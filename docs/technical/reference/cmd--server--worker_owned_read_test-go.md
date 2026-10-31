# cmd/server/worker_owned_read_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_owned_read_test.go)

Automated regression checks for worker_owned_read; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `blockingOwnedReadGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { lifetime context.Context entered, canceled, release, stopping, stopped chan struct{} }`.
- **`(*blockingOwnedReadGuard).ObjectReadAuthorityLifetime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() context.Context`.
- **`(*blockingOwnedReadGuard).AuthorizeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ workerJobRecord, _ workerObjectReadIntent, _ workerProbeEvent) error`.
  Calls: `close`, `ctx.Done`, `ctx.Err`.
- **`(*blockingOwnedReadGuard).AcknowledgeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, workerObjectReadRecord, workerProbeEvent) error`.
- **`(*blockingOwnedReadGuard).StopObjectReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `close`.
- **Type `ownedReadRelease`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ stopped bool }`.
- **`(*ownedReadRelease).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Scope, objectcustody.Receipt) error`.
- **`(*ownedReadRelease).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, objectcustody.Receipt) error`.
- **`(*ownedReadRelease).StopReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`TestOwnedReadJoinsRevocationBeforeCreatorUnlock`** — Regression contract: TestOwnedReadJoinsRevocationBeforeCreatorUnlock.
  Signature: `func(t *testing.T)`.
  Calls: `checkOwnedReadJoin`, `t.Parallel`, `t.Run`.
- **`checkOwnedReadJoin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, recovery bool)`.
  Calls: `assertHeld`, `close`, `errors.New`, `j.BeginObjectRead`, `j.BindObjectRead`, `j.FinishObjectRead`, `l.ownership.Close`, `lease.Close`, `len`, `make`, `newWorkerObjectReadRuntime`, `objectReadJournalFixture`, `readWorkerCaptureRecords`, `recoverRetainedWorkerObjectRead`, `runtime.Run`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `time.After`, `withWorkerWorkspaceOwnership`, `workerObjectReadSlots`, `workerTryLease`.
