# internal/workspacemount/observe_other.go

[Module index](internal.md) · [Open source](../../../internal/workspacemount/observe_other.go)

Go implementation for observe other; declarations and calls are listed below.

Package: `workspacemount`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`Observe`** — Unsupported platforms refuse; there is no host or simulated fallback.
  Signature: `func(context.Context, string) (Proof, error)`.
- **`ObserveObjects`** — ObserveObjects also refuses outside Linux; no simulated/host fallback.
  Signature: `func(context.Context, string) (Proof, error)`.
