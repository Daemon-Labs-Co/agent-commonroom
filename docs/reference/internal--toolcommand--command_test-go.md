# internal/toolcommand/command_test.go

[Module index](internal.md) · [Open source](../../internal/toolcommand/command_test.go)

Automated regression checks for command; read test names for individual contracts.

Package: `toolcommand`. Go uses types, structs, interfaces and receiver methods rather than classes.

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
  Calls: `Run`, `c.Err`, `cancel`, `context.Background`, `context.WithCancel`, `errors.New`, `io.WriteString`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `testRunner`.
