# internal/projectquota/set.go

[Module index](internal.md) · [Open source](../../../internal/projectquota/set.go)

Go implementation for set; declarations and calls are listed below.

Package: `projectquota`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Root`**: Root binds an already opened directory to an independently approved reserved capacity. Neither the first observed identity nor model metadata is policy.
  Declaration: `struct { Directory *os.File Expected Expected }`.
- **`InspectSet`** — InspectSet checks a bounded set of projects against a finite aggregate reserved budget. All root descriptors remain pinned until the entire observation joins. Callers MUST independently fence all writers and prove inventory completeness: matching repeated observations alone do not make an atomic snapshot, establish account ownership or authorize allocation/export. No filesystem is modified.
  Signature: `func(ctx context.Context, roots []Root, budget uint64) ([]Observation, error)`.
  Calls: `inspectSet`.
- **`inspectSet`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, roots []Root, budget uint64, open func(*os.File) (kernelReader, error)) (observations []Observation, err error)`.
  Calls: `append`, `ctx.Err`, `inspect`, `len`, `make`, `open`, `readers[i].Close`, `refuse`, `uint64`, `validExpected`.
