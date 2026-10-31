# internal/toolbroker/object_read_runtime_test.go

[Module index](internal.md) · [Open source](../../../internal/toolbroker/object_read_runtime_test.go)

Automated regression checks for object_read_runtime; read test names for individual contracts.

Package: `toolbroker`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `objectReadRuntimeFixture`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { *objectReadPlanRunner mode string exists, begun, bound, started, finished, absent, removed bool authorizations, acknowledgments, stopped int process *mockObjectReadProcess exitedChecks int cancel context.CancelFunc }`.
- **Type `mockObjectReadProcess`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { net.Conn f *objectReadRuntimeFixture done chan struct{} helperErr error closed atomic.Bool waited atomic.Bool }`.
- **`(*mockObjectReadProcess).Close`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() error`.
  Calls: `p.Conn.Close`, `p.closed.Store`.
- **`(*mockObjectReadProcess).Read`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(b []byte) (int, error)`.
  Calls: `p.Conn.Read`, `probeBool`.
- **`(*mockObjectReadProcess).Wait`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context) error`.
  Calls: `ctx.Done`, `ctx.Err`, `p.closed.Load`, `p.waited.Store`, `probeBool`.
- **`(*objectReadRuntimeFixture).BeginObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, v CaptureObjectVolume, image string) error`.
  Calls: `ctx.Err`, `f.t.Fatal`.
- **`(*objectReadRuntimeFixture).BindObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, id string) error`.
  Calls: `ctx.Err`, `f.t.Fatal`.
- **`(*objectReadRuntimeFixture).FinishObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, id string, absent bool) error`.
  Calls: `ctx.Err`, `f.cancel`, `f.process.closed.Load`, `f.t.Fatal`.
- **`(*objectReadRuntimeFixture).AuthorizeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, s objectcustody.Scope, r objectcustody.Receipt) error`.
  Calls: `ctx.Err`, `f.cancel`, `f.t.Fatal`.
- **`(*objectReadRuntimeFixture).AcknowledgeRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, r objectcustody.Receipt) error`.
  Calls: `ctx.Err`, `f.t.Fatal`.
- **`(*objectReadRuntimeFixture).StopReadAuthority`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
- **`(*objectReadRuntimeFixture).Run`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, args []string, in []byte, out io.Writer) error`.
  Calls: `f.objectReadPlanRunner.Run`, `f.p.CreateArgs`, `f.t.Fatal`, `io.WriteString`, `slices.Equal`, `strings.Repeat`.
- **Type `relayedSyntheticReader`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { journal *objectcustody.ReadJournalChannel corrupt bool }`.
- **`(relayedSyntheticReader).Get`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, s objectcustody.Scope, q objectcustody.Receipt) ([]byte, error)`.
  Calls: `r.journal.AcknowledgeRead`, `r.journal.AuthorizeRead`.
- **`(*objectReadRuntimeFixture).StartObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, id string) (ObjectReadProcess, error)`.
  Calls: `close`, `f.t.Fatal`, `helper.Close`, `make`, `net.Pipe`, `objectcustody.NewReadJournalChannel`, `objectcustody.ServeReadChannel`, `probeBool`.
- **`newObjectReadRuntimeFixture`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t *testing.T, mode string) (*ObjectReadRuntime, *objectReadRuntimeFixture)`.
  Calls: `NewObjectReadContainerPlan`, `NewObjectReadRuntime`, `hex.EncodeToString`, `objectReadPlanFixture`, `p.labels`, `sha256.Sum256`, `t.Fatal`, `t.Helper`.
- **`TestObjectReadRuntimeLifecycle`** — Regression contract: TestObjectReadRuntimeLifecycle.
  Signature: `func(t *testing.T)`.
  Calls: `f.process.closed.Load`, `f.process.waited.Load`, `newObjectReadRuntimeFixture`, `r.Run`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestObjectReadRuntimeFailuresWithholdBytes`** — Regression contract: TestObjectReadRuntimeFailuresWithholdBytes.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `errors.Is`, `f.process.closed.Load`, `newObjectReadRuntimeFixture`, `r.Run`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestObjectReadRuntimeRequiresDependencies`** — Regression contract: TestObjectReadRuntimeRequiresDependencies.
  Signature: `func(t *testing.T)`.
  Calls: `NewObjectReadRuntime`, `cancel`, `context.WithCancel`, `makeRuntime`, `newObjectReadRuntimeFixture`, `r.Run`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **`TestObjectReadRuntimeExitAtProtocolEOF`** — Regression contract: TestObjectReadRuntimeExitAtProtocolEOF.
  Signature: `func(t *testing.T)`.
  Calls: `newObjectReadRuntimeFixture`, `r.Run`, `string`, `t.Context`, `t.Fatal`, `t.Parallel`.
- **Type `objectReadDeadlineJournal`**: This journal observes the context actually supplied to the persistence boundary. Its embedded fixture still checks exact request/ID and process-close ordering.
  Declaration: `struct { *objectReadRuntimeFixture activeFinish bool }`.
- **`(*objectReadDeadlineJournal).FinishObjectRead`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, q objectread.Request, id string, absent bool) error`.
  Calls: `ctx.Deadline`, `ctx.Err`, `j.objectReadRuntimeFixture.FinishObjectRead`, `time.Until`.
- **`TestObjectReadCleanupExhaustionStillJournalsUncertainty`** — Regression contract: TestObjectReadCleanupExhaustionStillJournalsUncertainty.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.Background`, `context.WithCancel`, `context.WithDeadline`, `errors.Is`, `newObjectReadRuntimeFixture`, `r.cleanupAndFinish`, `stop`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.Now`, `time.Now().Add`.
