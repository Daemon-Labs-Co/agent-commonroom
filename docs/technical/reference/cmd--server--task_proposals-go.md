# cmd/server/task_proposals.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_proposals.go)

Go implementation for task proposals; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `proposalDecision`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ProposalID string 'json:"proposalId"' Revision int 'json:"revision"' Task room.Task 'json:"task"' }`.
- **`proposalSource`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, rid, mid string) (room.Message, bool)`.
- **`proposalText`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(title, criteria string) bool`.
  Calls: `len`, `strings.TrimSpace`.
- **`projectTaskProposal`** — Only the authenticated human capture/review endpoint can create these facts. A source message is untrusted content, never an approval or execution grant.
  Signature: `func(st *room.State, kind string, data []byte) error`.
  Calls: `append`, `errors.New`, `hasRoom`, `json.Unmarshal`, `membershipPartPattern.MatchString`, `projectActors`, `proposalSource`, `proposalText`, `strings.TrimSpace`.
- **`(*server).taskProposals`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `canAccess`, `errors.Is`, `fenceObjectFields`, `hasRoom`, `http.Error`, `http.MaxBytesReader`, `id`, `io.ReadAll`, `json.Unmarshal`, `len`, `operationError`, `prepareOperation`, `proposalSource`, `proposalText`, `r.Context`, `r.Header.Get`, `reply`, `requestActor`, `requestedRoom`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.validCSRF`, `s.workflowSafety`, `sort.Slice`, `strings.TrimSpace`, `withOperationResponse`.
