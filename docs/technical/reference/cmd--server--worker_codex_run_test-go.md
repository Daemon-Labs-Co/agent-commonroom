# cmd/server/worker_codex_run_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_codex_run_test.go)

Automated regression checks for worker_codex_run; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestOwnedCodexPreparationFailuresFenceAndRelease`** — Regression contract: TestOwnedCodexPreparationFailuresFenceAndRelease.
  Signature: `func(t *testing.T)`.
  Calls: `bundle.closes.Load`, `connectionAuthKey`, `context.WithValue`, `filepath.Join`, `len`, `lockWorkerCredential`, `net.Pipe`, `newWorkerToolSession`, `newWorkspaceCleanup`, `os.Chmod`, `os.MkdirAll`, `os.ReadDir`, `os.WriteFile`, `peer.Close`, `s.BeginCreate`, `s.RunCodex`, `session.Close`, `shortGatewayRoot`, `strings.NewReader`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `t.TempDir`, `time.Now`, `unlock`, `workspacePair`.
- **`TestOwnedCodexValidationRefusalRetainsCallerOwnership`** — Regression contract: TestOwnedCodexValidationRefusalRetainsCallerOwnership.
  Signature: `func(t *testing.T)`.
  Calls: `bundle.closes.Load`, `s.RunCodex`, `strings.NewReader`, `t.Context`, `t.Fatal`, `t.Parallel`, `workspacePair`.
- **`TestSupervisorRetainsCredentialUntilBothRolesAreFenced`** — Regression contract: TestSupervisorRetainsCredentialUntilBothRolesAreFenced.
  Signature: `func(t *testing.T)`.
  Calls: `lockWorkerCredential`, `newWorkspaceCleanup`, `pairContainerName`, `release`, `s.Shutdown`, `s.Start`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.TempDir`, `unlock`, `workspacePair`.
- **Type `failedCodexBundleClose`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { io.Reader err error }`.
- **`(*failedCodexBundleClose).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
- **`TestOwnedBundlePreservesFirstCloseError`** — Regression contract: TestOwnedBundlePreservesFirstCloseError.
  Signature: `func(t *testing.T)`.
  Calls: `bundle.Close`, `errors.Is`, `errors.New`, `strings.NewReader`, `t.Fatal`, `t.Parallel`.
