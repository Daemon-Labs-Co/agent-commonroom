# cmd/server/initiator_context.go

[Module index](cmd.md) · [Open source](../../cmd/server/initiator_context.go)

Go implementation for initiator context; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`detachedInitiatorContext`** — detachedInitiatorContext outlives the HTTP request while copying only its authenticated actor, never client payloads or request-scoped authority values. Missing provenance remains missing; it must never default to the owner.
  Signature: `func(request context.Context) context.Context`.
  Calls: `context.Background`, `request.Value`, `withInitiatingActor`.
- **`withInitiatingActor`** — withInitiatingActor binds a trusted request identity or recorded workflow initiator to execution. Its parent retains workflow cancellation/deadlines.
  Signature: `func(parent context.Context, actor string) context.Context`.
  Calls: `context.WithValue`.
