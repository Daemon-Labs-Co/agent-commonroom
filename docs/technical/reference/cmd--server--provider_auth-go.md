# cmd/server/provider_auth.go

[Module index](cmd.md) · [Open source](../../../cmd/server/provider_auth.go)

Go implementation for provider auth; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`localClaudeConnection`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(stConnectionProtocol, endpoint string) bool`.
- **`shellArgument`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(value string) string`.
  Calls: `strings.ReplaceAll`.
- **Type `cappedStatusOutput`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct{ bytes.Buffer }`.
- **`(*cappedStatusOutput).Write`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p []byte) (int, error)`.
  Calls: `b.Buffer.Write`, `b.Len`, `len`.
- **`participantHasUnsettledWork`** — providerAuth exposes owner-authorized connection login/status through the approved isolated SSH worker. Passwords and OAuth codes remain in Terminal. Disabling its feature flag cannot disable execution isolation or restore a host CLI fallback. Source forks CAN bypass this, BUT expose host credentials. Durable uncertainty survives controller restart. Configuration/authentication must retain the original connection until every bound job and task is settled.
  Signature: `func(st room.State, active map[string]bool, pid string) bool`.
- **`connectionHasActiveAgent`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, active map[string]bool, connection string) bool`.
  Calls: `participantHasUnsettledWork`, `verificationParticipant`.
- **`(*server).providerAuth`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `cancel`, `cmd.Run`, `connectionHasActiveAgent`, `connectionSignInEnabled`, `context.WithTimeout`, `decodeWorkerCredentialStatus`, `encodeWorkerAuth`, `err.Error`, `exec.CommandContext`, `exec.CommandContext(ctx, "/usr/bin/open", "-a", "Terminal", path).Run`, `filepath.Abs`, `http.Error`, `http.MaxBytesReader`, `interactiveWorkerLogin`, `json.NewDecoder`, `json.NewDecoder(r.Body).Decode`, `json.Unmarshal`, `os.CreateTemp`, `os.Getenv`, `os.MkdirAll`, `os.Remove`, `out.Bytes`, `output.Bytes`, `r.Context`, `remoteWorkerEndpoint`, `reply`, `requestActor`, `requestedRoom`, `s.agentsBlockedLocked`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.validCSRF`, `script.Chmod`, `script.Close`, `script.Name`, `script.WriteString`, `workerSSHCommand`.
- **`decodeWorkerCredentialStatus`** — decodeWorkerCredentialStatus never treats uncertain custody as a usable login, including contradictory worker replies. Older workers omit the optional bit.
  Signature: `func(raw []byte) (bool, bool, error)`.
  Calls: `errors.New`, `json.Unmarshal`.
