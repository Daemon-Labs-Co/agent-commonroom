# cmd/server/event_identity.go

[Module index](cmd.md) · [Open source](../../cmd/server/event_identity.go)

Go implementation for event identity; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`id`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func() string`.
  Calls: `hex.EncodeToString`, `make`, `panic`, `rand.Read`.
- **`checksum`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(previous, kind, payload string) string`.
  Calls: `hex.EncodeToString`, `sha256.Sum256`.
