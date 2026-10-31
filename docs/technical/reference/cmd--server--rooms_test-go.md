# cmd/server/rooms_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/rooms_test.go)

Automated regression checks for rooms; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRoomIsolation`** — Regression contract: TestRoomIsolation.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `scoped`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestChallengeGuidance`** — Regression contract: TestChallengeGuidance.
  Signature: `func(t *testing.T)`.
  Calls: `profileGuidance`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
