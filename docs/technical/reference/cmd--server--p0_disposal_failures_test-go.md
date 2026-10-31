# cmd/server/p0_disposal_failures_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_disposal_failures_test.go)

Automated regression checks for p0_disposal_failures; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `disposalLostReplyDocker`**: A successful remove followed by unavailable inventory must remain uncertain. This adapter only wraps synthetic Docker state; it cannot launch a process.
  Declaration: `struct{ *workspaceCleanupDocker }`.
- **`(disposalLostReplyDocker).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `d.workspaceCleanupDocker.Run`, `len`.
- **`driveP0DisposalFailure`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `(&server{db: f.db}).load`, `append`, `f.approve`, `f.calls.Load`, `f.db.snapshot`, `f.now.Store`, `f.propose`, `fenceWorkerJobContainers`, `len`, `managedStorageTotal`, `min`, `newDisposalFixture`, `newWorkspaceCleanup`, `original`, `p.NotBefore.UnixNano`, `readWorkerRecord`, `reflect.DeepEqual`, `stopWorkerPair`, `t.Context`, `t.Fatal`, `t.Helper`, `t.Setenv`, `t.TempDir`, `writeWorkerRecord`.
- **`driveP0CancelRetainsWorkspace`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `fenceWorkerJobContainers`, `len`, `newWorkspaceCleanup`, `pairRoleSettled`, `readWorkerRecord`, `s.Close`, `s.Start`, `stopWorkerPair`, `t.Context`, `t.Fatal`, `t.Helper`, `workspaceFence`, `workspacePair`.
- **`driveP0DisposalHistoryProtection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `f.approve`, `f.db.seed`, `f.db.snapshot`, `f.now.Store`, `f.propose`, `len`, `min`, `mockRequest`, `newDisposalFixture`, `p.NotBefore.UnixNano`, `reflect.DeepEqual`, `t.Fatal`, `t.Helper`, `t.Setenv`.
