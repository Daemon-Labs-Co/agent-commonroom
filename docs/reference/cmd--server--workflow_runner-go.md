# cmd/server/workflow_runner.go

[Module index](cmd.md) · [Open source](../../cmd/server/workflow_runner.go)

Go implementation for workflow runner; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`workflowExecutionEnabled`** — workflowExecutionEnabled defaults off. Owner-controlled orchestration never grants repository permissions or changes mandatory container isolation.
  Signature: `func(st room.State) bool`.
- **`validateRunnable`** — validateRunnable rejects unsupported semantics instead of treating saved graphs as executable guarantees. Every finish path requires human acceptance.
  Signature: `func(w room.Workflow, st room.State) error`.
  Calls: `append`, `errors.New`, `fmt.Errorf`, `hasMember`, `len`, `nextWorkflowNode`, `runnablePorts`, `startWorkflowNode`, `validateGraph`, `workflowNode`.
- **`runnablePorts`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(kind string) []string`.
- **`workflowNode`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w room.Workflow, id string) room.WorkflowNode`.
- **`startWorkflowNode`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w room.Workflow) string`.
- **`nextWorkflowNode`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w room.Workflow, id, port string) string`.
- **`workflowTerminal`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(status string) bool`.
- **`(*server).saveWorkflowRun`** — saveWorkflowRun creates an immutable revision; graph revisions remain pinned.
  Signature: `func(run *room.WorkflowRun) error`.
  Calls: `context.Background`, `s.append`, `time.Now`, `time.Now().UTC`.
- **`(*server).workflowRuns`** — workflowRuns handles owner-authorized starts, gates, cancellation and explicit recovery. No provider call is made in a request or without safety checks.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `appendWorkflowStep`, `cancel`, `err.Error`, `hasRoom`, `http.Error`, `http.MaxBytesReader`, `id`, `json.NewDecoder`, `json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024)).Decode`, `len`, `nextWorkflowNode`, `operationError`, `prepareOperation`, `r.Context`, `r.WithContext`, `reply`, `requestActor`, `requestedRoom`, `s.launchWorkflow`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.saveOperationRun`, `s.saveWorkflowRun`, `s.validCSRF`, `s.workflowAssignments`, `s.workflowSafety`, `scoped`, `startWorkflowNode`, `strings.TrimSpace`, `validateRunnable`, `w.WriteHeader`, `workflowExecutionEnabled`, `workflowNode`, `workflowTerminal`.
- **`(*server).workflowSafety`** — workflowSafety verifies the audit chain before every dispatch and human gate. Emergency-stop enforcement is shared with ordinary managed model turns.
  Signature: `func(ctx context.Context, st room.State, rid string) error`.
  Calls: `(&http.Request{Method: "GET"}).WithContext`, `body.Bytes`, `errors.New`, `json.Unmarshal`, `s.agentsBlockedLocked`, `s.integrityLocked`.
- **`workflowAgent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, rid, pid string) (room.Participant, string, error)`.
  Calls: `errors.New`, `room.SupportsAgent`.
- **`(*server).workflowAssignments`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, graph room.Workflow) error`.
  Calls: `errors.New`, `ready`, `workflowAgent`.
- **`(*server).launchWorkflow`** — launchWorkflow is called under the server lock. Only one live driver can own a run. Browser disconnection has no effect; restart requires explicit recovery.
  Signature: `func(runID string)`.
  Calls: `context.Background`, `context.WithCancel`, `s.driveWorkflow`.
- **`appendWorkflowStep`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(run *room.WorkflowRun, n room.WorkflowNode, output, evidence, port string)`.
  Calls: `append`, `hex.EncodeToString`, `sha256.Sum256`, `time.Now`, `time.Now().UTC`.
- **`workflowPrompt`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, run room.WorkflowRun, p room.Participant, n room.WorkflowNode) (string, error)`.
  Calls: `contextDestination`, `errors.New`, `exportReference`, `len`, `organizationGuidance`, `profileGuidance`.
- **`(*server).driveWorkflow`** — driveWorkflow persists intent before a model call and outcome afterward. Cancellation, flag changes, membership removal and integrity failure discard late outputs. No automatic retries replay potentially mutating work.
  Signature: `func(ctx context.Context, cancel context.CancelFunc, runID string)`.
  Calls: `append`, `appendWorkflowStep`, `cancel`, `context.Background`, `context.WithTimeout`, `ctx.Done`, `ctx.Err`, `delete`, `e.Error`, `errors.New`, `fail`, `hasMember`, `json.Unmarshal`, `len`, `nextWorkflowNode`, `ready`, `s.agentAdmission`, `s.appendBatch`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.persistWorkflowTransitionLocked`, `s.tripIntegrityLocked`, `s.workflowSafety`, `stop`, `strings.TrimSpace`, `time.Duration`, `time.NewTimer`, `time.Now`, `time.Now().UTC`, `timer.Stop`, `turn`, `turnCtx.Err`, `withInitiatingActor`, `workflowAgent`, `workflowExecutionEnabled`, `workflowNode`, `workflowPrompt`.
- **`(*server).saveOperationRun`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, run *room.WorkflowRun, status int) error`.
  Calls: `s.append`, `time.Now`, `time.Now().UTC`, `withOperationResponse`.
