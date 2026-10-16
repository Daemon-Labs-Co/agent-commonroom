# cmd/server/event_store.go

[Module index](cmd.md) · [Open source](../../cmd/server/event_store.go)

Go implementation for event store; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `auditChange`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Kind string Payload any }`.
- **`(*server).append`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, kind string, payload any) error`.
  Calls: `s.appendBatch`.
- **`(*server).appendBatch`** — appendBatch commits a whole business operation under the existing audit-chain lock. No dispatch may start until it succeeds; failure rolls back every event.
  Signature: `func(ctx context.Context, changes []auditChange) error`.
  Calls: `append`, `cancel`, `checksum`, `ctx.Value`, `err.Error`, `errors.New`, `id`, `json.Marshal`, `json.Unmarshal`, `len`, `protectAuditMetadata`, `s.db.Begin`, `s.storageContext`, `string`, `time.Now`, `time.Now().UTC`, `time.Now().UTC().Truncate`, `tx.Commit`, `tx.Exec`, `tx.QueryRow`, `tx.QueryRow(ctx, "SELECT $1::jsonb::text", string(raw)).Scan`, `tx.QueryRow(ctx, "SELECT count(*) FROM events").Scan`, `tx.QueryRow(ctx, "SELECT hash FROM events ORDER BY sequence DESC LIMIT 1").Scan`, `tx.QueryRow(ctx, "SELECT nextval('public.events_sequence_seq')").Scan`, `tx.Rollback`.
