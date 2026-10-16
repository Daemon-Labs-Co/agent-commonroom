# cmd/server/credential_boundary.go

[Module index](cmd.md) · [Open source](../../cmd/server/credential_boundary.go)

Go implementation for credential boundary; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`requireCredentialSeparatedDevelopment`** — requireCredentialSeparatedDevelopment closes the legacy shared-credential execution path. A fork CAN remove it, BUT model-controlled native tools then regain readable provider credentials. No feature flag, owner role or paid entitlement may restore this unsafe fallback. Chat/login remain separate.
  Signature: `func(repository string) error`.
