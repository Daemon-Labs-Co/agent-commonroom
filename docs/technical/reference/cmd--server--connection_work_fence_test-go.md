# cmd/server/connection_work_fence_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/connection_work_fence_test.go)

Automated regression checks for connection_work_fence; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestConnectionGuardRetainsDurableWorkAcrossControllerRestart`** — Regression contract: TestConnectionGuardRetainsDurableWorkAcrossControllerRestart.
  Signature: `func(t *testing.T)`.
  Calls: `connectionHasActiveAgent`, `participantHasUnsettledWork`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestConnectionGuardFencesPendingTaskActualRole`** — Regression contract: TestConnectionGuardFencesPendingTaskActualRole.
  Signature: `func(t *testing.T)`.
  Calls: `connectionHasActiveAgent`, `fmt.Sprint`, `participantHasUnsettledWork`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestConnectionMutationsWaitForDurableCleanup`** — Regression contract: TestConnectionMutationsWaitForDurableCleanup.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `fmt.Sprintf`, `json.Marshal`, `len`, `mockRequest`, `reg.Update`, `syntheticManagedJob`, `t.Fatal`, `t.Parallel`, `t.Run`, `w.Body.String`.
