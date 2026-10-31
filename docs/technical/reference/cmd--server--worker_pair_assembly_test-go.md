# cmd/server/worker_pair_assembly_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_pair_assembly_test.go)

Automated regression checks for worker_pair_assembly; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `assembledToolFake`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ *pairLaunchFake }`.
- **`(assembledToolFake).Create`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) (string, error)`.
  Calls: `f.pairLaunchFake.Create`.
- **`(assembledToolFake).VerifyCreated`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) error`.
  Calls: `f.Verify`.
- **`(assembledToolFake).Bootstrap`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, in io.ReadCloser) (workspacebootstrap.Receipt, error)`.
  Calls: `f.pairLaunchFake.Bootstrap`, `in.Close`, `strings.Repeat`.
- **Type `pairBundleFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { io.Reader closes atomic.Int32 }`.
- **`(*pairBundleFixture).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `b.closes.Add`.
- **`assembledPairFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*workerPairSupervisor, *workerAssembledPair, *pairLaunchFake, *pairBundleFixture)`.
  Calls: `strings.NewReader`, `t.Helper`, `workspacePair`.
- **`TestAssembledPairRoutesSupervisorStartupToSharedToolRuntime`** — Regression contract: TestAssembledPairRoutesSupervisorStartupToSharedToolRuntime.
  Signature: `func(t *testing.T)`.
  Calls: `a.Bootstrap`, `a.Start`, `assembledPairFixture`, `input.closes.Load`, `len`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestAssembledPairCannotStartEngineAfterFailedBootstrap`** — Regression contract: TestAssembledPairCannotStartEngineAfterFailedBootstrap.
  Signature: `func(t *testing.T)`.
  Calls: `a.Start`, `assembledPairFixture`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAssembledPairRejectsForeignScopeAndUnknownIdentities`** — Regression contract: TestAssembledPairRejectsForeignScopeAndUnknownIdentities.
  Signature: `func(t *testing.T)`.
  Calls: `a.Bootstrap`, `a.Create`, `a.Inspect`, `a.Start`, `assembledPairFixture`, `len`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestAssembledPairOriginalCancellationBlocksFreshContext`** — Regression contract: TestAssembledPairOriginalCancellationBlocksFreshContext.
  Signature: `func(t *testing.T)`.
  Calls: `a.Start`, `a.WorkspaceIntent`, `assembledPairFixture`, `context.Background`, `len`, `s.Close`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestAssemblePairBindsRealAuditedSessionAndClosesUnusedBundle`** — Regression contract: TestAssemblePairBindsRealAuditedSessionAndClosesUnusedBundle.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `input.closes.Load`, `net.Pipe`, `newWorkerToolSession`, `otherInput.closes.Load`, `peer.Close`, `s.AssemblePair`, `s.Close`, `session.Close`, `session.Scope`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `time.After`, `time.Now`, `workspacePair`.
- **`TestAssembledPairUncertainCreateAndStartAreNeverRetried`** — Regression contract: TestAssembledPairUncertainCreateAndStartAreNeverRetried.
  Signature: `func(t *testing.T)`.
  Calls: `a.Create`, `a.Start`, `assembledPairFixture`, `len`, `s.Start`, `strings.HasPrefix`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAssemblePairRejectsDifferentAuditedSessionScope`** — Regression contract: TestAssemblePairRejectsDifferentAuditedSessionScope.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `input.closes.Load`, `net.Pipe`, `newWorkerToolSession`, `peer.Close`, `s.AssemblePair`, `session.Close`, `strings.NewReader`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `time.Now`, `workspacePair`.
