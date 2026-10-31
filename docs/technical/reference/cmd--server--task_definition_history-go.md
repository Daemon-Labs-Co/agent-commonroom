# cmd/server/task_definition_history.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_definition_history.go)

Go implementation for task definition history; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`recordTaskDefinitionRevision`** — Called only after replay validates the complete acknowledged task event. Status/assignment-only revisions do not create fictional contract versions.
  Signature: `func(st *room.State, task room.Task)`.
  Calls: `append`, `len`, `taskcontract.Digest`.
- **`(*server).taskDefinitionHistory`** — Reading retained versions remains available with taskDefinitions off, just as execution history does. It grants no mutation, execution or acceptance power.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `canAccess`, `hasRoom`, `http.Error`, `len`, `r.Context`, `r.URL.Query`, `r.URL.Query().Get`, `reply`, `requestActor`, `requestedRoom`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `strconv.Itoa`.
