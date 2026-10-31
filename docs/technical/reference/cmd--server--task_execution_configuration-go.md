# cmd/server/task_execution_configuration.go

[Module index](cmd.md) · [Open source](../../../cmd/server/task_execution_configuration.go)

Go implementation for task execution configuration; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`taskParticipantConfigurationDigest`** — Hash stable configuration only. Session threads, status/telemetry, consent projection and derived company/actor fields legitimately change during a turn.
  Signature: `func(p room.Participant) string`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `sha256.Sum256`.
- **`taskConnectionConfigurationDigest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(c room.Connection) string`.
  Calls: `hex.EncodeToString`, `json.Marshal`, `sha256.Sum256`.
- **`bindTaskExecutionConfiguration`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, p room.Participant, e *room.TaskExecution) error`.
  Calls: `errors.New`, `taskConnectionConfigurationDigest`, `taskParticipantConfigurationDigest`.
- **`taskExecutionConfigurationCurrent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, e room.TaskExecution) bool`.
  Calls: `bindTaskExecutionConfiguration`.
- **`validTaskExecutionConfiguration`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, e room.TaskExecution) bool`.
  Calls: `taskExecutionConfigurationCurrent`.
- **`taskExecutionExportCurrent`** — Policy edits revoke pending authority even after an allow/deny/allow cycle.
  Signature: `func(st room.State, e room.TaskExecution) bool`.
  Calls: `contextDestination`.
