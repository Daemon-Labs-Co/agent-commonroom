# cmd/server/human_attention.go

[Module index](cmd.md) · [Open source](../../cmd/server/human_attention.go)

Go implementation for human attention; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`humanAttentionEnabled`** — humanAttentionEnabled controls detection, projection and acknowledgment, never authorization or workflow acceptance. Existing requests remain in history.
  Signature: `func(st room.State) bool`.
- **`humanRequestKind`** — humanRequestKind honors an explicit first-line protocol; conservative phrases are suggestions for older free-text responses, not authoritative decisions.
  Signature: `func(text string) string`.
  Calls: `strings.Contains`, `strings.HasPrefix`, `strings.SplitN`, `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace`.
- **`projectHumanAttention`** — projectHumanAttention adds display metadata without rewriting message history.
  Signature: `func(st *room.State, acknowledged map[string]string)`.
  Calls: `humanAttentionEnabled`, `humanRequestKind`.
