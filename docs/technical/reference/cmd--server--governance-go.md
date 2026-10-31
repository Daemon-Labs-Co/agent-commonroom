# cmd/server/governance.go

[Module index](cmd.md) · [Open source](../../../cmd/server/governance.go)

Go implementation for governance; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`eligibleGovernor`** — Governance records coordinate work; they are never execution capabilities. Every effect still requires the ordinary authenticated, scoped broker grant.
  Signature: `func(st room.State, rid, pid string) bool`.
  Calls: `room.SupportsAgent`.
- **`projectGovernance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, data []byte) error`.
  Calls: `eligibleGovernor`, `errors.New`, `hasRoom`, `json.Unmarshal`.
- **`(*server).governance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `canAccess`, `eligibleGovernor`, `errors.Is`, `fenceObjectFields`, `hasRoom`, `http.Error`, `http.MaxBytesReader`, `int`, `io.ReadAll`, `json.Unmarshal`, `len`, `operationError`, `prepareOperation`, `r.Context`, `r.Header.Get`, `reply`, `requestActor`, `requestedRoom`, `s.agentsBlockedLocked`, `s.append`, `s.cancelStaleCoordinationLocked`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.validCSRF`, `s.workflowSafety`, `uint`, `withOperationResponse`.
- **`initialGovernanceChanges`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, p room.Participant, c room.Connection) []auditChange`.
  Calls: `append`, `eligibleGovernor`.
- **`(*server).appendAppointmentBatch`** — appendAppointmentBatch requires s.mu, like all registration callers. A failed required appointment audit must stop admission even if no events committed.
  Signature: `func(ctx context.Context, changes []auditChange) error`.
  Calls: `errors.Is`, `s.appendBatch`, `s.tripIntegrityLocked`.
- **Type `coordinationContextKey`**: A role dependency can only narrow an authenticated human's work permission. It is controller context, never a bearer capability supplied by model output.
  Declaration: `struct{}`.
- **Type `coordinationBinding`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Appointment room.GovernanceAppointment Epoch int }`.
- **`currentCoordination`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, appointment room.GovernanceAppointment) bool`.
  Calls: `eligibleGovernor`, `hasRoom`.
- **`coordinationJobAllowed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, job room.ManagedJob) bool`.
  Calls: `currentCoordination`.
- **`coordinationContextAllowed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, p room.Participant) bool`.
  Calls: `ctx.Value`, `currentCoordination`.
- **`(*server).cancelStaleCoordinationLocked`** — Only cancels local role-bound turns. Worker cleanup retains its normal fenced proof protocol; cancellation is never reported as confirmed shutdown. Caller holds s.mu; a zero next value revokes all local coordination turns.
  Signature: `func(next room.GovernanceAppointment)`.
  Calls: `cancel`.
