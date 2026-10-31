# internal/toolbroker/execution_context.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/execution_context.go)

Go implementation for execution context; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*Authority).executionContext`** — executionContext binds cooperative work to both the original job deadline and revocation. Parent cancellation/earlier deadlines remain effective. Callers must cancel on return; this is not proof of descendant/container termination.
  Signature: `func(parent context.Context) (context.Context, context.CancelFunc)`.
  Calls: `cancel`, `context.WithDeadline`, `ctx.Done`.
