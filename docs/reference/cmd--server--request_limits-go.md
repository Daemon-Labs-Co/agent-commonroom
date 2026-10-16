# cmd/server/request_limits.go

[Module index](cmd.md) · [Open source](../../cmd/server/request_limits.go)

Go implementation for request limits; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **Type `actorRequestBudget`**: Structured data type; fields are shown in the declaration.
  Declaration: `struct { Tokens float64 Updated time.Time }`.
- **`(*server).admitActorMutation`** — admitActorMutation is a mandatory single-controller budget: twenty-request bursts, refilling at one request/second. Emergency stop bypasses this policy through actorGuard's existing authenticated-owner safety path.
  Signature: `func(actor string, now time.Time) bool`.
  Calls: `delete`, `len`, `now.Sub`, `now.Sub(b.Updated).Seconds`, `s.mu.Lock`, `s.mu.Unlock`.
