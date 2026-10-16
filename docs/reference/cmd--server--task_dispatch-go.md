# cmd/server/task_dispatch.go

[Module index](cmd.md) · [Open source](../../cmd/server/task_dispatch.go)

Go implementation for task dispatch; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).taskDispatch`** — taskDispatch commits assignment, message and intent before any managed turn. The optional capability can be disabled, but never falls back to client-owned split writes or uncontained execution. Caller identities are human-only.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `canAccess`, `cancel`, `chatPrompt`, `context.WithTimeout`, `dec.Decode`, `dec.DisallowUnknownFields`, `detachedInitiatorContext`, `err.Error`, `errors.New`, `hasRoom`, `http.Error`, `http.MaxBytesReader`, `id`, `json.NewDecoder`, `len`, `member`, `new`, `operationError`, `os.Getenv`, `prepareOperation`, `probe`, `r.Context`, `ready`, `reply`, `requestActor`, `requestedRoom`, `s.agentAdmission`, `s.appendBatch`, `s.finishAgentTurnWithTask`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.validCSRF`, `s.workflowSafety`, `scoped`, `strings.HasPrefix`, `strings.TrimSpace`, `taskDispatchEnabled`, `taskOutcome`, `time.Now`, `time.Now().UTC`, `turn`, `w.WriteHeader`, `withOperationResponse`, `workflowAgent`.
- **`taskDispatchEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) bool`.
- **`taskOutcome`** — taskOutcome never auto-accepts work or overwrites a different dispatch. Pending intents survive restart and require explicit owner reconciliation; they are not automatically replayed following uncertain external effects.
  Signature: `func(st room.State, taskID, dispatchID, status string) []auditChange`.
