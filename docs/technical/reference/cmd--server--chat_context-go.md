# cmd/server/chat_context.go

[Module index](cmd.md) · [Open source](../../../cmd/server/chat_context.go)

Go implementation for chat context; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`chatPrompt`** — chatPrompt bounds every concatenation before allocation. Untrusted context is reference material, never a grant of tool, host or recovery permissions.
  Signature: `func(st room.State, p room.Participant, text string) (string, error)`.
  Calls: `b.Len`, `b.String`, `b.WriteString`, `contextDestination`, `errors.New`, `exportReference`, `humanAttentionEnabled`, `len`, `memoryExportAllowed`, `organizationGuidance`, `profileGuidance`, `referenceJSON`, `write`.
- **`exportReference`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.Participant, policy *room.ContextExportPolicy) string`.
  Calls: `referenceJSON`.
