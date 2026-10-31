# cmd/server/connection_diagnostics.go

[Module index](cmd.md) · [Open source](../../../cmd/server/connection_diagnostics.go)

Go implementation for connection diagnostics; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `connectionCheck`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Name string 'json:"name"' Status string 'json:"status"' Detail string 'json:"detail"' }`.
- **Type `connectionTestResult`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Status string 'json:"status"' Checks []connectionCheck 'json:"checks"' CheckedAt time.Time 'json:"checkedAt"' ProviderResponseTested bool 'json:"providerResponseTested"' }`.
- **Type `workerProbe`**: Named data type; underlying representation is shown in the declaration.
  Declaration: `func(context.Context, string, string, string) ([]byte, error)`.
- **`connectionTestingEnabled`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(flags map[string]bool) bool`.
- **`probeWorker`** — probeWorker performs bounded, verified SSH diagnostics only. It never starts unrestricted commands, signs in, forwards credentials, or follows arbitrary URLs.
  Signature: `func(ctx context.Context, alias, action, value string) ([]byte, error)`.
  Calls: `cmd.Run`, `ctx.Err`, `diagnostics.String`, `errors.New`, `out.Bytes`, `strings.Contains`, `workerSSHCommand`.
- **`checkConnection`** — checkConnection distinguishes execution readiness from provider validation. Registration-only adapters never receive a misleading successful result.
  Signature: `func(ctx context.Context, c room.Connection, alias, enabled string, probe workerProbe) connectionTestResult`.
  Calls: `checkConnectionDraft`.
- **`checkConnectionDraft`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, c room.Connection, alias, enabled string, probe workerProbe, registered bool) connectionTestResult`.
  Calls: `checkConnectionDraftProgress`.
- **`checkConnectionDraftProgress`** — Progress reports real stage boundaries; SSH/Docker readiness remains one worker probe.
  Signature: `func(ctx context.Context, c room.Connection, alias, enabled string, probe workerProbe, registered bool, progress func(connectionCheck)) connectionTestResult`.
  Calls: `add`, `append`, `encodeWorkerAuth`, `err.Error`, `json.Unmarshal`, `probe`, `progress`, `remoteWorkerEndpoint`, `start`, `strings.HasPrefix`, `time.Now`, `time.Now().UTC`.
- **`(*server).connectionTest`** — connectionTest requires owner authorization, CSRF and incident clearance. Results are transient in the UI; only safe status metadata enters the audit.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `auditCancel`, `cancel`, `checkConnectionDraftProgress`, `connectionTestingEnabled`, `context.Background`, `context.WithTimeout`, `d.Decode`, `d.DisallowUnknownFields`, `encoder.Encode`, `err.Error`, `f.Flush`, `http.Error`, `http.MaxBytesReader`, `json.NewDecoder`, `json.NewEncoder`, `new`, `os.Getenv`, `probe`, `r.Context`, `r.Header.Get`, `reply`, `requestActor`, `requestedRoom`, `resolveConnectionTest`, `s.agentsBlockedLocked`, `s.append`, `s.load`, `s.managedVerification`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`, `s.validCSRF`, `send`, `strings.Contains`, `w.Header`, `w.Header().Set`.
- **Type `connectionDraft`**: Draft tests never persist edits or borrow another connection's credentials. Only a registered unchanged execution configuration can verify a model turn.
  Declaration: `struct { Name string 'json:"name"' Protocol string 'json:"protocol"' Endpoint string 'json:"endpoint"' Provider string 'json:"provider"' Location string 'json:"location"' }`.
- **`resolveConnectionTest`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(connections []room.Connection, connectionID string, draft *connectionDraft) (*room.Connection, bool, error)`.
  Calls: `errors.New`, `len`, `strings.TrimSpace`.
