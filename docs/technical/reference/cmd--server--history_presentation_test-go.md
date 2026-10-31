# cmd/server/history_presentation_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/history_presentation_test.go)

Automated regression checks for history_presentation; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestHistoryPresentationSafeContextAndSeverity`** — Regression contract: TestHistoryPresentationSafeContextAndSeverity.
  Signature: `func(t *testing.T)`.
  Calls: `describeAuditEvent`, `json.Marshal`, `string`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestHistoryRefinementServerFlagBothStates`** — Regression contract: TestHistoryRefinementServerFlagBothStates.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `fresh.historyPresentationEnabled`, `json.Unmarshal`, `mockRequest`, `mockServer`, `newMockDB`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `w.Body.Bytes`, `w.Body.String`.
