# internal/toolbroker/object_read_recovery.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/object_read_recovery.go)

Go implementation for object read recovery; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ObjectReadRecoveryJournal`**: ObjectReadRecoveryJournal MUST authenticate cleanup authority from a durable interrupted record pinning request, worker, image, volume and exact created ID. Begin authorizes no creation/read. Finish durably records absence/uncertainty; persistence failure must stop integrity. Storage charges and objects remain.
  Declaration: `interface { BeginObjectReadRecovery(context.Context, objectread.Request, CaptureObjectVolume, string, string) error FinishObjectReadRecovery(context.Context, objectread.Request, string, bool) error }`.
- **Type `ObjectReadRecovery`**: ObjectReadRecovery reconciles one recorded ID. It has no starter, release guard, reader, volume deletion, quota refund or name-adoption capability. There is no host fallback; source substitution of trusted dependencies invalidates isolation.
  Declaration: `struct { physical *objectReadInspector journal ObjectReadRecoveryJournal attempted atomic.Bool }`.
- **`NewObjectReadRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p *ObjectReadContainerPlan, r DockerRunner, j ObjectReadRecoveryJournal) (*ObjectReadRecovery, error)`.
  Calls: `missingProbeDependency`.
- **`(*ObjectReadRecovery).Reconcile`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (err error)`.
  Calls: `StopContainer`, `cancel`, `child.Err`, `context.Background`, `context.WithTimeout`, `ctx.Err`, `digest`, `errors.Join`, `r.attempted.CompareAndSwap`, `r.journal.BeginObjectReadRecovery`, `r.journal.FinishObjectReadRecovery`, `r.physical.absent`, `r.physical.known`, `stop`.
