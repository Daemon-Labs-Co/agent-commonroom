# internal/projectquota/kernel_linux_amd64.go

[Module index](internal.md) · [Open source](../../../internal/projectquota/kernel_linux_amd64.go)

Go implementation for kernel linux amd64; declarations and calls are listed below.

Package: `projectquota`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `linuxReader`**: Linux x86-64 uses syscall number 443 for quotactl_fd. Only fixed read-only XFS status/quota operations are reachable. ABI references are in the guide. Calls may block in the kernel; cancellation never detaches an in-flight query. No elevated helper, quota setter, command runner or privileged fallback exists.
  Declaration: `struct{ fd int }`.
- **`newKernelReader`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(root *os.File) (kernelReader, error)`.
  Calls: `int`, `raw.Control`, `root.SyscallConn`, `syscall.Close`, `syscall.Syscall`.
- **`(*linuxReader).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `syscall.Close`.
- **`(*linuxReader).Root`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (rootObservation, error)`.
  Calls: `binary.LittleEndian.Uint32`, `ctx.Err`, `runtime.KeepAlive`, `syscall.Fstat`, `syscall.Fstatfs`, `syscall.Syscall`, `uintptr`, `unsafe.Pointer`.
- **`(*linuxReader).query`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, command uintptr, project uint32, raw []byte) error`.
  Calls: `ctx.Err`, `runtime.KeepAlive`, `syscall.Syscall6`, `uintptr`, `unsafe.Pointer`.
- **`(*linuxReader).Status`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) ([]byte, error)`.
  Calls: `make`, `r.query`.
- **`(*linuxReader).Quota`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, project uint32) ([]byte, error)`.
  Calls: `make`, `r.query`.
