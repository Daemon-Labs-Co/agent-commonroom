# cmd/server/task_value_inputs_api.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_value_inputs_api.go)

Go implementation for task value inputs api; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `taskInputSource`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { TaskID string 'json:"taskId"' TaskTitle string 'json:"taskTitle"' TaskRevision int 'json:"taskRevision"' ExecutionID string 'json:"executionId"' OutputName string 'json:"outputName"' ValueJSON string 'json:"valueJson"' Digest string 'json:"digest"' }`.
- **`(*server).taskInputSources`** — Discovery exposes only accepted, current, compatible shared-room values. The eventual dispatch repeats resolution; discovery is not authorization.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `canAccess`, `http.Error`, `len`, `r.Context`, `r.URL.Query`, `r.URL.Query().Get`, `reply`, `requestActor`, `requestedRoom`, `resolveTaskValueInputs`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `sharedTaskValueRoom`, `sort.Slice`, `taskCurrentWorkExecution`.
