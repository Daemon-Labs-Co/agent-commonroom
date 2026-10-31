# internal/toolbroker/object_read_inspection.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/object_read_inspection.go)

Go implementation for object read inspection; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadInspector`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { plan *ObjectReadContainerPlan runner DockerRunner }`.
- **`(*objectReadInspector).capture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args ...string) ([]byte, error)`.
  Calls: `(&ContainerExecutor{runner: r.runner}).capture`.
- **`(*objectReadInspector).owned`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `decodeInspection`, `r.capture`, `r.plan.labels`, `r.plan.name`, `sameMap`.
- **`(*objectReadInspector).known`** — known accepts a durably recorded immutable ID only. Policy drift permits owned cleanup, never reads. Replacements/conflicts refuse; no ID discovery or adoption.
  Signature: `func(ctx context.Context, id string) (bool, error)`.
  Calls: `ctx.Err`, `r.capture`, `r.owned`, `r.plan.name`, `r.plan.verifyVolume`, `string`, `strings.TrimSpace`.
- **`(*objectReadInspector).absent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `r.known`.
- **`(*ObjectReadContainerPlan).VerifyNoReferences`** — VerifyNoReferences requires the retained exact custody volume without any job, helper-name or volume references. It grants neither creation nor release rights.
  Signature: `func(ctx context.Context, runner DockerRunner) error`.
  Calls: `(&objectReadInspector{p, runner}).absent`, `cancel`, `context.WithTimeout`, `ctx.Err`, `missingProbeDependency`.
- **`(*ObjectReadContainerPlan).InspectRecorded`** — InspectRecorded checks a separately durably bound immutable reader ID. It may observe owned policy drift for cleanup, never authorizes reads, discovers IDs or performs daemon mutations. Caller must hold actual creator/cleanup authority.
  Signature: `func(ctx context.Context, runner DockerRunner, id string) (bool, error)`.
  Calls: `(&objectReadInspector{p, runner}).known`, `cancel`, `context.WithTimeout`, `ctx.Err`, `digest`, `missingProbeDependency`.
