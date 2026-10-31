# internal/workspacecapture/frozen.go

[Module index](internal.md) · [Open source](../../../internal/workspacecapture/frozen.go)

Go implementation for frozen; declarations and calls are listed below.

Package: `workspacecapture`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Authority`**: Authority is the authenticated supervisor, not an agent or the kernel mount report alone. Every call MUST bind the exact job/creator/import/volume, reviewed running helper and current principal/task/export policy, hold creator ownership, verify complete writer exclusion and existing workspace storage accounting, and durably authorize bounded observation before success. Proof is local kernel evidence, not independent remote attestation or source release permission. Implementations must honor cancellation and must not reenter the same source. No production transport or permissive default authority is provided here.
  Declaration: `interface { AuthorizeFrozenSource(context.Context, workspacesnapshot.Binding, workspacemount.Proof) error }`.
- **Type `containerView`**: containerView is private: production always uses the fixed /workspace view and actual kernel checks. Tests may explicitly mock it against synthetic fixtures; external callers cannot choose a different root or mount observer.
  Declaration: `interface { OpenRoot() (*os.Root, error) RootIdentity() (os.FileInfo, error) Observe(context.Context, string) (workspacemount.Proof, error) }`.
- **Type `fixedContainerView`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(fixedContainerView).OpenRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (*os.Root, error)`.
  Calls: `os.OpenRoot`.
- **`(fixedContainerView).RootIdentity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (os.FileInfo, error)`.
  Calls: `os.Lstat`.
- **`(fixedContainerView).Observe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, challenge string) (workspacemount.Proof, error)`.
  Calls: `workspacemount.Observe`.
- **Type `frozenLease`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { mu sync.Mutex lifetime context.Context root *os.Root view containerView binding workspacesnapshot.Binding challenge string authority Authority proof workspacemount.Proof }`.
- **`missingAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a Authority) bool`.
  Calls: `isNilDependency`.
- **`isNilDependency`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value any) bool`.
  Calls: `reflect.ValueOf`, `v.IsNil`, `v.Kind`.
- **`(*frozenLease).VerifyFrozen`** — VerifyFrozen combines live supervisor fencing with actual fixed kernel mount evidence and pinned root identity on EVERY source verification/read. A stopped probe report cannot construct this lease or substitute for its running view.
  Signature: `func(ctx context.Context, root *os.Root, b workspacesnapshot.Binding) error`.
  Calls: `cancel`, `child.Err`, `context.AfterFunc`, `context.WithDeadline`, `ctx.Err`, `fixed.IsDir`, `fixed.Mode`, `l.authority.AuthorizeFrozenSource`, `l.lifetime.Deadline`, `l.lifetime.Err`, `l.mu.Lock`, `l.mu.Unlock`, `l.root.Lstat`, `l.view.Observe`, `l.view.RootIdentity`, `os.SameFile`, `stop`, `workspacemount.ValidProof`.
- **`WithFrozenSource`** — WithFrozenSource opens only the fixed isolated /workspace root, captures actual finite source bytes and lends the Source synchronously. No host path, mounting, provider invocation or fallback exists. Each lazy read still requires live authority; custody MUST reserve and enforce physical quota before consuming it. The lease expires at callback exit, error, cancellation or the lifetime ceiling. A retained Source cannot revive with a fresh context. No file is persisted here.
  Signature: `func(ctx context.Context, b workspacesnapshot.Binding, members []workspacesnapshot.Member, challenge string, authority Authority, use func(*workspacesnapshot.Source) error) error`.
  Calls: `withFrozenSource`.
- **`withFrozenSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, view containerView, b workspacesnapshot.Binding, members []workspacesnapshot.Member, challenge string, authority Authority, use func(*workspacesnapshot.Source) error) error`.
  Calls: `append`, `cancel`, `context.WithTimeout`, `ctx.Err`, `missingAuthority`, `root.Close`, `source.Current`, `use`, `view.OpenRoot`, `workspacemount.ValidChallenge`, `workspacesnapshot.Capture`.
