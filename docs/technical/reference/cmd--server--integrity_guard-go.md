# cmd/server/integrity_guard.go

[Module index](cmd.md) · [Open source](../../../cmd/server/integrity_guard.go)

Go implementation for integrity guard; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).tripIntegrityLocked`** — tripIntegrityLocked is independent of storage availability. Its latch must never clear solely because a later probe succeeds or a UI flag is disabled.
  Signature: `func()`.
  Calls: `cancel`, `file.Close`, `file.Sync`, `file.WriteString`, `log.Print`, `os.IsExist`, `os.OpenFile`, `syncParentDirectory`.
- **`(*server).restoreIntegrityStop`** — The marker preserves an outage stop across restarts even when the database cannot record an incident. It is operational state, not an independent archive.
  Signature: `func()`.
  Calls: `os.IsNotExist`, `os.Lstat`.
- **`(*server).recoverIntegrityLocked`** — recoverIntegrityLocked never repairs or rewrites history. It records owner approval only after a fresh valid chain and verified worker shutdown.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `(&http.Request{Method: "GET"}).WithContext`, `cancel`, `context.WithTimeout`, `http.Error`, `http.MaxBytesReader`, `json.NewDecoder`, `json.NewDecoder(r.Body).Decode`, `json.Unmarshal`, `len`, `os.Getenv`, `os.IsNotExist`, `os.Remove`, `probe`, `r.Context`, `rec.body.Bytes`, `reply`, `requestActor`, `s.append`, `s.integrityLocked`, `s.tripIntegrityLocked`, `s.validCSRF`, `strings.TrimSpace`, `syncParentDirectory`.
- **`(*server).checkIntegrityNow`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func()`.
  Calls: `(&http.Request{Method: "GET"}).WithContext`, `cancel`, `context.Background`, `context.WithTimeout`, `s.integrityLocked`, `s.mu.Lock`, `s.mu.Unlock`.
- **`(*server).watchIntegrity`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(ctx context.Context, interval time.Duration)`.
  Calls: `ctx.Done`, `s.checkIntegrityNow`, `ticker.Stop`, `time.NewTicker`.
