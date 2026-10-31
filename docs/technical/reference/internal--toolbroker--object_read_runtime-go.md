# internal/toolbroker/object_read_runtime.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/object_read_runtime.go)

Go implementation for object read runtime; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ObjectReadLifecycleJournal`**: ObjectReadLifecycleJournal is trusted current-authority infrastructure. Begin MUST authenticate the creator/worker, fence writers, enforce actual physical quotas/export policy and durably reserve capacity before any daemon operation. Bind journals the exact created ID before start. Finish records observed absence or uncertainty under current authority; failure must latch integrity. No default exists. Neither successful cleanup nor release acknowledges user acceptance.
  Declaration: `interface { BeginObjectRead(context.Context, objectread.Request, CaptureObjectVolume, string) error BindObjectRead(context.Context, objectread.Request, string) error FinishObjectRead(context.Context, objectread.Request, string, bool) error }`.
- **Type `ObjectReadProcess`**: ObjectReadProcess owns only non-TTY private helper stdin/stdout. Wait must join pumps and verify successful exit without closing the stream prematurely. Close promptly interrupts I/O and joins pumps on every path. Never use model pipes.
  Declaration: `interface { io.ReadWriteCloser Wait(context.Context) error }`.
- **Type `ObjectReadStarter`**: ObjectReadStarter starts ONLY this durably bound ID on the same authenticated rootless daemon as the runner. It must not create/replace/exec or adopt a name. Any process returned on failure is still owned and closed by the runtime.
  Declaration: `interface { StartObjectRead(context.Context, string) (ObjectReadProcess, error) }`.
- **Type `ObjectReadRuntime`**: ObjectReadRuntime is one-shot isolated helper orchestration, not a download or host launcher. All dependencies are mandatory trusted infrastructure. A source fork CAN fake runners/journals, invalidating isolation. Production must supply verified same-daemon dependencies; no permissive journal, host/SSH fallback or isolation toggle is supplied.
  Declaration: `struct { physical *objectReadInspector starter ObjectReadStarter lifecycle ObjectReadLifecycleJournal guard objectcustody.ReadReleaseGuard attempted atomic.Bool }`.
- **`NewObjectReadRuntime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p *ObjectReadContainerPlan, r DockerRunner, s ObjectReadStarter, l ObjectReadLifecycleJournal, g objectcustody.ReadReleaseGuard) (*ObjectReadRuntime, error)`.
  Calls: `missingProbeDependency`.
- **`(*ObjectReadRuntime).Run`** — Run withholds bytes until successful exit, exited policy, joined process, exact-ID cleanup, repeated absence checks, durable lifecycle finish and a final current release acknowledgment. Any uncertainty returns nil, stops this read authority and retains charges/objects. There is no automatic retry or adoption.
  Signature: `func(ctx context.Context) (raw []byte, err error)`.
  Calls: `cancel`, `child.Err`, `context.Background`, `context.WithTimeout`, `ctx.Err`, `digest`, `errors.Join`, `missingProbeDependency`, `objectcustody.ReadSupervisedChannel`, `p.CreateArgs`, `p.VerifyAdmission`, `p.VerifyNoReferences`, `process.Close`, `process.Wait`, `r.attempted.CompareAndSwap`, `r.cleanupAndFinish`, `r.guard.AcknowledgeRead`, `r.guard.AuthorizeRead`, `r.guard.StopReadAuthority`, `r.lifecycle.BeginObjectRead`, `r.lifecycle.BindObjectRead`, `r.physical.capture`, `r.starter.StartObjectRead`, `stop`, `string`, `strings.TrimSpace`.
- **`(*ObjectReadRuntime).cleanupAndFinish`** — cleanupAndFinish keeps observed cleanup uncertainty distinct from journal persistence failure. No cleanup result grants release or refunds storage.
  Signature: `func(cleanup context.Context, id string) error`.
  Calls: `StopContainer`, `context.Background`, `context.WithTimeout`, `error`, `errors.Join`, `r.lifecycle.FinishObjectRead`, `r.physical.absent`, `stop`.
- **Type `objectReadRunningGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { runtime *ObjectReadRuntime id string }`.
- **`(objectReadRunningGuard).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, s objectcustody.Scope, r objectcustody.Receipt) error`.
  Calls: `g.runtime.guard.AuthorizeRead`, `g.runtime.physical.plan.VerifyAdmission`.
- **`(objectReadRunningGuard).acknowledgeAdmission`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `g.runtime.physical.plan.VerifyAdmission`.
- **`(objectReadRunningGuard).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `g.acknowledgeAdmission`, `g.runtime.guard.AcknowledgeRead`.
- **`(objectReadRunningGuard).StopReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `g.runtime.guard.StopReadAuthority`.
