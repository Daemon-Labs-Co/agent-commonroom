# cmd/server/development_workspaces_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/development_workspaces_test.go)

Automated regression checks for development_workspaces; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestDevelopmentRoomFlagAndAllowlist`** — Regression contract: TestDevelopmentRoomFlagAndAllowlist.
  Signature: `func(t *testing.T)`.
  Calls: `controllerDevelopmentRepository`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestDevelopmentWorkerIndependentAuthorizationAndIntegrity`** — Regression contract: TestDevelopmentWorkerIndependentAuthorizationAndIntegrity.
  Signature: `func(t *testing.T)`.
  Calls: `filepath.EvalSymlinks`, `filepath.Join`, `hex.EncodeToString`, `json.Marshal`, `os.Chmod`, `os.Symlink`, `os.WriteFile`, `sha256.Sum256`, `string`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `workerDevelopmentBundle`.
- **`TestDevelopmentProtocolRequiresSnapshotAndExactToolPolicy`** — Regression contract: TestDevelopmentProtocolRequiresSnapshotAndExactToolPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `encodeWorkerRequest`, `providerTurnArgs`, `sampleWorkerRequest`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestDevelopmentMountPreservesContainment`** — Regression contract: TestDevelopmentMountPreservesContainment.
  Signature: `func(t *testing.T)`.
  Calls: `developmentRunArgs`, `isolatedRunArgs`, `providerTurnArgs`, `strings.Contains`, `strings.Join`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestDevelopmentSessionAuthorityNeverCrossesModesOrSnapshots`** — Regression contract: TestDevelopmentSessionAuthorityNeverCrossesModesOrSnapshots.
  Signature: `func(t *testing.T)`.
  Calls: `agentExecutionSession`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestDevelopmentConfigurationPersistenceAndEnvironmentOverride`** — Regression contract: TestDevelopmentConfigurationPersistenceAndEnvironmentOverride.
  Signature: `func(t *testing.T)`.
  Calls: `controllerDevelopmentRepository`, `filepath.Join`, `loadControllerDevelopmentConfiguration`, `os.Chmod`, `os.WriteFile`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
