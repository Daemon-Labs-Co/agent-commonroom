# cmd/server/workflow_templates.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workflow_templates.go)

Go implementation for workflow templates; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`workflowTemplatesEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (bool, error)`.
  Calls: `os.Getenv`, `strconv.ParseBool`.
- **`knownWorkflowTemplate`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(id string) bool`.
