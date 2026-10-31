# internal/room/object_read.go

[Module index](internal.md) · [Open source](../../../internal/room/object_read.go)

Go implementation for object read; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ObjectReadReservation`**: ObjectReadReservation is private controller capacity history, not a byte-release capability, physical observation, delivered artifact or task acceptance.
  Declaration: `struct { Version int 'json:"version"' CompanyID string 'json:"companyId"' RoomID string 'json:"roomId"' Worker string 'json:"worker"' ActorID string 'json:"actorId"' Intent toolbroker.ObjectReadIntent 'json:"intent"' Events []WorkspaceProbeEvent 'json:"events"' Released bool 'json:"released"' }`.
