# cmd/server/worker_engine_credential_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_engine_credential_test.go)

Automated regression checks for worker_engine_credential; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`credentialLeaseFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (context.Context, context.CancelFunc, toolbroker.Scope, string, string, workerJobRecord, string)`.
  Calls: `connectionAuthKey`, `context.WithCancel`, `context.WithValue`, `filepath.Join`, `os.MkdirAll`, `os.WriteFile`, `t.Cleanup`, `t.Context`, `t.Fatal`, `t.Helper`, `t.TempDir`, `workspacePair`.
- **`TestEngineCredentialLeaseSerializesAndRequiresFencing`** — Regression contract: TestEngineCredentialLeaseSerializesAndRequiresFencing.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `bytes.Equal`, `cancel`, `context.WithValue`, `credentialLeaseFixture`, `framedSyntheticCredential`, `lease.Provision`, `lease.Release`, `len`, `lockWorkerCredential`, `out.Bytes`, `out.Len`, `out.Reset`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `unlock`.
- **`TestEngineCredentialRejectsUnsafeInputAndUnlocksFailure`** — Regression contract: TestEngineCredentialRejectsUnsafeInputAndUnlocksFailure.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `context.WithValue`, `credentialLeaseFixture`, `filepath.Join`, `lease.Release`, `lockWorkerCredential`, `os.Chmod`, `os.Rename`, `os.Symlink`, `os.WriteFile`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `unlock`.
- **`TestEngineCredentialReleaseRequiresBothCreatedRolesAbsent`** — Regression contract: TestEngineCredentialReleaseRequiresBothCreatedRolesAbsent.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `credentialLeaseFixture`, `lease.Release`, `t.Fatal`, `t.Parallel`.
- **`framedSyntheticCredential`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() []byte`.
  Calls: `append`, `binary.BigEndian.PutUint32`, `len`, `make`, `uint32`.
- **Type `syntheticShortCredentialWriter`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ calls int }`.
- **`(*syntheticShortCredentialWriter).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `len`.
- **`TestEngineCredentialProvisionNeverReplaysPartialFrame`** — Regression contract: TestEngineCredentialProvisionNeverReplaysPartialFrame.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `credentialLeaseFixture`, `lease.Provision`, `lease.Release`, `retry.Len`, `t.Fatal`, `t.Parallel`.
- **`TestEngineCredentialJournalBlocksCrashSurvivors`** — Regression contract: TestEngineCredentialJournalBlocksCrashSurvivors.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `credentialLeaseFixture`, `filepath.Join`, `lease.Release`, `lockWorkerCredential`, `os.Remove`, `os.WriteFile`, `prior.Deadline.Add`, `prior.StartedAt.Add`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `unlock`, `workerMembership`, `writeWorkerRecord`.
- **`TestEngineRefreshCommitsOnlyAfterOriginalPairFenced`** — Regression contract: TestEngineRefreshCommitsOnlyAfterOriginalPairFenced.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `credentialLeaseFixture`, `filepath.Join`, `lease.Provision`, `lease.Release`, `lease.StageRefresh`, `len`, `os.ReadFile`, `readPrivateBounded`, `string`, `t.Fatal`, `t.Parallel`, `workerCredentialReady`, `writeWorkerRecord`.
- **`TestEngineRefreshNeverOverwritesChangedConnectionCredential`** — Regression contract: TestEngineRefreshNeverOverwritesChangedConnectionCredential.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `bytes.Equal`, `credentialLeaseFixture`, `filepath.Join`, `lease.Provision`, `lease.Release`, `lease.StageRefresh`, `lockWorkerCredential`, `os.ReadFile`, `os.WriteFile`, `t.Fatal`, `t.Parallel`, `unlock`, `workerCredentialReady`.
