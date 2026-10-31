# cmd/server/startup_command.go

[Module index](cmd.md) · [Open source](../../../cmd/server/startup_command.go)

Go implementation for startup command; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`startupCommand`** — Validate command intent before controller leases, configuration or DB startup. A typo in an operator preflight must never silently start the application or migrate a database. Diagnostics never echo arbitrary arguments or credentials. This startup safety boundary has no optional feature-flag bypass.
  Signature: `func(args []string, stdout, stderr io.Writer) (handled bool, code int)`.
  Calls: `captureProfileCheckMain`, `fmt.Fprintln`, `len`, `objectReadProfileCheckMain`, `usage`, `workerStorageCheckMain`.
