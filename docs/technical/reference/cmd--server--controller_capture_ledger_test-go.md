# cmd/server/controller_capture_ledger_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_capture_ledger_test.go)

Automated regression checks for controller_capture_ledger; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticCaptureLedgerEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { check func(context.Context, room.State, room.WorkspaceCaptureReservation, workerProbeEvent) error calls int }`.
- **`(*syntheticCaptureLedgerEvidence).ValidateCaptureTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.WorkspaceCaptureReservation, v workerProbeEvent) error`.
  Calls: `ctx.Err`, `e.check`.
- **`captureLedgerFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*controllerCaptureLedger, *server, *mockDB, context.Context, workerCaptureIntent, *syntheticCaptureLedgerEvidence)`.
  Calls: `controllerCaptureFixture`, `newControllerCaptureLedger`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`TestControllerCaptureLedgerDurableCapacityReplay`** — Regression contract: TestControllerCaptureLedgerDurableCapacityReplay.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `artifactState`, `auditSummary`, `captureLedgerFixture`, `controllerCaptureSlots`, `db.snapshot`, `j.Acknowledge`, `len`, `projectControllerCapture`, `strings.Contains`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerCaptureLedgerUnknownAndKnownRecovery`** — Regression contract: TestControllerCaptureLedgerUnknownAndKnownRecovery.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `captureLedgerFixture`, `controllerCaptureSlots`, `j.Acknowledge`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerCaptureLedgerDenialsAndFencing`** — Regression contract: TestControllerCaptureLedgerDenialsAndFencing.
  Signature: `func(t *testing.T)`.
  Calls: `captureLedgerFixture`, `context.WithValue`, `db.seed`, `db.snapshot`, `errors.Is`, `errors.New`, `j.Acknowledge`, `len`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerCaptureLedgerCleanupWithFlagsOffAndCombinedQuota`** — Regression contract: TestControllerCaptureLedgerCleanupWithFlagsOffAndCombinedQuota.
  Signature: `func(t *testing.T)`.
  Calls: `admitControllerProbeQuota`, `artifactState`, `captureLedgerFixture`, `controllerHelperSlots`, `controllerProbeKey`, `db.seed`, `j.Acknowledge`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerCaptureLedgerRejectsPublicAndMalformedHistory`** — Regression contract: TestControllerCaptureLedgerRejectsPublicAndMalformedHistory.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `captureLedgerFixture`, `controllerCaptureKey`, `controllerHelperSlots`, `j.Acknowledge`, `json.Marshal`, `mockRequest`, `mockServer`, `projectControllerCapture`, `t.Fatal`.
- **`TestControllerCaptureLedgerEvidenceCannotMutatePinnedIntent`** — Regression contract: TestControllerCaptureLedgerEvidenceCannotMutatePinnedIntent.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `captureLedgerFixture`, `controllerCaptureKey`, `j.Acknowledge`, `t.Fatal`, `t.Parallel`.
