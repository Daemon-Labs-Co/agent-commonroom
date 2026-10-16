# internal/toolbroker/container_inspection.go

[Module index](internal.md) · [Open source](../../internal/toolbroker/container_inspection.go)

Go implementation for container inspection; declarations and calls are listed below.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`toolEnvironment`** — toolEnvironment is fixed supervisor policy, not model input. Unknown inherited image environment is refused by inspection; image provenance remains separate.
  Signature: `func() map[string]string`.
- **`(*ToolContainerPlan).labels`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() map[string]string`.
- **`(*ToolContainerPlan).VolumeCreateArgs`** — VolumeCreateArgs generates only a plain local named volume. The future fenced supervisor must refuse existing volumes first, initialize its approved snapshot without credentials, and verify inspection before using it. This is not a launch.
  Signature: `func(now time.Time) ([]string, error)`.
  Calls: `append`, `p.RunArgs`, `p.WorkspaceVolume`, `p.labels`, `p.volume.Load`.
- **Type `toolInspection`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID string 'json:"Id"' Name string Image string RestartCount int State struct { Running, Paused, Restarting, Dead *bool Error string Status string } Config struct { Image, User, WorkingDir string Tty, OpenStdin, StdinOnce *bool Env, Entrypoint, Cmd []string Labels map[string]string Healthcheck struct{ Test []string } } HostConfig struct { Privileged, ReadonlyRootfs, Init *bool AutoRemove *bool RestartPolicy struct{ Name string } NetworkMode, PidMode, IpcMode, CgroupnsMode, UsernsMode string Memory, MemorySwap, NanoCpus, PidsLimit *int64 CapDrop, CapAdd, SecurityOpt, Binds, VolumesFrom []string Devices, DeviceRequests []json.RawMessage PortBindings map[string]json.RawMessage Tmpfs map[string]string LogConfig struct { Type string Config map[string]string } Ulimits []struct { Name string Soft, Hard int64 } } Mounts []struct { Type, Name, Source, Destination, Driver, Propagation string RW *bool } }`.
- **Type `toolVolumeInspection`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Name, Driver, Scope, Mountpoint, CreatedAt string Options, Labels map[string]string }`.
- **`isBool`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p *bool, value bool) bool`.
- **`isInt`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p *int64, value int64) bool`.
- **`sameMap`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(a, b map[string]string) bool`.
  Calls: `len`.
- **`(*ToolContainerPlan).VerifyInspection`** — VerifyInspection consumes bounded single-object Docker inspection metadata fetched directly by the trusted worker, never supplied by a model or plugin. dockerRoot must be the approved rootless daemon's canonical data root. Unknown descriptive Docker fields are ignored, but duplicate/case-alias fields refused. This verifies selected reported policy only: it cannot prove a trusted daemon, image contents, fresh volume history, atomic execution, or real shutdown.
  Signature: `func(container, volume []byte, dockerRoot string, now time.Time) error`.
  Calls: `p.verifyInspection`.
- **`(*ToolContainerPlan).VerifyCreatedInspection`** — VerifyCreatedInspection checks the same policy before any process may start. A stopped existing container is not automatically fresh; creation ownership must be established by the worker lifecycle and ToolContainerRuntime.
  Signature: `func(container, volume []byte, dockerRoot string, now time.Time) error`.
  Calls: `p.verifyInspection`.
- **`(*ToolContainerPlan).verifyInspection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(container, volume []byte, dockerRoot string, now time.Time, running bool) error`.
  Calls: `decodeInspection`, `digest`, `isBool`, `isInt`, `len`, `p.ContainerName`, `p.RunArgs`, `p.labels`, `p.verifyVolume`, `path.Clean`, `path.IsAbs`, `sameMap`, `slices.Equal`, `strings.Cut`, `strings.HasPrefix`, `strings.TrimPrefix`, `toolEnvironment`.
- **`decodeInspection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, target any) error`.
  Calls: `bytes.NewReader`, `bytes.TrimSpace`, `d.Token`, `inspectionValue`, `json.NewDecoder`, `json.Unmarshal`, `len`.
- **`inspectionValue`** — inspectionValue rejects ambiguous object members before Go's case-insensitive struct decoder. Null optional Docker fields are permitted; depth is bounded.
  Signature: `func(d *json.Decoder, depth int) error`.
  Calls: `d.More`, `d.Token`, `inspectionValue`, `strings.ToLower`.
- **`(*ToolContainerPlan).verifyVolume`** — verifyVolume rejects Docker local-driver options that could hide host binds.
  Signature: `func(v toolVolumeInspection, dockerRoot string) error`.
  Calls: `p.labels`, `p.verifyVolumeIdentity`, `p.verifyVolumeShape`, `p.volume.Load`.
- **`(*ToolContainerPlan).verifyVolumeShape`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(v toolVolumeInspection, dockerRoot string, labels map[string]string) error`.
  Calls: `len`, `p.WorkspaceVolume`, `path.Join`, `sameMap`.
