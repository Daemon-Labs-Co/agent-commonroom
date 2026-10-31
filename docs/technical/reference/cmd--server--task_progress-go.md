# cmd/server/task_progress.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_progress.go)

Go implementation for task progress; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `taskProgressDependency`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { InputName string 'json:"inputName"' SourceTaskID string 'json:"sourceTaskId"' SourceTaskRevision int 'json:"sourceTaskRevision"' ExecutionID string 'json:"executionId"' OutputName string 'json:"outputName"' Current bool 'json:"current"' }`.
- **Type `taskProgressView`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { DispatchStatus string 'json:"dispatchStatus,omitempty"' TaskID string 'json:"taskId"' Title string 'json:"title"' TaskRevision int 'json:"taskRevision"' TaskStatus string 'json:"taskStatus"' DefinitionRevision int 'json:"definitionRevision,omitempty"' Phase string 'json:"phase"' Detail string 'json:"detail"' ExecutionID string 'json:"executionId,omitempty"' ExecutionMode string 'json:"executionMode,omitempty"' ExecutionStatus string 'json:"executionStatus,omitempty"' WorkExecutionID string 'json:"workExecutionId,omitempty"' ReturnedValueCount int 'json:"returnedValueCount"' InputState string 'json:"inputState"' Dependencies []taskProgressDependency 'json:"dependencies"' CriterionReviewRevision int 'json:"criterionReviewRevision,omitempty"' }`.
- **`describeTaskProgress`** — Progress derives only from durable controller records. It never probes a provider or infers worker activity, correctness, observed checks or acceptance.
  Signature: `func(st room.State, t room.Task) taskProgressView`.
  Calls: `append`, `latestTaskAcceptance`, `len`, `taskCurrentWorkExecution`, `taskDefinitionDigest`, `taskExecutionInputsCurrent`, `taskHasCurrentAcceptance`.
