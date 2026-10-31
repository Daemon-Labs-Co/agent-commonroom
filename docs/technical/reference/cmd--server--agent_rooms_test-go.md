# cmd/server/agent_rooms_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/agent_rooms_test.go)

Automated regression checks for agent_rooms; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestAutomaticAgentRoomCatalogAndHistory`** — Regression contract: TestAutomaticAgentRoomCatalogAndHistory.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `hasRoom`, `len`, `load`, `member`, `mockServer`, `restarted.load`, `s.load`, `t.Fatal`, `t.Helper`, `t.Parallel`.
- **`TestAgentRoomReferenceRetention`** — Regression contract: TestAgentRoomReferenceRetention.
  Signature: `func(t *testing.T)`.
  Calls: `hasRoom`, `omitUnusedAgentRooms`, `rememberAgentRoomReferences`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAutomaticAgentRoomFlagOwnerOnly`** — Regression contract: TestAutomaticAgentRoomFlagOwnerOnly.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `len`, `mockRequest`, `mockServer`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestAgentRoomReferenceFastPathPreservesExistingReferences`** — Regression contract: TestAgentRoomReferenceFastPathPreservesExistingReferences.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `rememberAgentRoomReferences`, `t.Fatal`, `t.Parallel`.
