# cmd/server/task_delegation.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_delegation.go)

Go implementation for task delegation; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `taskJobContextKey`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{}`.
- **Type `taskJobBinding`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { TaskID, DispatchID, DelegationID, DefinitionDigest string DefinitionEpoch int }`.
- **`managedTaskCurrent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, j room.ManagedJob) bool`.
  Calls: `taskDefinitionDigest`, `taskExecutionPendingCurrent`.
- **`delegationTaskDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(t room.Task) string`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `sha256.Sum256`, `string`, `taskDefinitionDigest`.
- **`delegationPeople`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, d room.TaskDelegation) bool`.
  Calls: `canAccess`, `contextDestination`, `member`, `room.SupportsAgent`.
- **`delegationCurrent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, d room.TaskDelegation) bool`.
  Calls: `coordinationJobAllowed`, `d.ExpiresAt.After`, `delegationPeople`, `delegationTaskDigest`, `time.Now`.
- **`delegatedJobAllowed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, j room.ManagedJob) bool`.
  Calls: `delegationCurrent`, `j.Deadline.After`.
- **`delegationContextAllowed`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, st room.State, p room.Participant) bool`.
  Calls: `ctx.Value`, `delegationCurrent`.
- **`(*server).delegationRepositoryFor`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(rid string) (string, error)`.
  Calls: `controllerDevelopmentRepository`, `os.Getenv`, `s.delegationRepository`.
- **`(*controllerToolAuditor).RecordDelegation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, scope toolbroker.Scope, request toolbroker.DelegationRequest) error`.
  Calls: `a.check`, `context.WithValue`, `delegationPeople`, `delegationTaskDigest`, `errors.Is`, `hex.EncodeToString`, `member`, `pairedLifecycleVersion`, `s.append`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `sha256.Sum256`, `toolbroker.ValidDelegation`.
- **Type `delegationDecision`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { ID string 'json:"id"' Revision int 'json:"revision"' }`.
- **`projectDelegation`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st *room.State, kind string, raw []byte) error`.
  Calls: `coordinationJobAllowed`, `d.ExpiresAt.Equal`, `delegationPeople`, `delegationTaskDigest`, `errors.New`, `json.Unmarshal`, `managedTaskCurrent`, `pairedLifecycleVersion`, `projectActors`, `repositoryDigestPattern.MatchString`, `validWorkerJobID`.
- **`(*server).taskDelegations`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `canAccess`, `delegationCurrent`, `delegationTaskDigest`, `errors.Is`, `fenceObjectFields`, `hasRoom`, `http.Error`, `http.MaxBytesReader`, `io.ReadAll`, `json.Unmarshal`, `len`, `operationError`, `prepareOperation`, `r.Context`, `r.Header.Get`, `reply`, `requestActor`, `requestedRoom`, `s.append`, `s.delegationRepositoryFor`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.validCSRF`, `s.workflowSafety`, `sort.Slice`, `withOperationResponse`.
