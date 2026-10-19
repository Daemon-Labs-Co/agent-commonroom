# cmd/server/harness_review.go

[Module index](cmd.md) · [Open source](../../cmd/server/harness_review.go)

Go implementation for harness review; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `harnessReview`**: Only trusted deployment administrators supply reviews and pin signing keys. This source-level seam CAN be weakened by forks, BUT then image approval is untrusted. A signature proves review authority, not image safety or provenance.
  Declaration: `struct { Schema int 'json:"schema"' Sequence int64 'json:"sequence"' Provider string 'json:"provider"' Version string 'json:"version"' Image string 'json:"image"' Profile string 'json:"profile"' AuthProtocol int 'json:"authProtocol"' AuditProtocol int 'json:"auditProtocol"' WorkspaceProtocol int 'json:"workspaceProtocol"' IssuedAt time.Time 'json:"issuedAt"' ExpiresAt time.Time 'json:"expiresAt"' }`.
- **Type `harnessReviewEnvelope`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Review json.RawMessage 'json:"review"' Signature string 'json:"signature"' }`.
- **`canonicalHarnessJSON`** — Canonical bytes remove ambiguous duplicate, unknown and case-folded JSON fields. The signed payload must be exactly json.Marshal(harnessReview).
  Signature: `func(raw []byte, value any) bool`.
  Calls: `bytes.Equal`, `json.Marshal`, `json.Unmarshal`, `len`.
- **`verifyHarnessReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, key ed25519.PublicKey, now time.Time) (room.HarnessSelection, error)`.
  Calls: `base64.StdEncoding.Strict`, `base64.StdEncoding.Strict().DecodeString`, `canonicalHarnessJSON`, `ed25519.Verify`, `hex.EncodeToString`, `len`, `now.UTC`, `review.ExpiresAt.After`, `review.IssuedAt.After`, `review.IssuedAt.IsZero`, `sha256.Sum256`, `validHarnessSelection`.
- **`validHarnessSelection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v room.HarnessSelection) bool`.
  Calls: `harnessImagePattern.MatchString`, `harnessVersionPattern.MatchString`, `len`, `repositoryDigestPattern.MatchString`, `v.CheckedAt.IsZero`, `v.ExpiresAt.After`.
- **`validManagedHarness`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(job room.ManagedJob) bool`.
  Calls: `job.Harness.CheckedAt.After`, `job.Harness.CheckedAt.Before`, `validHarnessSelection`.
- **`selectHarnessReview`** — Selection runs only on new authorized allocations. Off leaves the existing explicit immutable-image contract intact; it never enables public dispatch.
  Signature: `func(enabled string, source func() ([]byte, error), key ed25519.PublicKey, now time.Time, jobs map[string]room.ManagedJob) (room.HarnessSelection, error)`.
  Calls: `source`, `verifyHarnessReview`.
- **`deploymentHarnessReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(now time.Time, jobs map[string]room.ManagedJob) (room.HarnessSelection, error)`.
  Calls: `base64.StdEncoding.Strict`, `base64.StdEncoding.Strict().DecodeString`, `ed25519.PublicKey`, `filepath.IsAbs`, `os.Getenv`, `readPrivateBounded`, `selectHarnessReview`.
- **`selectedHarnessImage`** — The allocation is trusted authenticated controller input, never model input. Recheck expiry only before startup; a running engine is never upgraded here.
  Signature: `func(job room.ManagedJob, configured string, now time.Time) (string, error)`.
  Calls: `job.Harness.ExpiresAt.After`, `now.Before`, `validManagedHarness`.
