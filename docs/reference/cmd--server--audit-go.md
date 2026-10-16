# cmd/server/audit.go

[Module index](cmd.md) · [Open source](../../cmd/server/audit.go)

Go implementation for audit; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`auditSummary`** — auditSummary derives a display description without changing the signed event payload or exposing message bodies, credentials, instructions, or raw errors.
  Signature: `func(kind string, payload []byte) string`.
  Calls: `json.Unmarshal`, `label`, `len`, `string`, `strings.Map`, `strings.ReplaceAll`, `strings.TrimSpace`, `unicode.IsControl`.
