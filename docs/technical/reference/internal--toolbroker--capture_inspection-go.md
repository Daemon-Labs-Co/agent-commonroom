# internal/toolbroker/capture_inspection.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/capture_inspection.go)

Go implementation for capture inspection; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureInspector`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { plan *CaptureContainerPlan runner DockerRunner }`.
- **`(*captureInspector).capture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args ...string) ([]byte, error)`.
  Calls: `(&ContainerExecutor{runner: r.runner}).capture`.
- **`(*captureInspector).owned`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `decodeInspection`, `r.capture`, `r.plan.labels`, `r.plan.name`, `sameMap`.
- **`(*captureInspector).absent`** — absent checks identities even after removal, and rejects any retained references.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `r.known`.
- **`(*captureInspector).known`** — known admits only the recorded immutable ID, never name discovery/adoption. A present owned container may have drifted policy and may only be stopped.
  Signature: `func(ctx context.Context, id string) (bool, error)`.
  Calls: `ctx.Err`, `r.capture`, `r.owned`, `r.plan.name`, `r.plan.verifyVolumes`, `string`, `strings.TrimSpace`.
- **`(*CaptureContainerPlan).VerifyNoReferences`** — VerifyNoReferences admits only both retained volumes with no helper/job/name/ volume references. It never discovers/adopts an ID, starts or removes anything. The caller must hold creator ownership and separately authorize the operation.
  Signature: `func(ctx context.Context, runner DockerRunner) error`.
  Calls: `(&captureInspector{p, runner}).absent`, `cancel`, `context.WithTimeout`, `ctx.Err`, `missingProbeDependency`.
- **`(*CaptureContainerPlan).InspectRecorded`** — InspectRecorded checks only a separately durably recorded immutable helper ID. A present owned policy-drifted helper is eligible for cleanup only, never source reads/execution. No deletion, release or name adoption occurs in this observer.
  Signature: `func(ctx context.Context, runner DockerRunner, id string) (bool, error)`.
  Calls: `(&captureInspector{p, runner}).known`, `cancel`, `context.WithTimeout`, `ctx.Err`, `digest`, `missingProbeDependency`.
