# cmd/server/controller_capture_authority_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_capture_authority_test.go)

Automated regression checks for controller_capture_authority; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticCaptureEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { check func(context.Context, room.State, workerJobRecord, workspacecapture.Request, workspacemount.Proof) error calls int }`.
- **`(*syntheticCaptureEvidence).ValidateCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r workerJobRecord, q workspacecapture.Request, p workspacemount.Proof) error`.
  Calls: `ctx.Err`, `e.check`.
- **`controllerCaptureFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, context.Context, workerJobRecord, workspacecapture.Request, workspacemount.Proof)`.
  Calls: `artifactJournalFixture`, `artifactState`, `retainedObservationFixture`, `retainedObservationRecord`, `strings.Repeat`, `t.Fatal`, `t.Helper`, `workerMembership`, `workspacecapture.ContractRequest`.
- **`TestControllerCaptureAuthorityDurableReplayAndPinnedActor`** — Regression contract: TestControllerCaptureAuthorityDurableReplayAndPinnedActor.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `auditSummary`, `context.WithValue`, `controllerCaptureFixture`, `db.snapshot`, `g.AuthorizeFrozenSource`, `json.Unmarshal`, `len`, `newControllerCaptureAuthority`, `projectControllerCaptureAuthorization`, `strings.Contains`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validControllerCaptureAuthorization`.
- **`TestControllerCaptureAuthorityDenials`** — Regression contract: TestControllerCaptureAuthorityDenials.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `context.WithValue`, `controllerCaptureFixture`, `db.seed`, `db.snapshot`, `errors.Is`, `errors.New`, `g.AuthorizeFrozenSource`, `len`, `newControllerCaptureAuthority`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerCaptureAuthorityCurrentPolicyAndHistoricalReplay`** — Regression contract: TestControllerCaptureAuthorityCurrentPolicyAndHistoricalReplay.
  Signature: `func(t *testing.T)`.
  Calls: `controllerCaptureFixture`, `copyCaptureRequest`, `g.current`, `j.Deadline.Add`, `liveControllerCaptureBinding`, `maps.Clone`, `newControllerCaptureAuthority`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `validControllerCaptureAuthorization`.
- **`TestControllerCaptureAuthorityRejectsPublicInjection`** — Regression contract: TestControllerCaptureAuthorityRejectsPublicInjection.
  Signature: `func(t *testing.T)`.
  Calls: `mockRequest`, `mockServer`, `t.Fatal`.
- **`TestControllerCaptureAuthorityMissingInfrastructureRefuses`** — Regression contract: TestControllerCaptureAuthorityMissingInfrastructureRefuses.
  Signature: `func(t *testing.T)`.
  Calls: `controllerCaptureFixture`, `g.AuthorizeFrozenSource`, `newControllerCaptureAuthority`, `t.Context`, `t.Fatal`.
