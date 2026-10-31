# cmd/server/execution_limits.go

[Module index](cmd.md) · [Open source](../../../cmd/server/execution_limits.go)

Go implementation for execution limits; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).agentAdmission`** — agentAdmission requires s.mu and the complete state, never a room projection. This mandatory resource boundary covers chat and workflow model turns on a single controller. It is not a durable queue or a multi-controller quota.
  Signature: `func(st room.State, rid string, requested int) error`.
  Calls: `errors.New`, `s.admitDailyTurns`, `time.Now`.
