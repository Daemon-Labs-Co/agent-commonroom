# cmd/server/p0_idle_uncertainty_acceptance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_idle_uncertainty_acceptance_test.go)

Automated regression checks for p0_idle_uncertainty_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0IdleCleanupUncertain`** — Combine actual worker cleanup, controller reconciliation and quota admission. The evidence-return adapter substitutes for SSH only; its result is derived from the worker's durable record, never from the catalog's expected decision.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `admitManagedContainerQuota`, `announce`, `append`, `errors.New`, `f.db.seed`, `f.db.snapshot`, `f.now.Add`, `f.s.load`, `f.s.reconcileManagedJob`, `f.worker.expireIdle`, `f.worker.queueTurn`, `len`, `managedJobFence`, `newIdleNoticeIntervalFixture`, `quotaJob`, `readWorkerRecord`, `reflect.DeepEqual`, `registration.Register`, `restarted.agentsBlockedLocked`, `restarted.load`, `restarted.managedJobContext`, `restarted.managedJobContext(t.Context(), p).Value`, `strconv.Atoi`, `t.Context`, `t.Fatal`, `t.Helper`, `workerJobSettled`.
