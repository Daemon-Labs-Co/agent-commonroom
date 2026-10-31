# cmd/server/controller_artifact_read_session_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/controller_artifact_read_session_test.go)

Automated regression checks for controller_artifact_read_session; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`artifactReadSessionFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*controllerArtifactReadSession, *mockDB, context.Context, *objectcustody.Store, *syntheticArtifactGuard)`.
  Calls: `artifactJournalFixture`, `artifactStore`, `newControllerArtifactJournal`, `newControllerArtifactReadSession`, `store.Put`, `strings.NewReader`, `t.Fatal`, `t.Helper`.
- **`TestControllerArtifactReadSessionPrivateByteTransfer`** — Regression contract: TestControllerArtifactReadSessionPrivateByteTransfer.
  Signature: `func(t *testing.T)`.
  Calls: `artifactReadSessionFixture`, `artifactState`, `context.WithValue`, `db.seed`, `db.snapshot`, `make`, `net.Pipe`, `objectcustody.ReadChannel`, `objectcustody.ServeReadChannel`, `string`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerArtifactReadSessionCurrentPolicy`** — Regression contract: TestControllerArtifactReadSessionCurrentPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `artifactReadSessionFixture`, `cancel`, `context.WithCancel`, `db.seed`, `db.snapshot`, `errors.New`, `g.AcknowledgeRead`, `g.AuthorizeRead`, `len`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestControllerArtifactReadSessionAuditFailureAndStop`** — Regression contract: TestControllerArtifactReadSessionAuditFailureAndStop.
  Signature: `func(t *testing.T)`.
  Calls: `artifactReadSessionFixture`, `g.AcknowledgeRead`, `g.AuthorizeRead`, `g.EstablishQuota`, `g.StopReadAuthority`, `g.ValidateSeal`, `g.ValidateSource`, `newControllerArtifactReadSession`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestControllerArtifactReadSessionPrivateJournalRelay`** — Regression contract: TestControllerArtifactReadSessionPrivateJournalRelay.
  Signature: `func(t *testing.T)`.
  Calls: `artifactJournalFixture`, `artifactState`, `artifactStore`, `cancel`, `context.WithCancel`, `make`, `net.Pipe`, `newControllerArtifactJournal`, `newControllerArtifactReadSession`, `objectcustody.New`, `objectcustody.NewReadJournalChannel`, `objectcustody.ReadChannel`, `objectcustody.ServeReadChannel`, `os.OpenRoot`, `relay.Close`, `root.Close`, `store.Put`, `string`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
