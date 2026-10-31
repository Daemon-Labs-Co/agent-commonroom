# cmd/server/controller_probe_ledger_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_probe_ledger_test.go)

Automated regression checks for controller_probe_ledger; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticControllerProbeEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { validate func(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error }`.
- **`(*syntheticControllerProbeEvidence).ValidateProbe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.WorkspaceProbeReservation, e workerProbeEvent) error`.
  Calls: `ctx.Err`, `g.validate`.
- **`controllerProbeFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, context.Context, toolbroker.WorkspaceProbeIntent)`.
  Calls: `context.WithValue`, `db.seed`, `mockServer`, `quotaJob`, `strings.Repeat`, `t.Context`, `t.Helper`.
- **`probeControllerState`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, s *server, ctx context.Context) room.State`.
  Calls: `s.load`, `t.Fatal`, `t.Helper`.
- **`TestControllerProbeLedgerRestartDedupAndQuota`** — Regression contract: TestControllerProbeLedgerRestartDedupAndQuota.
  Signature: `func(t *testing.T)`.
  Calls: `admitManagedContainerQuota`, `controllerProbeFixture`, `controllerProbeSlots`, `db.snapshot`, `j.Acknowledge`, `len`, `newControllerProbeLedger`, `other.Acknowledge`, `probeControllerState`, `quotaJob`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerProbeLedgerRetainsUnknownAndRecovery`** — Regression contract: TestControllerProbeLedgerRetainsUnknownAndRecovery.
  Signature: `func(t *testing.T)`.
  Calls: `controllerProbeFixture`, `controllerProbeSlots`, `fmt.Sprint`, `j.Acknowledge`, `newControllerProbeLedger`, `probeControllerState`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerProbeLedgerAuthorityEvidenceAndFailure`** — Regression contract: TestControllerProbeLedgerAuthorityEvidenceAndFailure.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `controllerProbeFixture`, `db.seed`, `db.snapshot`, `errors.New`, `j.Acknowledge`, `len`, `newControllerProbeLedger`, `quotaJob`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerProbeLedgerConcurrentControllersLastSlot`** — Regression contract: TestControllerProbeLedgerConcurrentControllersLastSlot.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `controllerProbeFixture`, `db.seed`, `errors.Is`, `j.Acknowledge`, `len`, `make`, `newControllerProbeLedger`, `probeControllerState`, `quotaJob`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestControllerProbeLedgerPrivateStateAndInjection`** — Regression contract: TestControllerProbeLedgerPrivateStateAndInjection.
  Signature: `func(t *testing.T)`.
  Calls: `controllerProbeFixture`, `j.Acknowledge`, `json.Marshal`, `mockRequest`, `mockServer`, `newControllerProbeLedger`, `probeControllerState`, `string`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestControllerProbeLedgerCleanupWhileLaunchesStopped`** — Regression contract: TestControllerProbeLedgerCleanupWhileLaunchesStopped.
  Signature: `func(t *testing.T)`.
  Calls: `controllerProbeFixture`, `db.seed`, `j.Acknowledge`, `newControllerProbeLedger`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerProbeLedgerEvidenceSnapshotConflict`** — Regression contract: TestControllerProbeLedgerEvidenceSnapshotConflict.
  Signature: `func(t *testing.T)`.
  Calls: `controllerProbeFixture`, `db.seed`, `errors.Is`, `j.Acknowledge`, `len`, `newControllerProbeLedger`, `probeControllerState`, `quotaJob`, `t.Fatal`, `t.Parallel`.
- **`TestControllerProbeLedgerRejectsForgedHistory`** — Regression contract: TestControllerProbeLedgerRejectsForgedHistory.
  Signature: `func(t *testing.T)`.
  Calls: `controllerProbeFixture`, `controllerProbeKey`, `controllerProbeSlots`, `delete`, `j.Acknowledge`, `json.Marshal`, `newControllerProbeLedger`, `probeControllerState`, `projectControllerProbe`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerProbeLedgerHumanRoomAccessRevocation`** — Regression contract: TestControllerProbeLedgerHumanRoomAccessRevocation.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `db.seed`, `j.Acknowledge`, `mockServer`, `newControllerProbeLedger`, `quotaJob`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
