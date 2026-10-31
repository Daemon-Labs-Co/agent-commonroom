# internal/toolcommand/command_test.go

[Module index](internal.md) · [Open source](../../../internal/toolcommand/command_test.go)

Automated regression checks for command; read test names for individual contracts.

Package: `toolcommand`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestCommandTimeoutWithholdsReceiptEvenIfRunnerReportsSuccess`** — Regression contract: TestCommandTimeoutWithholdsReceiptEvenIfRunnerReportsSuccess.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `ctx.Done`, `ctx.Err`, `io.WriteString`, `synctest.Test`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `testRunner`, `time.Now`, `time.Since`, `time.Sleep`.
- **Type `testRunner`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, Spec, io.Writer) (int, error)`.
- **`(testRunner).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, s Spec, w io.Writer) (int, error)`.
  Calls: `f`.
- **`TestProfilesAndFixedExecutionPolicy`** — Regression contract: TestProfilesAndFixedExecutionPolicy.
  Signature: `func(t *testing.T)`.
  Calls: `Digest`, `Run`, `base64.StdEncoding.DecodeString`, `context.Background`, `ctx.Deadline`, `io.WriteString`, `len`, `slices.Contains`, `slices.Equal`, `string`, `t.Fatal`, `t.Fatalf`, `t.Parallel`, `testRunner`, `time.Until`.
- **`TestOutputAndFailureBounds`** — Regression contract: TestOutputAndFailureBounds.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `c.Err`, `cancel`, `context.Background`, `context.WithCancel`, `errors.Is`, `errors.New`, `io.WriteString`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `testRunner`.
- **`TestOverflowHasDistinctFailureCauseAndWithholdsReceipt`** — Regression contract: TestOverflowHasDistinctFailureCauseAndWithholdsReceipt.
  Signature: `func(t *testing.T)`.
  Calls: `Run`, `ctx.Err`, `errors.Is`, `io.WriteString`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Parallel`, `testRunner`.
