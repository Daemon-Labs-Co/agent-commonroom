# internal/room/organization.go

[Module index](internal.md) · [Open source](../../../internal/room/organization.go)

Go implementation for organization; declarations and calls are listed below.

Package: `room`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`OrganizationallyOutranks`** — OrganizationallyOutranks compares reporting ancestry, not job titles or saved legacy ranks. Unrelated branches and multiple profiles of one identity are peers.
  Signature: `func(chart OrganizationChart, superior, report string) bool`.
- **`OrganizationDecisionPeers`** — OrganizationDecisionPeers identifies maximal available identities. Multiple candidates must reach consensus; an absent manager cannot resolve a conflict. Explicit human overrides remain final, independent of this automated candidate set.
  Signature: `func(chart OrganizationChart, available []string) []string`.
  Calls: `OrganizationallyOutranks`, `append`.
