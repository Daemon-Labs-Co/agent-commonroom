# cmd/server/storage_deadlines.go

[Module index](cmd.md) · [Open source](../../cmd/server/storage_deadlines.go)

Go implementation for storage deadlines; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`(*server).storageContext`** — storageContext bounds cooperative DB work, including work holding s.mu. The private shorter duration is for deterministic tests; it cannot disable the production cap or extend an earlier caller deadline. Source forks CAN remove it, BUT stalled storage can block cancellation and exhaust handlers.
  Signature: `func(parent context.Context) (context.Context, context.CancelFunc)`.
  Calls: `context.WithTimeout`.
