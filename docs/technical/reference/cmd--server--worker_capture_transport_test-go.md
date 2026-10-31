# cmd/server/worker_capture_transport_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_capture_transport_test.go)

Automated regression checks for worker_capture_transport; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestCaptureAuthorityTransportJournalLifecycle`** — Regression contract: TestCaptureAuthorityTransportJournalLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `beginCaptureJournal`, `c.StopCaptureAuthority`, `captureJournalFixture`, `j.BindCapture`, `j.FinishCapture`, `make`, `net.Pipe`, `newCaptureAuthorityClient`, `serveCaptureAuthority`, `stopped.Add`, `stopped.Load`, `strings.Join`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerCaptureSlots`.
- **`TestCaptureAuthorityTransportDenialRetainsCharge`** — Regression contract: TestCaptureAuthorityTransportDenialRetainsCharge.
  Signature: `func(t *testing.T)`.
  Calls: `c.AuthorizeCapture`, `captureJournalFixture`, `j.BeginCapture`, `make`, `net.Pipe`, `newCaptureAuthorityClient`, `serveCaptureAuthority`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerCaptureSlots`, `workerProbeStopped`.
- **`TestCaptureAuthorityTransportRejectsFramesBeforeGuard`** — Regression contract: TestCaptureAuthorityTransportRejectsFramesBeforeGuard.
  Signature: `func(t *testing.T)`.
  Calls: `captureJournalFixture`, `copyCaptureIntent`, `left.Close`, `len`, `make`, `net.Pipe`, `readCaptureAuthorityFrame`, `serveCaptureAuthority`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `writeCaptureAuthorityFrame`.
- **`TestCaptureAuthorityClientCancellationAndBadResponses`** — Regression contract: TestCaptureAuthorityClientCancellationAndBadResponses.
  Signature: `func(t *testing.T)`.
  Calls: `binary.BigEndian.PutUint32`, `c.AuthorizeCapture`, `c.StopCaptureAuthority`, `cancel`, `captureJournalFixture`, `close`, `context.WithCancel`, `len`, `make`, `net.Pipe`, `newCaptureAuthorityClient`, `readCaptureAuthorityFrame`, `right.Close`, `right.Read`, `right.Write`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `uint32`, `writeCaptureAuthorityFrame`.
- **`TestCaptureAuthorityFrameBoundsAndCanonicalEncoding`** — Regression contract: TestCaptureAuthorityFrameBoundsAndCanonicalEncoding.
  Signature: `func(t *testing.T)`.
  Calls: `b.Write`, `b.WriteString`, `binary.BigEndian.PutUint32`, `bytes.NewReader`, `len`, `readCaptureAuthorityFrame`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `uint32`, `writeCaptureAuthorityFrame`.
- **`TestCaptureAuthorityServerCancellationClosesOwnedStream`** — Regression contract: TestCaptureAuthorityServerCancellationClosesOwnedStream.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `captureJournalFixture`, `context.WithCancel`, `left.Close`, `make`, `net.Pipe`, `serveCaptureAuthority`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`.
- **Type `mutatingCaptureTransportGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls int stopped bool }`.
- **`(*mutatingCaptureTransportGuard).AuthorizeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord, i workerCaptureIntent, e workerProbeEvent) error`.
  Calls: `ctx.Err`.
- **`(*mutatingCaptureTransportGuard).AcknowledgeCapture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, workerCaptureRecord, workerProbeEvent) error`.
- **`(*mutatingCaptureTransportGuard).StopCaptureAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`TestCaptureAuthorityTransportDetachesCallbackMetadata`** — Regression contract: TestCaptureAuthorityTransportDetachesCallbackMetadata.
  Signature: `func(t *testing.T)`.
  Calls: `c.AuthorizeCapture`, `c.StopCaptureAuthority`, `captureJournalFixture`, `make`, `net.Pipe`, `newCaptureAuthorityClient`, `serveCaptureAuthority`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCaptureAuthorityClientRejectsSubstitutedBindings`** — Regression contract: TestCaptureAuthorityClientRejectsSubstitutedBindings.
  Signature: `func(t *testing.T)`.
  Calls: `c.AcknowledgeCapture`, `c.AuthorizeCapture`, `captureJournalFixture`, `copyCaptureIntent`, `detachedProbeWorkerRecord`, `net.Pipe`, `newCaptureAuthorityClient`, `right.Close`, `stopped.Add`, `stopped.Load`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCaptureAuthorityIdleEOFRevokesLifetimeAndJoinsPump`** — Regression contract: TestCaptureAuthorityIdleEOFRevokesLifetimeAndJoinsPump.
  Signature: `func(t *testing.T)`.
  Calls: `captureJournalFixture`, `client.AuthorizeCapture`, `client.CaptureAuthorityLifetime`, `client.CaptureAuthorityLifetime().Done`, `client.CaptureAuthorityLifetime().Err`, `client.StopCaptureAuthority`, `net.Pipe`, `newCaptureAuthorityClient`, `right.Close`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`.
