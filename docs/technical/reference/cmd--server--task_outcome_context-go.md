# cmd/server/task_outcome_context.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_outcome_context.go)

Go implementation for task outcome context; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`taskDispatchPrompt`** — Room deliverables are optional reference material, never task scope or tool authority. Opt-in export must also satisfy the room's provider disclosure policy. Ordinary chat and non-task execution paths do not use this helper.
  Signature: `func(st room.State, p room.Participant, text string, taskIDs ...string) (string, error)`.
  Calls: `chatPrompt`, `contextDestination`, `errors.New`, `len`, `referenceJSON`, `utf8.RuneCountInString`, `utf8.ValidString`.
