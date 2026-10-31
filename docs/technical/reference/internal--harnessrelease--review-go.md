# internal/harnessrelease/review.go

[Module index](internal.md) · [Open source](../../../internal/harnessrelease/review.go)

Go implementation for review; declarations and calls are listed below.

Package: `harnessrelease`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Review`**: Only trusted deployment administrators supply reviews and pin signing keys. This source-level seam CAN be weakened by forks, BUT then image approval is untrusted. A signature proves review authority, not image safety or provenance.
  Declaration: `struct { Schema int 'json:"schema"' Sequence int64 'json:"sequence"' Provider string 'json:"provider"' Version string 'json:"version"' Image string 'json:"image"' Profile string 'json:"profile"' AuthProtocol int 'json:"authProtocol"' AuditProtocol int 'json:"auditProtocol"' WorkspaceProtocol int 'json:"workspaceProtocol"' IssuedAt time.Time 'json:"issuedAt"' ExpiresAt time.Time 'json:"expiresAt"' }`.
- **Type `Envelope`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Review json.RawMessage 'json:"review"' Signature string 'json:"signature"' }`.
- **`canonicalJSON`** — Canonical bytes remove ambiguous duplicate, unknown and case-folded JSON fields. The signed payload must be exactly json.Marshal(Review).
  Signature: `func(raw []byte, value any) bool`.
  Calls: `bytes.Equal`, `json.Marshal`, `json.Unmarshal`, `len`.
- **`Verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, key ed25519.PublicKey, now time.Time) (room.HarnessSelection, error)`.
  Calls: `ValidSelection`, `base64.StdEncoding.Strict`, `base64.StdEncoding.Strict().DecodeString`, `canonicalJSON`, `ed25519.Verify`, `hex.EncodeToString`, `len`, `now.UTC`, `review.ExpiresAt.After`, `review.IssuedAt.After`, `review.IssuedAt.IsZero`, `sha256.Sum256`.
- **`ValidSelection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v room.HarnessSelection) bool`.
  Calls: `digestPattern.MatchString`, `imagePattern.MatchString`, `len`, `v.CheckedAt.IsZero`, `v.ExpiresAt.After`, `versionPattern.MatchString`.
