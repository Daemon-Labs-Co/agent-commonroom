# internal/workspacemount/observe_linux.go

[Module index](internal.md) · [Open source](../../../internal/workspacemount/observe_linux.go)

Go implementation for observe linux; declarations and calls are listed below.

Package: `workspacemount`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`Observe`** — Observe reads only fixed kernel metadata and directory identity, never source bytes. Read-only statfs flags are checked independently from mountinfo. The approved container must run as UID/GID1000 with a read-only rootfs and no other workspace writers. No host path, fallback or mount command is accepted.
  Signature: `func(ctx context.Context, challenge string) (Proof, error)`.
  Calls: `observeFixed`.
- **`ObserveObjects`** — ObserveObjects verifies only the fixed isolated read-only custody mount. The caller must independently bind its reviewed reader/volume; this proof does not authorize workspace access, exports, writer fencing or physical quota.
  Signature: `func(ctx context.Context, challenge string) (Proof, error)`.
  Calls: `observeFixed`.
- **`observeFixed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, challenge, fixed string) (Proof, error)`.
  Calls: `ValidChallenge`, `ctx.Err`, `f.Close`, `info.IsDir`, `info.Mode`, `info.Mode().Perm`, `info.Sys`, `io.LimitReader`, `io.ReadAll`, `mountIDAt`, `os.Getegid`, `os.Geteuid`, `os.Lstat`, `os.Open`, `syscall.Statfs`.
