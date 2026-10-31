# cmd/server/worker_workspace_probe.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_workspace_probe.go)

Go implementation for worker workspace probe; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `workerWorkspaceProbeGuard`**: workerWorkspaceProbeGuard binds trusted current controller authorization, durable observation/audit and account-wide admission. No production transport or permissive implementation exists. A model cannot supply this authority.
  Declaration: `interface { workerWorkspaceObservationGuard workerProbeGuard }`.
- **Type `guardedWorkspaceProbe`**: guardedWorkspaceProbe is confined to the synchronous creator-lock callback. Its journal methods track the one exact helper ID; no handle escapes to callers.
  Declaration: `struct { active bool intent toolbroker.WorkspaceProbeIntent id string lease *workerWorkspaceObservation probe *toolbroker.WorkspaceMountProbe journal *workerProbeJournal guard workerWorkspaceProbeGuard }`.
- **`(*guardedWorkspaceProbe).Authorize`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error`.
  Calls: `g.guard.Authorize`, `g.lease.checkProbe`.
- **`(*guardedWorkspaceProbe).Acknowledge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `g.guard.Acknowledge`.
- **`(*guardedWorkspaceProbe).StopAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `g.guard.StopAuthority`.
- **`(*guardedWorkspaceProbe).Begin`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error`.
  Calls: `g.journal.Begin`.
- **`(*guardedWorkspaceProbe).Bound`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string) error`.
  Calls: `g.journal.Bound`, `g.probe.VerifyAdmission`.
- **`(*guardedWorkspaceProbe).Finish`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, id string, absent bool) error`.
  Calls: `g.journal.Finish`, `probeHex`.
- **`probeRetainedWorkerWorkspace`** — probeRetainedWorkerWorkspace assembles creator ownership, exact retained workspace admission, durable local journal and isolated read-only helper. Caller supplies only trusted reviewed image/daemon/controller infrastructure. No public route, source bytes, host mount/fallback or FrozenLease is returned. A source fork CAN substitute dishonest guards/runners, BUT loses these claims.
  Signature: `func(ctx context.Context, dir string, record workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, image string, guard workerWorkspaceProbeGuard) (workspacemount.Proof, error)`.
  Calls: `probeRetainedWorkerWorkspaceAuthority`.
- **`probeRetainedWorkerWorkspaceChannel`** — probeRetainedWorkerWorkspaceChannel binds a fresh challenge to a private channel before observations. Creator ownership covers all guarded actions. connect MUST authenticate the approved peer and construct its controller guard independently; it must never use model streams. This is an internal assembly seam, not an enabled SSH/public export endpoint.
  Signature: `func(ctx context.Context, dir string, record workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, image string, connect func(context.Context, workerJobRecord, toolbroker.WorkspaceProbeIntent) (io.ReadWriteCloser, error), stop func()) (workspacemount.Proof, error)`.
  Calls: `connect`, `newProbeAuthorityClient`, `nilWorkerObservationDependency`, `probeRetainedWorkerWorkspaceAuthority`, `stream.Close`.
- **Type `pendingWorkspaceProbeGuard`**: pendingWorkspaceProbeGuard refuses until trusted setup completes under actual creator ownership. It never supplies default authorization while connecting.
  Declaration: `struct { guard workerWorkspaceObservationGuard }`.
- **`(*pendingWorkspaceProbeGuard).AuthorizeWorkspaceObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord) error`.
  Calls: `g.guard.AuthorizeWorkspaceObservation`, `nilWorkerObservationDependency`.
- **`probeRetainedWorkerWorkspaceAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, dir string, record workerJobRecord, dockerRoot string, runner toolbroker.DockerRunner, image string, authority func(*workerWorkspaceObservation, toolbroker.WorkspaceProbeIntent) (workerWorkspaceProbeGuard, func(), error)) (workspacemount.Proof, error)`.
  Calls: `authority`, `closeAuthority`, `ctx.Err`, `errors.New`, `hex.EncodeToString`, `lease.Check`, `make`, `newWorkerProbeJournal`, `nilWorkerObservationDependency`, `probe.Run`, `rand.Read`, `retainedObservationRecord`, `toolbroker.NewWorkspaceMountProbe`, `validWorkerProbeIntent`, `withWorkerWorkspaceOwnership`.
