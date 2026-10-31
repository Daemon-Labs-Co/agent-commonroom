# cmd/server/startup_command_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/startup_command_test.go)

Automated regression checks for startup_command; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestStartupCommandRejectsAmbiguousIntent`** — Regression contract: TestStartupCommandRejectsAmbiguousIntent.
  Signature: `func(t *testing.T)`.
  Calls: `diagnostic.Len`, `diagnostic.String`, `out.Len`, `startupCommand`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestStartupCommandHelpAndPrivateValidation`** — Regression contract: TestStartupCommandHelpAndPrivateValidation.
  Signature: `func(t *testing.T)`.
  Calls: `diagnostic.Len`, `objectReadProfileFixture`, `objectReadSessionFixture`, `out.String`, `startupCommand`, `strings.Contains`, `t.Fatal`, `t.Parallel`.
- **`TestStartupCommandProcessChild`** — Regression contract: TestStartupCommandProcessChild.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `main`, `os.Exit`, `os.Getenv`, `strings.Split`.
- **`TestStartupCommandProcessDoesNotStartController`** — Regression contract: TestStartupCommandProcessDoesNotStartController.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `captureProfileFixture`, `captureSessionFixture`, `cmd.CombinedOutput`, `context.WithTimeout`, `ctx.Err`, `errors.As`, `errors.Is`, `exec.CommandContext`, `exit.ExitCode`, `filepath.Join`, `os.Stat`, `string`, `strings.Contains`, `strings.Join`, `t.Context`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`, `t.TempDir`.
