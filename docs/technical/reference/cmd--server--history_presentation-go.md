# cmd/server/history_presentation.go

[Module index](cmd.md) · [Open source](../../../cmd/server/history_presentation.go)

Go implementation for history presentation; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).historyPresentationEnabled`** — historyPresentationEnabled reads only the latest setting, not the full history. Missing means on. Read/decoding failures stop the enriched history response.
  Signature: `func(ctx context.Context) (bool, error)`.
  Calls: `cancel`, `ctx.Err`, `json.Unmarshal`, `rows.Close`, `rows.Err`, `rows.Next`, `rows.Scan`, `s.db.Query`, `s.storageContext`.
- **Type `auditPresentation`**: auditPresentation contains allowlisted human-readable metadata, never a raw event payload, provider stderr, message body, endpoint or credential.
  Declaration: `struct { Category string 'json:"category"' Tone string 'json:"tone"' Routine bool 'json:"routine"' Actor string 'json:"actor,omitempty"' ActorID string 'json:"actorId,omitempty"' Subject string 'json:"subject,omitempty"' RoomID string 'json:"roomId,omitempty"' Outcome string 'json:"outcome,omitempty"' Explanation string 'json:"explanation,omitempty"' Next string 'json:"next,omitempty"' }`.
- **`auditSafeLabel`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value string) string`.
  Calls: `len`, `string`, `strings.Map`, `strings.TrimSpace`, `unicode.IsControl`.
- **`describeAuditEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(kind string, raw []byte) auditPresentation`.
  Calls: `auditSafeLabel`, `json.Unmarshal`, `label`, `safeAgentFailure`, `strings.HasPrefix`, `strings.ToLower`.
