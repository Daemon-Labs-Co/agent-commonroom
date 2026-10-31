# cmd/server/agent_boundaries_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/agent_boundaries_test.go)

Automated regression checks for agent_boundaries; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestAgentDispatchBoundariesWithMockDB`** — Regression contract: TestAgentDispatchBoundariesWithMockDB.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `context.WithCancel`, `ctx.Err`, `db.seed`, `db.snapshot`, `http.HandlerFunc`, `len`, `mockRequest`, `mockServer`, `s.load`, `strings.Contains`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `w.Body.String`.
- **`TestChallengeCapacityAndExpiry`** — Regression contract: TestChallengeCapacityAndExpiry.
  Signature: `func(t *testing.T)`.
  Calls: `httptest.NewRecorder`, `httptest.NewRequest`, `len`, `mockServer`, `s.challenge`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
