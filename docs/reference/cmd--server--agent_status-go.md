# cmd/server/agent_status.go

[Module index](cmd.md) · [Open source](../../cmd/server/agent_status.go)

Go implementation for agent status; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`safeAgentFailure`** — safeAgentFailure uses allowlisted explanations rather than displaying raw stderr, which can contain credentials, private paths or prompt material.
  Signature: `func(raw string) (reason, next string)`.
  Calls: `strings.Contains`, `strings.ToLower`.
- **`projectAgentElapsed`** — projectAgentElapsed reports controller dispatch age, not process health or useful progress. No heartbeat may extend the hard execution budget.
  Signature: `func(st *room.State, starts map[string]time.Time, now time.Time)`.
  Calls: `containsRoom`, `fmt.Sprintf`, `int`, `now.Sub`, `now.Sub(start).Seconds`, `start.After`, `start.IsZero`.
- **`statusActivity`** — statusActivity bounds operator-controlled workflow titles for safe display.
  Signature: `func(text string) string`.
  Calls: `len`, `string`, `strings.Map`, `strings.TrimSpace`, `unicode.IsControl`.
- **`projectAgentStatusDetails`** — projectAgentStatusDetails derives current explanations from retained events. Disabled delivery omits details server-side; historical audit records remain.
  Signature: `func(st *room.State, activities, failures map[string]string)`.
  Calls: `containsRoom`, `safeAgentFailure`, `statusActivity`.
- **`containsRoom`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(rooms []string, id string) bool`.
