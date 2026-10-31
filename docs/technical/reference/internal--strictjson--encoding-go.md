# internal/strictjson/encoding.go

[Module index](internal.md) · [Open source](../../../internal/strictjson/encoding.go)

Go implementation for encoding; declarations and calls are listed below.

Package: `strictjson`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`hex4`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte) (uint16, bool)`.
  Calls: `len`, `uint16`.
- **`ValidEncoding`** — ValidEncoding checks UTF-8 and escaped UTF-16 surrogate pairing without allocating or normalizing strings. It accepts explicit U+FFFD, not silent replacement of invalid bytes/escapes. Grammar errors outside this check must still be rejected by the caller's JSON decoder. Bound input before invocation.
  Signature: `func(raw []byte) bool`.
  Calls: `hex4`, `len`, `utf8.Valid`.
