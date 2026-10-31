# internal/toolbroker/object_read_plan.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/object_read_plan.go)

Go implementation for object read plan; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `ObjectReadContainerPlan`**: ObjectReadContainerPlan describes only the fixed credential-free reader. It supplies no lifecycle/capacity/release rights. The supervisor MUST authenticate the rootless worker/creator, durably reserve a helper slot and exact ID, fence writers, enforce physical quota and current export policy, and confirm exit/ cleanup before downstream release. Source forks CAN falsify observations; no host/SSH execution fallback or optional isolation toggle exists.
  Declaration: `struct { request objectread.Request image, root, encoded, binding string objects CaptureObjectVolume }`.
- **Type `ObjectReadIntent`**: ObjectReadIntent is finite immutable journal metadata, not permission to execute, read, release storage or bypass mandatory isolated admission.
  Declaration: `struct { Request objectread.Request 'json:"request"' Image string 'json:"image"' Objects CaptureObjectVolume 'json:"objects"' }`.
- **`ValidObjectReadMetadata`** — ValidObjectReadMetadata validates finite identities without inventing a daemon root observation. Actual worker, image, quota and release proof remain mandatory.
  Signature: `func(q objectread.Request, image string, objects CaptureObjectVolume) bool`.
  Calls: `created.IsZero`, `digest`, `json.Marshal`, `len`, `objectread.ValidRequest`, `strings.HasPrefix`, `strings.TrimPrefix`, `time.Parse`.
- **`NewObjectReadContainerPlan`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(q objectread.Request, image, root string, objects CaptureObjectVolume) (*ObjectReadContainerPlan, error)`.
  Calls: `ValidObjectReadMetadata`, `base64.RawURLEncoding.EncodeToString`, `hex.EncodeToString`, `json.Marshal`, `len`, `path.Clean`, `path.IsAbs`, `sha256.Sum256`.
- **`(*ObjectReadContainerPlan).name`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() string`.
- **`(*ObjectReadContainerPlan).labels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]string`.
- **`(*ObjectReadContainerPlan).objectLabels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]string`.
- **`(*ObjectReadContainerPlan).CreateArgs`** — CreateArgs is fixed policy metadata, never permission to launch a reader.
  Signature: `func() []string`.
  Calls: `append`, `p.labels`, `p.name`.
- **`(*ObjectReadContainerPlan).verifyVolume`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) (toolVolumeInspection, error)`.
  Calls: `decodeInspection`, `len`, `p.objectLabels`, `path.Join`, `sameMap`.
- **`(*ObjectReadContainerPlan).VerifyInspection`** — VerifyInspection validates actual bounded observations for a durably known exact ID. It neither discovers/adopts IDs nor proves quota or current release rights.
  Signature: `func(id, phase string, container, objects []byte) error`.
  Calls: `decodeInspection`, `digest`, `isBool`, `isInt`, `len`, `p.labels`, `p.name`, `p.verifyVolume`, `privateHelperHost`, `sameMap`, `slices.Equal`.
- **`(*ObjectReadContainerPlan).VerifyAdmission`** — VerifyAdmission performs two independent metadata/reference/catalog passes. Only read-only queries are issued; the caller must independently fence writers and hold actual creator/account authority throughout. Daemon admins remain trusted.
  Signature: `func(ctx context.Context, runner DockerRunner, id, phase string) error`.
  Calls: `(&ContainerExecutor{runner: runner}).capture`, `cancel`, `capture`, `child.Err`, `context.WithTimeout`, `ctx.Err`, `digest`, `missingProbeDependency`, `p.VerifyInspection`, `p.name`, `string`, `strings.TrimSpace`.
