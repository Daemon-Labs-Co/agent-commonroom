# internal/artifactmanifest/manifest_unix.go

[Module index](internal.md) · [Open source](../../../internal/artifactmanifest/manifest_unix.go)

Go implementation for manifest unix; declarations and calls are listed below.

Package: `artifactmanifest`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`Resolve`** — Resolve preflights the ENTIRE finite claim set before any object read. Actual byte integrity, authorization and audit acknowledgment remain mandatory.
  Signature: `func(ctx context.Context, definition taskcontract.Definition, scope objectcustody.ExecutionScope, outputName string, claims []Claim, store *objectcustody.Store) (Manifest, error)`.
  Calls: `ctx.Err`, `metadata`, `store.Get`, `validFormatBytes`.
- **`Revalidate`** — Revalidate repeats actual byte checks after exact metadata preflight. Canonical digests alone never authorize delivery or certify observed execution checks.
  Signature: `func(ctx context.Context, definition taskcontract.Definition, scope objectcustody.ExecutionScope, outputName string, claimed Manifest, store *objectcustody.Store) error`.
  Calls: `Resolve`, `ValidateMetadata`, `len`, `make`, `reflect.DeepEqual`.
