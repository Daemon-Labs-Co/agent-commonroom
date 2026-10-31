# cmd/server/workflows.go

[Module index](cmd.md) · [Open source](../../../cmd/server/workflows.go)

Go implementation for workflows; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`validateGraph`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(w room.Workflow) error`.
  Calls: `fmt.Errorf`, `len`, `math.IsNaN`, `strings.TrimSpace`, `validPort`.
- **`validPort`** — Source has no prose contract; use the signature and direct calls below to trace behavior.
  Signature: `func(kind, port string) bool`.
