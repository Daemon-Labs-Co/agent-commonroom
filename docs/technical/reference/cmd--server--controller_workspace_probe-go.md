# cmd/server/controller_workspace_probe.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_workspace_probe.go)

Go implementation for controller workspace probe; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerWorkspaceProbeEvidence`**: controllerWorkspaceProbeEvidence is independently authenticated physical worker infrastructure, never agent output. ValidateWorkspaceProbe must verify the exact approved worker/creator lease, retained volume, reviewed image and complete reference inventory at the requested observation. It must not reenter server locks. ValidateProbe additionally proves exact lifecycle IDs/absence before the ledger can release a slot. No permissive physical implementation is supplied.
  Declaration: `interface { controllerProbeEvidence ValidateWorkspaceProbe(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error }`.
- **Type `controllerProbeAuthorization`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Version int 'json:"version"' ActorID string 'json:"actorId"' CompanyID string 'json:"companyId"' Worker string 'json:"worker"' Operation string 'json:"operation"' RecordDigest string 'json:"recordDigest"' Intent toolbroker.WorkspaceProbeIntent 'json:"intent"' }`.
- **Type `controllerWorkspaceProbeAuthority`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server lifetime context.Context actor, worker string record workerJobRecord intent toolbroker.WorkspaceProbeIntent evidence controllerWorkspaceProbeEvidence ledger *controllerProbeLedger }`.
- **`probeWorkerRecordDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerJobRecord) string`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `sha256.Sum256`.
- **`detachedProbeWorkerRecord`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r workerJobRecord) workerJobRecord`.
- **`newControllerWorkspaceProbeAuthority`** — The context's human principal and worker identity MUST come from authenticated private session setup, not decoded peer frames. The immutable intent/record are pinned by that setup. Calls reauthorize their principal; a canceled session may not be revived by passing a fresh cleanup context. New cleanup needs a separately authenticated session. This authorizes metadata probes only, not file export.
  Signature: `func(ctx context.Context, s *server, worker string, r workerJobRecord, i toolbroker.WorkspaceProbeIntent, e controllerWorkspaceProbeEvidence) (*controllerWorkspaceProbeAuthority, error)`.
  Calls: `ctx.Err`, `ctx.Value`, `detachedProbeWorkerRecord`, `membershipPartPattern.MatchString`, `newControllerProbeLedger`, `nilWorkerObservationDependency`, `retainedObservationRecord`, `validWorkerProbeIntent`, `workerAliasPattern.MatchString`.
- **`validControllerProbeAuthorization`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v controllerProbeAuthorization) bool`.
  Calls: `artifactRecordedActor`, `probeHex`, `validManagedJob`, `validWorkerProbeIntent`.
- **`projectControllerProbeAuthorization`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `json.Unmarshal`, `validControllerProbeAuthorization`.
- **`(*controllerWorkspaceProbeAuthority).callContext`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (context.Context, context.CancelFunc, error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithCancel`, `context.WithValue`, `ctx.Err`, `g.lifetime.Err`, `stop`.
- **`(*controllerWorkspaceProbeAuthority).current`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, operation string) (room.State, controllerProbeAuthorization, error)`.
  Calls: `artifactActor`, `coordinationJobAllowed`, `ctx.Err`, `delegatedJobAllowed`, `g.lifetime.Err`, `g.server.agentsBlockedLocked`, `g.server.load`, `managedTaskCurrent`, `probeWorkerRecordDigest`, `sameWorkerBinding`, `validControllerProbeAuthorization`.
- **`(*controllerWorkspaceProbeAuthority).authorize`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, operation string) error`.
  Calls: `cancel`, `context.WithValue`, `ctx.Err`, `detachedProbeWorkerRecord`, `errors.Is`, `g.callContext`, `g.current`, `g.evidence.ValidateWorkspaceProbe`, `s.append`, `s.mu.Lock`, `s.mu.Unlock`, `s.storageContext`, `s.tripIntegrityLocked`, `storageCancel`.
- **`(*controllerWorkspaceProbeAuthority).Authorize`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error`.
  Calls: `g.authorize`.
- **`(*controllerWorkspaceProbeAuthority).AuthorizeWorkspaceObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord) error`.
  Calls: `g.authorize`, `probeWorkerRecordDigest`, `retainedObservationRecord`.
- **`(*controllerWorkspaceProbeAuthority).Acknowledge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `cancel`, `g.callContext`, `g.ledger.Acknowledge`.
- **`(*controllerWorkspaceProbeAuthority).ValidateProbe`** — Invoked inside the ledger's fenced authorization snapshot, so a worker change between calls cannot turn a session for one worker into authority over another.
  Signature: `func(ctx context.Context, st room.State, r room.WorkspaceProbeReservation, e workerProbeEvent) error`.
  Calls: `ctx.Err`, `g.evidence.ValidateProbe`, `g.lifetime.Err`, `sameWorkerBinding`.
- **`(*controllerWorkspaceProbeAuthority).StopAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `g.server.mu.Lock`, `g.server.mu.Unlock`, `g.server.tripIntegrityLocked`.
