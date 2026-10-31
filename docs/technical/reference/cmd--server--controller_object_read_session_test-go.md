# cmd/server/controller_object_read_session_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_object_read_session_test.go)

Automated regression checks for controller_object_read_session; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `syntheticObjectReadSessionEvidence`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { operation func(context.Context, room.State, workerJobRecord, workerObjectReadIntent, workerProbeEvent) error transition func(context.Context, room.State, room.ObjectReadReservation, workerProbeEvent) error calls, acks int }`.
- **`(*syntheticObjectReadSessionEvidence).ValidateObjectReadOperation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r workerJobRecord, i workerObjectReadIntent, v workerProbeEvent) error`.
  Calls: `ctx.Err`, `e.operation`.
- **`(*syntheticObjectReadSessionEvidence).ValidateObjectReadTransition`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, r room.ObjectReadReservation, v workerProbeEvent) error`.
  Calls: `ctx.Err`, `e.transition`.
- **`objectReadSessionFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*controllerObjectReadSession, *mockDB, context.Context, *syntheticObjectReadSessionEvidence)`.
  Calls: `artifactState`, `newControllerObjectReadSession`, `objectReadLedgerFixture`, `retainedObservationFixture`, `t.Fatal`, `t.Helper`, `workerMembership`.
- **`TestControllerObjectReadSessionPinnedLifecycle`** — Regression contract: TestControllerObjectReadSessionPinnedLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `artifactState`, `context.WithValue`, `controllerObjectReadKey`, `db.snapshot`, `g.AcknowledgeObjectRead`, `g.AuthorizeObjectRead`, `len`, `objectReadSessionFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerObjectReadSessionRejectsHistorySubstitution`** — Regression contract: TestControllerObjectReadSessionRejectsHistorySubstitution.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `db.snapshot`, `g.AcknowledgeObjectRead`, `len`, `objectReadSessionFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerObjectReadSessionPolicyAndLifetime`** — Regression contract: TestControllerObjectReadSessionPolicyAndLifetime.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `db.seed`, `db.snapshot`, `detachedProbeWorkerRecord`, `errors.Is`, `errors.New`, `g.AcknowledgeObjectRead`, `g.AuthorizeObjectRead`, `len`, `objectReadSessionFixture`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerObjectReadSessionCleanupAndPersistenceFault`** — Regression contract: TestControllerObjectReadSessionCleanupAndPersistenceFault.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `db.seed`, `g.AcknowledgeObjectRead`, `g.AuthorizeObjectRead`, `objectReadSessionFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerObjectReadSessionDetachesEvidenceAndStops`** — Regression contract: TestControllerObjectReadSessionDetachesEvidenceAndStops.
  Signature: `func(t *testing.T)`.
  Calls: `g.AuthorizeObjectRead`, `g.StopObjectReadAuthority`, `newControllerObjectReadSession`, `objectReadSessionFixture`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestControllerObjectReadSessionCanceledFinishCannotRelease`** — Regression contract: TestControllerObjectReadSessionCanceledFinishCannotRelease.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `artifactState`, `controllerObjectReadSlots`, `copyObjectReadRecord`, `db.snapshot`, `g.AcknowledgeObjectRead`, `g.AuthorizeObjectRead`, `g.StopObjectReadAuthority`, `len`, `objectReadSessionFixture`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestControllerObjectReadSessionThroughWorkerJournal`** — Regression contract: TestControllerObjectReadSessionThroughWorkerJournal.
  Signature: `func(t *testing.T)`.
  Calls: `artifactState`, `c.StopObjectReadAuthority`, `controllerObjectReadSlots`, `detachedProbeWorkerRecord`, `filepath.Join`, `g.lifetime.Err`, `j.BeginObjectRead`, `j.BindObjectRead`, `j.FinishObjectRead`, `lease.Close`, `len`, `make`, `net.Pipe`, `newObjectReadAuthorityClient`, `newWorkerObjectReadJournal`, `objectReadRecordName`, `objectReadSessionFixture`, `os.Chmod`, `os.WriteFile`, `readWorkerObjectReadRecords`, `right.Close`, `serveObjectReadAuthority`, `slices.Equal`, `stopped.Add`, `stopped.Load`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workerObjectReadSlots`, `workerProbeStopped`, `workerTryLease`, `writeWorkerRecord`.
- **Type `objectReadNthWriteDrop`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn writes int dropAt int }`.
- **`(*objectReadNthWriteDrop).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `errors.New`, `s.Conn.Close`, `s.Conn.Write`.
