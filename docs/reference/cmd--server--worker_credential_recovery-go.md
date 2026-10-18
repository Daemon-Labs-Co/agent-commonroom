# cmd/server/worker_credential_recovery.go

[Module index](cmd.md) · [Open source](../../cmd/server/worker_credential_recovery.go)

Go implementation for worker credential recovery; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`workerCredentialReady`** — workerCredentialReady treats any marker, unreadable entry or malformed path as unresolved custody. Its contents are deliberately not needed to deny reuse. This is a mandatory worker security boundary, not an optional feature.
  Signature: `func(root string) bool`.
  Calls: `filepath.Join`, `os.IsNotExist`, `os.Lstat`.
- **`markWorkerCredentialUncertain`** — markWorkerCredentialUncertain runs under the connection credential lock and durably blocks reuse BEFORE any bytes are supplied to an engine. Process exit, flock release and confirmed container absence cannot establish refreshed-token custody. No token, model output or customer data belongs in this marker.
  Signature: `func(root string) error`.
  Calls: `markWorkerCredentialUncertainDurably`.
- **`markWorkerCredentialUncertainDurably`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(root string, syncDirectory func(string) error) error`.
  Calls: `errors.New`, `filepath.Join`, `replacePrivateDurably`.
- **`clearWorkerCredentialRecovery`** — clearWorkerCredentialRecovery is permitted only after a new login or local revocation has been durably committed under the lock, with execution quiescent. Removal durability errors remain errors; never acknowledge uncertain recovery.
  Signature: `func(root string, syncDirectory func(string) error) error`.
  Calls: `errors.New`, `filepath.Join`, `os.IsNotExist`, `os.Remove`, `syncDirectory`.
- **`workerCredentialConnectionQuiescent`** — workerCredentialConnectionQuiescent permits explicit reauthentication after execution roles are fenced even while source is retained. Unknown or corrupt journal entries refuse recovery. A released flock alone is not shutdown proof.
  Signature: `func(dir, connection, provider string) bool`.
  Calls: `boundedWorkerRecords`, `pairRoleSettled`, `validWorkerRecord`, `workerJobSettled`.
