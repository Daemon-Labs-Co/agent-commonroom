# cmd/server/worker_credential_recovery_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_credential_recovery_test.go)

Automated regression checks for worker_credential_recovery; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `credentialOrderWriter`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { t *testing.T root string bytes.Buffer }`.
- **`(*credentialOrderWriter).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `w.Buffer.Write`, `w.t.Fatal`, `w.t.Helper`, `workerCredentialReady`.
- **`TestCredentialRecoveryIntentPrecedesProvisionAndSurvivesFencing`** — Regression contract: TestCredentialRecoveryIntentPrecedesProvisionAndSurvivesFencing.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `bytes.Equal`, `credentialLeaseFixture`, `credentialPresent`, `filepath.Join`, `framedSyntheticCredential`, `lease.Provision`, `lease.Release`, `out.Bytes`, `readPrivateBounded`, `string`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `workerCredentialConnectionQuiescent`, `workerCredentialReady`.
- **`TestCredentialRecoveryMarkerFailureSendsNoBytes`** — Regression contract: TestCredentialRecoveryMarkerFailureSendsNoBytes.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `credentialLeaseFixture`, `filepath.Join`, `lease.Provision`, `lease.Release`, `os.Mkdir`, `out.Len`, `t.Fatal`, `t.Parallel`.
- **`TestCredentialRecoveryFailuresRemainScopedAndClosed`** — Regression contract: TestCredentialRecoveryFailuresRemainScopedAndClosed.
  Signature: `func(t *testing.T)`.
  Calls: `acquireWorkerEngineCredential`, `connectionAuthKey`, `credentialLeaseFixture`, `credentialPresent`, `filepath.Join`, `os.Mkdir`, `os.MkdirAll`, `os.Symlink`, `os.WriteFile`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestCredentialRecoveryNeedsDurableCredentialBeforeMarkerClear`** — Regression contract: TestCredentialRecoveryNeedsDurableCredentialBeforeMarkerClear.
  Signature: `func(t *testing.T)`.
  Calls: `credentialPresent`, `err.Error`, `errors.New`, `filepath.Base`, `filepath.Join`, `markWorkerCredentialUncertain`, `os.WriteFile`, `persistWorkerCredentialDurably`, `strings.Contains`, `syncParentDirectory`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`, `workerCredentialReady`.
- **`TestCredentialRecoveryJournalRequiresBothRolesQuiescent`** — Regression contract: TestCredentialRecoveryJournalRequiresBothRolesQuiescent.
  Signature: `func(t *testing.T)`.
  Calls: `credentialLeaseFixture`, `filepath.Join`, `os.WriteFile`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `workerCredentialConnectionQuiescent`, `writeWorkerRecord`.
- **`TestCredentialRecoverySyncFailureRemainsBlocked`** — Regression contract: TestCredentialRecoverySyncFailureRemainsBlocked.
  Signature: `func(t *testing.T)`.
  Calls: `err.Error`, `errors.New`, `markWorkerCredentialUncertainDurably`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `workerCredentialReady`.
- **`TestWorkerCredentialStatusNeverOpensUncertainCustody`** — Regression contract: TestWorkerCredentialStatusNeverOpensUncertainCustody.
  Signature: `func(t *testing.T)`.
  Calls: `decodeWorkerCredentialStatus`, `err.Error`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
