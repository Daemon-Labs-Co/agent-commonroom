# cmd/server/state_projection.go

[Module index](cmd.md) · [Open source](../../cmd/server/state_projection.go)

Go implementation for state projection; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).load`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (room.State, error)`.
  Calls: `append`, `cancel`, `ctx.Err`, `delete`, `errors.New`, `fmt.Errorf`, `intent.Scope.Deadline.Add`, `json.Unmarshal`, `json.Valid`, `len`, `projectActors`, `projectAgentElapsed`, `projectAgentStatusDetails`, `projectBackground`, `projectHumanAttention`, `projectManagedStorage`, `projectToolAudit`, `providerForEndpoint`, `room.SupportsAgent`, `rows.Close`, `rows.Err`, `rows.Next`, `rows.Scan`, `s.db.Query`, `s.storageContext`, `time.Now`, `time.Now().UTC`, `toolbroker.ValidAuditIntent`, `toolbroker.ValidAuditOutcome`, `validManagedJob`.
