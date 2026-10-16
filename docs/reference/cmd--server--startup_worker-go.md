# cmd/server/startup_worker.go

[Module index](cmd.md) · [Open source](../../cmd/server/startup_worker.go)

Go implementation for startup worker; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).verifyWorkerAtStartup`** — verifyWorkerAtStartup refuses fresh dispatch over unverified orphan work. It does not adopt, kill or replay jobs. Source forks CAN skip it, BUT risk concurrent stale writers and bypassing controller admission limits.
  Signature: `func(parent context.Context, endpoint string)`.
  Calls: `auditCancel`, `cancel`, `context.WithTimeout`, `hasUnconfirmedManagedJobs`, `log.Print`, `probe`, `s.append`, `s.load`, `s.mu.Lock`, `s.mu.Unlock`, `s.tripIntegrityLocked`.
