# internal/toolbroker/workspace_mount_probe.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/workspace_mount_probe.go)

Go implementation for workspace mount probe; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `WorkspaceProbeIntent`**: WorkspaceProbeIntent is trusted worker metadata, never model launch input. The challenge is a fresh non-secret per-observation nonce, not a capability.
  Declaration: `struct { JobID, Repository, Image, Challenge string Volume WorkspaceVolumeIdentity }`.
- **Type `WorkspaceProbeJournal`**: WorkspaceProbeJournal MUST authenticate current authority, hold creator ownership and reserve a physical container slot plus durable creation intent before Begin succeeds. Bound journals the exact ID before execution. Finish durably records observed absence/uncertainty and rechecks current authority before proof release. Failure must trip controller integrity stop as required. Finish receives an empty ID and absent=false when creation identity is unknown. No production journal/authority assembly or default implementation exists.
  Declaration: `interface { Begin(context.Context, WorkspaceProbeIntent) error Bound(context.Context, WorkspaceProbeIntent, string) error Finish(context.Context, WorkspaceProbeIntent, string, bool) error }`.
- **Type `WorkspaceMountProbe`**: WorkspaceMountProbe is a one-shot isolated metadata observer, not a source exporter or FrozenLease. Only the reviewed credential-free scratch image may be supplied. runner MUST target the approved rootless daemon on every call. Source forks CAN provide dishonest runners/journals, BUT invalidate isolation. There is no host execution or mount fallback, and no source bytes are read.
  Declaration: `struct { intent WorkspaceProbeIntent runner DockerRunner journal WorkspaceProbeJournal root string attempted atomic.Bool }`.
- **`NewWorkspaceMountProbe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(intent WorkspaceProbeIntent, runner DockerRunner, journal WorkspaceProbeJournal, dockerRoot string) (*WorkspaceMountProbe, error)`.
  Calls: `ValidWorkspaceVolume`, `digest`, `len`, `missingProbeDependency`, `path.Clean`, `path.IsAbs`, `strings.HasPrefix`, `strings.TrimPrefix`, `workspacemount.ValidChallenge`.
- **`missingProbeDependency`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v any) bool`.
  Calls: `r.IsNil`, `r.Kind`, `reflect.ValueOf`.
- **`(*WorkspaceMountProbe).name`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() string`.
- **`(*WorkspaceMountProbe).labels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]string`.
- **`(*WorkspaceMountProbe).createArgs`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() []string`.
  Calls: `append`, `p.labels`, `p.name`.
- **`(*WorkspaceMountProbe).capture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, max int) ([]byte, error)`.
  Calls: `(&ContainerExecutor{runner: p.runner}).capture`.
- **`(*WorkspaceMountProbe).inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (toolInspection, error)`.
  Calls: `decodeInspection`, `p.capture`, `p.labels`, `p.name`, `sameMap`.
- **`(*WorkspaceMountProbe).verify`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string, exited bool) error`.
  Calls: `decodeInspection`, `isBool`, `isInt`, `len`, `p.capture`, `p.inspect`, `slices.Equal`, `validObservedWorkspaceVolume`.
- **`(*WorkspaceMountProbe).Run`** — Run withholds reports until the helper exits successfully, its exact ID is removed/observed absent, and final durable journal/current-authority checks pass. Lost create replies retain uncertainty; names are never adopted or retried. Cleanup uses a fresh bounded context while the caller must keep creator ownership held. This returns no live mount/root or enduring frozen capability.
  Signature: `func(ctx context.Context) (proof workspacemount.Proof, err error)`.
  Calls: `StopContainer`, `bytes.NewReader`, `bytes.TrimSpace`, `cancel`, `cancelCleanup`, `child.Err`, `context.Background`, `context.WithTimeout`, `ctx.Err`, `decodeInspection`, `decoder.Decode`, `decoder.DisallowUnknownFields`, `digest`, `error`, `errors.Join`, `json.NewDecoder`, `len`, `p.attempted.CompareAndSwap`, `p.capture`, `p.createArgs`, `p.inspect`, `p.journal.Begin`, `p.journal.Bound`, `p.journal.Finish`, `p.name`, `p.verify`, `string`, `strings.TrimSpace`, `workspacemount.ValidProof`.
- **`(*WorkspaceMountProbe).VerifyAdmission`** — VerifyAdmission verifies the exact stopped probe and retained-volume references while the caller holds creator ownership. It never authorizes other attached containers, running processes, source reads or an enduring FrozenLease. The approved daemon/image remain trusted; administrators can race names.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `ctx.Err`, `digest`, `p.capture`, `p.inspect`, `p.name`, `p.verify`, `string`, `strings.TrimSpace`.
- **`(*WorkspaceMountProbe).InspectRecovery`** — InspectRecovery admits only a durably known probe ID, present with exact ownership or already absent, and no other workspace/job/name references. Recovery may STOP an owned running or policy-drifted helper; it never starts it or authorizes bytes. The trusted caller MUST hold creator ownership and obtain durable current cleanup authorization before removal. No name adoption.
  Signature: `func(ctx context.Context, id string) (bool, error)`.
  Calls: `ctx.Err`, `decodeInspection`, `digest`, `p.capture`, `p.inspect`, `p.name`, `string`, `strings.TrimSpace`, `validObservedWorkspaceVolume`.
- **`(*WorkspaceMountProbe).RemoveRecorded`** — RemoveRecorded performs only exact-ID reconciliation, never create/start or volume deletion. Callers must durably authorize recovery first. Absence is reobserved with exact retained volume/reference checks before success.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `StopContainer`, `p.InspectRecovery`, `p.inspect`.
