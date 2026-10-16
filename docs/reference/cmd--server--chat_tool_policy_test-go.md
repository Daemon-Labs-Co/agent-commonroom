# cmd/server/chat_tool_policy_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/chat_tool_policy_test.go)

Automated regression checks for chat_tool_policy; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestChatCommandToolsCannotBeEnabledByWorkerRequest`** — Regression contract: TestChatCommandToolsCannotBeEnabledByWorkerRequest.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `sampleWorkerRequest`, `slices.Clone`, `slices.Index`, `t.Fatal`, `t.Parallel`, `validateWorkerRun`.
- **`TestChatPolicyAppliesToFreshAndResumedTurns`** — Regression contract: TestChatPolicyAppliesToFreshAndResumedTurns.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `providerTurnArgs`, `sampleWorkerRequest`, `slices.Contains`, `t.Fatal`, `t.Parallel`, `validateWorkerRun`.
