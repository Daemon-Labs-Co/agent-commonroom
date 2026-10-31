# internal/projectquota/observe.go

[Module index](internal.md) · [Open source](../../../internal/projectquota/observe.go)

Go implementation for observe; declarations and calls are listed below.

Package: `projectquota`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `Expected`**: Expected comes from independently approved volume/custody provisioning, never model metadata or the first observed root. Nonzero hard bounds are reservations, not permission to consume unaccounted space. Project zero is never accepted.
  Declaration: `struct { Device, Inode uint64 Project uint32 MaxBytes, MaxInodes uint64 }`.
- **Type `Observation`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Device, Inode uint64 Project uint32 HardBytes, UsedBytes, HardInodes, UsedInodes uint64 }`.
- **Type `rootObservation`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Device, Inode uint64 Type int64 Project, Flags uint32 }`.
- **Type `kernelReader`**: Behavior contract; required method signatures are shown in the declaration.
  Declaration: `interface { Root(context.Context) (rootObservation, error) Status(context.Context) ([]byte, error) Quota(context.Context, uint32) ([]byte, error) Close() error }`.
- **`validExpected`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(e Expected) bool`.
- **`Inspect`** — Inspect pins a duplicate of an already opened root directory descriptor for all observations. Callers must keep actual creator/writer fences and independently prove complete inventory, volume mapping and rootless containment. This is a point-in-time read-only check, not a physical quota provisioner or release guard. Permission/unsupported/cancellation/uncertainty refuse without fallback.
  Signature: `func(ctx context.Context, root *os.File, e Expected) (Observation, error)`.
  Calls: `ctx.Err`, `inspectOwned`, `newKernelReader`, `validExpected`.
- **`inspectOwned`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, reader kernelReader, e Expected) (observation Observation, err error)`.
  Calls: `inspect`, `reader.Close`.
- **`inspect`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r kernelReader, e Expected) (Observation, error)`.
  Calls: `bytes.Equal`, `ctx.Err`, `decodeQuota`, `fail`, `r.Quota`, `r.Root`, `r.Status`, `validExpected`, `validRoot`, `validStatus`.
- **`validRoot`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r rootObservation, e Expected) bool`.
- **`validStatus`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) bool`.
  Calls: `binary.LittleEndian.Uint16`, `binary.LittleEndian.Uint64`, `len`.
- **`decodeQuota`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, e Expected) (Observation, error)`.
  Calls: `binary.LittleEndian.Uint32`, `binary.LittleEndian.Uint64`, `fail`, `len`.
