# cmd/server/task_repository_links.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_repository_links.go)

Go implementation for task repository links; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`validTaskRepositoryLink`** — No network access or repository authority is obtained from a saved reference. This accepts a bounded safe URL subset, not proof an external object exists.
  Signature: `func(v room.TaskRepositoryLink) bool`.
  Calls: `len`, `repositoryBranch.MatchString`, `repositoryCommit.MatchString`, `repositoryNumber.MatchString`, `repositorySlug.MatchString`, `strings.ContainsAny`, `strings.HasPrefix`, `strings.Join`, `strings.Split`, `strings.TrimPrefix`, `url.Parse`.
- **`taskRepositoryLinksKey`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(company, rid, tid string) string`.
- **`taskForRepositoryLinks`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, rid, tid string) (room.Task, bool)`.
- **`validateTaskRepositoryLinks`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, v room.TaskRepositoryLinks) error`.
  Calls: `artifactRecordedActor`, `int`, `len`, `sharedTaskValueRoom`, `strings.ToLower`, `taskForRepositoryLinks`, `taskRepositoryLinksKey`, `uint`, `v.ChangedAt.IsZero`, `validTaskRepositoryLink`.
- **`projectTaskRepositoryLinks`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, raw []byte) error`.
  Calls: `json.Unmarshal`, `taskRepositoryLinksKey`, `validateTaskRepositoryLinks`.
- **Type `taskRepositoryLinksRequest`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { TaskID string 'json:"taskId"' TaskRevision int 'json:"taskRevision"' Revision int 'json:"revision"' Links []room.TaskRepositoryLink 'json:"links"' }`.
- **`(*server).taskRepositoryLinks`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `artifactActor`, `bytes.NewReader`, `dec.Decode`, `dec.DisallowUnknownFields`, `errors.Is`, `http.Error`, `http.MaxBytesReader`, `int`, `io.ReadAll`, `json.NewDecoder`, `len`, `new`, `operationError`, `prepareOperation`, `r.Context`, `r.Header.Get`, `r.URL.Query`, `r.URL.Query().Get`, `reply`, `requestActor`, `requestedRoom`, `s.appendBatch`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.validCSRF`, `sharedTaskValueRoom`, `taskForRepositoryLinks`, `taskRepositoryLinksKey`, `taskcontract.ValidJSON`, `time.Now`, `time.Now().UTC`, `uint`, `validateTaskRepositoryLinks`, `withOperationResponse`.
