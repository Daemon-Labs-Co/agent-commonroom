# cmd/server/p0_idle_handoff_acceptance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_idle_handoff_acceptance_test.go)

Automated regression checks for p0_idle_handoff_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0IdleArrivingWork`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `idleWorkspace`, `len`, `readWorkerRecord`, `reflect.DeepEqual`, `reset.Add`, `reset.After`, `reset.Before`, `since.Add`, `strconv.Atoi`, `t.Fatal`, `t.Helper`, `time.Duration`, `time.Now`, `worker.expireIdle`, `worker.finishQueuedTurn`, `worker.mu.Lock`, `worker.mu.Unlock`, `worker.queueTurn`.
- **`driveP0IdleOwnerInterval`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `announce`, `append`, `f.db.snapshot`, `f.now.Add`, `f.s.load`, `f.worker.ctx.Err`, `f.worker.expireIdle`, `idlePolicyForState`, `idleSettingsHTTP`, `json.Marshal`, `len`, `mockRequest`, `newIdleNoticeIntervalFixture`, `reflect.DeepEqual`, `s.load`, `strconv.Atoi`, `string`, `t.Context`, `t.Fatal`, `t.Helper`, `tick.Add`, `time.Duration`.
