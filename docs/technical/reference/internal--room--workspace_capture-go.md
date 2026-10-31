# internal/room/workspace_capture.go

[Module index](internal.md) · [Open source](../../../internal/room/workspace_capture.go)

Go implementation for workspace capture; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `WorkspaceCaptureReservation`**: WorkspaceCaptureReservation is private controller helper-capacity history. It is not a deliverable, byte-release capability or physical quota observation.
  Declaration: `struct { Version int 'json:"version"' CompanyID string 'json:"companyId"' RoomID string 'json:"roomId"' Worker string 'json:"worker"' ActorID string 'json:"actorId"' Intent toolbroker.CaptureIntent 'json:"intent"' Events []WorkspaceProbeEvent 'json:"events"' Released bool 'json:"released"' }`.
