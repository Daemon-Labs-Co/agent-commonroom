# cmd/server/p0_idle_audit_acceptance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_idle_audit_acceptance_test.go)

Automated regression checks for p0_idle_audit_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0IdleAuditLifecycle`** — Drive real idle admission with explicit clock ticks and mocked Docker. These refusal cases do not establish default-expiry audit or production admission.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `idleWorkspace`, `len`, `readWorkerIdlePolicy`, `readWorkerRecord`, `reflect.DeepEqual`, `since.Add`, `strconv.Atoi`, `strconv.FormatBool`, `t.Fatal`, `t.Helper`, `time.Duration`, `worker.ctx.Err`, `worker.expireIdle`, `worker.turn.Lock`, `worker.turn.Unlock`.
