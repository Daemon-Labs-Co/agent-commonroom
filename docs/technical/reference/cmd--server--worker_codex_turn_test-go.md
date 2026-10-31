# cmd/server/worker_codex_turn_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_codex_turn_test.go)

Automated regression checks for worker_codex_turn; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestCodexToolRequestBindsPrivateAuthorityOutsideModelArguments`** — Regression contract: TestCodexToolRequestBindsPrivateAuthorityOutsideModelArguments.
  Signature: `func(t *testing.T)`.
  Calls: `json.RawMessage`, `sampleToolScope`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `workerCodexToolRequest`.
- **`TestCodexTurnRefusesBeforeReviewedPairStartup`** — Regression contract: TestCodexTurnRefusesBeforeReviewedPairStartup.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `s.CodexTurn`, `t.Fatal`, `t.Parallel`, `workspacePair`.
- **`TestCodexCommandArgumentsStayWithinFixedProfiles`** — Regression contract: TestCodexCommandArgumentsStayWithinFixedProfiles.
  Signature: `func(t *testing.T)`.
  Calls: `json.RawMessage`, `len`, `sampleToolScope`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `workerCodexToolRequest`.
