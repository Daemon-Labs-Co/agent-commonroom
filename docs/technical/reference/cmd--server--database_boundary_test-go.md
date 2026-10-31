# cmd/server/database_boundary_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/database_boundary_test.go)

Automated regression checks for database_boundary; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRuntimeDatabaseRequiresAppendOnlyPrivileges`** — Regression contract: TestRuntimeDatabaseRequiresAppendOnlyPrivileges.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `newMockDB`, `rune`, `string`, `t.Fatal`, `t.Parallel`, `t.Run`, `validateRuntimeDatabase`.
- **`TestDatabaseStartupTimeoutAndCredentialSafeFailure`** — Regression contract: TestDatabaseStartupTimeoutAndCredentialSafeFailure.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `ctx.Deadline`, `ctx.Done`, `ctx.Err`, `err.Error`, `errors.New`, `newMockDB`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Since`, `time.Until`, `verifyRuntimeDatabaseStartup`.
- **`TestDatabaseStartupAcceptsVerifiedLeastPrivileges`** — Regression contract: TestDatabaseStartupAcceptsVerifiedLeastPrivileges.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `newMockDB`, `t.Fatal`, `t.Parallel`, `verifyRuntimeDatabaseStartup`.
- **`TestRuntimeDatabasePinsSearchPathAndFailsOnUnknownPrivileges`** — Regression contract: TestRuntimeDatabasePinsSearchPathAndFailsOnUnknownPrivileges.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `newMockDB`, `runtimeDatabaseConfig`, `t.Fatal`, `t.Parallel`, `validateRuntimeDatabase`.
