# internal/toolbroker/audit_idle.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/audit_idle.go)

Go implementation for audit idle; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `IdleStop`**: IdleStop is bounded lifecycle metadata on the private supervisor audit link, never a model tool or proof that containers/source are absent.
  Declaration: `struct { StartedAt time.Time 'json:"startedAt"' RequestedAt time.Time 'json:"requestedAt"' Minutes int 'json:"minutes"' PolicyRevision int 'json:"policyRevision"' }`.
- **Type `IdleAuditor`**: IdleAuditor is an optional controller protocol extension. Success MUST follow durable persistence under registered policy. A no-op source adapter invalidates auditing; this interface grants no executor, disposal or host-run authority.
  Declaration: `interface { RecordIdleStop(context.Context, Scope, IdleStop) error }`.
- **`ValidAuditIdleStop`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(scope Scope, stop IdleStop, now time.Time) bool`.
  Calls: `now.Add`, `stop.RequestedAt.After`, `stop.RequestedAt.Before`, `stop.StartedAt.Add`, `stop.StartedAt.IsZero`, `time.Duration`, `validScope`.
- **`(*AuditClient).RecordIdleStop`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, stop IdleStop) error`.
  Calls: `ValidAuditIdleStop`, `client.exchange`, `client.fail`, `client.mu.TryLock`, `client.mu.Unlock`, `time.Now`.
- **`(*ToolSession).RecordIdleStop`** — RecordIdleStop is supervisor-only and permanently closes tool authority before sending the lifecycle frame. It must share the dispatch slot; active operations cannot be relabelled idle. No automatic retry or audit transport downgrade.
  Signature: `func(ctx context.Context, stop IdleStop) error`.
  Calls: `authority.Revoke`, `s.audit.RecordIdleStop`.
