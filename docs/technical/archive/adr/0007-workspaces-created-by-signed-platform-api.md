# Create workspaces through a signed platform API, not a public form

Status: accepted

Sign-up, email verification, billing and plan limits belong to the operator's
private control plane. The open core exposes `/platform/v1/` (Ed25519-signed
requests, the same scheme as the extension contract) to create, suspend, resume and
delete workspaces and to set their limits; a created workspace returns its setup
link. A simple built-in sign-up page exists only when the operator sets
`COMMONROOM_SIGNUP=open`, is rate limited and off by default.

## Consequences

- The core carries no payment, plan or email logic and no mention of editions.
- Deleting a workspace through the API stops serving it at once and leaves a registry
  tombstone so the id is never reused; its append-only history is removed only by an
  administrator running `scripts/purge-workspace.sql`. A workspace's own owner cannot
  erase its audit history, and the application's database role cannot either.
