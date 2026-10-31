# cmd/server/room_outcomes_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/room_outcomes_test.go)

Automated regression checks for room_outcomes; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRoomOutcomeLifecycle`** — Regression contract: TestRoomOutcomeLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `call`, `context.Background`, `db.snapshot`, `json.Marshal`, `len`, `load`, `mockRequest`, `mockServer`, `restarted.load`, `s.load`, `string`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `t.Parallel`.
- **`TestRoomOutcomePermissionsAndLegacy`** — Regression contract: TestRoomOutcomePermissionsAndLegacy.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `fmt.Sprintf`, `json.Marshal`, `len`, `mockRequest`, `mockServer`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.String`.
