# cmd/server/http_state.go

[Module index](cmd.md) · [Open source](../../../cmd/server/http_state.go)

Go implementation for http state; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`reply`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, v any)`.
  Calls: `json.NewEncoder`, `json.NewEncoder(w).Encode`, `w.Header`, `w.Header().Set`.
- **`(*server).session`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `http.Error`, `reply`, `requestActor`, `s.auth.mu.Lock`, `s.auth.mu.Unlock`, `s.auth.session`.
- **`(*server).state`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `http.Error`, `humanProfileImagesEnabled`, `r.Context`, `reply`, `requestActor`, `requestedRoom`, `s.agentSafetyControlsEnabled`, `s.agentsBlockedLocked`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `visibleState`.
