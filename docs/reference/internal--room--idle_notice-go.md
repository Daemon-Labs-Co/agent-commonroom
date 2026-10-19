# internal/room/idle_notice.go

[Module index](internal.md) · [Open source](../../internal/room/idle_notice.go)

Go implementation for idle notice; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `IdleStopNotice`**: IdleStopNotice records an admitted worker's intention, never shutdown proof. It contains fixed lifecycle metadata only, never model text or credentials.
  Declaration: `struct { Version int 'json:"version"' ID string 'json:"id"' Job ManagedJob 'json:"job"' WorkerStartedAt time.Time 'json:"workerStartedAt"' RequestedAt time.Time 'json:"requestedAt"' RecordedAt time.Time 'json:"recordedAt"' }`.
