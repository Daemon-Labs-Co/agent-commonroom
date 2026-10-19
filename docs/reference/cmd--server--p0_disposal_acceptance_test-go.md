# cmd/server/p0_disposal_acceptance_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/p0_disposal_acceptance_test.go)

Automated regression checks for p0_disposal_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`driveP0DisposalRefusal`** — Negative cases drive the real owner HTTP boundary and observed journal/ledger state. Positive disposal separately exercises worker receipt and credit integration.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `f.calls.Load`, `f.db.seed`, `f.db.snapshot`, `f.now.Store`, `f.post`, `f.propose`, `f.s.load`, `len`, `managedStorageTotal`, `newDisposalFixture`, `p.NotBefore.UnixNano`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `t.Setenv`, `time.Now`, `time.Now().Add`.
- **`driveP0ApprovedDisposal`** — Exercises the owner HTTP flow, durable worker journal, creator ownership, exact receipt validation and controller refund with mocked Docker and database. This is hermetic component integration, not live filesystem quota evidence.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `(&server{db: f.db}).load`, `append`, `approved.ClaimedAt.Before`, `approved.ReviewObservedAt.Add`, `errors.New`, `f.approve`, `f.db.snapshot`, `f.now.Store`, `f.propose`, `f.s.load`, `fenceWorkerJobContainers`, `len`, `managedStorageTotal`, `min`, `newDisposalFixture`, `newWorkspaceCleanup`, `original`, `p.NotBefore.UnixNano`, `readWorkerRecord`, `request.record`, `stopWorkerPair`, `string`, `t.Context`, `t.Fatal`, `t.Helper`, `t.Setenv`, `t.TempDir`, `validWorkerFenceAck`, `workerFenceAcknowledgement`, `writeWorkerRecord`.
