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
- Deleting a workspace is a platform action that removes its stream through a
  privileged path and records a tombstone in the registry; a workspace's own owner
  cannot erase its audit history.
