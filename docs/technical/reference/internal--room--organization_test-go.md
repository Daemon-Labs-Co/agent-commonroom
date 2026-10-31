# internal/room/organization_test.go

[Module index](internal.md) · [Open source](../../../internal/room/organization_test.go)

Automated regression checks for organization; read test names for individual contracts.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`TestOrganizationDecisionAuthority`** — Regression contract: TestOrganizationDecisionAuthority.
  Signature: `func(t *testing.T)`.
  Calls: `OrganizationDecisionPeers`, `OrganizationallyOutranks`, `reflect.DeepEqual`, `t.Fatal`, `t.Parallel`.
