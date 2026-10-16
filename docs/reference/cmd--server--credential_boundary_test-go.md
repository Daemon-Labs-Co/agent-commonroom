# cmd/server/credential_boundary_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/credential_boundary_test.go)

Automated regression checks for credential_boundary; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestLegacyDevelopmentCannotReadCredentialsOrLaunch`** — Regression contract: TestLegacyDevelopmentCannotReadCredentialsOrLaunch.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `errors.Is`, `localIsolatedDevelopmentCommand`, `providerTurnArgs`, `strings.Repeat`, `t.Fatal`, `t.Parallel`.
- **`TestWorkerRefusesLegacyDevelopmentBeforeAdmission`** — Regression contract: TestWorkerRefusesLegacyDevelopmentBeforeAdmission.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `errors.Is`, `providerTurnArgs`, `sampleWorkerRequest`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `validateWorkerRun`.
- **`TestCredentialBoundaryDiagnosticDoesNotSuggestSignIn`** — Regression contract: TestCredentialBoundaryDiagnosticDoesNotSuggestSignIn.
  Signature: `func(t *testing.T)`.
  Calls: `errCredentialBoundary.Error`, `safeAgentFailure`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
