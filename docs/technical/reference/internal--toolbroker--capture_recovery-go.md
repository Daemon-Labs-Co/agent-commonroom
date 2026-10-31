# internal/toolbroker/capture_recovery.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/capture_recovery.go)

Go implementation for capture recovery; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `CaptureRecoveryJournal`**: CaptureRecoveryJournal MUST authenticate current cleanup authority against a durable interrupted capture record, its exact bound ID/request/image/volumes, and exclusive creator ownership. Begin never reserves new creation or exports bytes. Finish durably records observed absence/uncertainty under current policy; failed persistence must trip integrity stop. Storage charges/objects remain. No model-facing, permissive or production default implementation exists.
  Declaration: `interface { BeginCaptureRecovery(context.Context, workspacecapture.Request, CaptureObjectVolume, string, string) error FinishCaptureRecovery(context.Context, workspacecapture.Request, string, bool) error }`.
- **Type `CaptureRecovery`**: CaptureRecovery reconciles a durably known helper ID only. It has no starter, source reader, object-store writer, release, adoption or volume deletion API. Source forks may fake trusted evidence, invalidating containment guarantees. Mandatory isolation has no toggle or host/SSH fallback here.
  Declaration: `struct { physical *captureInspector journal CaptureRecoveryJournal attempted atomic.Bool }`.
- **`NewCaptureRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(plan *CaptureContainerPlan, runner DockerRunner, journal CaptureRecoveryJournal) (*CaptureRecovery, error)`.
  Calls: `missingProbeDependency`.
- **`(*CaptureRecovery).request`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() workspacecapture.Request`.
  Calls: `slices.Clone`.
- **`(*CaptureRecovery).Reconcile`** — Reconcile authorizes before any daemon call, refuses unknown/replaced/conflicting identities, and records uncertainty on failure. A nil result proves only observed exact helper absence and a successful durable acknowledgment, not released quota, source validity, complete objects or a recovered/publishable capture report.
  Signature: `func(ctx context.Context, id string) (err error)`.
  Calls: `StopContainer`, `cancel`, `child.Err`, `context.Background`, `context.WithTimeout`, `ctx.Err`, `digest`, `errors.Join`, `r.attempted.CompareAndSwap`, `r.journal.BeginCaptureRecovery`, `r.journal.FinishCaptureRecovery`, `r.physical.absent`, `r.physical.known`, `r.request`, `stop`.
