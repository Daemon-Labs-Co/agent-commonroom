# cmd/server/event_mutations_identity.go

[Module index](cmd.md) · [Open source](../../cmd/server/event_mutations_identity.go)

Go implementation for event mutations identity; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).prepareIdentityEvent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request, st room.State, roomID string, req eventRequest) (any, bool)`.
  Calls: `err.Error`, `http.Error`, `humanProfileImagesEnabled`, `json.Unmarshal`, `len`, `member`, `normalizeProfileImage`, `requestActor`, `strings.TrimSpace`, `validateOrganization`.
