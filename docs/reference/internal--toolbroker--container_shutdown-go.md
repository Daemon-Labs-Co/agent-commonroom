# internal/toolbroker/container_shutdown.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/container_shutdown.go)

Go implementation for container shutdown; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`StopContainer`** — StopContainer removes only an immutable container identity and requires a successful exact-ID absence observation. verify, when supplied, validates ownership before removal. This proves neither pair shutdown nor creator lease release; callers retain their stronger reconciliation checks. The trusted runner must enforce the approved local rootless daemon on every invocation. A source fork CAN replace that adapter, BUT invalidates containment proof.
  Signature: `func(ctx context.Context, runner DockerRunner, id string, verify func(context.Context, string) error) error`.
  Calls: `absent`, `cancel`, `capture`, `child.Err`, `context.WithCancel`, `ctx.Err`, `digest`, `out.snapshot`, `runner.Run`, `string`, `strings.TrimSpace`, `verify`.
