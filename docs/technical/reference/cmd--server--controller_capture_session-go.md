# cmd/server/controller_capture_session.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_capture_session.go)

Go implementation for controller capture session; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `controllerCaptureSessionEvidence`**: controllerCaptureSessionEvidence supplies independent operation-specific creator, physical quota/account and volume/image checks before a local journal transition, plus persisted local/physical evidence for controller acknowledgments. It must not reenter server/admission locks or the waiting control stream. No default exists.
  Declaration: `interface { controllerCaptureLedgerEvidence ValidateCaptureOperation(context.Context, room.State, workerJobRecord, workerCaptureIntent, workerProbeEvent) error }`.
- **Type `controllerCaptureSession`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { server *server lifetime context.Context cancel context.CancelFunc actor, worker string record workerJobRecord intent workerCaptureIntent evidence controllerCaptureSessionEvidence ledger *controllerCaptureLedger }`.
- **`(*controllerCaptureSession).CaptureAuthorityLifetime`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() context.Context`.
- **`newControllerCaptureSession`** — Session identity comes only from authenticated private setup, never peer fields. This adapter checks policy; it does not establish transport authentication or implement physical quotas. There is no host launch or optional isolation bypass.
  Signature: `func(ctx context.Context, s *server, worker string, r workerJobRecord, i workerCaptureIntent, e controllerCaptureSessionEvidence) (*controllerCaptureSession, error)`.
  Calls: `cancel`, `captureAuthorityBinding`, `context.WithCancel`, `copyCaptureIntent`, `ctx.Err`, `ctx.Value`, `detachedProbeWorkerRecord`, `membershipPartPattern.MatchString`, `newControllerCaptureLedger`, `nilWorkerObservationDependency`, `workerAliasPattern.MatchString`.
- **`(*controllerCaptureSession).callContext`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (context.Context, context.CancelFunc, error)`.
  Calls: `cancel`, `context.AfterFunc`, `context.WithCancel`, `context.WithValue`, `ctx.Err`, `g.lifetime.Err`, `stop`.
- **`(*controllerCaptureSession).pinned`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State) bool`.
  Calls: `ctx.Err`, `g.lifetime.Err`, `sameWorkerBinding`.
- **`(*controllerCaptureSession).candidate`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, event workerProbeEvent) (room.State, room.WorkspaceCaptureReservation, error)`.
  Calls: `g.ledger.current`, `g.pinned`, `len`, `validateControllerCaptureEvent`.
- **`(*controllerCaptureSession).AuthorizeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord, i workerCaptureIntent, event workerProbeEvent) error`.
  Calls: `cancel`, `captureAuthorityEvent`, `captureIntentEqual`, `child.Err`, `copyCaptureIntent`, `detachedProbeWorkerRecord`, `g.callContext`, `g.candidate`, `g.evidence.ValidateCaptureOperation`, `g.server.storageContext`, `probeWorkerRecordDigest`, `retainedObservationRecord`, `s.mu.Lock`, `s.mu.Unlock`, `storageCancel`.
- **`(*controllerCaptureSession).AcknowledgeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, local workerCaptureRecord, event workerProbeEvent) error`.
  Calls: `cancel`, `captureAuthorityRecord`, `copyCaptureRecord`, `g.callContext`, `ledger.Acknowledge`, `newControllerCaptureLedger`.
- **Type `captureSessionAcknowledgment`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { session *controllerCaptureSession local workerCaptureRecord }`.
- **`(*captureSessionAcknowledgment).ValidateCaptureTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.WorkspaceCaptureReservation, e workerProbeEvent) error`.
  Calls: `a.session.ValidateCaptureTransition`, `captureIntentEqual`, `reflect.DeepEqual`.
- **`(*controllerCaptureSession).ValidateCaptureTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.WorkspaceCaptureReservation, event workerProbeEvent) error`.
  Calls: `captureIntentEqual`, `copyControllerCaptureReservation`, `ctx.Err`, `g.evidence.ValidateCaptureTransition`, `g.lifetime.Err`, `g.pinned`.
- **`(*controllerCaptureSession).StopCaptureAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `g.cancel`, `g.server.mu.Lock`, `g.server.mu.Unlock`, `g.server.tripIntegrityLocked`.
