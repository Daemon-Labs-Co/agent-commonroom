# internal/toolbroker/capture_runtime.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/capture_runtime.go)

Go implementation for capture runtime; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `CaptureLifecycleJournal`**: CaptureLifecycleJournal is trusted current-authority infrastructure. Begin MUST hold creator ownership, fence writers, establish physical quotas and durably reserve a container slot before creation. Bound MUST journal the exact ID before start. Finish rechecks authority and records actual absence/uncertainty before report release. Failed/unknown work retains storage and charges. No default exists.
  Declaration: `interface { BeginCapture(context.Context, workspacecapture.Request, CaptureObjectVolume, string) error BindCapture(context.Context, workspacecapture.Request, string) error FinishCapture(context.Context, workspacecapture.Request, string, bool) error }`.
- **Type `CaptureProcess`**: CaptureProcess owns only the reviewed helper's non-TTY private stdin/stdout. Wait must observe actual exit success, respect cancellation and join all pumps; Close must promptly interrupt I/O and join pumps, including on failure. Implementations MUST NOT use model/provider pipes.
  Declaration: `interface { io.ReadWriteCloser Wait(context.Context) error }`.
- **Type `CaptureStarter`**: CaptureStarter starts ONLY the supplied durably bound container ID on the same approved rootless daemon as DockerRunner, returning an exclusively owned stream. StartCapture must not create/replace/exec commands, adopt names or return a model stream. On failure any returned process is still owned/closed by the runtime. Source forks CAN fake these dependencies, invalidating isolation guarantees.
  Declaration: `interface { StartCapture(context.Context, string) (CaptureProcess, error) }`.
- **Type `CaptureRuntime`**: CaptureRuntime is a one-shot lifecycle, never a host launcher or authorization API. All dependencies are mandatory trusted infrastructure. No real daemon/peer assembly, physical quota provisioner, fallback or boundary toggle is supplied.
  Declaration: `struct { *captureInspector starter CaptureStarter lifecycle CaptureLifecycleJournal authority workspacecapture.Authority custody objectcustody.Journal attempted atomic.Bool }`.
- **`NewCaptureRuntime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p *CaptureContainerPlan, r DockerRunner, s CaptureStarter, l CaptureLifecycleJournal, a workspacecapture.Authority, j objectcustody.Journal) (*CaptureRuntime, error)`.
  Calls: `missingProbeDependency`.
- **`(*CaptureRuntime).request`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() workspacecapture.Request`.
  Calls: `slices.Clone`.
- **`(*CaptureRuntime).Run`** — Run withholds its report until successful exit, exact-ID removal, repeated absence checks and final durable current-authority acknowledgment. Any failure returns a zero report and uncertainty; there is no automatic retry, refund or volume delete.
  Signature: `func(ctx context.Context) (report workspacecapture.Report, err error)`.
  Calls: `StopContainer`, `cancel`, `child.Err`, `context.Background`, `context.WithTimeout`, `ctx.Err`, `digest`, `error`, `errors.Join`, `missingProbeDependency`, `process.Close`, `process.Wait`, `r.absent`, `r.attempted.CompareAndSwap`, `r.capture`, `r.lifecycle.BeginCapture`, `r.lifecycle.BindCapture`, `r.lifecycle.FinishCapture`, `r.plan.CreateArgs`, `r.plan.VerifyAdmission`, `r.plan.name`, `r.request`, `r.starter.StartCapture`, `stop`, `string`, `strings.TrimSpace`, `workspacecapture.ServeSupervisedCaptureChannel`.
- **Type `captureRunningAuthority`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { runtime *CaptureRuntime id string }`.
- **`(captureRunningAuthority).AuthorizeFrozenSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, b workspacesnapshot.Binding, p workspacemount.Proof) error`.
  Calls: `a.runtime.authority.AuthorizeFrozenSource`, `a.runtime.plan.VerifyAdmission`.
