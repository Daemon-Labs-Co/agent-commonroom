# cmd/server/task_execution_api.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_execution_api.go)

Go implementation for task execution api; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).taskExecutions`** — Immutable execution records are room-scoped history, not agent capabilities.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `canAccess`, `hasRoom`, `http.Error`, `len`, `list[i].StartedAt.After`, `list[i].StartedAt.Equal`, `r.Context`, `r.URL.Query`, `r.URL.Query().Get`, `reply`, `requestActor`, `requestedRoom`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `sort.Slice`, `taskExecutionInputsCurrent`.
