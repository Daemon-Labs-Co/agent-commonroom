# cmd/server/agent_rooms.go

[Module index](cmd.md) · [Open source](../../../cmd/server/agent_rooms.go)

Go implementation for agent rooms; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`rememberAgentRoomReferences`** — Historical replay still derives legacy identities/rooms for scope validation. Only unused implicit agent rooms are omitted from the resulting room catalog. Any historical reference conservatively retains the room, including settings, jobs, audit records and removed work; hiding a room must never erase history.
  Signature: `func(data []byte, referenced map[string]bool)`.
  Calls: `bytes.Contains`, `json.Unmarshal`, `strings.HasPrefix`, `visit`.
- **`omitUnusedAgentRooms`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, referenced map[string]bool)`.
  Calls: `append`, `member`.
