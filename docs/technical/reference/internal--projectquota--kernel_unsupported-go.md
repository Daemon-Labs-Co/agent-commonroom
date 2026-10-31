# internal/projectquota/kernel_unsupported.go

[Module index](internal.md) · [Open source](../../../internal/projectquota/kernel_unsupported.go)

Go implementation for kernel unsupported; declarations and calls are listed below.

Package: `projectquota`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`newKernelReader`** — Unsupported platforms refuse. Never substitute directory size, filesystem free space, labels, polling or host execution for actual project enforcement.
  Signature: `func(*os.File) (kernelReader, error)`.
