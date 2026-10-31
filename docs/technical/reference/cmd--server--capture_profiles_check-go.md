# cmd/server/capture_profiles_check.go

[Module index](cmd.md) · [Open source](../../../cmd/server/capture_profiles_check.go)

Go implementation for capture profiles check; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`captureProfileCheckMain`** — Offline operator preflight: no controller lease, database, daemon, listener, provider, model or network access. Success validates private configuration only; it grants no execution/export authority and proves no physical/live acceptance.
  Signature: `func(args []string, stdout, stderr io.Writer) int`.
  Calls: `err.Error`, `fmt.Fprintf`, `fmt.Fprintln`, `len`, `loadCaptureWorkerProfiles`.
