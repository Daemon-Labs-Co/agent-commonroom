# cmd/server/events_handler.go

[Module index](cmd.md) · [Open source](../../../cmd/server/events_handler.go)

Go implementation for events handler; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `eventRequest`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { TemplateID string 'json:"templateId"' Kind string 'json:"kind"' Payload json.RawMessage 'json:"payload"' }`.
- **`(*server).events`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `auditPageBounds`, `auditSummary`, `cancel`, `context.WithValue`, `ctx.Err`, `describeAuditEvent`, `eligibleGovernor`, `err.Error`, `hasRoom`, `hasTaskValueSources`, `http.Error`, `http.MaxBytesReader`, `json.NewDecoder`, `json.NewDecoder(r.Body).Decode`, `json.Unmarshal`, `knownWorkflowTemplate`, `len`, `log.Print`, `member`, `operationError`, `prepareOperation`, `r.Context`, `r.URL.Query`, `r.URL.Query().Get`, `r.WithContext`, `reply`, `requestActor`, `requestedRoom`, `rows.Close`, `rows.Err`, `rows.Next`, `rows.Scan`, `s.append`, `s.appendAppointmentBatch`, `s.appendBatch`, `s.cancelStaleCoordinationLocked`, `s.db.Query`, `s.historyPresentationEnabled`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.prepareAgentsEvent`, `s.prepareContentEvent`, `s.prepareIdentityEvent`, `s.prepareRoomsEvent`, `s.storageContext`, `s.validCSRF`, `s.workflowSafety`, `withOperationResponse`.
