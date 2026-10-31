# cmd/server/object_read_authenticated_session_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/object_read_authenticated_session_test.go)

Automated regression checks for object_read_authenticated_session; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`boundObjectReadFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, g *controllerObjectReadSession) (objectReadPeerBinding, objectReadPeerBinding)`.
  Calls: `objectReadTestPeers`, `probeWorkerRecordDigest`, `t.Helper`.
- **`TestAuthenticatedObjectReadSessionLifecycle`** — Regression contract: TestAuthenticatedObjectReadSessionLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `append`, `artifactState`, `b.Close`, `boundObjectReadFixture`, `c.AcknowledgeObjectRead`, `c.AuthorizeObjectRead`, `c.StopObjectReadAuthority`, `controllerObjectReadKey`, `db.snapshot`, `g.lifetime.Err`, `len`, `make`, `net.Pipe`, `newAuthenticatedObjectReadClient`, `objectReadSessionFixture`, `serveAuthenticatedObjectReadSession`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAuthenticatedObjectReadSessionBindingRefusals`** — Regression contract: TestAuthenticatedObjectReadSessionBindingRefusals.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `b.Close`, `boundObjectReadFixture`, `c.StopObjectReadAuthority`, `cancel`, `context.WithTimeout`, `db.snapshot`, `len`, `make`, `net.Pipe`, `newAuthenticatedObjectReadClient`, `objectReadSessionFixture`, `serveAuthenticatedObjectReadSession`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestAuthenticatedObjectReadSessionIdleCancellation`** — Regression contract: TestAuthenticatedObjectReadSessionIdleCancellation.
  Signature: `func(t *testing.T)`.
  Calls: `a.Close`, `b.Close`, `boundObjectReadFixture`, `c.ObjectReadAuthorityLifetime`, `c.ObjectReadAuthorityLifetime().Done`, `c.StopObjectReadAuthority`, `cancel`, `context.WithCancel`, `db.snapshot`, `g.StopObjectReadAuthority`, `g.lifetime.Err`, `len`, `make`, `net.Pipe`, `newAuthenticatedObjectReadClient`, `objectReadSessionFixture`, `serveAuthenticatedObjectReadSession`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`.
