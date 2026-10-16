# cmd/server/integrity_handler.go

[Module index](cmd.md) · [Open source](../../cmd/server/integrity_handler.go)

Go implementation for integrity handler; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).integrity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `s.integrityLocked`, `s.mu.Lock`, `s.mu.Unlock`, `s.recoverIntegrityLocked`.
- **`(*server).integrityLocked`** — integrityLocked requires s.mu; every verification failure stops execution, including failures that prevent the incident from being committed.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `cancel`, `checksum`, `ctx.Err`, `http.Error`, `r.Context`, `r.WithContext`, `reply`, `rows.Close`, `rows.Err`, `rows.Next`, `rows.Scan`, `s.append`, `s.db.Query`, `s.load`, `s.storageContext`, `s.tripIntegrityLocked`, `verifyAuditMetadata`.
