# cmd/server/controller_object_read_ledger_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_object_read_ledger_test.go)

Automated regression checks for controller_object_read_ledger; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticObjectReadEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { check func(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error calls int }`.
- **`(*syntheticObjectReadEvidence).ValidateObjectReadTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.ObjectReadReservation, v workerProbeEvent) error`.
  Calls: `ctx.Err`, `e.check`.
- **`objectReadLedgerFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*controllerObjectReadLedger, *server, *mockDB, context.Context, workerObjectReadIntent, *syntheticObjectReadEvidence)`.
  Calls: `artifactJournalFixture`, `artifactStore`, `newControllerArtifactJournal`, `newControllerObjectReadLedger`, `store.Put`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `toolbroker.ValidObjectReadMetadata`.
- **`TestControllerObjectReadLedgerLifecycleAndCapacity`** — Regression contract: TestControllerObjectReadLedgerLifecycleAndCapacity.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `artifactState`, `auditSummary`, `controllerHelperSlots`, `controllerObjectReadSlots`, `db.snapshot`, `j.Acknowledge`, `len`, `objectReadLedgerFixture`, `projectControllerObjectRead`, `strings.Contains`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerObjectReadLedgerRecovery`** — Regression contract: TestControllerObjectReadLedgerRecovery.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `controllerObjectReadSlots`, `db.seed`, `db.snapshot`, `errors.New`, `j.Acknowledge`, `len`, `objectReadLedgerFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerObjectReadLedgerRefusals`** — Regression contract: TestControllerObjectReadLedgerRefusals.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `context.WithValue`, `db.seed`, `db.snapshot`, `errors.Is`, `errors.New`, `j.Acknowledge`, `len`, `objectReadLedgerFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerObjectReadLedgerEditingOffRetainedRead`** — Regression contract: TestControllerObjectReadLedgerEditingOffRetainedRead.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `controllerObjectReadSlots`, `db.seed`, `j.Acknowledge`, `objectReadLedgerFixture`, `t.Fatal`, `t.Parallel`.
- **`TestControllerObjectReadLedgerMalformedAndPrivateHistory`** — Regression contract: TestControllerObjectReadLedgerMalformedAndPrivateHistory.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `controllerHelperSlots`, `controllerObjectReadKey`, `j.Acknowledge`, `json.Marshal`, `mockRequest`, `mockServer`, `newControllerObjectReadLedger`, `objectReadLedgerFixture`, `projectControllerObjectRead`, `string`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestControllerObjectReadLedgerDetachedEvidence`** — Regression contract: TestControllerObjectReadLedgerDetachedEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `controllerObjectReadKey`, `j.Acknowledge`, `objectReadLedgerFixture`, `t.Fatal`, `t.Parallel`.
- **`TestControllerObjectReadLedgerConcurrentControllers`** — Regression contract: TestControllerObjectReadLedgerConcurrentControllers.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `close`, `errors.Is`, `j.Acknowledge`, `len`, `make`, `newControllerObjectReadLedger`, `objectReadLedgerFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestControllerObjectReadLedgerCaptureExclusionAndCombinedQuota`** — Regression contract: TestControllerObjectReadLedgerCaptureExclusionAndCombinedQuota.
  Signature: `func(t *testing.T)`.
  Calls: `admitControllerProbeQuota`, `artifactState`, `controllerCaptureKey`, `controllerHelperSlots`, `controllerObjectReadKey`, `controllerProbeKey`, `delete`, `j.Acknowledge`, `objectReadLedgerFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validateControllerCaptureEvent`, `validateControllerObjectReadEvent`, `workspacecapture.ContractRequest`.
