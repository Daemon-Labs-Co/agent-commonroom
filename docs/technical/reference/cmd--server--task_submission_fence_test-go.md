# cmd/server/task_submission_fence_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_submission_fence_test.go)

Automated regression checks for task_submission_fence; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestManagedTaskSubmissionRechecksAuthorityAfterAllocation`** — Regression contract: TestManagedTaskSubmissionRechecksAuthorityAfterAllocation.
  Signature: `func(t *testing.T)`.
  Calls: `cancel`, `context.WithCancel`, `db.seed`, `definedDispatchBody`, `dispatchRequest`, `errors.New`, `jobCtx.Value`, `len`, `make`, `reg.Register`, `reg.Update`, `s.load`, `s.managedExecutionJobContext`, `strings.Repeat`, `strings.ReplaceAll`, `structuredTaskFixture`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `w.Body.String`, `waitDispatch`.
- **`TestManagedDependencySubmissionRechecksSourceAfterAllocation`** — Regression contract: TestManagedDependencySubmissionRechecksSourceAfterAllocation.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `dispatchRequest`, `errors.New`, `jobCtx.Value`, `len`, `make`, `reg.Register`, `reg.Update`, `s.managedExecutionJobContext`, `strings.Repeat`, `strings.ReplaceAll`, `t.Fatal`, `t.Parallel`, `t.Run`, `time.After`, `valueDependencyFixture`, `valueDispatchBody`, `w.Body.String`, `waitDispatch`.
- **`TestManagedSubmissionCannotReuseConsumedOrTerminalAllocation`** — Regression contract: TestManagedSubmissionCannotReuseConsumedOrTerminalAllocation.
  Signature: `func(t *testing.T)`.
  Calls: `db.snapshot`, `len`, `reg.Update`, `s.load`, `syntheticManagedJob`, `t.Context`, `t.Fatal`, `t.Parallel`, `t.Run`.
