# cmd/server/event_mutations_agents.go

[Module index](cmd.md) · [Open source](../../../cmd/server/event_mutations_agents.go)

Go implementation for event mutations agents; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).prepareAgentsEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request, st room.State, roomID string, req eventRequest) (any, bool)`.
  Calls: `connectionHasActiveAgent`, `http.Error`, `id`, `json.Unmarshal`, `len`, `participantHasUnsettledWork`, `room.SupportsAgent`, `s.agentsBlockedLocked`, `strings.TrimSpace`.
