# internal/toolbroker/capture_plan.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/capture_plan.go)

Go implementation for capture plan; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `CaptureObjectVolume`**: CaptureObjectVolume is retained creation metadata, not storage authorization. A trusted supervisor MUST independently prove physical quotas, exclusive creator ownership and current custody authorization. Docker labels cannot prove those.
  Declaration: `struct{ Name, Nonce, CreatedAt string }`.
- **Type `CaptureContainerPlan`**: CaptureContainerPlan admits only the reviewed, credential-free capture helper. It does not launch containers or confer authority. A future runtime MUST journal creation and exact IDs, authenticate its private stream, fence other writers, enforce physical storage quotas, and verify exit/cleanup before report release. Source forks CAN fake inspection/provenance, BUT invalidate isolation guarantees. No host/SSH execution or mount fallback exists and this boundary has no toggle.
  Declaration: `struct { request workspacecapture.Request image, root, encoded, binding string source WorkspaceVolumeIdentity objects CaptureObjectVolume }`.
- **`ValidCaptureMetadata`** — ValidCaptureMetadata checks finite identities only, never daemon/root provenance, physical quotas or execution authority. Durable journal replay uses this without inventing an observed Docker data root.
  Signature: `func(request workspacecapture.Request, image string, objects CaptureObjectVolume) bool`.
  Calls: `captureMetadata`.
- **`captureMetadata`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(request workspacecapture.Request, image string, objects CaptureObjectVolume) (string, []byte, error)`.
  Calls: `ValidWorkspaceVolume`, `created.IsZero`, `digest`, `json.Marshal`, `len`, `strings.HasPrefix`, `strings.TrimPrefix`, `time.Parse`, `workspacecapture.RequestDigest`.
- **`NewCaptureContainerPlan`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(request workspacecapture.Request, image, root string, objects CaptureObjectVolume) (*CaptureContainerPlan, error)`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `captureMetadata`, `len`, `path.Clean`, `path.IsAbs`, `slices.Clone`.
- **`(*CaptureContainerPlan).name`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() string`.
- **`(*CaptureContainerPlan).labels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]string`.
- **`(*CaptureContainerPlan).objectLabels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]string`.
- **`(*CaptureContainerPlan).CreateArgs`** — CreateArgs describes fixed policy only; it MUST NOT be launched without the durable supervised lifecycle described on CaptureContainerPlan.
  Signature: `func() []string`.
  Calls: `append`, `p.labels`, `p.name`.
- **`(*CaptureContainerPlan).VerifyInspection`** — VerifyInspection checks directly fetched, bounded Docker metadata for a durably recorded exact ID. phase is created, running or exited; unknown phases fail. Repeated physical inventory/current-authority checks are still mandatory: this pure validator cannot exclude other mounts or races by daemon administrators.
  Signature: `func(id, phase string, container, source, objects []byte) error`.
  Calls: `decodeInspection`, `digest`, `isBool`, `isInt`, `len`, `p.labels`, `p.name`, `p.verifyVolumes`, `privateHelperHost`, `sameMap`, `slices.Equal`.
- **`(*CaptureContainerPlan).VerifyAdmission`** — VerifyAdmission fetches independent metadata and reference inventories twice. The trusted caller MUST hold exclusive creator ownership throughout, authenticate the approved rootless runner, and independently enforce quota/current policy. No creation, execution, adoption, removal or storage release occurs here.
  Signature: `func(ctx context.Context, runner DockerRunner, id, phase string) error`.
  Calls: `(&ContainerExecutor{runner: runner}).capture`, `cancel`, `capture`, `child.Err`, `context.WithTimeout`, `ctx.Err`, `digest`, `missingProbeDependency`, `p.VerifyInspection`, `p.name`, `string`, `strings.TrimSpace`.
- **`(*CaptureContainerPlan).verifyVolumes`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(source, objects []byte) (src, obj toolVolumeInspection, err error)`.
  Calls: `decodeInspection`, `len`, `p.objectLabels`, `path.Join`, `sameMap`, `validObservedWorkspaceVolume`.
- **`privateHelperHost`** — privateHelperHost verifies common credential-free helper limits, not image, mounts or authority.
  Signature: `func(c toolInspection) bool`.
  Calls: `isBool`, `isInt`, `len`, `slices.Equal`.
