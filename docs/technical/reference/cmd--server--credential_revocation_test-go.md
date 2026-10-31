# cmd/server/credential_revocation_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/credential_revocation_test.go)

Automated regression checks for credential_revocation; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestCredentialRevocationRequiresSyncEvenOnRetry`** — Regression contract: TestCredentialRevocationRequiresSyncEvenOnRetry.
  Signature: `func(t *testing.T)`.
  Calls: `clearWorkerCredential`, `clearWorkerCredentialDurably`, `errors.Is`, `errors.New`, `filepath.Join`, `os.IsNotExist`, `os.Lstat`, `os.WriteFile`, `providerCredentialName`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestUnknownCredentialProviderCannotDeleteCodexCredential`** — Regression contract: TestUnknownCredentialProviderCannotDeleteCodexCredential.
  Signature: `func(t *testing.T)`.
  Calls: `clearWorkerCredentialDurably`, `filepath.Join`, `os.ReadFile`, `os.WriteFile`, `string`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
