# cmd/server/tool_broker_audit.go

[Module index](cmd.md) · [Open source](../../cmd/server/tool_broker_audit.go)

Go implementation for tool broker audit; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `toolCallOutcomeRecord`**: toolCallOutcomeRecord decodes historical tool.call.outcome events only. New actions use worker.tool.intent/outcome via the scoped controller auditor. Preserve this wire shape and its projection; historical evidence is never rewritten or converted into current execution authority.
  Declaration: `struct { Intent toolbroker.Intent 'json:"intent"' Outcome toolbroker.Outcome 'json:"outcome"' }`.
