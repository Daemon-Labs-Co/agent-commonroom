# cmd/server/worker_capture_stream.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_stream.go)

Go implementation for worker capture stream; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(dockerCLIRunner).StartCapture`** — StartCapture starts only a durably bound, independently inspected capture ID. The caller MUST use CaptureRuntime, hold creator ownership and enforce physical quotas/current authority. Rootless daemon selection is checked per invocation. This is the fixed reviewed helper's private stream, never a provider/model pipe. No shell, host execution, alternative executable, SSH or automatic fallback is accepted. Source forks that replace this adapter invalidate isolation guarantees.
  Signature: `func(ctx context.Context, id string) (toolbroker.CaptureProcess, error)`.
  Calls: `captureStartCommand`, `ctx.Err`, `openCaptureCommand`, `repositoryDigestPattern.MatchString`, `verifiedLocalDockerContext`.
- **`captureStartCommand`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) *exec.Cmd`.
  Calls: `exec.CommandContext`, `os.Getenv`.
- **`openCaptureCommand`** — Factory injection is private to synthetic process tests. Production always uses captureStartCommand after rootless validation; no request may supply a factory.
  Signature: `func(ctx context.Context, factory func(context.Context) *exec.Cmd) (toolbroker.CaptureProcess, error)`.
  Calls: `openEngineCommand`.
