# cmd/server/managed_pair_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/managed_pair_test.go)

Automated regression checks for managed_pair; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`managedPairFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler, managedJobRegistration)`.
  Calls: `managedPairFixtureVersion`.
- **`managedPairFixtureVersion`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, version int) (*server, *mockDB, http.Handler, managedJobRegistration)`.
  Calls: `context.WithValue`, `ctx.Value`, `db.seed`, `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `registration.Register`, `s.actorGuard`, `s.auth.protect`, `s.managedExecutionJobContextVersion`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`.
- **`TestManagedPairJobPreservesVersionRepositoryAndPrincipal`** — Regression contract: TestManagedPairJobPreservesVersionRepositoryAndPrincipal.
  Signature: `func(t *testing.T)`.
  Calls: `managedJobFence`, `managedPairFixture`, `registration.Update`, `request.record`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestManagedPairRegistrationRechecksPolicyAndRequiresHuman`** — Regression contract: TestManagedPairRegistrationRechecksPolicyAndRequiresHuman.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `context.WithValue`, `ctx.Value`, `db.seed`, `errors.New`, `len`, `mockServer`, `reg.Register`, `s.load`, `s.managedExecutionJobContext`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestManagedPairRecoveryUsesMatchingWorkerVersionAndDigest`** — Regression contract: TestManagedPairRecoveryUsesMatchingWorkerVersionAndDigest.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `errors.New`, `managedJobFence`, `managedJobFence(job).record`, `managedPairFixture`, `mockRequest`, `response.Body.String`, `s.load`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Run`, `t.Setenv`, `view.Deadline.Add`.
- **`TestManagedPairProjectionRejectsVersionAndRepositoryChanges`** — Regression contract: TestManagedPairProjectionRejectsVersionAndRepositoryChanges.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `managedPairFixture`, `s.load`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestManagedPairAuditCannotFollowNewRepositoryPolicy`** — Regression contract: TestManagedPairAuditCannotFollowNewRepositoryPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `managedPairFixture`, `newControllerToolAuditor`, `registration.Update`, `s.load`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestPairedRegistrationCannotUseLegacyRemoteLaunch`** — Regression contract: TestPairedRegistrationCannotUseLegacyRemoteLaunch.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `context.WithValue`, `err.Error`, `managedPairFixture`, `providerTurnArgs`, `remoteAgentCommand`, `strings.Contains`, `t.Context`, `t.Fatal`, `t.Setenv`.
- **`TestManagedPairLifecycleUpdatesRetainInitiatingHuman`** — Regression contract: TestManagedPairLifecycleUpdatesRetainInitiatingHuman.
  Signature: `func(t *testing.T)`.
  Calls: `context.WithValue`, `ctx.Value`, `db.seed`, `db.snapshot`, `json.Unmarshal`, `mockServer`, `registration.Register`, `registration.Update`, `s.managedExecutionJobContext`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`.
