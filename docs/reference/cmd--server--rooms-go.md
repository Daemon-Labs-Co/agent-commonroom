# cmd/server/rooms.go

[Module index](cmd.md) · [Open source](../../cmd/server/rooms.go)

Go implementation for rooms; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`requestedRoom`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(r *http.Request) string`.
  Calls: `r.URL.Query`, `r.URL.Query().Get`.
- **`hasRoom`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, id string) bool`.
- **`hasMember`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, rid, pid string) bool`.
- **`scoped`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State, rid string) room.State`.
  Calls: `append`.
- **`profileGuidance`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(p room.Participant) string`.
  Calls: `fmt.Sprintf`.
- **`personalName`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(st room.State) string`.
  Calls: `big.NewInt`, `id`, `int`, `int64`, `len`, `offset.Int64`, `rand.Int`.
- **`providerForEndpoint`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(endpoint string) string`.
  Calls: `strings.HasSuffix`.
