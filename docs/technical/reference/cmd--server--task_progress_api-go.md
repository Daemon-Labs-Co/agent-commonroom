# cmd/server/task_progress_api.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_progress_api.go)

Go implementation for task progress api; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).taskProgress`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `canAccess`, `describeTaskProgress`, `http.Error`, `len`, `r.Context`, `r.URL.Query`, `r.URL.Query().Get`, `reply`, `requestActor`, `requestedRoom`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `sharedTaskValueRoom`, `sort.Slice`, `taskDispatchEnabled`.
