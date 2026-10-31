# internal/room/artifact_custody.go

[Module index](internal.md) · [Open source](../../../internal/room/artifact_custody.go)

Go implementation for artifact custody; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ArtifactCustodyObject`**: ArtifactCustodyObject retains the full reservation even when physical setup, sealing or release is uncertain. These records never authorize file delivery by themselves. There is no refund, replacement, deletion or adoption transition.
  Declaration: `struct { Version int 'json:"version"' Intent objectcustody.Intent 'json:"intent"' ReservationID string 'json:"reservationId"' QuotaEstablished bool 'json:"quotaEstablished"' Receipt *objectcustody.Receipt 'json:"receipt,omitempty"' ReadAcknowledgments int64 'json:"readAcknowledgments"' }`.
