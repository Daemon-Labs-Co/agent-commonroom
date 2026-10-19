# internal/toolbroker/container_plan.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/container_plan.go)

Go implementation for container plan; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ToolContainerPlan`**: ToolContainerPlan is trusted-supervisor policy, never a model request. It derives identities from the scoped job and exposes no caller-supplied mounts, network, environment, container names or Docker flags. It only generates arguments; there is no Docker launch, filesystem executor or host fallback. A source fork CAN add engine/host mounts, BUT invalidates credential separation.
  Declaration: `struct { authority *Authority image string workspace atomic.Pointer[workspaceReadiness] volume atomic.Pointer[WorkspaceVolumeIdentity] commandsEnabled bool commandUsed atomic.Bool }`.
- **`NewToolContainerPlan`** — NewToolContainerPlan requires an approved immutable image from private worker configuration. Pinning verifies spelling, not provenance or image contents: the operator must separately review/sign/scan the tool-only image. No provider credentials may be baked into it. There is no implicit image or tag fallback.
  Signature: `func(authority *Authority, image string, now time.Time) (*ToolContainerPlan, error)`.
  Calls: `NewToolContainerPlanWithCommands`.
- **`NewToolContainerPlanWithCommands`** — Commands are explicit trusted worker policy; this never grants authority.
  Signature: `func(authority *Authority, image string, now time.Time, enabled bool) (*ToolContainerPlan, error)`.
  Calls: `authority.mu.Lock`, `authority.mu.Unlock`, `now.Before`, `validScope`, `validToolImage`.
- **`validToolImage`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(image string) bool`.
  Calls: `digest`, `len`, `strings.HasPrefix`, `strings.Split`, `strings.TrimPrefix`, `toolImageRepository.MatchString`.
- **`(*ToolContainerPlan).ContainerName`** — ContainerName and WorkspaceVolume are unique job identities, not reusable connection homes. Before creation the future supervisor must prove the job is fresh and refuse pre-existing/unresolved containers or volumes.
  Signature: `func() string`.
- **`(*ToolContainerPlan).WorkspaceVolume`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() string`.
- **`(*ToolContainerPlan).RunArgs`** — RunArgs checks revocation/lifetime and returns a fresh slice. Planning is not an atomic Docker launch; the future fenced supervisor must revalidate authority. The idle container uses /workspace so the named volume stays empty until credential-free Bootstrap initializes its repository before engine startup. No host bundle/auth/socket is mounted. The fixed tool-only image must provide Node at the declared entrypoint; holding the process alive does not authorize any tool execution or establish readiness.
  Signature: `func(now time.Time) ([]string, error)`.
  Calls: `now.Before`, `p.ContainerName`, `p.WorkspaceVolume`, `p.authority.mu.Lock`, `p.authority.mu.Unlock`.
- **`(*ToolContainerPlan).RemoveArgs`** — RemoveArgs is scoped cleanup, not proof of shutdown. The future supervisor must inspect absence and release both-container leases before settling a pair.
  Signature: `func() []string`.
  Calls: `p.ContainerName`.
- **`(*ToolContainerPlan).CreateArgs`** — CreateArgs generates a stopped container with the same mandatory policy. Explicit immutable-ID cleanup replaces automatic removal so crash evidence is retained. Callers must not start it until inspection and pair journalling pass.
  Signature: `func(now time.Time) ([]string, error)`.
  Calls: `append`, `p.RunArgs`.
- **Type `workspaceReadiness`**: Readiness is issued only by successful bootstrap and retains its original job. A fresh per-call context must never revive a cancelled workspace authority.
  Declaration: `struct { lifetime context.Context original context.Context containerID string }`.
- **`(*ToolContainerPlan).workspaceReady`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() bool`.
  Calls: `digest`, `p.workspace.Load`, `ready.active`.
- **`(*workspaceReadiness).active`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() bool`.
  Calls: `ready.lifetime.Err`, `ready.original.Err`.
