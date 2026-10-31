# cmd/server/idle_stop_notices_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/idle_stop_notices_test.go)

Automated regression checks for idle_stop_notices; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `idleNoticeFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { s *server db *mockDB job room.ManagedJob worker *workerPairSupervisor docker *workspaceCleanupDocker announce workerIdleAnnounce now time.Time }`.
- **`newIdleNoticeFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, shortDirectory ...bool) *idleNoticeFixture`.
  Calls: `newIdleNoticeIntervalFixture`, `t.Helper`.
- **`newIdleNoticeIntervalFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, minutes int, shortDirectory ...bool) *idleNoticeFixture`.
  Calls: `admitWorkerPair`, `context.WithValue`, `db.seed`, `f.announce`, `job.AllocatedAt.Add`, `len`, `mockServer`, `newWorkspaceCleanup`, `now.Add`, `os.MkdirTemp`, `os.RemoveAll`, `pairContainerName`, `quotaJob`, `s.newIdleAnnouncer`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`, `t.TempDir`, `time.Duration`, `time.Now`, `time.Now().UTC`, `worker.Start`, `worker.configureJobIdle`, `worker.ctx.Err`, `worker.mu.Lock`, `worker.mu.Unlock`, `writeWorkerRecord`.
- **`TestIdleAnnouncementPrecedesCleanupAndRetainsEvidence`** — Regression contract: TestIdleAnnouncementPrecedesCleanupAndRetainsEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: f.db}).load`, `f.announce`, `f.db.snapshot`, `f.worker.expireIdle`, `len`, `newIdleNoticeFixture`, `record.IdleShutdownRequestedAt.Add`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestIdleAnnouncementFailuresDoNotAuthorizeCleanup`** — Regression contract: TestIdleAnnouncementFailuresDoNotAuthorizeCleanup.
  Signature: `func(t *testing.T)`.
  Calls: `errors.New`, `f.now.Add`, `f.s.load`, `f.worker.expireIdle`, `f.worker.mu.Lock`, `f.worker.mu.Unlock`, `len`, `newIdleNoticeFixture`, `original`, `panic`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestIdleNoticeVisibilityCannotDisableAudit`** — Regression contract: TestIdleNoticeVisibilityCannotDisableAudit.
  Signature: `func(t *testing.T)`.
  Calls: `f.db.seed`, `f.s.load`, `f.worker.expireIdle`, `len`, `newIdleNoticeFixture`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestPublicEventsCannotForgeIdleNotice`** — Regression contract: TestPublicEventsCannotForgeIdleNotice.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `idleSettingsHTTP`, `len`, `mockRequest`, `t.Fatal`, `t.Parallel`.
- **`TestIdleNoticeReplayRejectsChangedOrEarlyEvidence`** — Regression contract: TestIdleNoticeReplayRejectsChangedOrEarlyEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `f.job.AllocatedAt.Add`, `f.now.Add`, `f.s.load`, `id`, `json.Marshal`, `n.WorkerStartedAt.Add`, `newIdleNoticeFixture`, `projectIdleNotice`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestIdleNoticeConstructorPermissionsAndFlagCommands`** — Regression contract: TestIdleNoticeConstructorPermissionsAndFlagCommands.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `f.db.seed`, `f.db.snapshot`, `f.s.actorGuard`, `f.s.auth.protect`, `f.s.newIdleAnnouncer`, `http.NewServeMux`, `len`, `mockRequest`, `mux.HandleFunc`, `newIdleNoticeFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestIdleNoticeConcurrentControllersDeduplicate`** — Regression contract: TestIdleNoticeConcurrentControllersDeduplicate.
  Signature: `func(t *testing.T)`.
  Calls: `callback`, `close`, `context.WithValue`, `f.db.snapshot`, `len`, `make`, `newIdleNoticeFixture`, `other.newIdleAnnouncer`, `second`, `t.Context`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
- **`TestCodexRefusesUnregisteredIdleCallback`** — Regression contract: TestCodexRefusesUnregisteredIdleCallback.
  Signature: `func(t *testing.T)`.
  Calls: `bundle.closes.Load`, `context.WithValue`, `net.Pipe`, `newWorkerToolSession`, `newWorkspaceCleanup`, `peer.Close`, `session.Close`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `worker.RunCodex`, `workspacePair`.
