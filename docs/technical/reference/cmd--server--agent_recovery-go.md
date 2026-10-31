# cmd/server/agent_recovery.go

[Module index](cmd.md) · [Open source](../../../cmd/server/agent_recovery.go)

Go implementation for agent recovery; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`recoveryEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(flags map[string]bool) bool`.
- **`(*server).agentRecovery`** — agentRecovery never launches a model or replays a task. Quiescence is required even when an old worker job's identity was not retained. This deliberately blocks recovery while any other managed job runs on the shared worker.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `agentExecutionSession`, `cancel`, `context.WithTimeout`, `controllerDevelopmentRepository`, `http.Error`, `http.MaxBytesReader`, `json.NewDecoder`, `json.NewDecoder(r.Body).Decode`, `os.Getenv`, `probe`, `r.Context`, `recoveryEnabled`, `reply`, `requestActor`, `requestedRoom`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.validCSRF`, `s.workflowSafety`, `workflowAgent`.
- **`remoteWorkerIdle`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, endpoint string) (bool, error)`.
  Calls: `cmd.Run`, `errors.New`, `json.Unmarshal`, `os.Getenv`, `output.Bytes`, `remoteWorkerEndpoint`, `workerSSHCommand`.
