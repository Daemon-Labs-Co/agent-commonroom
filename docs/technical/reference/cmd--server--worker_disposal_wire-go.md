# cmd/server/worker_disposal_wire.go

[Module index](cmd.md) · [Open source](../../../cmd/server/worker_disposal_wire.go)

Go implementation for worker disposal wire; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`fenceObjectFields`** — Destructive request/receipt objects require exact keys and unique fields. encoding/json alone accepts aliases, duplicate keys and null scalar values. The wire bound is intentionally shared with existing cancellation transport.
  Signature: `func(raw []byte, keys ...string) (map[string]json.RawMessage, error)`.
  Calls: `bytes.Equal`, `bytes.NewReader`, `bytes.TrimSpace`, `d.Decode`, `d.More`, `d.Token`, `errors.New`, `json.Delim`, `json.NewDecoder`, `len`, `strictjson.ValidEncoding`.
- **`validFenceWire`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(raw []byte, ack bool) bool`.
  Calls: `fenceObjectFields`.
