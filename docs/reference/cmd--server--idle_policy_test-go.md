# cmd/server/idle_policy_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/idle_policy_test.go)

Automated regression checks for idle_policy; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`idleSettingsHTTP`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*server, *mockDB, http.Handler)`.
  Calls: `http.NewServeMux`, `mockServer`, `mux.HandleFunc`, `s.actorGuard`, `s.auth.protect`, `t.Helper`.
- **`TestOwnerIdleSettingsPermissionsAndRevision`** — Regression contract: TestOwnerIdleSettingsPermissionsAndRevision.
  Signature: `func(t *testing.T)`.
  Calls: `(&server{db: db}).load`, `db.seed`, `db.snapshot`, `idleSettingsHTTP`, `len`, `mockRequest`, `r.Body.String`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestIdlePolicySnapshotAndWorkerBinding`** — Regression contract: TestIdlePolicySnapshotAndWorkerBinding.
  Signature: `func(t *testing.T)`.
  Calls: `admitWorkerPair`, `context.WithValue`, `db.seed`, `json.Marshal`, `json.RawMessage`, `mockServer`, `registration.Register`, `s.load`, `s.managedExecutionJobContextVersion`, `s.managedExecutionJobContextVersion(ctx, p, repo, func() (string, error) { return repo, nil }, workerWorkspaceLifecycleVersion).Value`, `strings.Repeat`, `supervisor.Close`, `supervisor.configureJobIdle`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestIdleSettingsAuditFailureAndCSRF`** — Regression contract: TestIdleSettingsAuditFailureAndCSRF.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `h.ServeHTTP`, `httptest.NewRecorder`, `httptest.NewRequest`, `idlePolicyForState`, `idleSettingsHTTP`, `len`, `mockRequest`, `request.AddCookie`, `s.load`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestCodexAssemblyConsumesRegisteredIdlePolicy`** — Regression contract: TestCodexAssemblyConsumesRegisteredIdlePolicy.
  Signature: `func(t *testing.T)`.
  Calls: `bundle.closes.Load`, `context.WithValue`, `net.Pipe`, `newWorkerToolSession`, `newWorkspaceCleanup`, `peer.Close`, `s.RunCodex`, `session.Close`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `time.Now`, `workspacePair`.
- **`TestIdleFeatureFlagCommandsAndOffConfiguration`** — Regression contract: TestIdleFeatureFlagCommandsAndOffConfiguration.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `db.snapshot`, `idlePolicyForState`, `idleSettingsHTTP`, `json.Marshal`, `len`, `mockRequest`, `s.load`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`.
- **`TestIdleSettingsConcurrentControllers`** — Regression contract: TestIdleSettingsConcurrentControllers.
  Signature: `func(t *testing.T)`.
  Calls: `close`, `http.NewServeMux`, `idleSettingsHTTP`, `make`, `mockRequest`, `mux.HandleFunc`, `other.actorGuard`, `other.auth.protect`, `s.load`, `t.Context`, `t.Fatal`, `t.Parallel`, `wg.Add`, `wg.Done`, `wg.Wait`.
