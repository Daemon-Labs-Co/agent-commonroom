# cmd/server/audit_metadata.go

[Module index](cmd.md) · [Open source](../../../cmd/server/audit_metadata.go)

Go implementation for audit metadata; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `auditRecordMetadata`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' ID string 'json:"id"' Sequence int64 'json:"sequence"' CommittedAt time.Time 'json:"committedAt"' }`.
- **`protectAuditMetadata`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, meta auditRecordMetadata) ([]byte, error)`.
  Calls: `errors.New`, `json.Marshal`, `json.Unmarshal`.
- **`verifyAuditMetadata`** — New records bind identity/order/time inside their existing canonical payload checksum. Legacy records stay byte-for-byte unchanged and explicitly lack this protection. Chain verification forbids legacy records after this version appears. Neither version alone can detect prefix rollback.
  Signature: `func(payload string, id string, sequence int64, at time.Time) (protected, valid bool)`.
  Calls: `json.Unmarshal`, `meta.CommittedAt.Equal`, `meta.CommittedAt.IsZero`.
