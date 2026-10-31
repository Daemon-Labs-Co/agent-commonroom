# cmd/server/workspace_probe_physical_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_probe_physical_test.go)

Automated regression checks for workspace_probe_physical; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`physicalControllerFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, f *assembledProbeFixture) (*server, *mockDB, context.Context)`.
  Calls: `context.WithValue`, `db.seed`, `db.snapshot`, `json.Unmarshal`, `mockServer`, `t.Context`, `t.Helper`.
- **`TestWorkspaceProbePhysicalControllerLifecycle`** — Regression contract: TestWorkspaceProbePhysicalControllerLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `controllerProbeSlots`, `escaped.ValidateWorkspaceProbe`, `escaped.lifetime.Err`, `left.Close`, `len`, `make`, `net.Pipe`, `newAssembledProbeFixture`, `newControllerWorkspaceProbeAuthority`, `newProbeAuthorityClient`, `newWorkspaceProbePhysicalEvidence`, `physicalControllerFixture`, `probeControllerState`, `probeRetainedWorkerWorkspaceAuthority`, `readWorkerProbeRecords`, `right.Close`, `serveProbeAuthority`, `stopped.Add`, `stopped.Load`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerProbeSettled`, `workspacemount.ValidProof`.
- **`TestWorkspaceProbePhysicalEvidenceNeedsActualOwnership`** — Regression contract: TestWorkspaceProbePhysicalEvidenceNeedsActualOwnership.
  Signature: `func(t *testing.T)`.
  Calls: `e.lifetime.Err`, `newAssembledProbeFixture`, `newWorkspaceProbePhysicalEvidence`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `withWorkerWorkspaceObservation`.
- **`TestWorkspaceProbePhysicalRejectsSubstitutionAndFalseAbsence`** — Regression contract: TestWorkspaceProbePhysicalRejectsSubstitutionAndFalseAbsence.
  Signature: `func(t *testing.T)`.
  Calls: `e.ValidateProbe`, `e.ValidateWorkspaceProbe`, `filepath.Join`, `json.Marshal`, `newAssembledProbeFixture`, `newWorkspaceProbePhysicalEvidence`, `physicalControllerFixture`, `probeControllerState`, `probeRecordName`, `replacePrivateDurably`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `withWorkerWorkspaceObservation`.
