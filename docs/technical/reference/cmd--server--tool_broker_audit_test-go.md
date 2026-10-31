# cmd/server/tool_broker_audit_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/tool_broker_audit_test.go)

Automated regression checks for tool_broker_audit; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`sampleToolScope`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) toolbroker.Scope`.
  Calls: `strings.Repeat`, `t.Helper`, `time.Now`, `time.Now().Add`, `time.Now().Add(2 * time.Minute).UTC`, `time.Now().Add(2 * time.Minute).UTC().Round`.
- **`sampleIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) toolbroker.Intent`.
  Calls: `sampleToolScope`, `strings.Repeat`, `t.Helper`.
- **`TestLegacyToolCallHistoryRemainsReadable`** — Regression contract: TestLegacyToolCallHistoryRemainsReadable.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `len`, `mockServer`, `s.load`, `sampleIntent`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestToolCallOutcomeWithoutIntentFailsReplay`** — Regression contract: TestToolCallOutcomeWithoutIntentFailsReplay.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `mockServer`, `s.load`, `sampleIntent`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
