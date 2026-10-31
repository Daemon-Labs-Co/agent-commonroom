# cmd/server/worker_probe_transport_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_probe_transport_test.go)

Automated regression checks for worker_probe_transport; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `transportProbeGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls []string deny string }`.
- **`(*transportProbeGuard).Authorize`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent) error`.
  Calls: `append`, `ctx.Err`, `errors.New`.
- **`(*transportProbeGuard).Acknowledge`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i toolbroker.WorkspaceProbeIntent, e workerProbeEvent) error`.
  Calls: `append`, `ctx.Err`, `errors.New`.
- **`(*transportProbeGuard).AuthorizeWorkspaceObservation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord) error`.
  Calls: `append`, `ctx.Err`, `errors.New`.
- **`(*transportProbeGuard).StopAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`transportProbeBinding`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (workerJobRecord, toolbroker.WorkspaceProbeIntent)`.
  Calls: `retainedObservationFixture`, `strings.Repeat`.
- **`TestProbeAuthorityTransportRoundTrip`** — Regression contract: TestProbeAuthorityTransportRoundTrip.
  Signature: `func(t *testing.T)`.
  Calls: `call`, `cancel`, `client.Acknowledge`, `client.Authorize`, `client.AuthorizeWorkspaceObservation`, `context.WithCancel`, `left.Close`, `make`, `net.Pipe`, `newProbeAuthorityClient`, `serveProbeAuthority`, `stopped.Add`, `stopped.Load`, `strings.Join`, `t.Context`, `t.Fatal`, `t.Parallel`, `transportProbeBinding`.
- **`TestProbeAuthorityDenialPoisonsLink`** — Regression contract: TestProbeAuthorityDenialPoisonsLink.
  Signature: `func(t *testing.T)`.
  Calls: `client.Acknowledge`, `client.Authorize`, `client.AuthorizeWorkspaceObservation`, `client.StopAuthority`, `make`, `net.Pipe`, `newProbeAuthorityClient`, `serveProbeAuthority`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `transportProbeBinding`.
- **`TestProbeAuthorityRejectsInvalidFramesBeforeGuard`** — Regression contract: TestProbeAuthorityRejectsInvalidFramesBeforeGuard.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `binary.BigEndian.PutUint32`, `left.Close`, `left.Write`, `len`, `make`, `net.Pipe`, `serveProbeAuthority`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `transportProbeBinding`, `uint32`, `wire.Bytes`, `wire.Reset`, `wire.Write`, `writeProbeAuthorityFrame`.
- **`TestProbeAuthorityCancellationAndWrongResponse`** — Regression contract: TestProbeAuthorityCancellationAndWrongResponse.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `client.Authorize`, `close`, `context.WithCancel`, `make`, `net.Pipe`, `newProbeAuthorityClient`, `readProbeAuthorityFrame`, `right.Close`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `transportProbeBinding`, `uint64`, `writeProbeAuthorityFrame`.
- **`TestProbeAuthorityChannelRunsRealProbeLifecycle`** — Regression contract: TestProbeAuthorityChannelRunsRealProbeLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `make`, `net.Pipe`, `newAssembledProbeFixture`, `probeRetainedWorkerWorkspaceChannel`, `serveProbeAuthority`, `stopped.Add`, `stopped.Load`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `workerProbeStopped`, `workspacemount.ValidProof`.
- **`TestProbeAuthorityReplayDoesNotRepeatGuard`** — Regression contract: TestProbeAuthorityReplayDoesNotRepeatGuard.
  Signature: `func(t *testing.T)`.
  Calls: `left.Close`, `len`, `make`, `net.Pipe`, `readProbeAuthorityFrame`, `serveProbeAuthority`, `t.Context`, `t.Fatal`, `t.Parallel`, `transportProbeBinding`, `writeProbeAuthorityFrame`.
- **`TestProbeAuthorityRejectsUnboundRecord`** — Regression contract: TestProbeAuthorityRejectsUnboundRecord.
  Signature: `func(t *testing.T)`.
  Calls: `client.Authorize`, `client.AuthorizeWorkspaceObservation`, `left.Close`, `net.Pipe`, `newProbeAuthorityClient`, `right.Close`, `serveProbeAuthority`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `transportProbeBinding`.
