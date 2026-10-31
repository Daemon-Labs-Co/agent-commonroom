# cmd/server/workspace_object_read_evidence_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workspace_object_read_evidence_test.go)

Automated regression checks for workspace_object_read_evidence; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticObjectReadWitness`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { operations, lifecycles int deny, revoke bool change func() }`.
- **`(*syntheticObjectReadWitness).ValidateObjectReadOperation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ room.State, r workerJobRecord, _ workerObjectReadIntent, _ workerProbeEvent) error`.
  Calls: `ctx.Err`, `w.change`.
- **`(*syntheticObjectReadWitness).ValidateObjectReadLifecycle`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, _ room.State, r workerJobRecord, local workerObjectReadRecord, _ workerProbeEvent) error`.
  Calls: `ctx.Err`, `w.change`.
- **Type `physicalObjectReadFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { probe *assembledProbeFixture intent workerObjectReadIntent phase, mode string calls, inspections, effects int }`.
- **`(*physicalObjectReadFixture).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `ctx.Err`, `f.probe.checkLock`, `hex.EncodeToString`, `io.WriteString`, `json.Marshal`, `json.NewEncoder`, `json.NewEncoder(out).Encode`, `len`, `p.inspection`, `p.t.Error`, `p.t.Fatal`, `r.UpdatedAt.Add`, `sha256.Sum256`, `strings.Repeat`, `writeWorkerRecord`.
- **`objectReadPhysicalState`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, j *workerObjectReadJournal, p *assembledProbeFixture) room.State`.
  Calls: `physicalControllerFixture`, `probeControllerState`, `t.Helper`.
- **`TestWorkspaceObjectReadEvidenceOperation`** — Regression contract: TestWorkspaceObjectReadEvidenceOperation.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `beginObjectReadJournal`, `cancel`, `context.WithCancel`, `detachedProbeWorkerRecord`, `e.ValidateObjectReadOperation`, `filepath.Join`, `j.BeginObjectReadRecovery`, `j.BindObjectRead`, `j.FinishObjectRead`, `json.Marshal`, `newWorkspaceObjectReadEvidence`, `objectReadJournalFixture`, `objectReadPhysicalState`, `objectReadRecordName`, `os.WriteFile`, `readWorkerObjectReadRecords`, `strings.HasPrefix`, `strings.HasSuffix`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkspaceObjectReadEvidenceLifecycle`** — Regression contract: TestWorkspaceObjectReadEvidenceLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `beginObjectReadJournal`, `cancel`, `context.WithCancel`, `e.ValidateObjectReadTransition`, `filepath.Join`, `j.BeginObjectReadRecovery`, `j.BindObjectRead`, `j.FinishObjectRead`, `j.FinishObjectReadRecovery`, `json.Marshal`, `len`, `newWorkspaceObjectReadEvidence`, `objectReadControllerEvents`, `objectReadJournalFixture`, `objectReadPhysicalState`, `objectReadRecordName`, `os.WriteFile`, `readWorkerObjectReadRecords`, `releasedProbeEvents`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestWorkspaceObjectReadEvidenceRequiresDependencies`** — Regression contract: TestWorkspaceObjectReadEvidenceRequiresDependencies.
  Signature: `func(t *testing.T)`.
  Calls: `e.ValidateObjectReadOperation`, `e.ValidateObjectReadTransition`, `newWorkspaceObjectReadEvidence`, `objectReadJournalFixture`, `t.Context`, `t.Fatal`.
