# internal/toolbroker/object_read_recovery_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/object_read_recovery_test.go)

Automated regression checks for object_read_recovery; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadRecoveryFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *objectReadRuntimeFixture deny, failFinish bool }`.
- **`(*objectReadRecoveryFixture).BeginObjectReadRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, v CaptureObjectVolume, image, id string) error`.
  Calls: `ctx.Err`.
- **`(*objectReadRecoveryFixture).FinishObjectReadRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, id string, absent bool) error`.
  Calls: `ctx.Err`, `f.t.Fatal`.
- **`newObjectReadRecoveryFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*ObjectReadRecovery, *objectReadRecoveryFixture)`.
  Calls: `NewObjectReadRecovery`, `newObjectReadRuntimeFixture`, `t.Fatal`, `t.Helper`.
- **`TestObjectReadRecoveryKnownIdentity`** — Regression contract: TestObjectReadRecoveryKnownIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `newObjectReadRecoveryFixture`, `probeBool`, `r.Reconcile`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObjectReadRecoveryRefusesConflicts`** — Regression contract: TestObjectReadRecoveryRefusesConflicts.
  Signature: `func(t *testing.T)`.
  Calls: `errors.Is`, `newObjectReadRecoveryFixture`, `r.Reconcile`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObjectReadRecoveryAuthorizesBeforeDaemon`** — Regression contract: TestObjectReadRecoveryAuthorizesBeforeDaemon.
  Signature: `func(t *testing.T)`.
  Calls: `NewObjectReadRecovery`, `cancel`, `context.WithCancel`, `newObjectReadRecoveryFixture`, `r.Reconcile`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
