# cmd/server/harness_review.go

[Module index](cmd.md) · [Open source](../../cmd/server/harness_review.go)

Go implementation for harness review; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `harnessReview`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `harnessrelease.Review`.
- **Type `harnessReviewEnvelope`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `harnessrelease.Envelope`.
- **`verifyHarnessReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, key ed25519.PublicKey, now time.Time) (room.HarnessSelection, error)`.
  Calls: `harnessrelease.Verify`.
- **`validHarnessSelection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v room.HarnessSelection) bool`.
  Calls: `harnessrelease.ValidSelection`.
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
