# cmd/server/legacy_shutdown_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/legacy_shutdown_test.go)

Automated regression checks for legacy_shutdown; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestTerminalLabelCannotReplaceCurrentShutdownProof`** — Regression contract: TestTerminalLabelCannotReplaceCurrentShutdownProof.
  Signature: `func(t *testing.T)`.
  Calls: `fakeWorkerRecord`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `workerAdmission`, `workerJobSettled`.
- **`TestLegacyTerminalSnapshotRemainsUnresolvedAndUnmodified`** — Regression contract: TestLegacyTerminalSnapshotRemainsUnresolvedAndUnmodified.
  Signature: `func(t *testing.T)`.
  Calls: `fakeWorkerRecord`, `len`, `readWorkerRecord`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `workerSnapshot`, `writeWorkerRecord`.
