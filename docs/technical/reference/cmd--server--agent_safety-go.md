# cmd/server/agent_safety.go

[Module index](cmd.md) · [Open source](../../../cmd/server/agent_safety.go)

Go implementation for agent safety; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).agentsBlockedLocked`** — agentsBlockedLocked requires s.mu. Persisted stops survive restart; volatile latches stop dispatch locally even when recording the stop fails.
  Signature: `func(st room.State, rid string) bool`.
- **`(*server).agentSafety`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `(&http.Request{Method: "GET"}).WithContext`, `cancel`, `context.WithTimeout`, `delete`, `hasRoom`, `hasUnconfirmedManagedJobs`, `http.Error`, `http.MaxBytesReader`, `json.NewDecoder`, `json.NewDecoder(r.Body).Decode`, `json.Unmarshal`, `len`, `os.Getenv`, `probe`, `r.Context`, `rec.body.Bytes`, `reply`, `requestActor`, `requestedRoom`, `s.append`, `s.integrityLocked`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.validCSRF`, `strings.TrimSpace`.
- **`agentSafetyUIEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() (bool, error)`.
  Calls: `os.Getenv`, `strconv.ParseBool`.
- **`(*server).agentSafetyControlsEnabled`** — The stored owner setting overrides the environment default. Neither can disable cutoff enforcement; only the optional presentation script is gated.
  Signature: `func(st room.State) bool`.
- **`(*server).agentSafetyScript`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(root string) http.Handler`.
  Calls: `cancel`, `context.WithTimeout`, `http.Error`, `http.HandlerFunc`, `optionalScriptHandler`, `optionalScriptHandler(s.agentSafetyControlsEnabled(st), root, "agent-safety.js").ServeHTTP`, `r.Context`, `s.agentSafetyControlsEnabled`, `s.load`, `w.Header`, `w.Header().Set`.
