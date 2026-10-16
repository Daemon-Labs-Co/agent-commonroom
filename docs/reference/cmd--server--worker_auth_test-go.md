# cmd/server/worker_auth_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_auth_test.go)

Automated regression checks for worker_auth; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestConnectionAuthProtocol`** — Regression contract: TestConnectionAuthProtocol.
  Signature: `func(t *testing.T)`.
  Calls: `base64.RawURLEncoding.EncodeToString`, `decodeWorkerAuth`, `encodeWorkerAuth`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestConnectionCredentialScope`** — Regression contract: TestConnectionCredentialScope.
  Signature: `func(t *testing.T)`.
  Calls: `connectionAuthKey`, `credentialPresent`, `executionConnectionID`, `filepath.Join`, `os.Chmod`, `os.Remove`, `os.Symlink`, `os.WriteFile`, `t.Fatal`, `t.Parallel`, `t.TempDir`.
- **`TestConnectionSignInFlagAndOwner`** — Regression contract: TestConnectionSignInFlagAndOwner.
  Signature: `func(t *testing.T)`.
  Calls: `connectionSignInEnabled`, `db.seed`, `http.HandlerFunc`, `mockRequest`, `mockServer`, `s.actorGuard`, `s.auth.protect`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestInteractiveLoginCommand`** — Regression contract: TestInteractiveLoginCommand.
  Signature: `func(t *testing.T)`.
  Calls: `interactiveWorkerLogin`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestPersistWorkerCredential`** — Regression contract: TestPersistWorkerCredential.
  Signature: `func(t *testing.T)`.
  Calls: `credentialPresent`, `filepath.Glob`, `filepath.Join`, `info.Mode`, `info.Mode().Perm`, `len`, `os.ReadFile`, `os.Remove`, `os.Stat`, `os.Symlink`, `os.WriteFile`, `persistWorkerCredential`, `providerCredentialName`, `string`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `t.TempDir`.
- **`TestWorkerCredentialExportPolicy`** — Regression contract: TestWorkerCredentialExportPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `len`, `strings.Contains`, `strings.Join`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `workerCredentialExportArgs`.
- **`TestCredentialExportOutputLimit`** — Regression contract: TestCredentialExportOutputLimit.
  Signature: `func(t *testing.T)`.
  Calls: `output.Len`, `output.Write`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestClearWorkerCredentialScopeAndLock`** — Regression contract: TestClearWorkerCredentialScopeAndLock.
  Signature: `func(t *testing.T)`.
  Calls: `clearWorkerCredential`, `credentialPresent`, `filepath.Join`, `lockWorkerCredential`, `os.MkdirAll`, `os.WriteFile`, `second`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `unlock`.
- **`TestSignOutActiveConnectionGuard`** — Regression contract: TestSignOutActiveConnectionGuard.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `http.HandlerFunc`, `mockRequest`, `mockServer`, `s.actorGuard`, `s.auth.protect`, `strings.Contains`, `t.Fatal`, `t.Setenv`, `time.Now`, `time.Now().Add`, `w.Body.String`.
