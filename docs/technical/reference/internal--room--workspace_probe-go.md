# internal/room/workspace_probe.go

[Module index](internal.md) · [Open source](../../../internal/room/workspace_probe.go)

Go implementation for workspace probe; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `WorkspaceProbeEvent`**: WorkspaceProbeReservation is private controller audit/capacity state, never source-read authority. A worker's local journal remains independently required.
  Declaration: `struct { Stage string 'json:"stage"' ID string 'json:"id"' Absent bool 'json:"absent"' }`.
- **Type `WorkspaceProbeReservation`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' CompanyID string 'json:"companyId"' RoomID string 'json:"roomId"' Worker string 'json:"worker"' ActorID string 'json:"actorId"' Intent toolbroker.WorkspaceProbeIntent 'json:"intent"' Events []WorkspaceProbeEvent 'json:"events"' Released bool 'json:"released"' }`.
