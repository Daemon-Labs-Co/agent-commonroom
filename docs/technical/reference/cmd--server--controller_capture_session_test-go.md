# cmd/server/controller_capture_session_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_capture_session_test.go)

Automated regression checks for controller_capture_session; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticCaptureSessionEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { operation func(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error transition func(context.Context, room.State, room.WorkspaceCaptureReservation, workerProbeEvent) error calls, acks int }`.
- **`(*syntheticCaptureSessionEvidence).ValidateCaptureOperation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r workerJobRecord, i workerCaptureIntent, v workerProbeEvent) error`.
  Calls: `ctx.Err`, `e.operation`.
- **`(*syntheticCaptureSessionEvidence).ValidateCaptureTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.WorkspaceCaptureReservation, v workerProbeEvent) error`.
  Calls: `ctx.Err`, `e.transition`.
- **`captureSessionFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*controllerCaptureSession, *mockDB, context.Context, *syntheticCaptureSessionEvidence)`.
  Calls: `controllerCaptureFixture`, `newControllerCaptureSession`, `strings.Repeat`, `t.Fatal`, `t.Helper`.
- **`TestControllerCaptureSessionPinnedLifecycle`** — Regression contract: TestControllerCaptureSessionPinnedLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `artifactState`, `captureSessionFixture`, `context.WithValue`, `controllerCaptureKey`, `copyCaptureIntent`, `db.snapshot`, `g.AcknowledgeCapture`, `g.AuthorizeCapture`, `len`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerCaptureSessionRejectsHistorySubstitution`** — Regression contract: TestControllerCaptureSessionRejectsHistorySubstitution.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `captureSessionFixture`, `db.snapshot`, `g.AcknowledgeCapture`, `len`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerCaptureSessionPolicyAndLifetime`** — Regression contract: TestControllerCaptureSessionPolicyAndLifetime.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `captureSessionFixture`, `context.WithCancel`, `copyCaptureIntent`, `db.seed`, `db.snapshot`, `detachedProbeWorkerRecord`, `errors.Is`, `errors.New`, `g.AcknowledgeCapture`, `g.AuthorizeCapture`, `len`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerCaptureSessionCleanupAndPersistenceFault`** — Regression contract: TestControllerCaptureSessionCleanupAndPersistenceFault.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `captureSessionFixture`, `db.seed`, `g.AcknowledgeCapture`, `g.AuthorizeCapture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerCaptureSessionDetachesEvidenceAndStops`** — Regression contract: TestControllerCaptureSessionDetachesEvidenceAndStops.
  Signature: `func(t *testing.T)`.
  Calls: `captureSessionFixture`, `g.AuthorizeCapture`, `g.StopCaptureAuthority`, `newControllerCaptureSession`, `t.Context`, `t.Fatal`, `t.Parallel`.
