# cmd/server/turn_budget_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/turn_budget_test.go)

Automated regression checks for turn_budget; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestDailyTurnBudgetUTCWindowAndUniqueJobs`** — Regression contract: TestDailyTurnBudgetUTCWindowAndUniqueJobs.
  Signature: `func(t *testing.T)`.
  Calls: `b.ResetsAt.Equal`, `delete`, `midnight.Add`, `now.UTC`, `s.admitDailyTurns`, `s.dailyTurnBudget`, `t.Fatal`, `t.Parallel`, `time.Date`, `time.FixedZone`.
- **`TestDailyTurnLimitConfiguration`** — Regression contract: TestDailyTurnLimitConfiguration.
  Signature: `func(t *testing.T)`.
  Calls: `parseDailyTurnLimit`, `t.Fatal`, `t.Parallel`.
- **`TestDailyTurnBudgetPersistsAcrossControllerRestart`** — Regression contract: TestDailyTurnBudgetPersistsAcrossControllerRestart.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `reg.Update`, `restarted.admitDailyTurns`, `restarted.load`, `syntheticManagedJob`, `t.Fatal`, `t.Parallel`, `time.Now`.
- **`TestLastDailySlotIsAtomicAcrossRegistrations`** — Regression contract: TestLastDailySlotIsAtomicAcrossRegistrations.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `close`, `context.Background`, `db.seed`, `len`, `make`, `mockServer`, `reg.Register`, `s.load`, `s.managedJobContext`, `s.managedJobContext(context.Background(), p).Value`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestFailedAllocationDoesNotConsumeBudget`** — Regression contract: TestFailedAllocationDoesNotConsumeBudget.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `mockServer`, `reg.Register`, `s.dailyTurnBudget`, `s.load`, `s.managedJobContext`, `s.managedJobContext(context.Background(), p).Value`, `t.Fatal`, `t.Parallel`, `time.Now`.
- **`TestConnectionVerificationCannotBypassDailyBudget`** — Regression contract: TestConnectionVerificationCannotBypassDailyBudget.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `reg.Register`, `reg.Update`, `s.managedJobContext`, `s.managedJobContext(context.Background(), p).Value`, `s.managedVerification`, `t.Error`, `t.Fatal`, `t.Parallel`, `verificationFixture`.
