# cmd/server/task_execution.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_execution.go)

Go implementation for task execution; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`responseDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(text string) string`.
  Calls: `hex.EncodeToString`, `sha256.Sum256`.
- **`taskCurrentWorkExecution`** — Select only the latest successful work/review chain. Dependency validation is separate so recursion retains one shared depth/node budget and path stack.
  Signature: `func(st room.State, t room.Task) (room.TaskExecution, bool)`.
  Calls: `taskDefinitionDigest`.
- **`taskHasValidCurrentOutput`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, t room.Task) bool`.
  Calls: `taskCurrentWorkExecution`, `taskExecutionInputsCurrent`.
- **`prepareTaskExecution`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, t room.Task, mode string, inputs []taskcontract.Binding, actor string) (*room.TaskExecution, error)`.
  Calls: `errors.New`, `finalizeTaskValueInputs`, `len`, `reflect.DeepEqual`, `resolveTaskValueInputs`, `taskDefinitionDigest`, `taskExecutionInputsCurrent`, `taskcontract.Executable`, `taskcontract.Validate`, `time.Now`, `time.Now().UTC`.
- **`taskExecutionReference`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v room.TaskExecution) string`.
  Calls: `referenceJSON`.
- **`taskExecutionRoleMatches`** — Task/attempt identity and role must agree at admission and publication. Dispatch status/revision are checked separately for the atomic transition phase.
  Signature: `func(t room.Task, e room.TaskExecution) bool`.
  Calls: `taskDefinitionDigest`.
- **`taskExecutionPendingCurrent`** — Pending authority is phase-specific: completion replay has its own completed task transition fence. This helper never grants authority from a saved status.
  Signature: `func(st room.State, t room.Task, e room.TaskExecution) bool`.
  Calls: `taskExecutionConfigurationCurrent`, `taskExecutionExportCurrent`, `taskExecutionInputsCurrent`, `taskExecutionRoleMatches`.
- **`structuredTaskResult`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, t room.Task, modeStatus, text string) (room.TaskExecution, error)`.
  Calls: `errors.New`, `len`, `taskExecutionPendingCurrent`, `taskcontract.ReturnedValues`.
- **`taskExecutionOutcome`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, taskID, dispatchID, status string, text ...string) []auditChange`.
  Calls: `err.Error`, `len`, `responseDigest`, `taskcontract.ReturnedValues`, `time.Now`, `time.Now().UTC`.
- **`projectTaskExecution`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, kind string, raw []byte) error`.
  Calls: `append`, `errors.New`, `json.Marshal`, `json.Unmarshal`, `len`, `reflect.DeepEqual`, `taskExecutionConfigurationCurrent`, `taskExecutionInputsCurrent`, `taskExecutionRoleMatches`, `taskcontract.Bind`, `taskcontract.Digest`, `taskcontract.ReturnedValues`, `taskcontract.Validate`, `v.FinishedAt.Before`, `v.FinishedAt.IsZero`, `v.StartedAt.IsZero`, `validTaskExecutionConfiguration`.
