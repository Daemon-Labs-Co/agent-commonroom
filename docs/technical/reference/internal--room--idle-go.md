# internal/room/idle.go

[Module index](internal.md) · [Open source](../../../internal/room/idle.go)

Go implementation for idle; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `AgentIdleSettings`**: AgentIdleSettings is owner-controlled policy for newly registered paired jobs. A zero-value settings record means the five-minute default, revision zero.
  Declaration: `struct { Revision int 'json:"revision"' Minutes int 'json:"minutes"' }`.
