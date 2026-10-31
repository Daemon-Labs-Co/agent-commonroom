# cmd/server/p0_command_admission_acceptance_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/p0_command_admission_acceptance_test.go)

Automated regression checks for p0_command_admission_acceptance; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `commandAdmissionProbe`**: These drivers cover refusal before execution, not successful command isolation, compiler behavior, descendant shutdown or live worker/provider compatibility.
  Declaration: `struct{ calls int }`.
- **`(*commandAdmissionProbe).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(context.Context, []string, []byte, io.Writer) error`.
  Calls: `errors.New`.
- **Type `commandAdmissionLink`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ calls int }`.
- **`(*commandAdmissionLink).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func([]byte) (int, error)`.
- **`(*commandAdmissionLink).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func([]byte) (int, error)`.
- **`(*commandAdmissionLink).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
- **`driveP0CommandAdmission`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `context.WithValue`, `errors.Is`, `len`, `newWorkerToolSessionWithCommands`, `scopedAuditFixture`, `session.Close`, `session.CommandsEnabled`, `session.Dispatcher.Dispatch`, `strings.Repeat`, `t.Context`, `t.Fatal`, `t.Helper`, `time.Now`, `workerCodexToolRequest`.
- **`driveP0CommandIntentFailure`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, scenario pendingP0Case) pendingP0Observation`.
  Calls: `append`, `authority.Revoke`, `awaitAuditResult`, `cancel`, `clientLink.Close`, `context.WithCancel`, `controllerLink.Close`, `db.snapshot`, `dispatcher.Dispatch`, `errors.Is`, `errors.New`, `len`, `make`, `net.Pipe`, `reflect.DeepEqual`, `scopedAuditExecutor`, `scopedAuditFixture`, `t.Context`, `t.Fatal`, `t.Helper`, `time.Now`, `toolbroker.New`, `toolbroker.NewAuditClient`, `toolbroker.NewDispatcher`, `toolbroker.ServeAudit`, `workerCodexToolRequest`.
