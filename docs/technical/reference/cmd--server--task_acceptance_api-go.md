# cmd/server/task_acceptance_api.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_acceptance_api.go)

Go implementation for task acceptance api; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `taskAcceptanceRequest`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { TaskID string 'json:"taskId"' TaskRevision int 'json:"taskRevision"' ReviewRevision int 'json:"reviewRevision"' Snapshot string 'json:"snapshot"' Action string 'json:"action"' Note string 'json:"note"' Decisions []taskCriterionDecision 'json:"decisions"' }`.
- **`(*server).taskAcceptance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `assessTaskCriteria`, `bytes.NewReader`, `canAccess`, `dec.Decode`, `dec.DisallowUnknownFields`, `e.Error`, `errors.Is`, `http.Error`, `http.MaxBytesReader`, `int`, `io.ReadAll`, `json.NewDecoder`, `latestTaskAcceptance`, `len`, `new`, `operationError`, `prepareOperation`, `r.Context`, `r.Header.Get`, `r.URL.Query`, `r.URL.Query().Get`, `reply`, `requestActor`, `requestedRoom`, `s.appendBatch`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.validCSRF`, `s.workflowSafety`, `sharedTaskValueRoom`, `strconv.Atoi`, `strconv.Itoa`, `strings.TrimSpace`, `taskAcceptanceSnapshot`, `taskHasCurrentAcceptance`, `taskSchemaAssessments`, `taskcontract.ValidJSON`, `time.Now`, `time.Now().UTC`, `uint`, `utf8.RuneCountInString`, `validateTaskAcceptanceReview`, `withOperationResponse`.
