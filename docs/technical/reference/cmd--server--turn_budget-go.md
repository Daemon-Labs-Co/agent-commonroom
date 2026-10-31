# cmd/server/turn_budget.go

[Module index](cmd.md) · [Open source](../../../cmd/server/turn_budget.go)

Go implementation for turn budget; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `turnBudget`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Limit int 'json:"limit"' Used int 'json:"used"' Remaining int 'json:"remaining"' ResetsAt time.Time 'json:"resetsAt"' }`.
- **`parseDailyTurnLimit`** — parseDailyTurnLimit is an operator setting, not an optional security flag. Invalid/zero values fail startup rather than silently removing the budget.
  Signature: `func(raw string) (int, error)`.
  Calls: `errors.New`, `strconv.Atoi`, `strings.TrimSpace`.
- **`(*server).dailyTurnBudget`** — dailyTurnBudget counts unique controller-journalled intents, not lifecycle updates, successful completions or dollars. Restart cannot reset allocations.
  Signature: `func(st room.State, now time.Time) turnBudget`.
  Calls: `job.AllocatedAt.Before`, `now.Day`, `now.Month`, `now.UTC`, `now.Year`, `start.AddDate`, `time.Date`.
- **`(*server).admitDailyTurns`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, now time.Time, requested int) error`.
  Calls: `budget.ResetsAt.Format`, `fmt.Errorf`, `s.dailyTurnBudget`.
