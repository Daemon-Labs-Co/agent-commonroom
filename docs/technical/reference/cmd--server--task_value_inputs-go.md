# cmd/server/task_value_inputs.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_value_inputs.go)

Go implementation for task value inputs; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`hasTaskValueSources`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(inputs []taskcontract.BoundValue) bool`.
- **`sharedTaskValueRoom`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, rid string) bool`.
- **`taskValueSource`** — Select source metadata only. Callers must validate complete lineage using one shared path/depth/node budget; acceptance metadata must not start another walk.
  Signature: `func(st room.State, rid, executionID, outputName string) (room.Task, room.TaskExecution, taskcontract.BoundValue, bool)`.
  Calls: `sharedTaskValueRoom`, `taskCurrentWorkExecution`, `taskHasRecordedAcceptance`.
- **`resolveTaskValueInputs`** — Resolution uses only stored returned bytes. A client-supplied value or version accompanying a reference is refused, never promoted into trusted provenance.
  Signature: `func(st room.State, t room.Task, inputs []taskcontract.Binding) ([]taskcontract.BoundValue, error)`.
  Calls: `errors.New`, `len`, `make`, `sharedTaskValueRoom`, `taskValueLineageCurrent`, `taskValueSource`, `taskcontract.Bind`.
- **`finalizeTaskValueInputs`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v *room.TaskExecution) (*room.TaskExecution, error)`.
  Calls: `errors.New`, `hasTaskValueSources`, `taskExecutionInputsCurrent`.
- **`taskExecutionInputsCurrent`** — Every use retains one path and one budget across the whole dependency DAG. Completed history survives flag cycling; in-flight authority never revives.
  Signature: `func(st room.State, v room.TaskExecution) bool`.
  Calls: `hasTaskValueSources`, `taskValueLineageCurrent`.
- **`taskValueLineageCurrent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v room.TaskExecution, path map[string]bool, depth int, nodes *int) bool`.
  Calls: `bytes.Equal`, `delete`, `taskValueLineageCurrent`, `taskValueSource`.
