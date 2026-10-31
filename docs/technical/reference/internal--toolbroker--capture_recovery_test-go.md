# internal/toolbroker/capture_recovery_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/capture_recovery_test.go)

Automated regression checks for capture_recovery; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `captureRecoveryFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *captureRuntimeFixture denied, finishDenied bool expectedImage string }`.
- **`(*captureRecoveryFixture).BeginCaptureRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q workspacecapture.Request, o CaptureObjectVolume, image, id string) error`.
  Calls: `ctx.Err`, `workspacecapture.RequestDigest`.
- **`(*captureRecoveryFixture).FinishCaptureRecovery`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q workspacecapture.Request, id string, absent bool) error`.
  Calls: `ctx.Err`.
- **`newCaptureRecoveryFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*CaptureRecovery, *captureRecoveryFixture)`.
  Calls: `NewCaptureRecovery`, `newCaptureRuntimeFixture`, `t.Fatal`, `t.Helper`.
- **`TestCaptureRecoveryKnownPresentAndAbsent`** — Regression contract: TestCaptureRecoveryKnownPresentAndAbsent.
  Signature: `func(t *testing.T)`.
  Calls: `newCaptureRecoveryFixture`, `probeBool`, `r.Reconcile`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureRecoveryRefusesConflictAndRetainsUncertainty`** — Regression contract: TestCaptureRecoveryRefusesConflictAndRetainsUncertainty.
  Signature: `func(t *testing.T)`.
  Calls: `errors.Is`, `newCaptureRecoveryFixture`, `r.Reconcile`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureRecoveryRequiresDurableAuthorityBeforeDaemon`** — Regression contract: TestCaptureRecoveryRequiresDurableAuthorityBeforeDaemon.
  Signature: `func(t *testing.T)`.
  Calls: `NewCaptureRecovery`, `cancel`, `capturePlanFixture`, `context.WithCancel`, `newCaptureRecoveryFixture`, `r.Reconcile`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
