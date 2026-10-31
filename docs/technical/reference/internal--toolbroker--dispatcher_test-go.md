# internal/toolbroker/dispatcher_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/dispatcher_test.go)

Automated regression checks for dispatcher; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `mockAudit`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { intent func(context.Context, Intent) error outcome func(context.Context, Intent, Outcome) error }`.
- **`(mockAudit).RecordIntent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i Intent) error`.
  Calls: `a.intent`.
- **`(mockAudit).RecordOutcome`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, i Intent, o Outcome) error`.
  Calls: `a.outcome`.
- **Type `mockExecutor`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, Scope, Action, io.Writer) error`.
- **`(mockExecutor).Execute`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c context.Context, s Scope, a Action, w io.Writer) error`.
  Calls: `e`.
- **`liveFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T) (*Authority, Request)`.
  Calls: `New`, `fixture`, `now.Add`, `t.Fatal`, `t.Helper`, `time.Now`.
- **`dispatcher`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, a *Authority, audit Auditor, exec Executor) *Dispatcher`.
  Calls: `NewDispatcher`, `t.Fatal`, `t.Helper`.
- **`TestAuditBeforeExecutionAndOutcome`** — Regression contract: TestAuditBeforeExecutionAndOutcome.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `d.Dispatch`, `digest`, `dispatcher`, `io.WriteString`, `liveFixture`, `mockExecutor`, `string`, `t.Error`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestAuditFailureFencesJob`** — Regression contract: TestAuditFailureFencesJob.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `d.Dispatch`, `dispatcher`, `errors.New`, `liveFixture`, `mockExecutor`, `t.Fatal`, `t.Parallel`.
- **`TestCancellationBeforeExecution`** — Regression contract: TestCancellationBeforeExecution.
  Signature: `func(t *testing.T)`.
  Calls: `a.Revoke`, `cancel`, `context.Background`, `context.WithCancel`, `d.Dispatch`, `dispatcher`, `liveFixture`, `mockExecutor`, `t.Fatal`, `t.Parallel`.
- **`TestSharedJobSlotAndRunningRevocation`** — Regression contract: TestSharedJobSlotAndRunningRevocation.
  Signature: `func(t *testing.T)`.
  Calls: `a.Revoke`, `close`, `context.Background`, `ctx.Done`, `ctx.Err`, `d.Dispatch`, `dispatcher`, `intents.Add`, `intents.Load`, `liveFixture`, `make`, `mockExecutor`, `other.Dispatch`, `t.Fatal`, `t.Parallel`, `time.After`.
- **`TestBoundedOutputAndClosedWriter`** — Regression contract: TestBoundedOutputAndClosedWriter.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `ctx.Err`, `d.Dispatch`, `dispatcher`, `io.WriteString`, `liveFixture`, `mockExecutor`, `retained.Write`, `strings.Repeat`, `t.Error`, `t.Fatal`, `t.Parallel`, `w.Write`.
- **`TestOriginalDeadlineAndInputSnapshot`** — Regression contract: TestOriginalDeadlineAndInputSnapshot.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `ctx.Deadline`, `d.Dispatch`, `deadline.Equal`, `dispatcher`, `liveFixture`, `mockExecutor`, `parent.Deadline`, `t.Error`, `t.Fatal`, `t.Parallel`.
- **`TestExecutionFailureCannotReplay`** — Regression contract: TestExecutionFailureCannotReplay.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `d.Dispatch`, `dispatcher`, `errors.New`, `liveFixture`, `mockExecutor`, `t.Fatal`, `t.Parallel`.
- **`TestDeadlineCancelsRunningWork`** — Regression contract: TestDeadlineCancelsRunningWork.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithTimeout`, `ctx.Done`, `d.Dispatch`, `dispatcher`, `liveFixture`, `mockExecutor`, `t.Fatal`, `t.Parallel`.
- **`TestMissingDependencies`** — Regression contract: TestMissingDependencies.
  Signature: `func(t *testing.T)`.
  Calls: `NewDispatcher`, `liveFixture`, `mockExecutor`, `t.Fatal`, `t.Parallel`.
- **`TestExecutorReportedOutputLimitPreservesFailureEvidence`** — Regression contract: TestExecutorReportedOutputLimitPreservesFailureEvidence.
  Signature: `func(t *testing.T)`.
  Calls: `d.Dispatch`, `dispatcher`, `errors.Join`, `liveFixture`, `mockExecutor`, `t.Context`, `t.Fatal`, `t.Parallel`.
