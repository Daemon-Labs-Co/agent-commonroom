# cmd/server/controller_workspace_probe_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_workspace_probe_test.go)

Automated regression checks for controller_workspace_probe; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticWorkspaceProbeEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { observe func(context.Context, room.State, workerJobRecord, toolbroker.WorkspaceProbeIntent, string) error stage func(context.Context, room.State, room.WorkspaceProbeReservation, workerProbeEvent) error }`.
- **`(*syntheticWorkspaceProbeEvidence).ValidateWorkspaceProbe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r workerJobRecord, i toolbroker.WorkspaceProbeIntent, op string) error`.
  Calls: `ctx.Err`, `e.observe`.
- **`(*syntheticWorkspaceProbeEvidence).ValidateProbe`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.WorkspaceProbeReservation, event workerProbeEvent) error`.
  Calls: `ctx.Err`, `e.stage`.
- **`controllerWorkspaceAuthorityFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, context.Context, workerJobRecord, toolbroker.WorkspaceProbeIntent)`.
  Calls: `context.WithValue`, `db.seed`, `mockServer`, `retainedObservationFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `validManagedJob`.
- **`TestControllerWorkspaceProbeAuthorityDurableAndReplay`** — Regression contract: TestControllerWorkspaceProbeAuthorityDurableAndReplay.
  Signature: `func(t *testing.T)`.
  Calls: `auditSummary`, `context.WithValue`, `controllerProbeSlots`, `controllerWorkspaceAuthorityFixture`, `db.snapshot`, `g.Acknowledge`, `g.Authorize`, `g.AuthorizeWorkspaceObservation`, `json.Unmarshal`, `len`, `newControllerWorkspaceProbeAuthority`, `probeControllerState`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestControllerWorkspaceProbeAuthorityRefusalAndFencing`** — Regression contract: TestControllerWorkspaceProbeAuthorityRefusalAndFencing.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `controllerWorkspaceAuthorityFixture`, `db.seed`, `db.snapshot`, `errors.New`, `g.Authorize`, `g.AuthorizeWorkspaceObservation`, `len`, `newControllerWorkspaceProbeAuthority`, `probeControllerState`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerWorkspaceProbeStagePinsWorkerInsideLedger`** — Regression contract: TestControllerWorkspaceProbeStagePinsWorkerInsideLedger.
  Signature: `func(t *testing.T)`.
  Calls: `controllerWorkspaceAuthorityFixture`, `db.seed`, `g.Acknowledge`, `newControllerWorkspaceProbeAuthority`, `probeControllerState`, `t.Fatal`, `t.Parallel`.
- **`TestControllerWorkspaceProbeChannelIntegrated`** — Regression contract: TestControllerWorkspaceProbeChannelIntegrated.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `controllerProbeSlots`, `db.seed`, `f.Acknowledge`, `f.checkLock`, `len`, `make`, `mockServer`, `net.Pipe`, `newAssembledProbeFixture`, `newControllerWorkspaceProbeAuthority`, `probeControllerState`, `probeRetainedWorkerWorkspaceChannel`, `serveProbeAuthority`, `strings.Repeat`, `t.Context`, `t.Error`, `t.Fatal`, `t.Parallel`, `workspacemount.ValidProof`.
- **`TestControllerWorkspaceProbeHumanPermissionsAndRevocation`** — Regression contract: TestControllerWorkspaceProbeHumanPermissionsAndRevocation.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `context.WithValue`, `controllerWorkspaceAuthorityFixture`, `db.seed`, `db.snapshot`, `g.Authorize`, `len`, `newControllerWorkspaceProbeAuthority`, `probeControllerState`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerWorkspaceProbeInternalOnlyAndFeatureIndependence`** — Regression contract: TestControllerWorkspaceProbeInternalOnlyAndFeatureIndependence.
  Signature: `func(t *testing.T)`.
  Calls: `controllerWorkspaceAuthorityFixture`, `db.seed`, `fmt.Sprint`, `g.Authorize`, `mockRequest`, `mockServer`, `newControllerWorkspaceProbeAuthority`, `probeControllerState`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerWorkspaceProbeChannelAuditFailureWithholdsProof`** — Regression contract: TestControllerWorkspaceProbeChannelAuditFailureWithholdsProof.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `controllerProbeSlots`, `db.seed`, `f.Acknowledge`, `f.checkLock`, `make`, `mockServer`, `net.Pipe`, `newAssembledProbeFixture`, `newControllerWorkspaceProbeAuthority`, `probeControllerState`, `probeRetainedWorkerWorkspaceChannel`, `serveProbeAuthority`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerProbeStopped`.
