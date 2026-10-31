# internal/toolbroker/delegation.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/delegation.go)

Go implementation for delegation; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `DelegationRequest`**: DelegationRequest contains no claimed agent identity, resources or permissions. Identity and scope come exclusively from the authenticated supervisor link.
  Declaration: `struct { // Supervisor-owned snapshot, overwritten from the local Authority by // ToolSession.ProposeDelegation. Never copy this from model input. ParentOperations uint8 'json:"parentOperations"' RequestID string 'json:"requestId"' TaskID string 'json:"taskId"' TaskRevision int 'json:"taskRevision"' DelegateID string 'json:"delegateId"' }`.
- **Type `DelegationAuditor`**: Behavior contract; required method signatures are shown in the declaration.
  Declaration: `interface { RecordDelegation(context.Context, Scope, DelegationRequest) error }`.
- **`ValidDelegation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r DelegationRequest) bool`.
  Calls: `identifier`.
- **`(*AuditClient).ProposeDelegation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r DelegationRequest) error`.
  Calls: `ValidDelegation`, `client.exchange`, `client.fail`, `client.mu.TryLock`, `client.mu.Unlock`, `time.Now`, `validScope`.
- **`(*ToolSession).ProposeDelegation`** — ProposeDelegation shares the supervisor's active-work slot and authenticated control link. It does not mint capabilities, execute work or revoke cleanup. Provider adapters must explicitly translate a supported typed proposal here; model stdout and claims of identity must never substitute for this channel.
  Signature: `func(ctx context.Context, r DelegationRequest) error`.
  Calls: `OperationMask`, `authority.mu.Lock`, `authority.mu.Unlock`, `s.audit.ProposeDelegation`, `time.Now`, `time.Now().Before`.
- **`OperationMask`** — OperationMask is a fixed policy representation, never an executable command.
  Signature: `func(operation string) uint8`.
