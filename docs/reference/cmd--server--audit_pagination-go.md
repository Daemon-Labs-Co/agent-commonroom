# cmd/server/audit_pagination.go

[Module index](cmd.md) · [Open source](../../cmd/server/audit_pagination.go)

Go implementation for audit pagination; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`auditPageBounds`** — auditPageBounds validates keyset pagination and half-open date windows. No caller input is interpolated into SQL; page size is fixed and bounded.
  Signature: `func(query url.Values) ([]any, error)`.
  Calls: `dates[0].Before`, `errors.New`, `query.Get`, `strconv.ParseInt`, `time.Parse`.
