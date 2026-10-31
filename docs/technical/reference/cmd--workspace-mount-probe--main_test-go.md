# cmd/workspace-mount-probe/main_test.go

[Module index](cmd.md) · [Open source](../../../cmd/workspace-mount-probe/main_test.go)

Automated regression checks for main; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestRunWithholdsInvalidObservation`** — Regression contract: TestRunWithholdsInvalidObservation.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `cancel`, `context.Background`, `context.WithCancel`, `errors.New`, `out.Len`, `out.String`, `run`, `strings.Contains`, `strings.Repeat`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `t.Run`.
