# cmd/server/room_outcomes.go

[Module index](cmd.md) · [Open source](../../../cmd/server/room_outcomes.go)

Go implementation for room outcomes; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`prepareRoomOutcome`** — Outcome metadata defines intent, never execution permission or completion. Historical events remain readable regardless of the current feature setting.
  Signature: `func(w http.ResponseWriter, st room.State, req eventRequest, v *room.Room) bool`.
  Calls: `http.Error`, `json.Unmarshal`, `strings.TrimSpace`, `utf8.RuneCountInString`.
