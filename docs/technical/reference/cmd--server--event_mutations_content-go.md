# cmd/server/event_mutations_content.go

[Module index](cmd.md) · [Open source](../../../cmd/server/event_mutations_content.go)

Go implementation for event mutations content; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).prepareContentEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request, st room.State, roomID string, req eventRequest) (any, bool)`.
  Calls: `contextPolicyEditorEnabled`, `http.Error`, `humanAttentionEnabled`, `id`, `json.Unmarshal`, `member`, `referenceJSON`, `reply`, `requestActor`, `strings.TrimSpace`, `time.Now`, `time.Now().UTC`, `validExportProviders`.
