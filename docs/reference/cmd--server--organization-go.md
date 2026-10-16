# cmd/server/organization.go

[Module index](cmd.md) · [Open source](../../cmd/server/organization.go)

Go implementation for organization; declarations and calls are listed below.

Package: `main`. Go uses types, structs, interfaces and receiver methods rather than classes.

- **`validateOrganization`** — validateOrganization checks a bounded company graph, independently of layout. Reporting lines replace numbered authority, never account roles or room access.
  Signature: `func(chart room.OrganizationChart, actors []room.Actor) error`.
  Calls: `errors.New`, `len`, `math.IsInf`, `math.IsNaN`.
- **`(*server).organization`** — organization exposes only company identity metadata to the owner. It omits keys, endpoints, conversations and agent instructions/context.
  Signature: `func(w http.ResponseWriter, r *http.Request)`.
  Calls: `append`, `http.Error`, `r.Context`, `reply`, `requestActor`, `s.load`, `strings.TrimSpace`, `w.Header`, `w.Header().Set`.
- **`organizationGuidance`** — organizationGuidance supplies only relationships between identities currently participating in this room. It is context, not automatic conflict arbitration.
  Signature: `func(st room.State, participant room.Participant, rid string) string`.
  Calls: `room.OrganizationallyOutranks`.
