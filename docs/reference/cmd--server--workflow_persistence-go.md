# cmd/server/workflow_persistence.go

[Module index](cmd.md) · [Open source](../../cmd/server/workflow_persistence.go)

Go implementation for workflow persistence; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).persistWorkflowTransitionLocked`** — persistWorkflowTransitionLocked commits run and participant outcomes together. Callers hold s.mu. A storage failure latches the mandatory execution stop; neither optional flags nor a later successful write may clear that latch.
  Signature: `func(run *room.WorkflowRun, effects ...auditChange) error`.
  Calls: `append`, `context.Background`, `log.Print`, `s.appendBatch`, `s.tripIntegrityLocked`, `time.Now`, `time.Now().UTC`.
