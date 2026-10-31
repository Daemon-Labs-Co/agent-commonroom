# internal/toolbroker/container_command_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/container_command_test.go)

Automated regression checks for container_command; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `commandDocker`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { DockerRunner command func(context.Context, []string, io.Writer) (int, error) }`.
- **`(commandDocker).RunCommand`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, a []string, w io.Writer) (int, error)`.
  Calls: `r.command`.
- **`TestCommandRequiresFlagAndGrantBeforeDocker`** — Regression contract: TestCommandRequiresFlagAndGrantBeforeDocker.
  Signature: `func(t *testing.T)`.
  Calls: `New`, `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `executorFixture`, `mockDocker`, `out.Len`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`.
- **`TestCommandReceiptRequiresContainerAbsenceAndConsumesContainer`** — Regression contract: TestCommandReceiptRequiresContainerAbsenceAndConsumesContainer.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `containerExecutor`, `context.Background`, `context.WithDeadline`, `e.Execute`, `errors.New`, `executorFixture`, `io.WriteString`, `json.Unmarshal`, `len`, `output.Bytes`, `output.Len`, `runtimeMock`, `slices.Contains`, `slices.Equal`, `strings.Repeat`, `t.Fatal`, `t.Parallel`, `t.Run`, `validCommandReceipt`.
- **`TestCommandReceiptRejectsTampering`** — Regression contract: TestCommandReceiptRejectsTampering.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `json.Marshal`, `mutate`, `string`, `strings.Repeat`, `strings.Replace`, `t.Fatal`, `t.Parallel`, `toolcommand.Run`, `validCommandReceipt`.
- **Type `receiptRunner`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **`(receiptRunner).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, toolcommand.Spec, io.Writer) (int, error)`.
- **`TestCommandAuditAcknowledgementsGateEffectsAndOutput`** — Regression contract: TestCommandAuditAcknowledgementsGateEffectsAndOutput.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `dispatcher`, `dispatcher(t, a, audit, executor).Dispatch`, `json.Marshal`, `len`, `liveFixture`, `mockExecutor`, `slices.Equal`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `toolcommand.Run`, `w.Write`.
