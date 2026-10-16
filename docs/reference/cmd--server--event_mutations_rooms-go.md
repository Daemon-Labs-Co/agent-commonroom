# cmd/server/event_mutations_rooms.go

[Module index](cmd.md) · [Open source](../../cmd/server/event_mutations_rooms.go)

Go implementation for event mutations rooms; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).prepareRoomsEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request, st room.State, roomID string, req eventRequest) (any, bool)`.
  Calls: `e.Error`, `hasMember`, `hasRoom`, `http.Error`, `id`, `json.Unmarshal`, `strings.TrimSpace`, `time.Now`, `time.Now().UTC`, `validateGraph`.
