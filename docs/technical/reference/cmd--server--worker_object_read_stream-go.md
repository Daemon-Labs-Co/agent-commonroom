# cmd/server/worker_object_read_stream.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_object_read_stream.go)

Go implementation for worker object read stream; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(dockerCLIRunner).StartObjectRead`** — StartObjectRead starts only a separately durably bound, independently inspected immutable helper ID. The caller MUST use ObjectReadRuntime with actual creator ownership, physical quotas and current export authority. Rootless local daemon selection is revalidated per invocation. No shell, model executable, host/SSH fallback, socket exposure or optional isolation bypass exists. Source forks can replace this adapter and invalidate the mandatory execution boundary.
  Signature: `func(ctx context.Context, id string) (toolbroker.ObjectReadProcess, error)`.
  Calls: `captureStartCommand`, `ctx.Err`, `openObjectReadCommand`, `repositoryDigestPattern.MatchString`, `verifiedLocalDockerContext`.
- **`openObjectReadCommand`** — Factory injection is private to synthetic process tests; no request can select a command. The shared process owner preserves final stdout through Wait and sanitizes stderr/errors, while Close interrupts I/O and joins the process.
  Signature: `func(ctx context.Context, factory func(context.Context) *exec.Cmd) (toolbroker.ObjectReadProcess, error)`.
  Calls: `openCaptureCommand`.
