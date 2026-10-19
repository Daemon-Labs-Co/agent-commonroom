# cmd/server/worker_disposal_binding_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_disposal_binding_test.go)

Automated regression checks for worker_disposal_binding; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestWorkspaceDiscardRequiresCompleteExactIdentityBeforeEffects`** — Regression contract: TestWorkspaceDiscardRequiresCompleteExactIdentityBeforeEffects.
  Signature: `func(t *testing.T)`.
  Calls: `fenceWorkerJobContainers`, `filepath.Join`, `os.IsNotExist`, `os.Remove`, `os.Stat`, `s.Close`, `s.Start`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`, `time.Now().Add(time.Hour).UTC`, `time.Now().Add(time.Hour).UTC().Format`, `workspaceFence`, `workspacePair`.
- **`TestWorkspaceDiscardRechecksIdentityAfterCreatorLeaseWait`** — Regression contract: TestWorkspaceDiscardRechecksIdentityAfterCreatorLeaseWait.
  Signature: `func(t *testing.T)`.
  Calls: `ctx.Value`, `fenceWorkerJobContainers`, `s.Close`, `s.Start`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `workspaceFence`, `workspacePair`, `writeWorkerRecord`.
- **`TestWorkspaceDiscardAcknowledgementBindsRemovedIdentity`** — Regression contract: TestWorkspaceDiscardAcknowledgementBindsRemovedIdentity.
  Signature: `func(t *testing.T)`.
  Calls: `fenceWorkspace`, `newWorkspaceCleanup`, `request.record`, `s.Close`, `s.Start`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `time.Now().Add(time.Hour).UTC`, `time.Now().Add(time.Hour).UTC().Format`, `validWorkerFenceAck`, `workerFenceAcknowledgement`, `workspaceFence`, `workspacePair`.
- **`TestFenceWireRefusesAmbiguousDisposalJSON`** — Regression contract: TestFenceWireRefusesAmbiguousDisposalJSON.
  Signature: `func(t *testing.T)`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `decodeFenceRequest`, `json.Marshal`, `s.Close`, `s.Start`, `string`, `strings.Replace`, `t.Context`, `t.Fatal`, `t.Parallel`, `validFenceWire`, `workerFenceAcknowledgement`, `workspaceFence`, `workspacePair`.
