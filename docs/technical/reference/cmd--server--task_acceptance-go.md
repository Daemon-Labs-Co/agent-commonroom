# cmd/server/task_acceptance.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_acceptance.go)

Go implementation for task acceptance; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `taskCriterionDecision`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { CriterionID string 'json:"criterionId"' Decision string 'json:"decision"' Note string 'json:"note"' }`.
- **`latestTaskAcceptance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, tid string) room.TaskAcceptanceReview`.
  Calls: `len`.
- **`taskAcceptanceContentSnapshot`** — Status/revision change in the same atomic acceptance operation. Request revision and post-decision revision are fenced separately from this content snapshot so that acceptance does not immediately invalidate its own record.
  Signature: `func(st room.State, t room.Task) (string, room.TaskExecution, bool)`.
  Calls: `hasTaskValueSources`, `hex.EncodeToString`, `json.Marshal`, `sha256.Sum256`, `taskCurrentWorkExecution`.
- **`taskAcceptanceSnapshot`** — Snapshot construction never traverses lineage: source acceptance metadata is checked inside the single bounded dependency walk without resetting its budget.
  Signature: `func(st room.State, t room.Task) (string, room.TaskExecution, bool)`.
  Calls: `taskAcceptanceContentSnapshot`, `taskExecutionInputsCurrent`.
- **`taskHasRecordedAcceptance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, t room.Task) bool`.
  Calls: `latestTaskAcceptance`, `taskAcceptanceContentSnapshot`.
- **`taskHasCurrentAcceptance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, t room.Task) bool`.
  Calls: `taskCurrentWorkExecution`, `taskExecutionInputsCurrent`, `taskHasRecordedAcceptance`.
- **`assessTaskCriteria`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(d taskcontract.Definition, execution room.TaskExecution, decisions []taskCriterionDecision) ([]room.TaskCriterionAssessment, error)`.
  Calls: `append`, `delete`, `errors.New`, `len`, `strings.TrimSpace`, `taskSchemaAssessment`, `utf8.RuneCountInString`.
- **`taskSchemaAssessment`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c taskcontract.Criterion, contract taskcontract.DataContract, output taskcontract.BoundValue, present bool) room.TaskCriterionAssessment`.
  Calls: `int64`, `len`, `taskcontract.ValidContractValue`.
- **`taskSchemaAssessments`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(d taskcontract.Definition, execution room.TaskExecution) []room.TaskCriterionAssessment`.
  Calls: `append`, `taskSchemaAssessment`.
- **`taskReviewAllPassed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(assessments []room.TaskCriterionAssessment) bool`.
  Calls: `len`.
- **`validateTaskAcceptanceReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v room.TaskAcceptanceReview) error`.
  Calls: `append`, `assessTaskCriteria`, `errors.New`, `int`, `latestTaskAcceptance`, `reflect.DeepEqual`, `sharedTaskValueRoom`, `strings.TrimSpace`, `taskAcceptanceSnapshot`, `taskReviewAllPassed`, `uint`, `utf8.RuneCountInString`, `v.ReviewedAt.IsZero`.
- **`projectTaskAcceptanceReview`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `append`, `errors.New`, `json.Unmarshal`, `validateTaskAcceptanceReview`.
