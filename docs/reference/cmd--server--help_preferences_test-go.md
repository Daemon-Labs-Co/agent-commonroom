# cmd/server/help_preferences_test.go

[Module index](cmd.md) · [Open source](../../cmd/server/help_preferences_test.go)

Automated regression checks for help_preferences; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestHelpPreferenceSelfOnly`** — Regression contract: TestHelpPreferenceSelfOnly.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `member`, `mockRequest`, `mockServer`, `s.load`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.String`.
- **`TestHelpPreferenceValidationAndFlag`** — Regression contract: TestHelpPreferenceValidationAndFlag.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `member`, `mockRequest`, `mockServer`, `s.load`, `strings.Contains`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
