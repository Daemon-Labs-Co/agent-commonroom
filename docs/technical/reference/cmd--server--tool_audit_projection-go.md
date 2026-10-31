# cmd/server/tool_audit_projection.go

[Module index](cmd.md) · [Open source](../../../cmd/server/tool_audit_projection.go)

Go implementation for tool audit projection; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`toolAuditKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(job, call string) string`.
- **`toolAuditIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(call room.ManagedToolCall) toolbroker.Intent`.
- **`projectToolAudit`** — projectToolAudit checks historical transitions, not current access policy. Permission changes must not invalidate previously valid audit evidence.
  Signature: `func(st *room.State, kind string, raw []byte) error`.
  Calls: `call.RecordedAt.Equal`, `call.RecordedAt.IsZero`, `call.UpdatedAt.Before`, `errors.New`, `json.Unmarshal`, `toolAuditIntent`, `toolAuditKey`, `toolbroker.ValidAuditIntent`, `toolbroker.ValidAuditOutcome`.
