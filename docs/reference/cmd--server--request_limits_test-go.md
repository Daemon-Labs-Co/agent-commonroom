# cmd/server/request_limits_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/request_limits_test.go)

Automated regression checks for request_limits; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestMutationBudgetIsConcurrentAndActorScoped`** — Regression contract: TestMutationBudgetIsConcurrentAndActorScoped.
  Signature: `func(t *testing.T)`.
  Calls: `allowed.Add`, `allowed.Load`, `jobs.Add`, `jobs.Done`, `jobs.Wait`, `mockServer`, `now.Add`, `s.admitActorMutation`, `t.Fatal`, `t.Parallel`, `time.Date`.
- **`TestMutationBudgetCapsTrackingAndEmergencyStopBypassesIt`** — Regression contract: TestMutationBudgetCapsTrackingAndEmergencyStopBypassesIt.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `ctx.Err`, `db.snapshot`, `fmt.Sprint`, `len`, `mockRequest`, `now.Add`, `s.admitActorMutation`, `safetyHarness`, `t.Fatal`, `t.Parallel`, `time.Now`, `w.Body.String`.
