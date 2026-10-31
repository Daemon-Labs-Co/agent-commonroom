# cmd/server/connection_diagnostics_test.go

[Module index](cmd.md) · [Open source](../../../cmd/server/connection_diagnostics_test.go)

Automated regression checks for connection_diagnostics; read test names for individual contracts.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestConnectionReadinessStages`** — Regression contract: TestConnectionReadinessStages.
  Signature: `func(t *testing.T)`.
  Calls: `checkConnection`, `context.Background`, `decodeWorkerAuth`, `errors.New`, `json.Marshal`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestConnectionChecksDisabledTransport`** — Regression contract: TestConnectionChecksDisabledTransport.
  Signature: `func(t *testing.T)`.
  Calls: `checkConnection`, `context.Background`, `t.Fatal`, `t.Parallel`.
- **`TestConnectionTestAPIControls`** — Regression contract: TestConnectionTestAPIControls.
  Signature: `func(t *testing.T)`.
  Calls: `connectionTestingEnabled`, `db.seed`, `http.HandlerFunc`, `json.Unmarshal`, `mockRequest`, `mockServer`, `s.actorGuard`, `s.auth.protect`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.Bytes`, `w.Body.String`.
- **`TestDraftConnectionScope`** — Regression contract: TestDraftConnectionScope.
  Signature: `func(t *testing.T)`.
  Calls: `resolveConnectionTest`, `t.Fatal`, `t.Parallel`, `t.Run`.
- **`TestDraftRuntimeNeverReadsCredentials`** — Regression contract: TestDraftRuntimeNeverReadsCredentials.
  Signature: `func(t *testing.T)`.
  Calls: `checkConnectionDraft`, `context.Background`, `t.Fatal`, `t.Parallel`.
- **`TestDraftAPILeavesConnectionUnchanged`** — Regression contract: TestDraftAPILeavesConnectionUnchanged.
  Signature: `func(t *testing.T)`.
  Calls: `context.Background`, `db.seed`, `http.HandlerFunc`, `len`, `mockRequest`, `mockServer`, `s.actorGuard`, `s.auth.protect`, `s.load`, `t.Fatal`, `t.Parallel`, `w.Body.String`.
- **`TestConnectionProgressRealStageBoundaries`** — Regression contract: TestConnectionProgressRealStageBoundaries.
  Signature: `func(t *testing.T)`.
  Calls: `append`, `checkConnectionDraftProgress`, `context.Background`, `len`, `t.Fatal`, `t.Fatalf`, `t.Parallel`.
- **`TestConnectionProgressStreamControls`** — Regression contract: TestConnectionProgressStreamControls.
  Signature: `func(t *testing.T)`.
  Calls: `db.seed`, `h.ServeHTTP`, `http.HandlerFunc`, `httptest.NewRecorder`, `httptest.NewRequest`, `json.Unmarshal`, `len`, `mockServer`, `r.AddCookie`, `r.Header.Set`, `request`, `s.actorGuard`, `s.auth.protect`, `strings.NewReader`, `strings.Split`, `strings.TrimSpace`, `t.Fatal`, `t.Parallel`, `time.Now`, `time.Now().Add`, `w.Body.String`, `w.Header`, `w.Header().Get`.
