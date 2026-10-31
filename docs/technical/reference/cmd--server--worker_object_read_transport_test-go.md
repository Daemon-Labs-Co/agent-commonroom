# cmd/server/worker_object_read_transport_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_object_read_transport_test.go)

Automated regression checks for worker_object_read_transport; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestObjectReadAuthorityTransportJournalLifecycle`** — Regression contract: TestObjectReadAuthorityTransportJournalLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `beginObjectReadJournal`, `c.StopObjectReadAuthority`, `j.BindObjectRead`, `j.FinishObjectRead`, `make`, `net.Pipe`, `newObjectReadAuthorityClient`, `objectReadTransportFixture`, `serveObjectReadAuthority`, `stopped.Add`, `stopped.Load`, `strings.Join`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerObjectReadSlots`.
- **`TestObjectReadAuthorityTransportDenialRetainsCharge`** — Regression contract: TestObjectReadAuthorityTransportDenialRetainsCharge.
  Signature: `func(t *testing.T)`.
  Calls: `c.AuthorizeObjectRead`, `j.BeginObjectRead`, `make`, `net.Pipe`, `newObjectReadAuthorityClient`, `objectReadTransportFixture`, `serveObjectReadAuthority`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `workerObjectReadSlots`, `workerProbeStopped`.
- **`TestObjectReadAuthorityTransportRejectsFramesBeforeGuard`** — Regression contract: TestObjectReadAuthorityTransportRejectsFramesBeforeGuard.
  Signature: `func(t *testing.T)`.
  Calls: `left.Close`, `len`, `make`, `net.Pipe`, `objectReadTransportFixture`, `readObjectReadAuthorityFrame`, `serveObjectReadAuthority`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `writeObjectReadAuthorityFrame`.
- **`TestObjectReadAuthorityClientCancellationAndBadResponses`** — Regression contract: TestObjectReadAuthorityClientCancellationAndBadResponses.
  Signature: `func(t *testing.T)`.
  Calls: `binary.BigEndian.PutUint32`, `c.AuthorizeObjectRead`, `c.StopObjectReadAuthority`, `cancel`, `close`, `context.WithCancel`, `len`, `make`, `net.Pipe`, `newObjectReadAuthorityClient`, `objectReadTransportFixture`, `readObjectReadAuthorityFrame`, `right.Close`, `right.Read`, `right.Write`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `uint32`, `writeObjectReadAuthorityFrame`.
- **`TestObjectReadAuthorityFrameBoundsAndCanonicalEncoding`** — Regression contract: TestObjectReadAuthorityFrameBoundsAndCanonicalEncoding.
  Signature: `func(t *testing.T)`.
  Calls: `b.Write`, `b.WriteString`, `binary.BigEndian.PutUint32`, `bytes.NewReader`, `len`, `readObjectReadAuthorityFrame`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `uint32`, `writeObjectReadAuthorityFrame`.
- **`TestObjectReadAuthorityServerCancellationClosesOwnedStream`** — Regression contract: TestObjectReadAuthorityServerCancellationClosesOwnedStream.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `left.Close`, `make`, `net.Pipe`, `objectReadTransportFixture`, `serveObjectReadAuthority`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`.
- **Type `mutatingObjectReadTransportGuard`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { calls int stopped bool }`.
- **`(*mutatingObjectReadTransportGuard).AuthorizeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r workerJobRecord, i workerObjectReadIntent, e workerProbeEvent) error`.
  Calls: `ctx.Err`.
- **`(*mutatingObjectReadTransportGuard).AcknowledgeObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, workerObjectReadRecord, workerProbeEvent) error`.
- **`(*mutatingObjectReadTransportGuard).StopObjectReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`TestObjectReadAuthorityTransportDetachesCallbackMetadata`** — Regression contract: TestObjectReadAuthorityTransportDetachesCallbackMetadata.
  Signature: `func(t *testing.T)`.
  Calls: `c.AuthorizeObjectRead`, `c.StopObjectReadAuthority`, `make`, `net.Pipe`, `newObjectReadAuthorityClient`, `objectReadTransportFixture`, `serveObjectReadAuthority`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestObjectReadAuthorityClientRejectsSubstitutedBindings`** — Regression contract: TestObjectReadAuthorityClientRejectsSubstitutedBindings.
  Signature: `func(t *testing.T)`.
  Calls: `c.AcknowledgeObjectRead`, `c.AuthorizeObjectRead`, `detachedProbeWorkerRecord`, `net.Pipe`, `newObjectReadAuthorityClient`, `objectReadTransportFixture`, `right.Close`, `stopped.Add`, `stopped.Load`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`objectReadTransportFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*workerObjectReadJournal, *objectReadJournalGuard)`.
  Calls: `objectReadJournalFixture`, `t.Helper`.
- **`TestObjectReadAuthorityTransportFrameBudget`** — Regression contract: TestObjectReadAuthorityTransportFrameBudget.
  Signature: `func(t *testing.T)`.
  Calls: `c.AuthorizeObjectRead`, `c.StopObjectReadAuthority`, `close`, `make`, `net.Pipe`, `newObjectReadAuthorityClient`, `objectReadTransportFixture`, `serveObjectReadAuthority`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestObjectReadAuthorityIdleEOFRevokesLifetime`** — Regression contract: TestObjectReadAuthorityIdleEOFRevokesLifetime.
  Signature: `func(t *testing.T)`.
  Calls: `c.AuthorizeObjectRead`, `c.ObjectReadAuthorityLifetime`, `c.ObjectReadAuthorityLifetime().Done`, `c.StopObjectReadAuthority`, `net.Pipe`, `newObjectReadAuthorityClient`, `objectReadTransportFixture`, `right.Close`, `stopped.Add`, `stopped.Load`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.After`.
